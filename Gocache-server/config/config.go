package config

import (
	"fmt"

	"github.com/spf13/viper"
)

func ConfigInit() (*viper.Viper, error) {
	config := viper.New()
	config.SetConfigName("GoCache")
	config.SetConfigType("yaml")
	config.AddConfigPath("./config")

	if err := config.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("config error: %s", err.Error())
	}
	return config, nil
}
