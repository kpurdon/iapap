FROM golang:1.12 as builder
ENV GO111MODULE=on

# go.mod dependencies layer
COPY ./go.mod /src/go.mod
COPY ./go.sum /src/go.sum
WORKDIR /src
RUN go mod download

# go build layer
COPY . /src
WORKDIR /src
RUN CGO_ENABLED=0 GOOS=linux go build -o /src/iapap .

FROM alpine:latest
COPY --from=builder /src/iapap .
CMD ./iapap
