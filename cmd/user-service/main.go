package main

import (
	"context"
	"fmt"
	"github.com/deimossy/order-processing-system/internal/user/auth"
	"github.com/deimossy/order-processing-system/internal/user/config"
	"github.com/deimossy/order-processing-system/internal/user/controller/grpc"
	pgrepo "github.com/deimossy/order-processing-system/internal/user/repository/postgres"
	"github.com/deimossy/order-processing-system/internal/user/usecase"
	"github.com/deimossy/order-processing-system/pkg/postgres"
	user_v1 "github.com/deimossy/order-processing-system/protobuf/user/v1"
	rpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"log/slog"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := slog.New(
		slog.NewTextHandler(
			os.Stdout, &slog.HandlerOptions{
				Level:     slog.LevelDebug,
				AddSource: true,
			},
		),
	)

	logger.Info("program started")

	cfg := config.NewConfig(logger)

	rand.Seed(time.Now().UnixNano())

	pgClient := postgres.NewPgClient(ctx, cfg)
	defer func() {
		_ = pgClient.Close()
	}()

	privKey, err := auth.LoadPrivateKey(cfg.RSAPrivateKeyPath)
	if err != nil {
		log.Fatal(err)
	}

	uow := pgrepo.NewPgUnitOfWork(pgClient, cfg)

	userService := usecase.NewUserService(cfg, uow, privKey)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.ServerGRPCPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := rpc.NewServer()
	controller := grpc.NewController(userService)
	user_v1.RegisterUserServiceServer(s, controller)

	reflection.Register(s)
	
	server := grpc.NewServer(logger, s, lis)

	go func() {
		if err = server.Run(); err != nil {
			return
		}
	}()

	//redisClient := redis.NewRedisClient(ctx, cfg)
	//defer func() {
	//	_ = redisClient.Close()
	//}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGTERM, syscall.SIGINT)
	<-shutdown

	logger.Info("shutting down gracefully")

	server.Stop()

	cancel()
}
