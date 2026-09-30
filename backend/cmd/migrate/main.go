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
	"regexp"
	"sort"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
)

const (
	migrationTable       = "schema_migrations"
	migrationFilePattern = `^(\d+)_.*\.up\.sql$`
)

// main initializes the database schema from the versioned table definitions.
// It intentionally does not use the application config initializer because a
// database migration must not require unrelated application settings such as
// the JWT secret.
func main() {
	loadDotEnv()

	// 允许通过命令行指定配置文件路径。
	configPath := flag.String("config", "./config/config.local.yaml", "path to the YAML configuration file")
	migrationsPath := flag.String("migrations", "./database/migrations", "path to the migration files")
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
	if err := applyMigrations(db, *migrationsPath); err != nil {
		log.Fatal(err)
	}

	log.Println("database migrations are up to date")
}

func loadDotEnv() {
	for _, path := range []string{".env", "../.env"} {
		if _, err := os.Stat(path); err == nil {
			_ = gotenv.Load(path)
			return
		}
	}
}

func openDatabase(configPath string) (*sql.DB, error) {
	// 迁移程序只读取 MySQL 相关配置，不依赖其他业务配置。
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.BindEnv("mysql.password", "MYSQL_PASSWORD")
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
	if password == "" {
		return nil, errors.New("MYSQL_PASSWORD must be configured")
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

func applyMigrations(db *sql.DB, migrationDir string) error {
	// 读取目录中的 SQL 文件，并按文件名中的版本号确定迁移顺序。
	entries, err := os.ReadDir(migrationDir)
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}

	type migrationFile struct {
		name    string
		version int64
	}

	pattern := regexp.MustCompile(migrationFilePattern)
	migrationFiles := make([]migrationFile, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			matches := pattern.FindStringSubmatch(entry.Name())
			if len(matches) != 2 {
				return fmt.Errorf("invalid migration filename %q: expected NNN_name.up.sql", entry.Name())
			}
			version, err := strconv.ParseInt(matches[1], 10, 64)
			if err != nil || version <= 0 {
				return fmt.Errorf("invalid migration version in filename %q", entry.Name())
			}
			migrationFiles = append(migrationFiles, migrationFile{name: entry.Name(), version: version})
		}
	}
	sort.Slice(migrationFiles, func(i, j int) bool {
		return migrationFiles[i].version < migrationFiles[j].version
	})
	if len(migrationFiles) == 0 {
		return errors.New("no migrations found")
	}
	for index := 1; index < len(migrationFiles); index++ {
		if migrationFiles[index].version == migrationFiles[index-1].version {
			return fmt.Errorf("duplicate migration version: %d", migrationFiles[index].version)
		}
	}

	if err := ensureMigrationTableExists(db); err != nil {
		return err
	}

	currentVersion, dirty, err := migrationState(db)
	if err != nil {
		return err
	}
	if dirty {
		return fmt.Errorf("migration state is dirty at version %d; inspect the database and repair it manually before retrying", currentVersion)
	}

	for _, migration := range migrationFiles {
		version := migration.version
		fileName := migration.name
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
