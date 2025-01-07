package test

import (
	"fmt"
	"log"
	"time"

	postgres "github.com/Bofry/worker-postgres"
)

var (
	defaultLogger *log.Logger = log.New(log.Writer(), "[worker-postgres-test] ", log.LstdFlags|log.Lmsgprefix|log.LUTC)
)

type (
	Host postgres.Worker

	Config struct {
		PostgresHost     string `env:"*TEST_POSTGRES_HOST"       yaml:"-"`
		PostgresPort     uint16 `env:"*TEST_POSTGRES_PORT"       yaml:"-"`
		PostgresDatabase string `env:"*TEST_POSTGRES_DATABASE"   yaml:"-"`
		PostgresUser     string `env:"*TEST_POSTGRES_USER"       yaml:"-"`
		PostgresPassword string `env:"*TEST_POSTGRES_PASSWORD"   yaml:"-"`

		// jaeger
		JaegerTraceUrl string `yaml:"jaegerTraceUrl"`
		JaegerQueryUrl string `yaml:"jaegerQueryUrl"`

		CreateReplicationSlotSource string `yaml:"createReplicationSlotSource"`
	}
)

func (h *Host) Init(conf *Config) {
	config := postgres.Config{
		Host:           conf.PostgresHost,
		Port:           conf.PostgresPort,
		Database:       conf.PostgresDatabase,
		User:           conf.PostgresUser,
		Password:       conf.PostgresPassword,
		PollingTimeout: 1 * time.Second,
		ReplicationOptions: postgres.ConfigureReplicationOptions().
			WithPluginArgs(`"pretty-print" 'true'`),
	}

	if len(conf.CreateReplicationSlotSource) > 0 {
		err := h.ReplicationSlotSourceProvider.ScanString(conf.CreateReplicationSlotSource)
		if err != nil {
			panic(err)
		}
		fmt.Println("ReplicationSlotSourceProvider::", h.ReplicationSlotSourceProvider.Sources())
	}

	h.DisableAutoAck = false
	h.Config = &config
}

func (h *Host) OnError(err error) (disposed bool) {
	return false
}
