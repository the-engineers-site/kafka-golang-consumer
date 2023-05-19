FROM golang:1.20-alpine AS builder
RUN apk add alpine-sdk
WORKDIR /go/app
COPY . .
RUN go mod download
RUN GOOS=linux GOARCH=amd64 go build -o rest-api -tags musl main.go

FROM alpine:latest as runner
WORKDIR /root/
COPY --from=builder /go/app/rest-api .
ENTRYPOINT /root/rest-api