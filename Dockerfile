FROM golang:1.22 AS builder
WORKDIR /go/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /go/bin/oracle ./cmd/oracle

FROM scratch
COPY --chown=65534:65534 --from=builder /go/bin/oracle .
USER 65534

ENTRYPOINT [ "./oracle" ]