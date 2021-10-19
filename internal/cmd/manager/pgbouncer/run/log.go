/*
This file is part of Cloud Native PostgreSQL.

Copyright (C) 2019-2021 EnterpriseDB Corporation.
*/

package run

import (
	"bufio"
	"bytes"
	"regexp"
	"strings"

	"github.com/EnterpriseDB/cloud-native-postgresql/pkg/management/log"
)

var pgBouncerLogRegex = regexp.
	MustCompile(`(?s)(?P<Timestamp>^.*) \[(?P<Pid>[0-9]+)\] (?P<Level>[A-Z]+) (?P<Msg>.+)$`)

type pgBouncerLogWriter struct {
	Logger log.Logger
}

func (p *pgBouncerLogWriter) Write(in []byte) (n int, err error) {
	// pgbouncer can write multi-line logs, and each continuation line starts
	// with "\t". This code will parse that and work on each log line separately

	currentLogLine := ""

	sc := bufio.NewScanner(bytes.NewReader(in))
	for sc.Scan() {
		logLine := sc.Text()
		if strings.HasPrefix(logLine, "\t") {
			currentLogLine += logLine
			continue
		}

		if currentLogLine != "" {
			p.writePgbouncerLogLine(currentLogLine)
		}

		currentLogLine = logLine
	}

	if currentLogLine != "" {
		p.writePgbouncerLogLine(currentLogLine)
	}

	return len(in), nil
}

func (p *pgBouncerLogWriter) writePgbouncerLogLine(line string) {
	matches := pgBouncerLogRegex.FindStringSubmatch(line)
	switch {
	case matches == nil:
		p.Logger.WithValues("matched", false).Info(line)
	case len(matches) != 5:
		p.Logger.WithValues("matched", false, "matches", matches).Info(line)
	default:
		p.Logger.Info("record", "record",
			pgBouncerLogRecord{
				Timestamp: matches[1],
				Pid:       matches[2],
				Level:     matches[3],
				Msg:       matches[4],
			})
	}
}

type pgBouncerLogRecord struct {
	Timestamp string `json:"timestamp"`
	Pid       string `json:"pid"`
	Level     string `json:"level"`
	Msg       string `json:"msg"`
}