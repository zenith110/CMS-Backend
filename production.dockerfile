FROM golang:1.17.5-alpine3.14 as builder


WORKDIR /home/backend/v2
COPY . /home/backend/v2

RUN go mod download
CMD  ["go", "build", "server.go"]

# Copy the binary and run the binary
FROM alpine:latest
COPY --from=builder /home/backend/v2/ .
CMD ["./server"]