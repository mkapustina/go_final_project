package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	Port     string
	DbFile   string
	Password string
	WebDir   string
}

func ParseConfig() (Config, error) {
	cfg := Config{}

	cfg.Port = os.Getenv("TODO_PORT")
	if len(cfg.Port) == 0 {
		cfg.Port = "7540"
	}

	cfg.DbFile = os.Getenv("TODO_DBFILE")
	if len(cfg.DbFile) == 0 {
		cfg.DbFile = filepath.Join("..", "..", "db", "scheduler.db")
	}

	cfg.Password = os.Getenv("TODO_PASSWORD")

	cfg.WebDir = os.Getenv("TODO_WEBDIR")
	if len(cfg.WebDir) == 0 {
		cfg.WebDir = filepath.Join("..", "..", "web")
	}

	// TODO check dbfile and webdir existance

	return cfg, nil
}
