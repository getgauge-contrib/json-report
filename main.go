package main

import (
	"net"
	"os"

	"github.com/getgauge-contrib/json-report/gauge_messages"
	"github.com/getgauge-contrib/json-report/logger"
	"google.golang.org/grpc"
)

const tenGB = 1024 * 1024 * 1024 * 10

func main() {
	findProjectRoot()
	action := os.Getenv(pluginActionEnv)
	if action == setupAction {
		addDefaultPropertiesToProject()
	} else if action == executionAction {
		os.Chdir(projectRoot)
		address, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
		if err != nil {
			logger.Fatal("failed to start server")
		}
		l, err := net.ListenTCP("tcp", address)
		if err != nil {
			logger.Fatal("failed to start server")
		}
		server := grpc.NewServer(grpc.MaxRecvMsgSize(tenGB))
		h := &handler{server: server}
		gauge_messages.RegisterReporterServer(server, h)
		logger.Info("Listening on port:%d", l.Addr().(*net.TCPAddr).Port)
		server.Serve(l)
	}
}
