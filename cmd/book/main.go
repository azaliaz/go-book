package main

import (
	"fmt"
	"helloapp/internal/config"
	"helloapp/internal/logger"
	"helloapp/internal/server"

	"helloapp/internal/storage"
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
