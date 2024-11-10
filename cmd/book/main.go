package main

import (
	"fmt"

	"github.com/azaliaz/go-book/internal/config"
	"github.com/azaliaz/go-book/internal/logger"
	"github.com/azaliaz/go-book/internal/server"
	"github.com/azaliaz/go-book/internal/storage"
)

func main() {
	cfg := config.ReadConfig()
	fmt.Println(cfg)
	log := logger.Get(cfg.Debug)
	log.Debug().Any("cfg", cfg).Send()
	stor := storage.New()

	serv := server.New(*cfg, stor)
	err := serv.Run()
	if err != nil {
		log.Fatal().Err(err).Msg("server fatal error")
	}
	log.Info().Msg("server stopped")
}
