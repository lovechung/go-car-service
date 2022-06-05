package data

import (
	"car-service/internal/biz"
	"car-service/internal/data/ent"
	"car-service/internal/data/ent/car"
	"car-service/internal/data/ent/predicate"
	ex "car-service/internal/pkg/errors"
	"context"
	"github.com/go-kratos/kratos/v2/log"
	userV1 "github.com/lovechung/api-base/api/user"
	"github.com/lovechung/go-kit/util/pagination"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type carRepo struct {
	data *Data
	log  *log.Helper
}

func NewCarRepo(data *Data, logger log.Logger) biz.CarRepo {
	return &carRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r carRepo) ListCar(ctx context.Context, page, pageSize int, model *string) ([]*biz.CarReply, int, error) {
	var list []*biz.CarReply
	// 组装查询条件
	cond := make([]predicate.Car, 0)
	if model != nil {
		cond = append(cond, car.ModelContains(*model))
	}

	q := r.data.db.Car.Query().Where(cond...)
	// 查询总数
	total := q.CountX(ctx)
	// 查询列表
	cars := q.Offset(pagination.GetOffset(page, pageSize)).
		Limit(pageSize).
		Order(ent.Desc(car.FieldRegisteredAt)).
		AllX(ctx)

	if len(cars) == 0 {
		return list, 0, nil
	}
	// 查询用户名称
	userIds := make([]int64, 0)
	for _, c := range cars {
		userIds = append(userIds, c.UserID)
	}

	// grpc调用
	reply, err := r.data.uc.GetUserNameMap(ctx, &userV1.UserIdsReq{Ids: userIds})
	if err != nil {
		return list, 0, err
	}
	for _, c := range cars {
		list = append(list, &biz.CarReply{
			Id:           c.ID,
			Model:        c.Model,
			RegisteredAt: c.RegisteredAt,
			UserName:     reply.NameMap[c.UserID],
		})
	}

	return list, total, nil
}

func (r carRepo) GetById(ctx context.Context, id int64) (*biz.CarReply, error) {
	c, err := r.data.db.Car.Get(ctx, id)
	if err != nil {
		return nil, ex.CarNotFound
	}

	// grpc调用
	reply, err := r.data.uc.GetUserName(ctx, &wrapperspb.Int64Value{Value: c.UserID})
	if err != nil {
		return nil, err
	}

	return &biz.CarReply{
		Id:           c.ID,
		Model:        c.Model,
		RegisteredAt: c.RegisteredAt,
		UserName:     reply.Value,
	}, nil
}

func (r carRepo) Save(ctx context.Context, c *biz.Car) (int64, error) {
	rsp, err := r.data.db.Car.
		Create().
		SetCar(c).
		Save(ctx)
	return rsp.ID, err
}

func (r carRepo) Update(ctx context.Context, c *biz.Car) error {
	return r.data.db.Car.
		Update().
		Where(car.ID(c.ID)).
		SetCar(c).
		Exec(ctx)
}

func (r carRepo) Delete(ctx context.Context, id int64) error {
	return r.data.db.Car.
		DeleteOneID(id).
		Exec(ctx)
}
