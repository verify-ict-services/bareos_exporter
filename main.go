package main

import (
	"github.com/verify-ict-services/bareos_exporter/dataaccess"

	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

const (
  host     = "___POSTGRESQL_HOST___"
  port     = 5432
  user     = "___POSTGRESQL_READ_ONLY_USER___"
  password = "___POSTGRESQL_PASSWORD___"
  dbname   = "___POSTGRESQL_DB___"
)

var connectionString string

var (
	exporterPort     = flag.Int("port", 9625, "Bareos exporter port")
	exporterEndpoint = flag.String("endpoint", "/metrics", "Bareos exporter endpoint")
)

func init() {
	flag.Usage = func() {
		fmt.Println("Usage: bareos_exporter [ ... ]\n\nParameters:")
		fmt.Println()
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
	  "password=%s dbname=%s sslmode=disable",
	  host, port, user, password, dbname)

	connection, err := dataaccess.GetConnection("postgres", psqlInfo)
	if err != nil {
		panic(err.Error())
	}
	defer connection.Close()

	collector := bareosCollector(connection)
	prometheus.MustRegister(collector)

	http.Handle(*exporterEndpoint, promhttp.Handler())
	log.Info("Beginning to serve on port ", *exporterPort)

	addr := fmt.Sprintf(":%d", *exporterPort)
	log.Fatal(http.ListenAndServe(addr, nil))
}
