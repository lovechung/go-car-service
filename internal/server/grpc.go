package server

import (
	"car-service/internal/conf"
	"car-service/internal/service"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/lovechung/api-base/api/car"
	contrib "github.com/lovechung/go-kit/contrib/metrics"
	"github.com/lovechung/go-kit/middleware/metrics"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/unit"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(c *conf.Server, cs *service.CarService, logger log.Logger) *grpc.Server {
	meter := global.Meter("car-serivce")
	requestHistogram, _ := meter.SyncInt64().Histogram("request_seconds", instrument.WithUnit(unit.Milliseconds))

	var opts = []grpc.ServerOption{
		grpc.Middleware(
			middleware.Chain(
				recovery.Recovery(),
				tracing.Server(),
				metrics.Server(
					metrics.WithSeconds(contrib.NewHistogram(requestHistogram)),
				),
				logging.Server(logger),
			),
		),
	}
	if c.Grpc.Network != "" {
		opts = append(opts, grpc.Network(c.Grpc.Network))
	}
	if c.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Grpc.Addr))
	}
	if c.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
	}
	srv := grpc.NewServer(opts...)
	car.RegisterCarServer(srv, cs)
	return srv
}
