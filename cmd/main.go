package main

import (
	"context"
	"flag"
	"github.com/BurntSushi/toml"
	"github.com/mrsova/go-migrate/config"
	"github.com/mrsova/go-migrate/pkg/logger"
	"github.com/mrsova/go-migrate/pkg/migrate"
	"github.com/mrsova/go-migrate/pkg/postgres"
	"time"
)

var (
	configPath string
	rollback   bool
)

func main() {
	parseConfig()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pgClient, err := postgres.NewClient(ctx, 5, time.Second*5, postgres.NewPgConfig(
		config.GetConfig().Database.Username,
		config.GetConfig().Database.Password,
		config.GetConfig().Database.Host,
		config.GetConfig().Database.Port,
		config.GetConfig().Database.Database,
	))
	if err != nil {
		logger.Logger.Fatal(err)
	}
	migration := migrate.New(pgClient, config.GetConfig().Migrate.Dir, config.GetConfig().Migrate.TableName)
	if !rollback {
		err = migration.Migrate(ctx)
	} else {
		err = migration.Rollback(ctx)
	}
	if err != nil {
		panic(err)
	}
}

func init() {
	logger.New()
	flag.StringVar(
		&configPath,
		"config-path",
		"./example/config/config.toml",
		"path to config file")
	flag.BoolVar(
		&rollback,
		"rollback",
		false,
		"rollback flag")
}

func parseConfig() *config.Config {
	flag.Parse()
	conf := config.NewConfig()
	_, err := toml.DecodeFile(configPath, conf)
	if err != nil {
		logger.Logger.Fatal(err)
	}
	config.SetConfig(conf)
	return conf
}
