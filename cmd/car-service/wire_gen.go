// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"car-service/internal/biz"
	"car-service/internal/conf"
	"car-service/internal/data"
	"car-service/internal/server"
	"car-service/internal/service"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
)

// Injectors from wire.go:

// wireApp init kratos application.
func wireApp(confServer *conf.Server, confData *conf.Data, registry *conf.Registry, logger log.Logger) (*kratos.App, func(), error) {
	client := data.NewDB(confData, logger)
	rueidisClient := data.NewRedis(confData, logger)
	discovery := data.NewDiscovery(registry)
	userClient := data.NewUserServiceClient(confServer, discovery)
	dataData, cleanup, err := data.NewData(client, rueidisClient, userClient, logger)
	if err != nil {
		return nil, nil, err
	}
	carRepo := data.NewCarRepo(dataData, logger)
	transaction := data.NewTransaction(dataData)
	carUseCase := biz.NewCarUseCase(carRepo, transaction, logger)
	carService := service.NewCarService(carUseCase, logger)
	grpcServer := server.NewGRPCServer(confServer, carService, logger)
	registrar := data.NewRegistrar(registry)
	app := newApp(logger, grpcServer, registrar)
	return app, func() {
		cleanup()
	}, nil
}
