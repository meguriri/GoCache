package config

import (
	"github.com/meguriri/GoCache/replacement"
	"github.com/spf13/viper"
)

func Configinit() error {
	viper.SetConfigName("GoCache")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	replacement.ReplacementPolicy = viper.GetString("replacement-policy")
	replacement.MaxBytes = viper.GetInt64("max-bytes")

	return nil
}
