FROM golang:1-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go .
RUN go build -o bridge .

FROM alpine
COPY --from=builder --chmod=755 /app/bridge /usr/bin/bridge
CMD ["bridge"]