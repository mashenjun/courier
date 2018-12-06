package endpoint

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/mashenjun/courier/pkg/service"
)

type Endpoints struct {
	PingEndpoint endpoint.Endpoint
}

func New(srv service.Service, logger log.Logger) Endpoints {
	var pingEndpoint endpoint.Endpoint
	{
		pingEndpoint = MakePingEndpoint(srv)
		pingEndpoint = LoggingMiddleware(log.With(logger, "method", "Ping"))(pingEndpoint)
	}

	return Endpoints{
		PingEndpoint: pingEndpoint,
	}
}

func MakePingEndpoint(srv service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		pong, err := srv.Ping(ctx)
		return pong, err
	}
}
