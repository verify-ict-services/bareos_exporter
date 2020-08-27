## bareos_exporter
[![Go Report Card](https://goreportcard.com/badge/github.com/vierbergenlars/bareos_exporter)](https://goreportcard.com/report/github.com/vierbergenlars/bareos_exporter)

[Prometheus](https://github.com/prometheus) exporter for [bareos](https://github.com/bareos) data recovery system with PostgreSQL as database

### [`Dockerfile`](./Dockerfile)

### Usage with [docker](https://hub.docker.com/r/vierbergenlnars/bareos_exporter)
1. Replace 4 variables in the file `main.go`:
- `host     = "___POSTGRESQL_HOST___"`
- `user     = "___POSTGRESQL_READ_ONLY_USER___"`
- `password = "___POSTGRESQL_PASSWORD___"`
- `dbname   = "___POSTGRESQL_DB___"`
2. Build the image as follows:
```bash
docker image build -t verify-ict-services/bareos_exporter:latest .
```
3. Run docker image as follows
```bash
docker run --name bareos_exporter -p 9625:9625 -d verify-ict-services/bareos_exporter:latest
```
### Metrics

- Total amout of bytes and files saved
- Latest executed job metrics (level, errors, execution time, bytes and files saved)
- Latest full job (level = F) metrics
- Amount of scheduled jobs

