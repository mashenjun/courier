package main

import (
	"flag"
	"fmt"
	"github.com/mashenjun/courier/pkg/endpoint"
	"github.com/mashenjun/courier/pkg/transport"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/go-kit/kit/log"
	"github.com/mashenjun/courier/pkg/service"
	"os"
	"text/tabwriter"
)

func main() {

	fs := flag.NewFlagSet("courier", flag.ExitOnError)
	var (
		httpAddr = fs.String("http-addr", ":8081", "HTTP listen address")
	)
	fs.Usage = usageFor(fs, os.Args[0]+" [flags]")
	fs.Parse(os.Args[1:])

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}
	var (
		srv         = service.New(logger)
		endpoints   = endpoint.New(srv, logger)
		httpHandler = transport.NewHTTPHandler(endpoints, logger)
	)
	// Interrupt handler.
	errc := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	// HTTP transport.
	go func() {
		logger.Log("transport", "HTTP", "addr", *httpAddr)
		errc <- http.ListenAndServe(*httpAddr, httpHandler)
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
