package postgres

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/DIMO-INC/devices-api/internal/config"
)

// instance holds a single instance of the database
var instance *DBS

var ready bool

// once is used to ensure that there is only a single instance of the database
var once sync.Once

// DbStore holds the database connection and other stuff.
type DbStore struct {
	db    func() *sql.DB
	dbs   *DBS
	ready *bool
}

// NewDbStore either generates or returns the connection to the database.
func NewDbStore(ctx context.Context, settings config.Settings) DbStore {
	once.Do(func() {
		instance = NewDBS(
			ctx,
			&ready,
			Options{
				Retries:            5,
				Delay:              time.Second * 15,
				Timeout:            time.Minute * 5,
				DSN:                settings.WriterConnectionString(),
				MaxOpenConnections: settings.DbMaxOpenConnections,
				MaxIdleConnections: settings.DbMaxIdleConnections,
				ConnMaxLifetime:    time.Minute * 5,
			},
			Options{
				Retries:            5,
				Delay:              time.Second * 15,
				Timeout:            time.Minute * 5,
				DSN:                settings.WriterConnectionString(),
				MaxOpenConnections: settings.DbMaxOpenConnections,
				MaxIdleConnections: settings.DbMaxIdleConnections,
				ConnMaxLifetime:    time.Minute * 5,
			},
			settings.ServiceName,
		)
	})

	return DbStore{db: instance.GetWriterConn, dbs: instance, ready: &ready}
}

//Ready returns if db is ready to connect to
func (store *DbStore) Ready() bool {
	return *store.ready
}

//DBS returns the dbs to connect to
func (store *DbStore) DBS() *DBS {
	return store.dbs
}