package main

import (
	"flag"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/spf13/viper"
	"github.com/tosone/logging"
	"github.com/unknwon/com"

	"transfer/database"
	"transfer/router"
)

// ConfigFile default config path
const ConfigFile = "/etc/transfer/config.yaml"

// Config ..
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

func main() {
	var err error

	if err = Config(); err != nil {
		logging.Fatal(err)
	}

	if err = database.Initialize(); err != nil {
		logging.Fatal(err)
	}
	defer func() {
		if err = database.Teardown(); err != nil {
			logging.Error(err)
		}
	}()

	if err = RunTask(); err != nil {
		logging.Fatal(err)
	}

	var app = fiber.New()

	app.Use(compress.New())
	app.Use(requestid.New())

	if err = router.Task(app); err != nil {
		logging.Fatal(err)
	}

	go func() {
		if err = app.Listen(":3000"); err != nil {
			logging.Fatal(err)
		}
	}()

	signalChanel := make(chan os.Signal, 1)
	signal.Notify(signalChanel, os.Interrupt)

	<-signalChanel

	logging.Info("transfer has been stopped")
}
