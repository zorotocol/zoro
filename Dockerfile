FROM golang:1.22
WORKDIR /app
COPY . .
ENV GOPROXY=https://goproxy.io,direct
RUN go build ./cmd/oracle
RUN mv ./oracle /oracle
CMD ["oracle"]