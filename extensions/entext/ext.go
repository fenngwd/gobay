package entext

import (
	"database/sql"
	"errors"
	"math/rand"
	"time"

	"github.com/facebook/ent/dialect"
	entsql "github.com/facebook/ent/dialect/sql"
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

	IsNotFound          func(error) bool
	IsConstraintFailure func(error) bool
	IsNotSingular       func(error) bool

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
		// 增加随机延时，避免部署时多个 Pod 连接到期时间接近，导致数据库新建连接飙升
		rand.Seed(int64(time.Now().Second()))
		db.SetConnMaxLifetime(config.GetDuration("conn_max_lifetime") + time.Duration(rand.Intn(30)) * time.Second)
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
