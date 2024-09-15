FROM golang:1.22-alpine AS builder
WORKDIR /cluster3
COPY go.mod go.mod
COPY go.sum go.sum
COPY main.go main.go
RUN go mod download
RUN go build main.go

FROM alpine as helloworld-server
WORKDIR /
COPY --from=builder /cluster3/main .
USER 65532:65532
ENTRYPOINT ["/main"]