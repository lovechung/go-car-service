package ex

import "github.com/lovechung/api-base/api/car"

var (
	CarNotFound = car.ErrorCarNotFound("该汽车不存在")
)
