package app

import (
	"fmt"
	"log/slog"
	"net"

	"github.com/vshevchenk0/kv-storage/internal/api"
	"github.com/vshevchenk0/kv-storage/internal/config"
	"github.com/vshevchenk0/kv-storage/internal/filestorage"
	"github.com/vshevchenk0/kv-storage/internal/storage"
	"github.com/vshevchenk0/kv-storage/pkg/logger"
	"google.golang.org/grpc"
)

type App struct {
	logger     *slog.Logger
	server     *grpc.Server
	serverAddr string
	storage    storage.Storage
}

func NewApp(config *config.Config) (*App, error) {
	logger := logger.NewLogger(config.Env)

	filestorage, err := filestorage.NewJsonStorage()
	if err != nil {
		return nil, fmt.Errorf("error creating file storage: %v", err)
	}
	storage := storage.NewStorage(logger, filestorage, config.FlushFrequency)

	server := grpc.NewServer()
	api.Register(server, storage)

	return &App{
		logger:     logger,
		server:     server,
		serverAddr: net.JoinHostPort(config.GrpcHost, config.GrpcPort),
		storage:    storage,
	}, nil
}

func (a *App) Run() error {
	listener, err := net.Listen("tcp", a.serverAddr)
	if err != nil {
		return err
	}
	a.logger.Info(fmt.Sprintf("server is running on %s", a.serverAddr))

	err = a.server.Serve(listener)
	if err != nil {
		return err
	}

	return nil
}

func (a *App) Stop(doneChan chan<- struct{}) error {
	a.server.GracefulStop()
	if err := a.storage.Stop(); err != nil {
		return err
	}
	doneChan <- struct{}{}
	return nil
}
