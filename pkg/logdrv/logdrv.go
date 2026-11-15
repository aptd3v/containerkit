// Package logdrv is the Engine log-driver vocabulary.
package logdrv

// Driver is a built-in logging driver name.
type Driver string

const (
	JSONFile Driver = "json-file"
	Local    Driver = "local"
	Syslog   Driver = "syslog"
	Journald Driver = "journald"
	GELF     Driver = "gelf"
	Fluentd  Driver = "fluentd"
	AWSLogs  Driver = "awslogs"
	Splunk   Driver = "splunk"
	ETWLogs  Driver = "etwlogs"
	GCPLogs  Driver = "gcplogs"
	None     Driver = "none"
)

// All returns every known built-in log driver.
func All() []Driver {
	return []Driver{JSONFile, Local, Syslog, Journald, GELF, Fluentd, AWSLogs, Splunk, ETWLogs, GCPLogs, None}
}
