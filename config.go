package main

import (
	"flag"
	"os"

	"github.com/spf13/viper"
	"github.com/tosone/logging"
	"github.com/unknwon/com"
)

const ConfigFile = "/etc/transfer/config.yaml"

func Config() {
	var configFile string
	flag.StringVar(&configFile, "config", ConfigFile, "config file")
	flag.Parse()

	if !com.IsFile(configFile) {
		logging.Fatalf("cannot find config file: %s", configFile)
	}

	var err error
	var configReader *os.File
	if configReader, err = os.Open(configFile); err != nil {
		logging.Fatal(err)
	}
	if err = viper.ReadConfig(configReader); err != nil {
		logging.Fatal(err)
	}
}
