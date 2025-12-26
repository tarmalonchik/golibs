package config

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
	"github.com/tarmalonchik/golibs/trace"
)

func Load(conf interface{}, configFile string) error {
	configs := make([]string, 0, 2)

	path := "./configs/.env"

	if _, err := os.Stat(path); err == nil {
		configs = append(configs, path)
	}
	if _, err := os.Stat(configFile); err == nil {
		configs = append(configs, configFile)
	}

	if err := godotenv.Load(configs...); err != nil {
		log.WithField("filenames", configs).Info("config file not found, using defaults")
	}

	if err := envconfig.Process("", conf); err != nil {
		return trace.FuncNameWithErrorMsg(err, "env config process")
	}
	return nil
}
