package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
)

const (
	migrationTable = "schema_migrations"
	migrationDir   = "./database/migrations"
)

// main initializes the database schema from the versioned table definitions.
// It intentionally does not use the application config initializer because a
// database migration must not require unrelated application settings such as
// the JWT secret.
func main() {
	// 允许通过命令行指定配置文件路径。
	configPath := flag.String("config", "./config/config.yaml", "path to the YAML configuration file")
	repair := flag.Bool("repair", false, "repair dirty migration state")
	flag.Parse()

	// 打开数据库连接，并在程序结束时关闭连接。
	db, err := openDatabase(*configPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("connect to MySQL: %v", err)
	}

	// 按版本顺序执行尚未完成的迁移脚本。
	if err := applyMigrations(db, migrationDir, *repair); err != nil {
		log.Fatal(err)
	}

	log.Println("database migrations are up to date")
}

func openDatabase(configPath string) (*sql.DB, error) {
	// 迁移程序只读取 MySQL 相关配置，不依赖其他业务配置。
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	host := v.GetString("mysql.host")
	port := v.GetInt("mysql.port")
	database := v.GetString("mysql.database")
	user := v.GetString("mysql.user")
	password := v.GetString("mysql.password")
	if host == "" {
		return nil, errors.New("mysql.host must be configured")
	}
	if port <= 0 {
		return nil, errors.New("mysql.port must be configured")
	}
	if user == "" {
		return nil, errors.New("mysql.user must be configured")
	}
	if database == "" {
		return nil, errors.New("mysql.database must be configured")
	}

	// 组装 MySQL 数据源名称（DSN）并建立数据库连接。
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local&multiStatements=true",
		url.QueryEscape(user), url.QueryEscape(password), host, port, url.PathEscape(database))
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open MySQL: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	return db, nil
}

func applyMigrations(db *sql.DB, migrationDir string, repair bool) error {
	// 读取目录中的 SQL 文件，并按文件名排序来确定迁移顺序。
	entries, err := os.ReadDir(migrationDir)
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}

	fileNames := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			fileNames = append(fileNames, entry.Name())
		}
	}
	sort.Strings(fileNames)
	if len(fileNames) == 0 {
		return errors.New("no migrations found")
	}

	if err := ensureMigrationTableExists(db); err != nil {
		return err
	}

	currentVersion, dirty, err := migrationState(db)
	if err != nil {
		return err
	}
	if dirty {
		if repair {
			if err := repairMigration(db, currentVersion-1); err != nil {
				return err
			}
			log.Printf("repaired dirty migration: version=%d", currentVersion)
			return nil
		}
		return fmt.Errorf("migration version %d is dirty; fix the database before continuing", currentVersion)
	}

	for index, fileName := range fileNames {
		version := int64(index + 1)
		if version <= currentVersion {
			continue
		}

		// 执行前记录目标版本并标记为 dirty，执行失败时保留该状态。
		if err := markMigrationDirty(db, version); err != nil {
			return fmt.Errorf("mark migration %d as dirty: %w", version, err)
		}

		sqlBytes, err := os.ReadFile(filepath.Join(migrationDir, fileName))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", fileName, err)
		}
		// 一个迁移文件对应一次数据库结构变更。
		if _, err := db.Exec(string(sqlBytes)); err != nil {
			return fmt.Errorf("execute table definition %s: %w", fileName, err)
		}
		// 执行成功后更新版本号并清除 dirty 标记。
		if _, err := db.Exec(
			"UPDATE `schema_migrations` SET `version` = ?, `dirty` = 0",
			version,
		); err != nil {
			return fmt.Errorf("mark migration %d as complete: %w", version, err)
		}
		currentVersion = version
		log.Printf("applied migration: version=%d file=%s", version, fileName)
	}
	return nil
}

func ensureMigrationTableExists(db *sql.DB) error {
	// 迁移状态表必须预先存在，且不能由迁移程序自动创建。
	var tableName string
	err := db.QueryRow(
		"SELECT TABLE_NAME FROM information_schema.tables WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?",
		migrationTable,
	).Scan(&tableName)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("migration table %q does not exist", migrationTable)
	}
	if err != nil {
		return fmt.Errorf("check migration table: %w", err)
	}
	return nil
}

func migrationState(db *sql.DB) (int64, bool, error) {
	// 读取当前迁移版本和执行状态。
	var version int64
	var dirty bool
	err := db.QueryRow(
		"SELECT `version`, `dirty` FROM `schema_migrations` LIMIT 1",
	).Scan(&version, &dirty)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, fmt.Errorf("migration table %q has no state record", migrationTable)
	}
	if err != nil {
		return 0, false, fmt.Errorf("read migration state: %w", err)
	}
	return version, dirty, nil
}

func markMigrationDirty(db *sql.DB, version int64) error {
	// 将目标迁移版本标记为执行中；只有成功后才会清除 dirty 标记。
	_, err := db.Exec("UPDATE `schema_migrations` SET `version` = ?, `dirty` = 1", version)
	return err
}

func repairMigration(db *sql.DB, version int64) error {
	// 回退到上一个版本，并清除 dirty 状态；修复后需再次运行迁移程序。
	_, err := db.Exec(
		"UPDATE `schema_migrations` SET `version` = ?, `dirty` = 0",
		version,
	)
	if err != nil {
		return fmt.Errorf("repair migration state: %w", err)
	}
	return nil
}
