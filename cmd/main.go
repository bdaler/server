package main

import (
	"github.com/bdaler/server/pkg/server"
	"net"
	"os"
)

func main() {
	if err := execute(server.HOST, server.PORT); err != nil {
		os.Exit(1)
	}
}

func execute(host, port string) (err error) {
	srv := server.NewServer(net.JoinHostPort(host, port))
	srv.Register("/", srv.RouteHandler("Welcome to our web-site"))
	srv.Register("/about", srv.RouteHandler("About Golang Academy"))
	srv.Register("/payments", srv.RouteHandler("boolshit"))
	return srv.Start()
}
