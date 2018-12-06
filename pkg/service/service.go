package service

import (
	"context"
	"github.com/go-kit/kit/log"
)

type Service interface {
	// todo
	Ping(ctx context.Context) (PingResp, error)
}

type service struct {
}

func New(logger log.Logger) Service {
	return &service{}
}

type PingResp struct {
	Data string `json:"data"`
}

func (srv *service) Ping(ctx context.Context) (PingResp, error) {
	return PingResp{Data: "pong"}, nil
}
