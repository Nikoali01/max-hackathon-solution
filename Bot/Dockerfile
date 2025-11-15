FROM golang:1.23.8 AS builder

WORKDIR /build

COPY go.mod go.sum ./
COPY max-bot-api-client-go ./max-bot-api-client-go
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o go-binary

FROM alpine:3.14

WORKDIR /app

COPY --from=builder /build/go-binary .

CMD [ "./go-binary"]
