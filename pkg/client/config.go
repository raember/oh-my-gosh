package client

import (
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
)

func setDefaults(config *viper.Viper) {
	log.WithField("config", config).Traceln("--> client.setDefaults")
	config.SetDefault("Client.Port", common.PORT)
	config.SetDefault("Client.Protocol", common.TCP)
	config.SetDefault("Logging.LogLevel", "info")
	config.SetDefault("Authentication.KeyStore", "~/.gosh")
}

func LoadConfig(configpath string) *viper.Viper {
	log.WithField("configpath", configpath).Traceln("--> client.LoadConfig")
	config := viper.New()
	config.SetConfigName(common.CLIENTNAME + "_config")
	config.AddConfigPath(configpath)
	config.SetConfigType(common.CONFIGFORMAT)
	setDefaults(config)
	err := config.ReadInConfig()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Warnln("Failed to read config file.")
		return config
	}
	config.WatchConfig()
	config.OnConfigChange(func(e fsnotify.Event) {
		log.WithFields(log.Fields{
			"name": e.Name,
		}).Warnln("LoadConfig file changed.")
	})
	return config
}
