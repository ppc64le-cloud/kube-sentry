package utils

import (
	"encoding/json"
	"os"
	"time"
)

// Map SvcLogSeverity maps the severity level with the associated text in ascending
// order.
var SvcLogSeverity = map[uint8]string{
	0: "",
	1: "DEBUG",
	2: "INFO",
	3: "EVENT",
	4: "WARNING",
	5: "ERROR_LOCAL",
	6: "ERROR",
	7: "FATAL",
}

// ServiceLogConf stores the configuration of the servicelog parser.
type ServiceLogConf struct {
	// ServicelogPath contains the path to the servicelog.db file (default: /var/lib/servicelog.db)
	ServicelogPath string `json:"ServicelogDBPath"`
	// PollInterval(Seconds) determines the frequency of checking the DB.
	PollInterval time.Duration `json:"PollInterval"`
	// Severity retrieves the entries whose severity is beyond a set threshold.
	Severity int `json:"Severity"`
}

// Readconfig reads the data present in the config file to setup kube-sentry.
func ReadConfig(configPath string) (*ServiceLogConf, error) {
	var initCfg ServiceLogConf
	file, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(file, &initCfg); err != nil {
		return nil, err
	}
	return &initCfg, nil
}
