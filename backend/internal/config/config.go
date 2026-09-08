package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config 全局配置结构体（修正字段名、标签与YAML匹配）
type Config struct {
	JWT struct {
		Secret string `yaml:"secret"`
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
	} `yaml:"app"`
}

// Cfg 全局配置实例，供外部调用
var Cfg Config

// Init 初始化配置（读取并解析yaml）
func Init() error {
	// 设置viper参数
	viper.SetConfigName("config")   // 配置文件名（无后缀）
	viper.SetConfigType("yaml")     // 配置文件类型
	viper.AddConfigPath("./config") // 当前推荐的配置文件目录
	viper.AddConfigPath(".")        // 兼容已有部署中的 config.yaml
	viper.BindEnv("jwt.secret", "JWT_SECRET")
	viper.SetDefault("app.package_dir", "/www/wwwroot/Leave/releases")

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置失败: %w", err)
	}

	// 将配置解析到Cfg结构体
	if err := viper.Unmarshal(&Cfg); err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}
	Cfg.JWT.Secret = strings.TrimSpace(viper.GetString("jwt.secret"))
	Cfg.App.PackageDir = strings.TrimSpace(viper.GetString("app.package_dir"))

	return nil
}

// JWTSecret 返回初始化后用于 JWT 签名与验签的密钥。
func JWTSecret() []byte { return []byte(Cfg.JWT.Secret) }
