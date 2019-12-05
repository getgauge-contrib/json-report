package main

import (
	"context"
	"os"

	"github.com/getgauge-contrib/json-report/gauge_messages"
	"google.golang.org/grpc"
)

type handler struct {
	server *grpc.Server
}

func (h *handler) NotifySuiteResult(c context.Context, m *gauge_messages.SuiteExecutionResult) (*gauge_messages.Empty, error) {
	createReport(m)
	return &gauge_messages.Empty{}, nil
}
func (h *handler) Kill(c context.Context, m *gauge_messages.KillProcessRequest) (*gauge_messages.Empty, error) {
	defer h.stopServer()
	return &gauge_messages.Empty{}, nil
}

func (h *handler) stopServer() {
	h.server.Stop()
	os.Exit(0)
}
