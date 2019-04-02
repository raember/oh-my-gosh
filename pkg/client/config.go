package client

import (
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
)

func setDefaults(config *viper.Viper) {
	config.SetDefault("Client.Port", common.PORT)
	config.SetDefault("Client.Protocol", common.TCP)
	config.SetDefault("Logging.LogLevel", "info")
}

func Config(configpath string) *viper.Viper {
	config := viper.New()
	config.SetConfigName(common.CLIENTNAME + "_config")
	config.AddConfigPath(configpath)
	config.SetConfigType(common.CONFIGFORMAT)
	setDefaults(config)
	err := config.ReadInConfig()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatalln("Couldn't read config file.")
	}
	config.WatchConfig()
	config.OnConfigChange(func(e fsnotify.Event) {
		log.WithFields(log.Fields{
			"name": e.Name,
		}).Warnln("Config file changed.")
	})
	return config
}
