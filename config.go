package main

import (
	"flag"
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/tosone/logging"
	"github.com/unknwon/com"
)

const ConfigFile = "/etc/transfer/config.yaml"

func Config() (err error) {
	var configFile string
	flag.StringVar(&configFile, "config", ConfigFile, "config file")
	flag.Parse()

	if !com.IsFile(configFile) {
		logging.Fatalf("cannot find config file: %s", configFile)
	}
	viper.SetConfigType("yaml")
	viper.SetConfigName(filepath.Base(configFile))
	viper.AddConfigPath(filepath.Dir(configFile))

	if err = viper.ReadInConfig(); err != nil {
		return
	}
	return
}
