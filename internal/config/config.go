package config

import (
	"cmp"
	"flag"
	"fmt"
	"os"
)

const (
	defaltAddr         = "localhost"
	defaultPort        = "8080"
	defaultDBDsn       = "postgres://user:password@localhost:5432/project?sslmode=disable"
	defaultMigratePath = "migrations"
)

type Config struct {
	Addr        string
	Debug       bool
	DBDsn       string
	MigratePath string
}

func ReadConfig() *Config {
	var host, dbDsn, migratePath, port string
	// var port int
	var debug bool
	flag.StringVar(&host, "host", defaltAddr, "flag to set the server startup host")
	flag.StringVar(&port, "port", defaultPort, "flag to set the server startup port")
	flag.BoolVar(&debug, "debug", false, "flag to set debug logger level")
	flag.StringVar(&dbDsn, "db", defaultDBDsn, "database connection address")
	flag.StringVar(&migratePath, "m", defaultMigratePath, "path to migrations")

	flag.Parse()

	host = cmp.Or(os.Getenv("SERVER_HOST"), host)
	port = cmp.Or(os.Getenv("SERVER_PORT"), port)
	dbDsn = cmp.Or(os.Getenv("DB_DSN"), dbDsn)
	migratePath = cmp.Or(os.Getenv("MIGRATE_PATH"), migratePath)
	return &Config{
		Addr:        fmt.Sprintf("%s:%s", host, port),
		Debug:       debug,
		DBDsn:       dbDsn,
		MigratePath: migratePath,
	}
}
