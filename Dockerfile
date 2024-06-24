FROM golang:1.22-alpine as builder
WORKDIR /build
COPY . .
RUN go mod download
RUN go build -o ./bin/app ./cmd/app/main.go

FROM golang:1.22-alpine as app
WORKDIR /app
COPY --from=builder /build/bin/app .
RUN chmod +x ./app
EXPOSE 3000
CMD ["./app"]
