FROM golang:1.19 AS builder

WORKDIR /app
COPY go.* ./

RUN go mod tidy

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -a -o out .

FROM alpine:latest
COPY --from=builder /app/out ./

CMD ["./out"]
