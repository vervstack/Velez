package main

import (
	"bufio"
	"os"
	"strings"
)

const (
	failMarker = "--- FAIL:"
	tailLines  = 5

	scanBufInitial = 64 * 1024
	scanBufMax     = 1024 * 1024
)

// failureLines is the group's "--- FAIL:" lines, or the log tail when the
// process died without reporting one (build error, panic, timeout, setup).
func failureLines(logPath string) string {
	file, err := os.Open(logPath)
	if err != nil {
		return "(log unreadable: " + err.Error() + ")"
	}

	defer func() { _ = file.Close() }()

	var fails, tail []string

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, scanBufInitial), scanBufMax)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, failMarker) {
			fails = append(fails, strings.TrimSpace(line))
		}

		tail = append(tail, line)
		if len(tail) > tailLines {
			tail = tail[1:]
		}
	}

	if len(fails) > 0 {
		return strings.Join(fails, "\n")
	}

	return strings.Join(tail, "\n")
}
