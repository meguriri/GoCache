package config

import (
	"github.com/meguriri/GoCache/peer/cache"
	"github.com/meguriri/GoCache/peer/replacement"
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
	cache.SaveSeconds = viper.GetInt("save.seconds")
	cache.SaveModify = viper.GetInt("save.modify")
	cache.SlaveAddress = viper.GetString("slave.address")
	cache.SlavePort = viper.GetString("slave.port")
	cache.PeerAddress = viper.GetString("peer.address")
	cache.PeerPort = viper.GetString("peer.port")
	return nil
}
