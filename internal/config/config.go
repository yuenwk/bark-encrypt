package config

import (
	"errors"
	"log"

	"github.com/spf13/viper"
)

type BarkConfig struct {
	AesKey string `mapstructure:"AES_KEY"`
	Domain string `mapstructure:"DOMAIN"`
}

type Config struct {
	Port string     `mapstructure:"PORT"`
	Bark BarkConfig `mapstructure:"BARK"`
}

func Load() (cfg *Config, err error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	viper.SetDefault("port", "9090")

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			// 配置文件未找到错误；如果需要可以忽略
			log.Fatalf("No config file found: %v", err)
		}
	}

	// 将读取到的配置反序列化到 Config 结构体中
	err = viper.Unmarshal(&cfg)

	return
}
