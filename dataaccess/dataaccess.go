package dataaccess

import (
	"database/sql"
	"fmt"
	"github.com/vierbergenlars/bareos_exporter/types"
	"time"
)

// Connection to database, and database specific queries
type Connection struct {
	db      *sql.DB
	queries *sqlQueries
}

type sqlQueries struct {
	ServerList    string
	TotalBytes    string
	TotalFiles    string
	LastJob       string
	LastFullJob   string
	ScheduledJobs string
}

var mysqlQueries *sqlQueries = &sqlQueries{
	ServerList:    "SELECT DISTINCT Name FROM Job WHERE SchedTime >= ?",
	TotalBytes:    "SELECT SUM(JobBytes) FROM Job WHERE Name=? AND PurgedFiles=0AND JobStatus = 'T'",
	TotalFiles:    "SELECT SUM(JobFiles) FROM Job WHERE Name=? AND PurgedFiles=0 AND JobStatus = 'T'",
	LastJob:       "SELECT Level,JobBytes,JobFiles,JobErrors,StartTime FROM Job WHERE Name = ? AND JobStatus = 'T' ORDER BY StartTime DESC LIMIT 1",
	LastFullJob:   "SELECT Level,JobBytes,JobFiles,JobErrors,StartTime FROM Job WHERE Name = ? AND Level = 'F' AND JobStatus = 'T' ORDER BY StartTime DESC LIMIT 1",
	ScheduledJobs: "SELECT COUNT(SchedTime) AS JobsScheduled FROM Job WHERE Name = ? AND SchedTime >= ?",
}

var postgresQueries *sqlQueries = &sqlQueries{
	ServerList:    "SELECT DISTINCT Name FROM job WHERE SchedTime >= ?",
	TotalBytes:    "SELECT SUM(JobBytes) FROM job WHERE Name=? AND PurgedFiles=0AND JobStatus = 'T'",
	TotalFiles:    "SELECT SUM(JobFiles) FROM job WHERE Name=? AND PurgedFiles=0 AND JobStatus = 'T'",
	LastJob:       "SELECT Level,JobBytes,JobFiles,JobErrors,StartTime FROM job WHERE Name = ? AND JobStatus = 'T' ORDER BY StartTime DESC LIMIT 1",
	LastFullJob:   "SELECT Level,JobBytes,JobFiles,JobErrors,StartTime FROM job WHERE Name = ? AND Level = 'F' AND JobStatus = 'T' ORDER BY StartTime DESC LIMIT 1",
	ScheduledJobs: "SELECT COUNT(SchedTime) AS JobsScheduled FROM job WHERE Name = ? AND SchedTime >= ?",
}

// GetConnection opens a new db connection
func GetConnection(databaseType string, connectionString string) (*Connection, error) {
	var queries *sqlQueries
	switch databaseType {
	case "mysql":
		queries = mysqlQueries
	case "postgres":
		queries = postgresQueries
	default:
		return nil, fmt.Errorf("Unknown database type %s", databaseType)
	}

	db, err := sql.Open(databaseType, connectionString)

	if err != nil {
		return nil, err
	}

	return &Connection{
		db:      db,
		queries: queries,
	}, nil
}

// GetServerList reads all servers with scheduled backups for current date
func (connection Connection) GetServerList() ([]string, error) {
	date := fmt.Sprintf("%s%%", time.Now().AddDate(0, 0, -7).Format("2006-01-02"))
	results, err := connection.db.Query(connection.queries.ServerList, date)

	if err != nil {
		return nil, err
	}

	var servers []string

	for results.Next() {
		var server string
		err = results.Scan(&server)
		servers = append(servers, server)
	}

	return servers, err
}

// TotalBytes returns total bytes saved for a server since the very first backup
func (connection Connection) TotalBytes(server string) (*types.TotalBytes, error) {
	results, err := connection.db.Query(connection.queries.TotalBytes, server)

	if err != nil {
		return nil, err
	}

	var totalBytes types.TotalBytes
	if results.Next() {
		err = results.Scan(&totalBytes.Bytes)
		results.Close()
	}

	return &totalBytes, err
}

// TotalFiles returns total files saved for a server since the very first backup
func (connection Connection) TotalFiles(server string) (*types.TotalFiles, error) {
	results, err := connection.db.Query(connection.queries.TotalFiles, server)

	if err != nil {
		return nil, err
	}

	var totalFiles types.TotalFiles
	if results.Next() {
		err = results.Scan(&totalFiles.Files)
		results.Close()
	}

	return &totalFiles, err
}

// LastJob returns metrics for latest executed server backup
func (connection Connection) LastJob(server string) (*types.LastJob, error) {
	results, err := connection.db.Query(connection.queries.LastJob, server)

	if err != nil {
		return nil, err
	}

	var lastJob types.LastJob
	if results.Next() {
		err = results.Scan(&lastJob.Level, &lastJob.JobBytes, &lastJob.JobFiles, &lastJob.JobErrors, &lastJob.JobDate)
		results.Close()
	}

	return &lastJob, err
}

// LastFullJob returns metrics for latest executed server backup with Level F
func (connection Connection) LastFullJob(server string) (*types.LastJob, error) {
	results, err := connection.db.Query(connection.queries.LastFullJob, server)

	if err != nil {
		return nil, err
	}

	var lastJob types.LastJob
	if results.Next() {
		err = results.Scan(&lastJob.Level, &lastJob.JobBytes, &lastJob.JobFiles, &lastJob.JobErrors, &lastJob.JobDate)
		results.Close()
	}

	return &lastJob, err
}

// ScheduledJobs returns amount of scheduled jobs
func (connection Connection) ScheduledJobs(server string) (*types.ScheduledJob, error) {
	date := fmt.Sprintf("%s%%", time.Now().Format("2006-01-02"))
	results, err := connection.db.Query(connection.queries.ScheduledJobs, server, date)

	if err != nil {
		return nil, err
	}

	var schedJob types.ScheduledJob
	if results.Next() {
		err = results.Scan(&schedJob.ScheduledJobs)
		results.Close()
	}

	return &schedJob, err
}

// Close the database connection
func (connection Connection) Close() error {
	return connection.db.Close()
}
