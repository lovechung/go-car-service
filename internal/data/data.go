package data

import (
	"car-service/internal/biz"
	"car-service/internal/conf"
	"car-service/internal/data/ent"
	"context"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"fmt"
	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/wire"
	consulApi "github.com/hashicorp/consul/api"
	"github.com/lovechung/api-base/api/user"
	"github.com/rueian/rueidis"
	"github.com/rueian/rueidis/rueidiscompat"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var ProviderSet = wire.NewSet(
	NewTransaction,
	NewData,
	NewDB,
	NewRedis,
	NewRegistrar,
	NewDiscovery,
	NewCarRepo,
	NewUserServiceClient,
)

type Data struct {
	db     *ent.Client
	rds    rueidis.Client
	rdsCmd rueidiscompat.Cmdable
	uc     user.UserClient
}

type contextTxKey struct{}

func (d *Data) ExecTx(ctx context.Context, f func(ctx context.Context) error) error {
	tx, err := d.db.Tx(ctx)
	if err != nil {
		return err
	}
	ctx = context.WithValue(ctx, contextTxKey{}, tx)
	if err := f(ctx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (d *Data) Car(ctx context.Context) *ent.CarClient {
	tx, ok := ctx.Value(contextTxKey{}).(*ent.Tx)
	if ok {
		return tx.Car
	}
	return d.db.Car
}

func NewTransaction(d *Data) biz.Transaction {
	return d
}

func NewData(db *ent.Client, rds rueidis.Client, uc user.UserClient, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
		if err := db.Close(); err != nil {
			log.Error(err)
		}
		rds.Close()
	}

	return &Data{
		db:     db,
		rds:    rds,
		rdsCmd: rueidiscompat.NewAdapter(rds),
		uc:     uc,
	}, cleanup, nil
}

func NewDB(conf *conf.Data, logger log.Logger) *ent.Client {
	thisLog := log.NewHelper(logger)

	drv, err := sql.Open(
		conf.Database.Driver,
		conf.Database.Source,
	)
	// 打印sql日志
	sqlDrv := dialect.DebugWithContext(drv, func(ctx context.Context, i ...interface{}) {
		thisLog.WithContext(ctx).Debug(i...)
		// 开启db trace
		tracer := otel.Tracer("entgo.io/ent")
		_, span := tracer.Start(ctx,
			"query",
			trace.WithAttributes(
				attribute.String("sql", fmt.Sprint(i...)),
			),
		)
		defer span.End()
	})
	db := ent.NewClient(ent.Driver(sqlDrv))

	if err != nil {
		thisLog.Fatalf("数据库连接失败: %v", err)
	}
	// 运行自动创建表
	//if err := db.Schema.Create(context.Background(), migrate.WithForeignKeys(false)); err != nil {
	//	thisLog.Fatalf("创建表失败: %v", err)
	//}
	return db
}

func NewRedis(conf *conf.Data, logger log.Logger) rueidis.Client {
	thisLog := log.NewHelper(logger)

	client, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress:      []string{conf.Redis.Addr},
		Password:         conf.Redis.Password,
		SelectDB:         int(conf.Redis.Db),
		ConnWriteTimeout: conf.Redis.WriteTimeout.AsDuration(),
	})
	// todo rueidisotel支持的otel版本过低，导致启动报错
	//client = rueidisotel.WithClient(client)

	if err != nil {
		thisLog.Fatalf("redis连接失败: %v", err)
	}
	return client
}

func NewRegistrar(conf *conf.Registry) registry.Registrar {
	c := consulApi.DefaultConfig()
	c.Address = conf.Consul.Address
	c.Scheme = conf.Consul.Scheme
	c.Token = conf.Consul.Token
	cli, err := consulApi.NewClient(c)
	if err != nil {
		panic(err)
	}
	r := consul.New(cli, consul.WithHealthCheck(conf.Consul.HealthCheck))
	return r
}

func NewDiscovery(conf *conf.Registry) registry.Discovery {
	c := consulApi.DefaultConfig()
	c.Address = conf.Consul.Address
	c.Scheme = conf.Consul.Scheme
	c.Token = conf.Consul.Token
	cli, err := consulApi.NewClient(c)
	if err != nil {
		panic(err)
	}
	r := consul.New(cli, consul.WithHealthCheck(false))
	return r
}

func NewUserServiceClient(r registry.Discovery) user.UserClient {
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("discovery:///user-service"),
		grpc.WithDiscovery(r),
		grpc.WithMiddleware(
			recovery.Recovery(),
			tracing.Client(),
		),
	)
	if err != nil {
		panic(err)
	}
	return user.NewUserClient(conn)
}
