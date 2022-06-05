package service

import (
	"car-service/internal/biz"
	"context"
	"github.com/go-kratos/kratos/v2/log"
	v1 "github.com/lovechung/api-base/api/car"
	"github.com/lovechung/go-kit/util/pagination"
	"github.com/lovechung/go-kit/util/time"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type CarService struct {
	v1.UnimplementedCarServer

	uc  *biz.CarUseCase
	log *log.Helper
}

func NewCarService(uc *biz.CarUseCase, logger log.Logger) *CarService {
	return &CarService{uc: uc, log: log.NewHelper(logger)}
}

func (s *CarService) ListCar(ctx context.Context, req *v1.ListCarReq) (*v1.ListCarReply, error) {
	page, pageSize := pagination.GetPage(req.Page, req.PageSize)
	list, total, err := s.uc.ListCar(ctx, page, pageSize, req.Model)

	rsp := &v1.ListCarReply{}
	rsp.Total = int32(total)
	for _, car := range list {
		carInfo := ConvertToCarReply(car)
		rsp.List = append(rsp.List, carInfo)
	}
	return rsp, err
}

func (s *CarService) GetCar(ctx context.Context, req *wrapperspb.Int64Value) (*v1.CarReply, error) {
	c, err := s.uc.GetCarById(ctx, req.Value)
	if err != nil {
		return nil, err
	}
	return ConvertToCarReply(c), err
}

func ConvertToCarReply(c *biz.CarReply) *v1.CarReply {
	return &v1.CarReply{
		Id:           c.Id,
		Username:     c.UserName,
		Model:        c.Model,
		RegisteredAt: t.Format(c.RegisteredAt),
	}
}

func (s *CarService) SaveCar(ctx context.Context, req *v1.SaveCarReq) (*emptypb.Empty, error) {
	err := s.uc.SaveCar(ctx, &biz.Car{
		UserID: &req.UserId,
		Model:  &req.Model,
	})
	return nil, err
}

func (s *CarService) TradeCar(ctx context.Context, req *v1.TradeCarReq) (*emptypb.Empty, error) {
	err := s.uc.UpdateCar(ctx, &biz.Car{
		ID:     req.Id,
		UserID: &req.UserId,
	})
	return nil, err
}

func (s *CarService) DeleteCar(ctx context.Context, req *wrapperspb.Int64Value) (*emptypb.Empty, error) {
	err := s.uc.DeleteCar(ctx, req.Value)
	return nil, err
}
