package client

import (
	"context"
	"github.com/openzipkin/zipkin-go"
	zipkinhttpclient "github.com/openzipkin/zipkin-go/middleware/http"
	"github.com/openzipkin/zipkin-go/model"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
	"time"
)

func TestClient_Ping(t *testing.T) {
	reporter := zipkinhttp.NewReporter("http://100.100.62.190:9411/api/v2/spans")

	// create our tracer's local endpoint (how the service is identified in Zipkin).
	localEndpoint, err := zipkin.NewEndpoint("pong", "localhost:0")
	if err != nil {
		t.Fatalf("could not create endpoint: %v", err)
	}
	remoteEndpoint, err := zipkin.NewEndpoint("courierv3", "100.100.62.190:8081")
	// create our tracer instance.
	tracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(localEndpoint))
	// create global zipkin traced http client
	client, err := zipkinhttpclient.NewClient(tracer, zipkinhttpclient.ClientTrace(true))
	if err != nil {
		t.Fatalf("could not create client: %+v\n", err)
	}
	req, err := http.NewRequest("GET", "http://localhost:8081/ping", nil)
	if err != nil {
		t.Fatalf("could not create http request: %+v\n", err)
	}
	span := tracer.StartSpan("ping", zipkin.RemoteEndpoint(remoteEndpoint))
	req = req.WithContext(zipkin.NewContext(req.Context(), span))
	res, err := client.DoWithAppSpan(req, "pong")
	if err != nil {
		t.Fatalf("could not ping: %v", err)
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("could not read: %v", err)
	}
	log.Printf("%+v", string(b))
	time.Sleep(5*time.Second)
}

func TestTracer(t *testing.T) {
	var (
		serviceName        = "courierv4"
		serviceHostPort    = "localhost:8000"
		zipkinHTTPEndpoint = "http://100.100.62.190:9411/api/v2/spans"
	)

	// create an instance of the HTTP Reporter.
	reporter := zipkinhttp.NewReporter(zipkinHTTPEndpoint)

	// create our tracer's local endpoint (how the service is identified in Zipkin).
	localEndpoint, err := zipkin.NewEndpoint(serviceName, serviceHostPort)
	if err != nil {
		t.Fatalf("could not new endpoint: %+v", err)
	}
	// create our tracer instance.
	tracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(localEndpoint))
	if err != nil {
		t.Fatalf("could not new tracer: %+v", err)
	}
	span, _ := tracer.StartSpanFromContext(context.Background(), "ping", zipkin.Kind(model.Client))
	remoteEndpoint, err := zipkin.NewEndpoint("myservicev1", "100.100.62.190:8081")
	span.SetRemoteEndpoint(remoteEndpoint)
	log.Printf("%+v\n", span)
	span.Finish()
	span.Flush()
	time.Sleep(5*time.Second)
}