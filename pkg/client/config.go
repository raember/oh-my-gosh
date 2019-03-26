package client

import (
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
)

var config = viper.New()

func setDefaults(config *viper.Viper) {
	config.SetDefault("Client.Port", common.PORT)
	config.SetDefault("Client.Protocol", common.TCP)
	config.SetDefault("Logging.LogLevel", "info")
}

func init() {
	config.SetConfigName(common.CLIENTNAME + "_config")
	config.AddConfigPath("/etc/" + common.CLIENTNAME + "/")
	config.AddConfigPath("./configs")
	config.AddConfigPath("../../configs")
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
}

func Config() *viper.Viper {
	return config
}
