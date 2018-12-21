package main

import (
	"flag"
	"fmt"
	log "github.com/go-kit/kit/log/logrus"
	"github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
	"github.com/mashenjun/courier/com/wsStorage"
	"github.com/mashenjun/courier/pkg/endpoint"
	"github.com/mashenjun/courier/pkg/transport"
	"github.com/sirupsen/logrus"
	"net/http"
	"os/signal"
	"strings"
	"syscall"

	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/mashenjun/courier/pkg/service"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"os"
	"text/tabwriter"
	kitlog "github.com/go-kit/kit/log"
)

const (
	zipkinPath = "/api/v2/spans"
	serviceName = "courierV2"
)

func main() {

	fs := flag.NewFlagSet("courier", flag.ExitOnError)
	var (
		httpAddr  = fs.String("http-addr", ":8081", "HTTP listen address")
		zipkinAddr = fs.String("zipkin-addr", "100.100.62.190:9411", "Enable Zipkin v2 tracing (zipkin-go) using a Reporter URL e.g. http://localhost:9411/api/v2/spans")
		)
	fs.Usage = usageFor(fs, os.Args[0]+" [flags]")
	fs.Parse(os.Args[1:])
	var logger kitlog.Logger
	{
		logruslogger := logrus.New()
		logruslogger.SetFormatter(&logrus.JSONFormatter{})
		logruslogger.SetOutput(os.Stderr)
		logger = log.NewLogrusLogger(logruslogger)
		//logger = kitlog.With(logger, "ts", kitlog.DefaultTimestampUTC)
		logger = kitlog.With(logger, "caller", kitlog.DefaultCaller)
	}

	storage, err := wsStorage.NewStorage()
	if err != nil {
		logger.Log("new ws storage err", err)
	}
	// init prometheus client
	fieldKeys := []string{"method"}
	requestCount := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "courier",
		Subsystem: "courier",
		Name:      "request_count",
		Help:      "Number of requests received.",
	}, fieldKeys)
	requestLatency := kitprometheus.NewHistogramFrom(stdprometheus.HistogramOpts{
		Namespace: "courier",
		Subsystem: "courier",
		Name:      "request_latency_microseconds",
		Help:      "Total duration of requests in microseconds.",
	}, fieldKeys)
	var tracer *zipkin.Tracer
	{
		// init zipkin client
		zipkinHTTPEndpoint := fmt.Sprintf("http://%v%v", *zipkinAddr, zipkinPath)
		// create an instance of the HTTP Reporter.
		reporter := zipkinhttp.NewReporter(zipkinHTTPEndpoint)

		// create our tracer's local endpoint (how the service is identified in Zipkin).
		localEndpoint, err := zipkin.NewEndpoint(serviceName, *httpAddr)
		if err != nil {
			logger.Log("zipkin err", err)
		}
		// create our tracer instance.
		tracer, err = zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(localEndpoint))
	}
	var srv service.Service
	srv = service.New(logger, storage)
	// set InstrumentMiddleware as prometheus provider
	srv = service.InstrumentMiddleware{requestCount, requestLatency, srv}
	endpoints := endpoint.New(srv, logger)
	httpHandler := transport.NewHTTPHandler(endpoints, logger, tracer)

	// Interrupt handler.
	errc := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	// HTTP transport.
	go func() {
		sp := strings.Split(*httpAddr, ":")
		port := sp[len(sp)-1]
		logger.Log("transport", "HTTP", "addr", port)
		errc <- http.ListenAndServe(":"+port, httpHandler)
	}()

	// Run!
	logger.Log("exit", <-errc)

}

func usageFor(fs *flag.FlagSet, short string) func() {
	return func() {
		fmt.Fprintf(os.Stderr, "USAGE\n")
		fmt.Fprintf(os.Stderr, "  %s\n", short)
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "FLAGS\n")
		w := tabwriter.NewWriter(os.Stderr, 0, 2, 2, ' ', 0)
		fs.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(w, "\t-%s %s\t%s\n", f.Name, f.DefValue, f.Usage)
		})
		w.Flush()
		fmt.Fprintf(os.Stderr, "\n")
	}
}
