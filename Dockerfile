FROM golang:1.23-alpine as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -o book cmd/book/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/book .
CMD [ "./book", "-debug" ]
EXPOSE 8080