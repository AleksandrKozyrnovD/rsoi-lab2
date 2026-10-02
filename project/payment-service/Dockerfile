FROM golang:1.27.0-alpine AS builder

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY ./ ./
RUN go mod tidy && go build -o paymentservice .

FROM alpine:3.20

WORKDIR /app

COPY --from=builder --chmod=755 /app/paymentservice /paymentservice
COPY --from=builder /app/config.yml /app/config.yml


EXPOSE 8060

ENTRYPOINT ["/paymentservice"]
