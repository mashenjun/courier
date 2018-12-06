package transport

import (
	"context"
	"encoding/json"
	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/mashenjun/courier/pkg/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"net/http"
)

func decodeHTTPGenericRequest(ctx context.Context, r *http.Request) (interface{}, error){
	return nil, nil
}
func encodeHTTPGenericResponse(ctx context.Context, w http.ResponseWriter, resp interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	response := struct {
		ErrorCode int `json:"error_code"`
		Data interface{} `json:"data"`
	}{}
	response.Data = resp
	return json.NewEncoder(w).Encode(response)
}


func NewHTTPHandler(endpoints endpoint.Endpoints, logger log.Logger) http.Handler {
	options := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(func(ctx context.Context, err error, w http.ResponseWriter){
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Write([]byte(err.Error()))
		}),

		}
	r := mux.NewRouter()
	r.Handle("/ping", httptransport.NewServer(
		endpoints.PingEndpoint,
		decodeHTTPGenericRequest,
		encodeHTTPGenericResponse,
		options...,
	))
	return r
}
