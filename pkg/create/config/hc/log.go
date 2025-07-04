package hc

import (
	"github.com/aptd3v/go-contain/pkg/create"
	"github.com/docker/docker/api/types/container"
)

// WithLogDriver sets a custom logging driver and options
func WithLogDriver(driver string, options map[string]string) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		opt.LogConfig = container.LogConfig{
			Type:   driver,
			Config: options,
		}
		return nil
	}
}

// json-file logger (default)
// parameters:
//   - maxSize: max size before rotation (e.g. "10m")
//   - maxFile: max number of log files to keep (e.g. "3")
func WithJSONFileLogger(maxSize, maxFile string) create.SetHostConfig {
	return WithLogDriver("json-file", map[string]string{
		"max-size": maxSize,
		"max-file": maxFile,
	})
}

// syslog logger
// parameters:
//   - address: syslog server address (e.g. "tcp://1.2.3.4:514")
//   - tag: syslog tag string
func WithSyslogLogger(address, tag string) create.SetHostConfig {
	return WithLogDriver("syslog", map[string]string{
		"syslog-address": address,
		"tag":            tag,
	})
}

// journald logger
// no options typically needed
func WithJournaldLogger() create.SetHostConfig {
	return WithLogDriver("journald", nil)
}

// gelf logger
// parameters:
//   - address: gelf server address (e.g. "udp://1.2.3.4:12201")
//   - maxSize: max message size in bytes (optional, "" to omit)
func WithGelfLogger(address, maxSize string) create.SetHostConfig {
	opts := map[string]string{
		"gelf-address": address,
	}
	if maxSize != "" {
		opts["max-message-size"] = maxSize
	}
	return WithLogDriver("gelf", opts)
}

// fluentd logger
// parameters:
//   - address: fluentd server address (e.g. "localhost:24224")
func WithFluentdLogger(address string) create.SetHostConfig {
	return WithLogDriver("fluentd", map[string]string{
		"fluentd-address": address,
	})
}

// awslogs logger
// parameters:
//   - region: AWS region (e.g. "us-east-1")
//   - group: CloudWatch log group name
//   - stream: CloudWatch log stream name
func WithAWSLogsLogger(region, group, stream string) create.SetHostConfig {
	return WithLogDriver("awslogs", map[string]string{
		"awslogs-region":       region,
		"awslogs-group":        group,
		"awslogs-stream":       stream,
		"awslogs-create-group": "true", // optionally auto-create group
	})
}

// splunk logger
// parameters:
//   - url: splunk HEC URL (e.g. "https://splunk.example.com:8088")
//   - token: splunk HEC token
//   - source: source type (optional)
//   - sourcetype: sourcetype (optional)
func WithSplunkLogger(url, token, source, sourcetype string) create.SetHostConfig {
	opts := map[string]string{
		"splunk-url":   url,
		"splunk-token": token,
	}
	if source != "" {
		opts["splunk-source"] = source
	}
	if sourcetype != "" {
		opts["splunk-sourcetype"] = sourcetype
	}
	return WithLogDriver("splunk", opts)
}

// none logger disables logging
func WithNoneLogger() create.SetHostConfig {
	return WithLogDriver("none", nil)
}
