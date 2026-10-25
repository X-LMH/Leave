package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
)

// 配置按启动时加载的实例传递，各组件只接收需要的部分。
type JWTConfig struct {
	Secret         string `yaml:"secret"`
	ExpirationDays int    `yaml:"expiration_days"`
}
type MySQLConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}
type AppConfig struct {
	PackageDir string `yaml:"package_dir"`
	Port       int    `yaml:"port"`
}
type StorageConfig struct {
	RootDir string `yaml:"root_dir"`
	BaseURL string `yaml:"base_url"`
}
type LogConfig struct {
	Level    string `yaml:"level"`
	Encoding string `yaml:"encoding"`
}
type Config struct {
	Redis   RedisConfig   `yaml:"redis"`
	Log     LogConfig     `yaml:"log"`
	JWT     JWTConfig     `yaml:"jwt"`
	Mysql   MySQLConfig   `yaml:"mysql"`
	App     AppConfig     `yaml:"app"`
	Storage StorageConfig `yaml:"storage"`
}

// Load 读取并校验配置，不维护包级配置实例。
func Load() (Config, error) {
	var cfg Config
	v := viper.New()
	if err := loadDotEnv(); err != nil {
		return Config{}, fmt.Errorf("加载环境变量文件失败: %w", err)
	}

	configPath, err := FindConfigFile()
	if err != nil {
		return Config{}, err
	}
	v.SetConfigFile(configPath)
	v.AllowEmptyEnv(true)
	v.SetDefault("app.port", 10000)
	v.BindEnv("redis.password", "REDIS_PASSWORD")
	v.SetDefault("log.level", "info")
	v.SetDefault("log.encoding", "console")
	v.BindEnv("mysql.password", "MYSQL_PASSWORD")
	v.BindEnv("jwt.secret", "JWT_SECRET")

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("读取配置失败: %w", err)
	}

	// 将配置解析到cfg结构体
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("解析配置失败: %w", err)
	}
	cfg.JWT.Secret = strings.TrimSpace(v.GetString("jwt.secret"))
	cfg.JWT.ExpirationDays = v.GetInt("jwt.expiration_days")
	cfg.Mysql.Password = v.GetString("mysql.password")
	cfg.Redis.Host = strings.TrimSpace(cfg.Redis.Host)
	cfg.Redis.Password = v.GetString("redis.password")
	cfg.App.PackageDir = strings.TrimSpace(v.GetString("app.package_dir"))
	cfg.App.Port = v.GetInt("app.port")
	cfg.Storage.RootDir = strings.TrimSpace(v.GetString("storage.root_dir"))
	cfg.Storage.BaseURL = strings.TrimSpace(v.GetString("storage.base_url"))

	if cfg.Redis.Host == "" {
		return Config{}, fmt.Errorf("redis.host cannot be empty")
	}
	if cfg.Redis.Port <= 0 || cfg.Redis.Port > 65535 {
		return Config{}, fmt.Errorf("redis.port must be between 1 and 65535")
	}
	if cfg.Redis.DB < 0 {
		return Config{}, fmt.Errorf("redis.db cannot be negative")
	}
	if cfg.Mysql.Password == "" {
		return Config{}, fmt.Errorf("MySQL password cannot be empty; configure MYSQL_PASSWORD")
	}
	if cfg.JWT.Secret == "" {
		return Config{}, fmt.Errorf("JWT secret cannot be empty; configure jwt.secret or JWT_SECRET")
	}
	if cfg.JWT.ExpirationDays <= 0 {
		return Config{}, fmt.Errorf("jwt.expiration_days must be greater than 0")
	}
	if cfg.App.Port <= 0 || cfg.App.Port > 65535 {
		return Config{}, fmt.Errorf("app.port must be between 1 and 65535")
	}
	if cfg.Storage.RootDir == "" {
		return Config{}, fmt.Errorf("storage.root_dir cannot be empty")
	}
	if cfg.Storage.BaseURL == "" {
		return Config{}, fmt.Errorf("storage.base_url cannot be empty")
	}

	return cfg, nil
}

// FindConfigFile 为服务和迁移工具选择同一个环境配置，不匹配模板或任意 YAML。
func FindConfigFile() (string, error) {
	// 宝塔会从程序所在目录启动；兼容新旧发布布局及根目录启动。
	paths := []string{
		"./config/config.local.yaml",
		"./config/config.test.yaml",
		"./config/config.server.yaml",
		"../../config/config.test.yaml",
		"../../config/config.server.yaml",
		"../../../config/config.test.yaml",
		"../../../config/config.server.yaml",
	}
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		} else if !os.IsNotExist(err) {
			return "", fmt.Errorf("检查配置文件 %q 失败: %w", path, err)
		}
	}
	return "", fmt.Errorf("未找到配置文件，请提供 config.local.yaml、config.test.yaml 或 config.server.yaml")
}

func loadDotEnv() error {
	// current 会解析为 releases/vN 或 backend/releases/vN，分别向上两级或三级。
	for _, path := range []string{".env", "../.env", "../../.env", "../../../.env"} {
		if _, err := os.Stat(path); err == nil {
			return gotenv.Load(path)
		} else if !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}
