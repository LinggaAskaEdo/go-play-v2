package main

import (
	"github.com/spf13/viper"

	auth "github.com/linggaaskaedo/go-play-v2/stdlib/auth"
	httpmux "github.com/linggaaskaedo/go-play-v2/stdlib/httpmux"
	log "github.com/linggaaskaedo/go-play-v2/stdlib/logger"
	parser "github.com/linggaaskaedo/go-play-v2/stdlib/parser"
	redis "github.com/linggaaskaedo/go-play-v2/stdlib/redis"
	libsql "github.com/linggaaskaedo/go-play-v2/stdlib/sql"
)

type Config struct {
	App     Options
	Log     log.Options
	Parser  parser.Options
	Redis   redis.Options
	SQL     map[string]libsql.Options
	HTTPMux httpmux.Options
	Auth    auth.Options
}

type Options struct {
	port int
}

func InitConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("conf")
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)

	return
}
