FROM golang:1.27-alpine AS builder 

WORKDIR /notebook

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o notebook-service ./cmd

FROM alpine

COPY --from=builder /notebook/notebook-service .

RUN chmod +x notebook-service