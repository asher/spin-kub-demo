#FROM golang
#
#ADD . /go/src/github.com/lwander/k8s-demo
#
#RUN go install github.com/lwander/k8s-demo
#
#ADD ./content /content
#
#ENTRYPOINT /go/bin/k8s-demo

FROM golang:latest
RUN mkdir /app
ADD . /app/
WORKDIR /app
RUN go get ; go build -o main .
EXPOSE 80/tcp
CMD ["/app/main"]

