package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/verify-ict-services/bareos_exporter/dataaccess"

	log "github.com/sirupsen/logrus"
)

type bareosMetrics struct {
	TotalFiles       *prometheus.Desc
	TotalBytes       *prometheus.Desc
	LastJobBytes     *prometheus.Desc
	LastJobFiles     *prometheus.Desc
	LastJobErrors    *prometheus.Desc
	LastJobTimestamp *prometheus.Desc
	LastJobStatus    *prometheus.Desc

	LastFullJobBytes     *prometheus.Desc
	LastFullJobFiles     *prometheus.Desc
	LastFullJobErrors    *prometheus.Desc
	LastFullJobTimestamp *prometheus.Desc

	ScheduledJob *prometheus.Desc

	connection *dataaccess.Connection
}

func bareosCollector(conn *dataaccess.Connection) *bareosMetrics {
	return &bareosMetrics{
		TotalFiles: prometheus.NewDesc("bareos_files_saved_total",
			"Total files saved for server during all backups for hostname combined",
			[]string{"hostname"}, nil,
		),
		TotalBytes: prometheus.NewDesc("bareos_bytes_saved_total",
			"Total bytes saved for server during all backups for hostname combined",
			[]string{"hostname"}, nil,
		),
		LastJobBytes: prometheus.NewDesc("bareos_last_backup_job_bytes_saved_total",
			"Total bytes saved during last backup for hostname",
			[]string{"hostname", "level"}, nil,
		),
		LastJobFiles: prometheus.NewDesc("bareos_last_backup_job_files_saved_total",
			"Total files saved during last backup for hostname",
			[]string{"hostname", "level"}, nil,
		),
		LastJobErrors: prometheus.NewDesc("bareos_last_backup_job_errors_occurred_while_saving_total",
			"Total errors occurred during last backup for hostname",
			[]string{"hostname", "level"}, nil,
		),
		LastJobTimestamp: prometheus.NewDesc("bareos_last_backup_job_unix_timestamp",
			"Execution timestamp of last backup for hostname",
			[]string{"hostname", "level"}, nil,
		),
		LastJobStatus: prometheus.NewDesc("bareos_last_backup_job_status",
			"Termination status of the last backup for hostname",
			[]string{"hostname", "status"}, nil,
		),
		LastFullJobBytes: prometheus.NewDesc("bareos_last_full_backup_job_bytes_saved_total",
			"Total bytes saved during last full backup (Level = F) for hostname",
			[]string{"hostname"}, nil,
		),
		LastFullJobFiles: prometheus.NewDesc("bareos_last_full_backup_job_files_saved_total",
			"Total files saved during last full backup (Level = F) for hostname",
			[]string{"hostname"}, nil,
		),
		LastFullJobErrors: prometheus.NewDesc("bareos_last_full_backup_job_errors_occurred_while_saving_total",
			"Total errors occurred during last full backup (Level = F) for hostname",
			[]string{"hostname"}, nil,
		),
		LastFullJobTimestamp: prometheus.NewDesc("bareos_last_full_backup_job_unix_timestamp",
			"Execution timestamp of last full backup (Level = F) for hostname",
			[]string{"hostname"}, nil,
		),
		ScheduledJob: prometheus.NewDesc("bareos_scheduled_jobs_total",
			"Probable execution timestamp of next backup for hostname",
			[]string{"hostname"}, nil,
		),
		connection: conn,
	}
}

func (collector *bareosMetrics) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.TotalFiles
	ch <- collector.TotalBytes
	ch <- collector.LastJobBytes
	ch <- collector.LastJobFiles
	ch <- collector.LastJobErrors
	ch <- collector.LastJobTimestamp
	ch <- collector.LastJobStatus
	ch <- collector.LastFullJobBytes
	ch <- collector.LastFullJobFiles
	ch <- collector.LastFullJobErrors
	ch <- collector.LastFullJobTimestamp
	ch <- collector.ScheduledJob
}

var bareosTerminationStates = []string{
	"C", "R", "B", "T", "E", "e", "f", "D", "A", "I", "L", "W", "l", "q", "F", "S", "m", "M", "s", "j", "c", "d", "t", "p", "i", "a",
}

