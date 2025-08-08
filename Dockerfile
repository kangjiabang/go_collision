FROM golang:1.23.4-alpine AS builder

RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/main ./
COPY --from=builder /app/.env* ./

RUN chmod +x main

ENV ENV=test
EXPOSE 8800
CMD ["./main"]