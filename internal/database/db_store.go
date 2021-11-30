package database

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/DIMO-INC/devices-api/internal/config"
)

const driverName = "postgres"

// instance holds a single instance of the database
var instance *DBReaderWriter

var ready bool

// once is used to ensure that there is only a single instance of the database
var once sync.Once

// DbStore holds the database connection and other stuff.
type DbStore struct {
	db    func() *sql.DB
	dbs   *DBReaderWriter
	ready *bool
}

// NewDbConnectionFromSettings sets up a db connection from the settings, only once
func NewDbConnectionFromSettings(ctx context.Context, settings *config.Settings) DbStore {
	once.Do(func() {
		instance = NewDbConnection(
			ctx,
			&ready,
			ConnectOptions{
				Retries:            5,
				RetryDelay:         time.Second * 10,
				ConnectTimeout:     time.Minute * 5,
				DSN:                settings.GetWriterDSN(true),
				MaxOpenConnections: settings.DbMaxOpenConnections,
				MaxIdleConnections: settings.DbMaxIdleConnections,
				ConnMaxLifetime:    time.Minute * 5,
				DriverName: driverName,
			},
			ConnectOptions{
				Retries:            5,
				RetryDelay:         time.Second * 10,
				ConnectTimeout:     time.Minute * 5,
				DSN:                settings.GetWriterDSN(true),
				MaxOpenConnections: settings.DbMaxOpenConnections,
				MaxIdleConnections: settings.DbMaxIdleConnections,
				ConnMaxLifetime:    time.Minute * 5,
				DriverName: driverName,
			},
		)
	})

	return DbStore{db: instance.GetWriterConn, dbs: instance, ready: &ready}
}

// IsReady returns if db is ready to connect to
func (store *DbStore) IsReady() bool {
	return *store.ready
}

// DBS returns the reader and writer databases to connect to
func (store *DbStore) DBS() *DBReaderWriter {
	return store.dbs
}