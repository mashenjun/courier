package service

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/mashenjun/courier/com"
	"github.com/mashenjun/courier/com/wsStorage"
	"github.com/rs/xid"
	"net/http"
	"time"
)

type Service interface {
	// todo
	Ping(ctx context.Context, req struct{}) (*PingResp, error)
	Subscribe(ctx context.Context, req SubscribeReq) (*SubscribeResp, error)
	Send(ctx context.Context, req SendReq) (*SendResp, error)
	Close(ctx context.Context, req CloseReq) (*CloseResp, error)
}

// service func also fit the endpoint interface
type service struct {
	logger  log.Logger
	storage *wsStorage.WsStorage
}

func New(logger log.Logger, wsStorage *wsStorage.WsStorage) Service {
	srv := service{
		logger:  logger,
		storage: wsStorage,
	}
	return &srv
}

type PingResp struct {
	Data string `json:"data"`
}

func (srv *service) Ping(ctx context.Context, req struct{}) (*PingResp, error) {
	return &PingResp{Data: "pong"}, nil
}

type SubscribeReq struct {
	W              http.ResponseWriter
	R              *http.Request
	ResponseHeader http.Header
}

type SubscribeResp struct {
	Key string `json:"key"`
}

func (srv *service) Subscribe(ctx context.Context, req SubscribeReq) (resp *SubscribeResp, err error) {
	key := xid.New().String()
	conn, err := wsStorage.NewConn(req.W, req.R, req.ResponseHeader)
	if err != nil {
		srv.logger.Log("err", err)
		return nil, com.InternalError
	}
	defer func() {
		if err != nil {
			conn.CloseWithMessage(err.Error())
		}
	}()
	conn.WriteJSON(key)
	conn.KeepLive(1 * time.Second)
	if err := srv.storage.Store(key, conn); err != nil {
		return nil, com.InternalError
	}
	return &SubscribeResp{Key: key}, nil
}

type SendReq struct {
	Key     string
	Data interface{}
}

type SendResp struct {
	// nothing to return
}

func (srv *service) Send(ctx context.Context, req SendReq) (*SendResp, error) {
	conn, err := srv.storage.Load(req.Key)
	if err != nil {
		srv.logger.Log("could not load key",req.Key ,"err", err)
		return nil, com.ParameterError
	}
	if err := conn.WriteJSON(req.Data); err != nil {
		srv.logger.Log("could not send json to conn",req.Data ,"err", err)
		return nil, com.InternalError
	}
	return &SendResp{}, nil
}

type CloseReq struct {
	Key string `json:"key"`
}

type CloseResp struct {
	Key string `json:"key"`
}

func (srv *service) Close(ctx context.Context, req CloseReq) (*CloseResp, error) {
	defer srv.storage.Delete(req.Key)
	conn, err := srv.storage.Load(req.Key)
	if err != nil {
		return nil, com.InternalError
	}
	conn.Close()
	return &CloseResp{Key: req.Key}, nil
}
