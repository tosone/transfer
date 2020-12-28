package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/tosone/logging"
	"github.com/unknwon/com"

	"transfer/database"
	"transfer/router"
)

// ConfigFile default config path
const ConfigFile = "/etc/transfer/config.yaml"

var appStopped bool

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

	if !viper.GetBool("Debug") {
		logging.Setting(logging.Config{LogLevel: logging.InfoLevel})
	}

	if err = Config(); err != nil {
		logging.Fatal(err)
	}

	if err = database.Initialize(); err != nil {
		logging.Fatal(err)
	}

	if err = RunTask(); err != nil {
		logging.Fatal(err)
	}

	if !viper.GetBool("Debug") {
		gin.SetMode(gin.ReleaseMode)
	}
	var app = gin.Default()
	app.Use(cors.Default())
	app.Use(gzip.Gzip(gzip.DefaultCompression))
	app.Use(requestid.New())

	var srv = &http.Server{
		Addr:    ":8080",
		Handler: app,
	}

	if err = router.Initialize(app); err != nil {
		logging.Fatal(err)
	}

	go func() {
		if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logging.Fatal(err)
		}
	}()

	signalChanel := make(chan os.Signal, 1)
	signal.Notify(signalChanel, os.Interrupt)

	<-signalChanel

	logging.Info("transfer is stopping")

	appStopped = true

	stopWaitGroup.Wait()

	logging.Info("database is stopping")
	if err = database.Teardown(); err != nil {
		logging.Error(err)
	}

	logging.Info("server is stopping")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = srv.Shutdown(ctx); err != nil {
		logging.Error(err)
	}
}
