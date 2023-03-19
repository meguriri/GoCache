package config

import (
	"github.com/meguriri/GoCache/data"
	"github.com/spf13/viper"
)

func Configinit() error {
	viper.SetConfigName("GoCache")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	data.ReplacementPolicy = viper.GetString("replacement-policy")
	data.MaxBytes = viper.GetInt64("max-bytes")

	return nil
}
