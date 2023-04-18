package config

import (
	m "github.com/meguriri/GoCache/server/manager"
	"github.com/meguriri/GoCache/server/replacement"
	"github.com/spf13/viper"
)

func Configinit() error {
	viper.SetConfigName("GoCache")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	replacement.ReplacementPolicy = viper.GetString("replacement.policy")
	replacement.MaxBytes = viper.GetInt64("replacement.max-bytes")
	m.DefaultReplicas = viper.GetInt("defaultReplicas")
	m.ManagerIP = viper.GetString("server.ip")
	m.ManagerPort = viper.GetString("server.port")
	return nil
}
