package main

import (
	"sync"

	"github.com/caarlos0/env/v11"
)

type config struct {
	DbURL string `env:"DB_URL,required,notEmpty"`
	Port  string `env:"PORT" envDefault:"8080"`
}

var (
	once sync.Once
	cfg  config
)

func loadConfig() (config, error) {
	var err error
	once.Do(func() {
		err = env.Parse(&cfg)
	})
	return cfg, err
}
