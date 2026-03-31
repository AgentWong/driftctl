FROM golang:1.24 AS builder

WORKDIR /go/src/app
COPY go.mod go.sum Makefile ./
RUN go mod download
COPY . .
RUN make build

FROM alpine:3.21.2

WORKDIR /app
COPY --from=builder /go/src/app/bin/driftctl /bin/driftctl
RUN chmod +x /bin/driftctl
ENTRYPOINT ["/bin/driftctl"]
