FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git build-base
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o codesage ./cmd/server


FROM alpine:3.19
WORKDIR /app

# golangci-lint
COPY --from=builder /go/bin/golangci-lint /usr/local/bin/golangci-lint

COPY --from=builder /app/codesage .

EXPOSE 8081
CMD ["./codesage"]
