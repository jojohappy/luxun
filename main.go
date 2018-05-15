package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jojohappy/luxun/pkg/controller"
	luxunhttp "github.com/jojohappy/luxun/pkg/http"
	"github.com/jojohappy/luxun/pkg/stream"
)

var listenIp = flag.String("listen_ip", "", "IP to listen on, defaults all")
var listenPort = flag.Int("port", 9280, "listen port")
var prometheusEndpoint = flag.String("prometheus_endpoint", "/metrics", "Endpoint to expose Prometheus metrics on")

func main() {
	flag.Parse()
	stream.Init()
	mux := http.NewServeMux()
	luxunhttp.RegisterHandler(mux, *prometheusEndpoint)

	go controller.Execute()

	s := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", *listenIp, *listenPort),
		Handler: mux,
	}
	go log.Fatal(s.ListenAndServe())

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGTERM)
	signal.Notify(sigterm, syscall.SIGINT)
	<-sigterm
	controller.Stop()
	s.Shutdown(context.Background())
}
