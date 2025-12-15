FROM golang:1.24 AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w -extldflags=-static" -o main .

# Run stage
FROM alpine:3.19
RUN apk add --update tzdata netcat-openbsd && \
    cp /usr/share/zoneinfo/Asia/Ho_Chi_Minh /etc/localtime && \
    echo "Asia/Ho_Chi_Minh" > /etc/timezone && \
    apk del tzdata && \
    rm -rf /var/cache/apk/*
WORKDIR /app
COPY --from=builder /app/main .
COPY config.yml .
COPY db/migration ./db/migration
COPY start.sh .
COPY wait-for.sh .

# Set execute permissions for scripts
RUN chmod +x start.sh wait-for.sh

EXPOSE 8080 53 53/udp

ENTRYPOINT ["./start.sh"]