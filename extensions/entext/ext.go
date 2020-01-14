package entext

import (
	"database/sql"
	"errors"
	"github.com/facebookincubator/ent/dialect"
	entsql "github.com/facebookincubator/ent/dialect/sql"
	"github.com/shanbay/gobay"
	"go.elastic.co/apm/module/apmsql"
)

const (
	defaultMaxOpenConns = 15
	defaultMaxIdleConns = 5
)

type Client interface {
	Close() error
}

type EntExt struct {
	NS        string
	NewClient func(interface{}) Client
	Driver    func(dialect.Driver) interface{}

	drv    *entsql.Driver
	client Client
	app    *gobay.Application
}

func (d *EntExt) Object() interface{} { return d.client }

func (d *EntExt) Application() *gobay.Application { return d.app }

func (d *EntExt) Init(app *gobay.Application) error {
	if d.NS == "" {
		return errors.New("lack of NS")
	}
	d.app = app
	config := gobay.GetConfigByPrefix(app.Config(), d.NS, true)
	config.SetDefault("max_open_conns", defaultMaxOpenConns)
	config.SetDefault("max_idle_conns", defaultMaxIdleConns)
	dbURL := config.GetString("url")
	dbDriver := config.GetString("driver")

	var db *sql.DB
	var err error
	if app.Config().GetBool("elastic_apm_enable") {
		db, err = apmsql.Open(dbDriver, dbURL)
	} else {
		db, err = sql.Open(dbDriver, dbURL)
	}
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(config.GetInt("max_open_conns"))
	db.SetMaxIdleConns(config.GetInt("max_idle_conns"))
	if config.IsSet("conn_max_lifetime") {
		db.SetConnMaxLifetime(config.GetDuration("conn_max_lifetime"))
	}
	drv := entsql.OpenDB(dbDriver, db)
	d.drv = drv
	d.client = d.NewClient(d.Driver(drv))
	return nil
}

func (d *EntExt) Close() error { return d.client.Close() }

// DB 获取数据库，ent目前还不够完善，某些场景下还需要执行sql
func (d *EntExt) DB() *sql.DB {
	return d.drv.DB()
}