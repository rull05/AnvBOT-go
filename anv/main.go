package anv

import (
	"context"
	"errors"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)

type DBDialect string

const (
	SQLite3  DBDialect = "sqlite3"
	Postgres DBDialect = "postgres"
)

var log = waLog.Stdout("Anv", "DEBUG", true)

var validDBDialects = map[DBDialect]bool{
	SQLite3:  true,
	Postgres: true,
}

// Option represents a configuration option for Anv.
type Option func(*Config)

// Config represents the configuration for Anv.
type Config struct {
	DBName          string
	LogLevel        string
	RequestFullSync bool
	DBDialect       string
}

// WithDBName sets the database name in the configuration.
func WithDBName(dbName string) Option {
	return func(c *Config) {
		c.DBName = fmt.Sprintf("file:%s?_foreign_keys=on", dbName)
	}
}

// WithLogLevel sets the log level in the configuration.
func WithLogLevel(logLevel string) Option {
	return func(c *Config) {
		c.LogLevel = logLevel
	}
}

// WithRequestFullSync enables full sync in the configuration.
func WithRequestFullSync() Option {
	return func(c *Config) {
		c.RequestFullSync = true
	}
}

// WithDBDialect sets the database dialect in the configuration.
func WithDBDialect(dialect string) Option {
	return func(c *Config) {
		if _, ok := validDBDialects[DBDialect(dialect)]; !ok {
			log.Errorf("Invalid DBDialect: %s\n", dialect, "Valid options are: %v", validDBDialects)
			os.Exit(1)
		}
		c.DBDialect = dialect
	}
}

// Anv represents the Anv bot.
type Anv struct {
	Client     *whatsmeow.Client
	EvtHandler *EvtHandler
}

// NewConfig creates a new configuration with the given options.
func NewConfig(options ...Option) *Config {
	config := &Config{
		DBName:          "file:anv.db?_foreign_keys=on",
		LogLevel:        "INFO",
		RequestFullSync: false,
		DBDialect:       "sqlite3",
	}

	for _, option := range options {
		option(config)
	}

	return config
}

// Init initializes the Anv bot with the given configuration.
func NewAnv(config *Config) (*Anv, error) {
	if config == nil {
		config = NewConfig()
	}
	anv := &Anv{}
	// log info about the config
	// output: "Starting Anv with config:
	// DATABASE: anv.db?_foreign_keys=on LOG: INFO RequestFullSync: false DBDialect: sqlite3"
	log.Infof("Starting Anv with config DATABASE: %s LOG: %s RequestFullSync: %t DBDialect: %s", config.DBName, config.LogLevel, config.RequestFullSync, config.DBDialect)
	if config.RequestFullSync {
		store.DeviceProps.RequireFullSync = proto.Bool(true)
		store.DeviceProps.HistorySyncConfig = &waProto.DeviceProps_HistorySyncConfig{
			FullSyncDaysLimit:   proto.Uint32(3650),
			FullSyncSizeMbLimit: proto.Uint32(102400),
			StorageQuotaMb:      proto.Uint32(102400),
		}
	}
	log := waLog.Stdout("Client", config.LogLevel, true)
	dbLog := waLog.Stdout("Database", config.LogLevel, true)
	storeContainer, err := sqlstore.New(config.DBDialect, config.DBName, dbLog)
	if err != nil {
		log.Errorf("Error connecting to database: %v", err)
		return anv, err
	}
	device, err := storeContainer.GetFirstDevice()
	if err != nil {
		log.Errorf("Error getting first device: %v", err)
		return anv, err
	}
	anv.Client = whatsmeow.NewClient(device, log)
	anv.EvtHandler = NewEvtHandler(anv.Client, DEBUG)

	anv.Client.AddEventHandler(anv.EvtHandler.HandleEvent)
	return anv, nil
}

// anv Connect
func (anv *Anv) Start() error {
	ch, err := anv.Client.GetQRChannel(context.Background())
	if err != nil {
		// This error means that we're already logged in, so ignore it.
		if !errors.Is(err, whatsmeow.ErrQRStoreContainsID) {
			anv.Client.Log.Errorf("Failed to get QR channel: %v", err)
			return err
		}
	} else {
		go func() {
			for evt := range ch {
				log.Debugf("QR channel event: %v", evt)
				if evt.Event == "code" {
					qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				} else {
					anv.Client.Log.Infof("QR channel result: %s", evt.Event)
				}
			}
		}()
	}
	log.Infof("Connecting to WhatsApp...")
	err = anv.Client.Connect()
	if err != nil {
		anv.Client.Log.Errorf("Error connecting to WhatsApp: %v", err)
		return err
	}
	return nil

}
