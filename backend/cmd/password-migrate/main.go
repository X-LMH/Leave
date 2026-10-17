// password-migrate 将 users 表中的历史明文密码转换为 Argon2id 哈希。
package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	passwordutil "backend/internal/utils/password"

	driver "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, out io.Writer) error {
	flags := flag.NewFlagSet("password-migrate", flag.ContinueOnError)
	flags.SetOutput(out)
	configPath := flags.String("config", "", "MySQL YAML 配置文件（必填）")
	envPath := flags.String("env", ".env", "环境变量文件；已有环境变量优先，可传空字符串跳过")
	apply := flags.Bool("apply", false, "执行写入；默认只预览")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if *configPath == "" || flags.NArg() != 0 {
		return errors.New("请通过 -config 指定配置文件，不接受位置参数")
	}
	if *envPath != "" {
		if err := gotenv.Load(*envPath); err != nil && !os.IsNotExist(err) {
			return errors.New("读取环境变量文件失败")
		}
	}
	dsn, err := readDatabaseConfig(*configPath)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "目标：%s/%s，执行写入：%t\n", dsn.Addr, dsn.DBName, *apply)
	db, err := sql.Open("mysql", dsn.FormatDSN())
	if err != nil {
		return errors.New("打开 MySQL 连接失败")
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		return errors.New("连接 MySQL 失败，请检查配置和网络")
	}
	return migrate(db, *apply, out)
}

// 只读取数据库配置，不要求 JWT、上传目录等服务配置。
func readDatabaseConfig(path string) (*driver.Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return nil, errors.New("读取 MySQL 配置文件失败")
	}
	if err := v.BindEnv("mysql.password", "MYSQL_PASSWORD"); err != nil {
		return nil, err
	}
	host, port := v.GetString("mysql.host"), v.GetInt("mysql.port")
	dsn := driver.NewConfig()
	dsn.User = v.GetString("mysql.user")
	dsn.Passwd = v.GetString("mysql.password")
	dsn.DBName = v.GetString("mysql.database")
	if host == "" || port < 1 || port > 65535 || dsn.User == "" || dsn.Passwd == "" || dsn.DBName == "" {
		return nil, errors.New("MySQL 主机、端口、用户、密码和数据库必须配置完整")
	}
	dsn.Net = "tcp"
	dsn.Addr = net.JoinHostPort(host, strconv.Itoa(port))
	dsn.Timeout = 10 * time.Second
	dsn.ReadTimeout = 30 * time.Second
	dsn.WriteTimeout = 30 * time.Second
	return dsn, nil
}

type account struct {
	id       uint
	password string
}

// 已有合法哈希保持原样；疑似其他哈希或损坏哈希必须人工确认，避免二次哈希。
func isHashed(value string) (bool, error) {
	if strings.HasPrefix(value, "$argon2id$") {
		if _, err := passwordutil.Verify(value, "migration-format-check"); err != nil {
			return false, err
		}
		return true, nil
	}
	if strings.HasPrefix(value, "$") || value == "" {
		return false, errors.New("发现空密码或未知哈希格式，请人工检查")
	}
	return false, nil
}

func migrate(db *sql.DB, apply bool, out io.Writer) error {
	tx, err := db.Begin()
	if err != nil {
		return errors.New("开启事务失败")
	}
	defer tx.Rollback()
	query := "SELECT id, password FROM users ORDER BY id"
	if apply {
		// 与两端改密共用行锁，锁保持至整个迁移提交；执行前应暂停账号写入。
		query += " FOR UPDATE"
	}
	rows, err := tx.Query(query)
	if err != nil {
		return errors.New("读取 users 表失败")
	}
	var pending []account
	total, hashed := 0, 0
	for rows.Next() {
		var a account
		if err := rows.Scan(&a.id, &a.password); err != nil {
			rows.Close()
			return errors.New("读取账号失败")
		}
		total++
		ok, err := isHashed(a.password)
		if err != nil {
			rows.Close()
			return fmt.Errorf("账号 id=%d 密码格式异常，未提交任何修改", a.id)
		}
		if ok {
			hashed++
		} else {
			pending = append(pending, a)
		}
	}
	readErr := rows.Err()
	rows.Close()
	if readErr != nil {
		return errors.New("读取账号失败，未提交任何修改")
	}
	fmt.Fprintf(out, "账号总数：%d，已有哈希：%d，待迁移：%d\n", total, hashed, len(pending))
	if !apply {
		fmt.Fprintln(out, "预览完成，未写入数据库。使用 -apply 执行迁移。")
		return nil
	}
	for _, a := range pending {
		hash, err := passwordutil.Hash(a.password)
		if err != nil {
			return errors.New("生成哈希失败，未提交任何修改")
		}
		result, err := tx.Exec("UPDATE users SET password = ? WHERE id = ? AND BINARY password = BINARY ?", hash, a.id, a.password)
		if err != nil {
			return fmt.Errorf("更新账号 id=%d 失败，未提交任何修改", a.id)
		}
		n, err := result.RowsAffected()
		if err != nil || n != 1 {
			return fmt.Errorf("账号 id=%d 状态发生变化，未提交任何修改", a.id)
		}
		var stored string
		if err := tx.QueryRow("SELECT password FROM users WHERE id = ?", a.id).Scan(&stored); err != nil {
			return errors.New("回读密码失败，未提交任何修改")
		}
		matched, err := passwordutil.Verify(stored, a.password)
		if err != nil || !matched {
			return errors.New("回读校验失败，未提交任何修改")
		}
	}
	if err := tx.Commit(); err != nil {
		return errors.New("事务提交结果未知，请先预览检查数据库状态再重试")
	}
	fmt.Fprintf(out, "迁移完成：%d 个账号，原密码保持不变。\n", len(pending))
	return nil
}
