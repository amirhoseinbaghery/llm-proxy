# Build stage
FROM golang:1.24 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/llm-proxy ./cmd/llm-proxy

# Run stage (distroless)
FROM gcr.io/distroless/base-debian12
COPY --from=build /out/llm-proxy /llm-proxy
EXPOSE 8080
ENTRYPOINT ["/llm-proxy"]
