package main

import (
	"flag"

	"github.com/linggaaskaedo/go-play-v2/stdlib/auth"
	"github.com/linggaaskaedo/go-play-v2/stdlib/httpmux"
	"github.com/linggaaskaedo/go-play-v2/stdlib/logger"
	"github.com/linggaaskaedo/go-play-v2/stdlib/parser"
	"github.com/linggaaskaedo/go-play-v2/stdlib/redis"
	"github.com/linggaaskaedo/go-play-v2/stdlib/sql"
)

var (
	confPath  string
	minJitter int
	maxJitter int
)

func init() {
	// Flag Settings Initialization
	flag.StringVar(&confPath, "staticConfPath", "./etc/conf", "config path")
	flag.IntVar(&minJitter, "minSleep", DefaultMinJitter, "min. sleep duration during app initialization")
	flag.IntVar(&maxJitter, "maxSleep", DefaultMaxJitter, "max. sleep duration during app initialization")
	flag.Parse()

	// Add Sleep with Jitter to drag the the initialization time among instances
	sleepWithJitter(minJitter, maxJitter)

	// Config Initialization
	conf, err := InitConfig(confPath)
	if err != nil {
		panic(err)
	}

	// Logger Initialization
	logger := logger.Init(conf.Log)

	// Parser Initialization
	parse := parser.Init(logger, conf.Parser)

	// Redis Initialization
	_ = redis.Init(logger, conf.Redis)

	// SQL Initialization
	_ = sql.Init(logger, conf.SQL["sql-0"])
	_ = sql.Init(logger, conf.SQL["sql-1"])

	// HTTPMux Initialization
	httpmux.Init(logger, conf.HTTPMux)

	// Auth Initialization
	_ = auth.Init(logger, conf.Auth, parse)

	logger.Info().Msg("Yuhuuu !!!!")
}

func main() {
	// log.Debug().Msg("Yeaahhhh.... !!!")
	// logger.Debug("Yeaahhhh.... !!!")
	// logger.logger.Debug().Msg("AAAA")
}
