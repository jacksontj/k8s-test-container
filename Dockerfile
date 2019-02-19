FROM       golang:alpine as builder

COPY . /go/src/github.com/jacksontj/k8s-test-container
RUN cd /go/src/github.com/jacksontj/k8s-test-container/cmd/k8s-test-container && CGO_ENABLED=0 go build

FROM golang:alpine
EXPOSE 8080

COPY --from=builder /go/src/github.com/jacksontj/k8s-test-container/cmd/k8s-test-container/k8s-test-container /bin/k8s-test-container

ENTRYPOINT [ "/bin/k8s-test-container" ]
