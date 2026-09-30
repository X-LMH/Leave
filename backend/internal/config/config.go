package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
)

// Config 全局配置结构体（修正字段名、标签与YAML匹配）
type Config struct {
	JWT struct {
		Secret         string `yaml:"secret"`
		ExpirationDays int    `yaml:"expiration_days"`
	} `yaml:"jwt"`
	Mysql struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Database string `yaml:"database"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
	} `yaml:"mysql"`
	App struct {
		PackageDir string `yaml:"package_dir"`
		Port       int    `yaml:"port"`
	} `yaml:"app"`
}

// Cfg 全局配置实例，供外部调用
var Cfg Config

// Init 初始化配置（读取并解析yaml）
func Init() error {
	if err := loadDotEnv(); err != nil {
		return fmt.Errorf("加载环境变量文件失败: %w", err)
	}

	// 支持在 backend 目录本地运行，以及在服务器项目根目录或 release 目录运行。
	configPaths := []string{
		"./config/config.local.yaml",
		"./config/config.server.yaml",
		"../../config/config.server.yaml",
	}
	configFound := false
	for _, configPath := range configPaths {
		if _, err := os.Stat(configPath); err == nil {
			viper.SetConfigFile(configPath)
			configFound = true
			break
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("检查配置文件 %q 失败: %w", configPath, err)
		}
	}
	if !configFound {
		return fmt.Errorf("未找到配置文件，请提供 config.local.yaml 或 config.server.yaml")
	}
	viper.AllowEmptyEnv(true)
	viper.SetDefault("app.port", 10000)
	viper.BindEnv("mysql.password", "MYSQL_PASSWORD")
	viper.BindEnv("jwt.secret", "JWT_SECRET")

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置失败: %w", err)
	}

	// 将配置解析到Cfg结构体
	if err := viper.Unmarshal(&Cfg); err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}
	Cfg.JWT.Secret = strings.TrimSpace(viper.GetString("jwt.secret"))
	Cfg.JWT.ExpirationDays = viper.GetInt("jwt.expiration_days")
	Cfg.Mysql.Password = viper.GetString("mysql.password")
	Cfg.App.PackageDir = strings.TrimSpace(viper.GetString("app.package_dir"))
	Cfg.App.Port = viper.GetInt("app.port")
	if Cfg.Mysql.Password == "" {
		return fmt.Errorf("MySQL password cannot be empty; configure MYSQL_PASSWORD")
	}
	if Cfg.JWT.Secret == "" {
		return fmt.Errorf("JWT secret cannot be empty; configure jwt.secret or JWT_SECRET")
	}
	if Cfg.JWT.ExpirationDays <= 0 {
		return fmt.Errorf("jwt.expiration_days must be greater than 0")
	}
	if Cfg.App.Port <= 0 || Cfg.App.Port > 65535 {
		return fmt.Errorf("app.port must be between 1 and 65535")
	}

	return nil
}

func loadDotEnv() error {
	// current 通常是软链接，进程工作目录可能解析为 releases/vX.Y.Z；
	// 因此还需要向上两级读取 Leave 根目录下的 .env。
	for _, path := range []string{".env", "../.env", "../../.env"} {
		if _, err := os.Stat(path); err == nil {
			return gotenv.Load(path)
		} else if !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

// JWTSecret 返回初始化后用于 JWT 签名与验签的密钥。
func JWTSecret() []byte { return []byte(Cfg.JWT.Secret) }
