FROM golang:1.27-alpine AS builder 

WORKDIR /guestbook

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o guestbook-service ./cmd

FROM alpine

COPY --from=builder /guestbook/guestbook-service .

RUN chmod +x guestbook-service