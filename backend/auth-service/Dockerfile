FROM golang:1.27-alpine AS builder 

WORKDIR /auth

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o auth-service ./cmd

FROM alpine 

WORKDIR /auth

COPY --from=builder /auth/auth-service .

RUN chmod +x auth-service