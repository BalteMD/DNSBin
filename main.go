package main

import (
	"context"
	"dnsbin/api"
	db "dnsbin/db/sqlc"
	"dnsbin/dns"
	"dnsbin/notify"
	"dnsbin/util"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var interruptSignals = []os.Signal{
	os.Interrupt,
	syscall.SIGTERM,
	syscall.SIGINT,
}

var version = "v1.1.0"

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load config")
	}

	if config.Environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
	ctx, stop := signal.NotifyContext(context.Background(), interruptSignals...)
	defer stop()

	connPool, err := pgxpool.New(ctx, config.DBSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to db")
	}

	runDBMigration(config.MigrationURL, config.DBSource)
	store := db.NewStore(connPool)

	telegram, err := initTelegram(config)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot initialize telegram notification")
	}

	go runDNSServer(config, store, telegram)
	runGinServer(config, store, telegram)
}

func initTelegram(config util.Config) (*notify.Telegram, error) {
	tgNotify, err := notify.NewTelegram(config)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create telegram notification")
	}
	msg := &notify.InitialMessage{
		Version:  version,
		Interval: config.Interval,
		DNSLog:   config.DNSDomain,
		HTTPLog:  config.HTTPLog,
	}

	if err := tgNotify.SendMarkdown("", notify.RenderInitialMsg(msg)); err != nil {
		return tgNotify, err
	}

	return tgNotify, err
}

func runDBMigration(migrationURL string, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create new migrate instance")
	}

	if err = migration.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatal().Err(err).Msg("failed to run migrate up")
	}

	log.Info().Msg("db migrated successfully")
}

func runDNSServer(config util.Config, store db.Store, telegram *notify.Telegram) {
	dnsServer, err := dns.NewDNSServer(config, store, telegram)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create DNS server")
	}

	err = dnsServer.Start()
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start DNS server")
	}
}

func runGinServer(config util.Config, store db.Store, telegram *notify.Telegram) {
	server, err := api.NewServer(config, store, telegram)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server")
	}

	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start server")
	}
}
