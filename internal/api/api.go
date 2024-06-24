package api

import (
	"context"
	"errors"

	"github.com/vshevchenk0/kv-storage/internal/storage"
	"github.com/vshevchenk0/kv-storage/pkg/kv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serverApi struct {
	kv.UnimplementedKVServiceServer
	storage storage.Storage
}

func Register(server *grpc.Server, storage storage.Storage) {
	kv.RegisterKVServiceServer(server, &serverApi{storage: storage})
}

func (s *serverApi) Set(ctx context.Context, in *kv.SetRequest) (*kv.EmptyResponse, error) {
	if in.Key == "" {
		return nil, status.Error(codes.InvalidArgument, "key should not be empty")
	}
	if in.Value == "" {
		return nil, status.Error(codes.InvalidArgument, "value should not be empty")
	}
	if err := s.storage.Set(in.GetKey(), in.GetValue(), in.GetTtl(), in.GetSyncCommit()); err != nil {
		return nil, status.Error(codes.Aborted, "failed to add record")
	}
	return &kv.EmptyResponse{}, nil
}

func (s *serverApi) Get(ctx context.Context, in *kv.GetRequest) (*kv.GetResponse, error) {
	if in.Key == "" {
		return nil, status.Error(codes.InvalidArgument, "key should not be empty")
	}
	value, err := s.storage.Get(in.GetKey())
	if errors.Is(err, storage.ErrRecordDoesNotExist) {
		return nil, status.Error(codes.NotFound, "record not found")
	}
	if errors.Is(err, storage.ErrRecordExpired) {
		return nil, status.Error(codes.NotFound, "record expired")
	}
	return &kv.GetResponse{Value: value}, nil
}

func (s *serverApi) Delete(ctx context.Context, in *kv.DeleteRequest) (*kv.EmptyResponse, error) {
	if in.Key == "" {
		return nil, status.Error(codes.InvalidArgument, "key should not be empty")
	}
	if err := s.storage.Delete(in.GetKey(), in.GetSyncCommit()); err != nil {
		return nil, status.Error(codes.Aborted, "failed to delete record")
	}
	return &kv.EmptyResponse{}, nil
}
