FROM golang:1.24 as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o app .

FROM scratch

COPY --from=builder /app/main /app/main

ENTRYPOINT ["/app/main"]