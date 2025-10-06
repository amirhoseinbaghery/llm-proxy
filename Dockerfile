# Build stage
FROM golang:1.24 AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN apt-get update && apt-get install -y gcc libsqlite3-dev

RUN CGO_ENABLED=1 GOOS=linux go build -o /out/llm-proxy ./cmd/llm-proxy

# Run stage
FROM debian:bookworm-slim
RUN adduser --disabled-password --gecos "" appuser
RUN mkdir /data && chown appuser /data

COPY --from=build /out/llm-proxy /llm-proxy
USER appuser
EXPOSE 8080
ENTRYPOINT ["/llm-proxy"]
