package sql

import (
	"errors"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
)

const (
	PGSQL string = `postgres`
	MYSQL string = `mysql`
)

type Options struct {
	Enabled     bool
	Driver      string
	Host        string
	Port        string
	DB          string
	User        string
	Password    string
	SSL         bool
	ConnOptions ConnOptions
}

type ConnOptions struct {
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime int
}

func Init(logger zerolog.Logger, opt Options) *sqlx.DB {
	driver, host, err := getURI(opt)
	if err != nil {
		logger.Panic().Err(err).Str("when", "getURI").Send()
	}

	db, err := sqlx.Connect(driver, host)
	if err != nil {
		logger.Panic().Err(err).Str("when", "connect").Send()
	}

	return db
}

func getURI(opt Options) (string, string, error) {
	switch opt.Driver {
	case PGSQL:
		ssl := `disable`
		if opt.SSL {
			ssl = `require`
		}

		return opt.Driver, fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", opt.Host, opt.Port, opt.User, opt.Password, opt.DB, ssl), nil

	case MYSQL:
		ssl := `false`
		if opt.SSL {
			ssl = `true`
		}

		return opt.Driver, fmt.Sprintf("%s:%s@tcp(%s:%v)/%s?tls=%s", opt.User, opt.Password, opt.Host, opt.Port, opt.DB, ssl), nil

	default:
		return "", "", errors.New("DB Driver is not supported ")
	}
}
