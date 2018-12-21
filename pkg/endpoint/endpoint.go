package endpoint

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/mashenjun/courier/pkg/service"
)

type Endpoints struct {
	PingEndpoint endpoint.Endpoint
	SubscribeEndpoint endpoint.Endpoint
	SendEndpoint endpoint.Endpoint
	CloseEndpoint endpoint.Endpoint
}

func New(srv service.Service, logger log.Logger) Endpoints {
	var eps Endpoints
	eps.PingEndpoint = MakePingEndpoint(srv)
	eps.SubscribeEndpoint = MakeSubscribeEndpoint(srv)
	eps.SendEndpoint = MakeSendEndpoint(srv)
	eps.CloseEndpoint = MakeCloseEndpoint(srv)
	// add logging middleware
	eps.PingEndpoint = LoggingMiddleware(log.With(logger, "method", "ping"))(eps.PingEndpoint)
	eps.SubscribeEndpoint = LoggingMiddleware(log.With(logger, "method", "subscribe"))(eps.SubscribeEndpoint)
	eps.SendEndpoint = LoggingMiddleware(log.With(logger, "method", "send"))(eps.SendEndpoint)
	eps.CloseEndpoint = LoggingMiddleware(log.With(logger, "method", "close"))(eps.CloseEndpoint)
	return eps
}

func MakePingEndpoint(srv service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		pong, err := srv.Ping(ctx, struct {}{})
		return pong, err
	}
}

func MakeSubscribeEndpoint(srv service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		return srv.Subscribe(ctx, request.(service.SubscribeReq))
	}
}

func MakeSendEndpoint(srv service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		return srv.Send(ctx, request.(service.SendReq))
	}
}

func MakeCloseEndpoint(srv service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		return srv.Close(ctx, request.(service.CloseReq))
	}
}
