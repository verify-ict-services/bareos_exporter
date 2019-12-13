FROM golang as builder
COPY . /go/src/github.com/vierbergenlars/bareos_exporter
WORKDIR /go/src/github.com/vierbergenlars/bareos_exporter
RUN go get -v .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bareos_exporter .

FROM busybox:latest

ENV endpoint /metrics
ENV port 9625
ENV dsn mysql://bareos@unix()/bareos

WORKDIR /bareos_exporter
COPY --from=builder /go/src/github.com/vierbergenlars/bareos_exporter/bareos_exporter bareos_exporter

ENTRYPOINT ["./bareos_exporter"]
CMD ["-port", "$port", "-endpoint", "$endpoint", "-dsn", "$dsn"]
EXPOSE $port