func (collector *bareosMetrics) Collect(ch chan<- prometheus.Metric) {

	var servers, getServerListErr = collector.connection.GetServerList()

	if getServerListErr != nil {
		log.WithFields(log.Fields{
			"method": "GetServerList",
		}).Error(getServerListErr)
		return
	}

	for _, server := range servers {
		serverFiles, filesErr := collector.connection.TotalFiles(server)
		serverBytes, bytesErr := collector.connection.TotalBytes(server)
		lastServerJob, jobErr := collector.connection.LastJob(server)
		lastFullServerJob, fullJobErr := collector.connection.LastFullJob(server)
		scheduledJob, scheduledJobErr := collector.connection.ScheduledJobs(server)
		lastJobStatus, lastJobStatusErr := collector.connection.LastJobStatus(server)

		if filesErr != nil || bytesErr != nil || jobErr != nil || fullJobErr != nil || scheduledJobErr != nil || lastJobStatusErr != nil {
			log.Info(server)
		}

		if filesErr != nil {
			log.WithFields(log.Fields{
				"method": "TotalFiles",
			}).Error(filesErr)
		}

		if bytesErr != nil {
			log.WithFields(log.Fields{
				"method": "TotalBytes",
			}).Error(bytesErr)
		}

		if jobErr != nil {
			log.WithFields(log.Fields{
				"method": "LastJob",
			}).Error(jobErr)
		}

		if fullJobErr != nil {
			log.WithFields(log.Fields{
				"method": "LastFullJob",
			}).Error(fullJobErr)
		}

		if scheduledJobErr != nil {
			log.WithFields(log.Fields{
				"method": "ScheduledJobs",
			}).Error(scheduledJobErr)
		}

		if lastJobStatus != nil {

			log.WithFields(log.Fields{
				"method": "LastJobStatus",
			}).Error(lastJobStatusErr)
		}

		ch <- prometheus.MustNewConstMetric(collector.TotalFiles, prometheus.CounterValue, float64(serverFiles.Files), server)
		ch <- prometheus.MustNewConstMetric(collector.TotalBytes, prometheus.CounterValue, float64(serverBytes.Bytes), server)

		ch <- prometheus.MustNewConstMetric(collector.LastJobBytes, prometheus.CounterValue, float64(lastServerJob.JobBytes), server, lastServerJob.Level)
		ch <- prometheus.MustNewConstMetric(collector.LastJobFiles, prometheus.CounterValue, float64(lastServerJob.JobFiles), server, lastServerJob.Level)
		ch <- prometheus.MustNewConstMetric(collector.LastJobErrors, prometheus.CounterValue, float64(lastServerJob.JobErrors), server, lastServerJob.Level)
		ch <- prometheus.MustNewConstMetric(collector.LastJobTimestamp, prometheus.CounterValue, float64(lastServerJob.JobDate.Unix()), server, lastServerJob.Level)
		for _, terminationState := range bareosTerminationStates {
			var state = float64(0)
			if lastJobStatus != nil {
				if terminationState == *lastJobStatus {
					state = 1
				}
			}

			ch <- prometheus.MustNewConstMetric(collector.LastJobStatus, prometheus.CounterValue, state, server, terminationState)

		}

		ch <- prometheus.MustNewConstMetric(collector.LastFullJobBytes, prometheus.CounterValue, float64(lastFullServerJob.JobBytes), server)
		ch <- prometheus.MustNewConstMetric(collector.LastFullJobFiles, prometheus.CounterValue, float64(lastFullServerJob.JobFiles), server)
		ch <- prometheus.MustNewConstMetric(collector.LastFullJobErrors, prometheus.CounterValue, float64(lastFullServerJob.JobErrors), server)
		ch <- prometheus.MustNewConstMetric(collector.LastFullJobTimestamp, prometheus.CounterValue, float64(lastFullServerJob.JobDate.Unix()), server)

		ch <- prometheus.MustNewConstMetric(collector.ScheduledJob, prometheus.CounterValue, float64(scheduledJob.ScheduledJobs), server)

	}
}
