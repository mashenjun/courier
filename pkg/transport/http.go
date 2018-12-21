package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/zipkin"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/mashenjun/courier/com"
	"github.com/mashenjun/courier/pkg/endpoint"
	"github.com/mashenjun/courier/pkg/service"
	stdzipkin "github.com/openzipkin/zipkin-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"html/template"
	"net/http"
)

const (
	responseWriterKey = "rw"
)

func decodeHTTPGenericRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	// todo
	fmt.Printf("%#v\n", ctx)
	return nil, nil
}

func decodeHTTPSubscribeRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var req service.SubscribeReq
	req.W = ctx.Value(responseWriterKey).(http.ResponseWriter)
	req.R = r
	return req, nil
}

func decodeHTTPSendRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var req service.SendReq
	key := mux.Vars(r)["key"]
	req.Key = key
	tmp :=  map[string]interface{}{}
	err := json.NewDecoder(r.Body).Decode(&tmp)
	req.Data = tmp
	return req, err
}

func encodeHTTPGenericResponse(ctx context.Context, w http.ResponseWriter, resp interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	response := struct {
		ErrorCode int         `json:"error_code"`
		Data      interface{} `json:"data"` // data should contain page info
	}{}
	response.Data = resp
	return json.NewEncoder(w).Encode(response)
}

func encodeError(ctx context.Context, err error, w http.ResponseWriter) {
	// process the given error and set status code

	if sErr, ok := err.(*com.ServiceError); ok {
		w.WriteHeader(sErr.StatusCode)

	}else {
		w.WriteHeader(http.StatusInternalServerError)
		err = com.ParameterError
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(err)
}

func NewHTTPHandler(endpoints endpoint.Endpoints, logger log.Logger, tracer *stdzipkin.Tracer) http.Handler {
	options := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(encodeError),
		httptransport.ServerErrorLogger(logger),
		zipkin.HTTPServerTrace(tracer),
	}
	r := mux.NewRouter()
	r.Handle("/ping", httptransport.NewServer(
		endpoints.PingEndpoint,
		decodeHTTPGenericRequest,
		encodeHTTPGenericResponse,
		options...,
	)).Methods(http.MethodGet)
	r.HandleFunc("/subscribe",  func(w http.ResponseWriter,
		r *http.Request){
			httpSrv := httptransport.NewServer(
				endpoints.SubscribeEndpoint,
				decodeHTTPSubscribeRequest,
				encodeHTTPGenericResponse,
				append(options, httptransport.ServerBefore())...,
			)
			newCtx := context.WithValue(r.Context(),"rw", w)
			req := r.WithContext(newCtx)
			httpSrv.ServeHTTP(w, req)
	}).Methods(http.MethodGet)
	r.Handle("/send/{key}", httptransport.NewServer(
		endpoints.SendEndpoint,
		decodeHTTPSendRequest,
		encodeHTTPGenericResponse,
		options...,
		)).Methods(http.MethodPost)
	r.HandleFunc("/home",func(w http.ResponseWriter,
		r *http.Request){
		homeTemplate.Execute(w, "ws://"+r.Host+"/subscribe")
	}).Methods(http.MethodGet)
	// enable metrics exporter for prometheus
	r.Handle("/metrics", promhttp.Handler())
	return r
}

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<script>  
window.addEventListener("load", function(evt) {
    var output = document.getElementById("output");
    var input = document.getElementById("input");
    var ws;
    var print = function(message) {
        var d = document.createElement("div");
        d.innerHTML = message;
        output.appendChild(d);
    };
    document.getElementById("open").onclick = function(evt) {
        if (ws) {
            return false;
        }
        ws = new WebSocket("{{.}}");
        ws.onopen = function(evt) {
            print("OPEN");
        }
        ws.onclose = function(evt) {
            print("CLOSE");
            ws = null;
        }
        ws.onmessage = function(evt) {
            print("RESPONSE: " + evt.data);
        }
        ws.onerror = function(evt) {
            print("ERROR: " + evt.data);
        }
        return false;
    };
    document.getElementById("send").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        print("SEND: " + input.value);
        ws.send(input.value);
        return false;
    };
    document.getElementById("close").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        ws.close();
        return false;
    };
});
</script>
</head>
<body>
<table>
<tr><td valign="top" width="50%">
<p>Click "Open" to create a connection to the server, 
"Send" to send a message to the server and "Close" to close the connection. 
You can change the message and send multiple times.
<p>
<form>
<button id="open">Open</button>
<button id="close">Close</button>
<p><input id="input" type="text" value="Hello world!">
<button id="send">Send</button>
</form>
</td><td valign="top" width="50%">
<div id="output"></div>
</td></tr></table>
</body>
</html>
`))
