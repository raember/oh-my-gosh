package server

import (
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
)

func setDefaults(config *viper.Viper) {
	log.WithField("config", config).Traceln("--> server.setDefaults")
	config.SetDefault("Serve.Port", common.PORT)
	config.SetDefault("Serve.Protocol", common.TCP)
	config.SetDefault("Logging.LogLevel", "info")
	config.SetDefault("Authentication.KeyStore", "~/.gosh/authorized_keys")
	config.SetDefault("Authentication.LoginGraceTime", 120)
	config.SetDefault("Authentication.PermitRootLogin", false)
	config.SetDefault("Authentication.MaxTries", 6)
	config.SetDefault("Authentication.MaxSessions", 10)
}

func init() {
}

func LoadConfig(configpath string) *viper.Viper {
	log.WithField("configpath", configpath).Traceln("--> server.LoadConfig")
	config := viper.New()
	config.SetConfigName(common.SERVERNAME + "_config")
	config.AddConfigPath(configpath)
	config.SetConfigType(common.CONFIGFORMAT)
	setDefaults(config)
	err := config.ReadInConfig()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Warnln("Failed to read config file. Using defaults instead.")
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
