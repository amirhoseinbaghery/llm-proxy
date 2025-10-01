# Build stage
FROM golang:1.24 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/llm-proxy ./cmd/llm-proxy

# Run stage
FROM alpine:3.20
RUN adduser -D -u 10001 appuser && apk add --no-cache ca-certificates
COPY --from=build /out/llm-proxy /llm-proxy
USER appuser
EXPOSE 8080
ENTRYPOINT ["/llm-proxy"]
