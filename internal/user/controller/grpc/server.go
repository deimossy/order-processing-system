package grpc

import (
	rpc "google.golang.org/grpc"
	"log/slog"
	"net"
)

type Server struct {
	logger   *slog.Logger
	srv      *rpc.Server
	listener net.Listener
}

func NewServer(logger *slog.Logger, srv *rpc.Server, listener net.Listener) *Server {
	return &Server{
		logger:   logger,
		srv:      srv,
		listener: listener,
	}
}

func (s *Server) Run() error {
	s.logger.Info("starting server",
		slog.String("port", s.listener.Addr().String()),
	)

	if err := s.srv.Serve(s.listener); err != nil {
		s.logger.Error("failed to serve",
			slog.String("error", err.Error()),
		)
		return err
	}

	return nil
}

func (s *Server) Stop() {
	s.logger.Info("shutting down server")

	s.srv.GracefulStop()
}
