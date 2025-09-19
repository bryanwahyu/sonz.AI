# syntax=docker/dockerfile:1

FROM golang:1.21 AS builder
WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o sandai-api ./cmd/api

FROM gcr.io/distroless/base-debian12
COPY --from=builder /workspace/sandai-api /bin/sandai-api
ENV SANDAI_HTTP_ADDR=:8080
EXPOSE 8080
ENTRYPOINT ["/bin/sandai-api"]
