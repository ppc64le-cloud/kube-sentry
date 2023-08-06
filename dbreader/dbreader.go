package dbreader

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"kubertas/knode"

	"k8s.io/klog/v2"

	// SQLite driver.
	_ "rsc.io/sqlite"
)

type DbReader interface {
	ReadServicelogDB()
}

// ServiceLog is a template to retrieve relevant fields from the servicelog.db.
// This can further be extended to retrieve other fields that are available in the servicelog DB.
type ServiceLog struct {
	Id   int    `json:"Id"`
	Desc string `json:"Description"`
	Sev  uint8  `json:"Severity"`
}

// ServicelogParser implements the DBReader interface.
type ServicelogParser struct {
	dbPath                     string
	severity                   int
	currLogCount, prevLogCount int
}

func NewReader(dbFilePath string, eventSeverity int) *ServicelogParser {
	return &ServicelogParser{
		dbPath:   dbFilePath,
		severity: eventSeverity,
	}
}

// ReadServicelogDB accesses the servicelog.db and retrieves the rows, based on
// the severity threshold configured. The retrieved entries are logged and published as
// an event to the kubernetes API server.
func (reader *ServicelogParser) ReadServicelogDB(notifier *knode.K8sNotifier) error {
	var config ServiceLog
	klog.V(1).Infof("Log cursor is at row %v \n", reader.prevLogCount)
	db, err := sql.Open("sqlite3", reader.dbPath)
	if err != nil {
		return fmt.Errorf("error while opening DB. %v", err)
	}
	// Estimate the number of rows that are needed to be processed
	if err = db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM EVENTS WHERE ID > %v",
		reader.currLogCount)).Scan(&reader.currLogCount); err != nil {
		return fmt.Errorf("error while retrieving number of rows %v", err)
	}
	// Check if the number of rows have increased in DB since last read and process pending logs, if any.
	if reader.currLogCount-reader.prevLogCount > 0 {
		klog.V(1).Infof("Processing %d logs", reader.currLogCount-reader.prevLogCount)
		rows, err := db.Query(fmt.Sprintf("SELECT ID,SEVERITY,DESCRIPTION FROM EVENTS WHERE ID > %v AND SEVERITY > %v",
			reader.prevLogCount, reader.severity))
		if err != nil {
			return fmt.Errorf("error occured while reading rows of DB %v", err)
		}
		for rows.Next() {
			err = rows.Scan(&config.Id, &config.Sev, &config.Desc)
			if err != nil {
				return fmt.Errorf("cannot read rows from DB %v", err)
			}
			if logEntry, err := json.MarshalIndent(config, "", " "); err != nil {
				return fmt.Errorf("cannot marshal data %v", err)
			} else {
				klog.Infof("Retrieved RTAS event:\n%v", string(logEntry))
			}
			// Publish as node event by Description to the kubernetes API server.
			notifier.NotifyAPIServer(config.Desc)
		}
		reader.prevLogCount = reader.currLogCount
	} else {
		klog.V(1).Infof("There are no pending servicelogs to be processed.")
	}
	return nil
}
