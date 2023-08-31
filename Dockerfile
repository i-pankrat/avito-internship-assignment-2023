FROM golang:latest AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o segment-management cmd/segment-management/main.go

FROM ubuntu
ENV CONFIG_PATH /local.yaml
COPY --from=builder /app/segment-management segment-management
COPY --from=builder /app/config/local.yaml local.yaml
EXPOSE 8080
CMD ["./segment-management"]
