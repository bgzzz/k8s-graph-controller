FROM golang:1.12.4-alpine as builder
RUN apk add --update make
WORKDIR /go/src/github.com/pavelgonchukov/k8s-graph-controller
COPY . .

RUN make

FROM alpine as service

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
COPY --from=builder /go/src/github.com/pavelgonchukov/k8s-graph-controller/bin/k8s-graph-controller /k8s-graph-controller
ENTRYPOINT ["/k8s-graph-controller"]
