generate:
	mkdir -p pkg/kv
	protoc --proto_path api/kv \
		--go_out=pkg/kv/ --go_opt=paths=source_relative \
		--go-grpc_out=pkg/kv/ --go-grpc_opt=paths=source_relative \
		api/kv/kv.proto

test:
	go test -v ./...

vet:
	go vet ./...

build:
	docker build --tag=kv_storage .

run:
	docker run -v cache:/app/cache --env-file=.env -p 3000:3000 kv_storage
