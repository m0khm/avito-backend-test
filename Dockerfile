# Step 1: Builder
FROM golang:1.24-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git ca-certificates tzdata
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/conference-mock ./cmd/conference-mock

# Step 2: Final
FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata curl
COPY --from=builder /out/api /usr/local/bin/api
COPY --from=builder /out/conference-mock /usr/local/bin/conference-mock
COPY db ./db
COPY docs ./docs
EXPOSE 8080 8090
CMD ["/usr/local/bin/api"]
