package dbreader

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"k8s.io/klog/v2"

	// SQLite driver.
	_ "rsc.io/sqlite"

	"github.com/ppc64le-cloud/kube-rtas/pkg/knode"
)

type DbReader interface {
	ParseServiceLogDB()
}

// ServiceLog is a template to retrieve relevant fields from the servicelog.db.
// This can further be extended to retrieve other fields that are available in the servicelog DB.
type ServiceLog struct {
	ID          int    `json:"Id"`
	Description string `json:"Description"`
	Severity    uint8  `json:"Severity"`
}

// ServicelogParser implements the DBReader interface.
type ServiceLogParser struct {
	db                         *sql.DB
	dbPath                     string
	thresholdSeverity          int
	currLogCount, prevLogCount int
}

// NewReader creates an instance of ServerLogParser.
func NewReader(dbFilePath string, eventSeverity int) *ServiceLogParser {
	return &ServiceLogParser{
		dbPath:            dbFilePath,
		thresholdSeverity: eventSeverity,
	}
}

// ParseServiceLogDB accesses the servicelog.db and retrieves the rows based on
// the severity threshold configured. The retrieved entries are logged and published as
// an event to the Kubernetes API server.
func (parser *ServiceLogParser) ParseServiceLogDB(notifier *knode.Notifier) error {
	if err := parser.openDatabase(); err != nil {
		return err
	}
	defer parser.closeDatabase()

	if err := parser.estimateRowCount(); err != nil {
		return err
	}
	klog.V(1).Infof("Processed %v logs so far", parser.prevLogCount)

	if parser.currLogCount == parser.prevLogCount {
		klog.V(1).Info("There are no pending service logs to be processed.")
		return nil
	}

	klog.Infof("Processing %d logs", parser.currLogCount-parser.prevLogCount)
	return parser.processLogs(notifier)
}

// openDatabase opens the servicelog.db file from the configured path.
func (parser *ServiceLogParser) openDatabase() error {
	db, err := sql.Open("sqlite3", parser.dbPath)
	if err != nil {
		return fmt.Errorf("error while opening DB: %v", err)
	}
	parser.db = db
	return nil
}

func (parser *ServiceLogParser) closeDatabase() {
	if parser.db != nil {
		parser.db.Close()
	}
}

// estimateRowCount calculates the number of rows present in the servicelog.db file.
// and updates the current number of lines read, serving as a pointer to the latest logs.
func (parser *ServiceLogParser) estimateRowCount() error {
	err := parser.db.QueryRow("SELECT COUNT(*) FROM EVENTS").Scan(&parser.currLogCount)
	if err != nil {
		return fmt.Errorf("error while retrieving number of rows: %v", err)
	}
	return nil
}

// processLogs retrieves the servicelogs, if present, based on the configured severity.
// The service logs are then notified as events in Kubernetes.
func (parser *ServiceLogParser) processLogs(notifier *knode.Notifier) error {
	rows, err := parser.db.Query(fmt.Sprintf("SELECT ID, SEVERITY, DESCRIPTION FROM EVENTS WHERE ID > %v AND SEVERITY > %v", parser.prevLogCount, parser.thresholdSeverity))
	if err != nil {
		return fmt.Errorf("error occurred while reading rows from DB: %v", err)
	}
	defer rows.Close()

	var config ServiceLog
	for rows.Next() {
		if err := rows.Scan(&config.ID, &config.Severity, &config.Description); err != nil {
			return fmt.Errorf("cannot read rows from DB: %v", err)
		}
		if logEntry, err := json.MarshalIndent(config, "", " "); err != nil {
			return fmt.Errorf("cannot marshal data: %v", err)
		} else {
			klog.Infof("Retrieved RTAS event:\n%v", string(logEntry))
			// Publish as a node event by Description to the Kubernetes API server.
			notifier.NotifyAPIServer(config.Description, config.Severity)
		}
	}
	parser.prevLogCount = parser.currLogCount
	return nil
}
