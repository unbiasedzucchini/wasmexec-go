FROM golang:1.24-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags='-s -w' -o /wasmexec-go .

FROM scratch
COPY --from=builder /wasmexec-go /wasmexec-go
EXPOSE 8000
ENTRYPOINT ["/wasmexec-go"]
