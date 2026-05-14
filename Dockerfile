#stage:1 Build
FROM            docker.io/library/golang:1.26 AS builder
WORKDIR         /app
COPY            ./ /app/
RUN             CGO_ENABLED=0 go build -o auth-service ./cmd/server

#stage:2 RUN

FROM            docker.io/redhat/ubi9
COPY            --from=builder /app/auth-service .
ENTRYPOINT      ["./auth-service"]