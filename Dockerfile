FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /gcs-bucket-create main.go

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=builder /gcs-bucket-create /gcs-bucket-create
ENTRYPOINT ["/gcs-bucket-create"]