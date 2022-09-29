// Code generated by ent, DO NOT EDIT.

package ent

import "car-service/internal/biz"

func (cu *CarUpdate) SetCar(input *biz.Car) *CarUpdate {

	cu.SetNillableUserID(input.UserID)

	cu.SetNillableModel(input.Model)

	cu.SetNillableRegisteredAt(input.RegisteredAt)
	return cu
}

func (cc *CarCreate) SetCar(input *biz.Car) *CarCreate {

	cc.SetNillableUserID(input.UserID)

	cc.SetNillableModel(input.Model)

	cc.SetNillableRegisteredAt(input.RegisteredAt)
	return cc
}
