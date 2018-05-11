package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/jojohappy/luxun/pkg/controller"
	luxunhttp "github.com/jojohappy/luxun/pkg/http"
	"github.com/jojohappy/luxun/pkg/stream"
)

var listenIp = flag.String("listen_ip", "", "IP to listen on, defaults all")
var listenPort = flag.Int("port", 9280, "listen port")
var prometheusEndpoint = flag.String("prometheus_endpoint", "/metrics", "Endpoint to expose Prometheus metrics on")

func main() {
	stream.Init()
	mux := http.NewServeMux()
	luxunhttp.RegisterHandler(mux, *prometheusEndpoint)

	controller.Execute()

	server := fmt.Sprintf("%s:%d", *listenIp, *listenPort)
	http.ListenAndServe(server, mux)
}
