package server

import (
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
)

var config = viper.New()

func setDefaults(config *viper.Viper) {
	config.SetDefault("Server.Port", common.PORT)
	config.SetDefault("Server.Protocol", common.TCP)
	config.SetDefault("Server.HostKeys", "/etc/gosh/host_key")
	config.SetDefault("Logging.LogLevel", "info")
	config.SetDefault("Authentication.Certificate", "../../test/certificate.pem")
	config.SetDefault("Authentication.KeyFile", "../../test/key.pem")
	config.SetDefault("Authentication.LoginGraceTime", 120)
	config.SetDefault("Authentication.PermitRootLogin", false)
	config.SetDefault("Authentication.MaxTries", 6)
	config.SetDefault("Authentication.MaxSessions", 10)
}

func init() {
	config.SetConfigName(common.SERVERNAME + "_config")
	config.AddConfigPath("/etc/" + common.SERVERNAME + "/")
	config.AddConfigPath("./configs")
	config.AddConfigPath("../../configs")
	config.SetConfigType(common.CONFIGFORMAT)
	setDefaults(config)
	err := config.ReadInConfig()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatalln("Couldn't read server_config file.")
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
