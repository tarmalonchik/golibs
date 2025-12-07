package config

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
	"github.com/tarmalonchik/golibs/trace"
)

func Load(conf interface{}, configFile string) error {
	if err := godotenv.Load(configFile); err != nil {
		log.WithField("filenames", configFile).Info("config file not found, using defaults")
	}

	if err := envconfig.Process("", conf); err != nil {
		return trace.FuncNameWithErrorMsg(err, "env config process")
	}
	return nil
}
