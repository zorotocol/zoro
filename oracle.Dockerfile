FROM golang:1.22.1 AS build
WORKDIR /oracle
COPY . .
RUN GOPROXY=https://goproxy.io,direct go mod download
RUN CGO_ENABLED=0 go build ./cmd/oracle
FROM alpine
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /
COPY --from=build /oracle/oracle /
ENTRYPOINT ["/oracle"]