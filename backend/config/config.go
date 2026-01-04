package config

import (
	"backend/models"
	"fmt"

	"github.com/spf13/viper"
)

// Config 全局配置结构体（修正字段名、标签与YAML匹配）
type Config struct {
	Version struct {
		Latest      string   `yaml:"latest"`
		DownloadURL string   `yaml:"downloadURL"`
		UpdateLog   []string `yaml:"updateLog"`
	} `yaml:"version"`

	Mysql struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Database string `yaml:"database"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
	} `yaml:"mysql"`
}

// Cfg 全局配置实例，供外部调用
var Cfg Config

// Init 初始化配置（读取并解析yaml）
func Init() error {
	// 设置viper参数
	viper.SetConfigName("config") // 配置文件名（无后缀）
	viper.SetConfigType("yaml")   // 配置文件类型
	viper.AddConfigPath(".")      // 配置文件所在路径（当前根目录）

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置失败: %w", err)
	}

	// 将配置解析到Cfg结构体
	if err := viper.Unmarshal(&Cfg); err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}

	return nil
}

func GetVersionInfo() models.VersionInfo {
	return models.VersionInfo{
		Latest:      Cfg.Version.Latest,
		DownloadURL: Cfg.Version.DownloadURL,
		UpdateLog:   Cfg.Version.UpdateLog,
	}
}
