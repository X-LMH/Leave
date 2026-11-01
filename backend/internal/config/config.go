package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
)

// 配置按启动时加载的实例传递，各组件只接收需要的部分。
type JWTConfig struct {
	Secret         string `mapstructure:"secret"`
	ExpirationDays int    `mapstructure:"expiration_days"`
}

type SessionConfig struct {
	StudentTTL time.Duration `mapstructure:"student_ttl"`
	AdminTTL   time.Duration `mapstructure:"admin_ttl"`
}

type MySQLConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"database"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}
type AppConfig struct {
	Namespace  string `mapstructure:"namespace"`
	PackageDir string `mapstructure:"package_dir"`
	Port       int    `mapstructure:"port"`
}
type StorageConfig struct {
	RootDir string `mapstructure:"root_dir"`
	BaseURL string `mapstructure:"base_url"`
}
type LogConfig struct {
	Level    string `mapstructure:"level"`
	Encoding string `mapstructure:"encoding"`
	Dir      string `mapstructure:"dir"`
}
type Config struct {
	Session SessionConfig `mapstructure:"session"`
	Redis   RedisConfig   `mapstructure:"redis"`
	Log     LogConfig     `mapstructure:"log"`
	JWT     JWTConfig     `mapstructure:"jwt"`
	Mysql   MySQLConfig   `mapstructure:"mysql"`
	App     AppConfig     `mapstructure:"app"`
	Storage StorageConfig `mapstructure:"storage"`
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
	for _, binding := range [][2]string{
		{"mysql.password", "MYSQL_PASSWORD"},
		{"redis.password", "REDIS_PASSWORD"},
		{"jwt.secret", "JWT_SECRET"},
	} {
		if err := v.BindEnv(binding[0], binding[1]); err != nil {
			return Config{}, fmt.Errorf("绑定环境变量失败: %w", err)
		}
	}

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("读取配置失败: %w", err)
	}

	// 将配置解析到cfg结构体
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("解析配置失败: %w", err)
	}
	cfg.JWT.Secret = strings.TrimSpace(cfg.JWT.Secret)
	cfg.Redis.Host = strings.TrimSpace(cfg.Redis.Host)
	cfg.App.Namespace = strings.TrimSpace(cfg.App.Namespace)
	cfg.App.PackageDir = strings.TrimSpace(cfg.App.PackageDir)
	cfg.Storage.RootDir = strings.TrimSpace(cfg.Storage.RootDir)
	cfg.Storage.BaseURL = strings.TrimSpace(cfg.Storage.BaseURL)

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
	if cfg.App.Namespace == "" || strings.ContainsAny(cfg.App.Namespace, ": \t\r\n") {
		return Config{}, fmt.Errorf("app.namespace must be nonempty and contain no colon or whitespace")
	}
	if cfg.Session.StudentTTL <= 0 || cfg.Session.AdminTTL <= 0 {
		return Config{}, fmt.Errorf("session TTLs must be positive durations")
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
