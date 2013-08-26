package main

import (
	"testing"
	"regexp"
)

func BeforeTest () {
	OverwatchConfiguration.ServerConfigFile = "test/test.json"
	OverwatchConfiguration.LoadConfiguration()
}


func TestOverwatchConfigurationAll (t *testing.T) {
	BeforeTest()

	list := OverwatchConfiguration.LogsToWatch.All()
	// 3x test-service-1
	// 3x test-service-2
	// 1x basic-log-file-1
	// 1x basic-log-file-2
	// =8
	if len(list) != 8 {
		t.Errorf("Expected 8, found %d", len(list))
	}

	var matches []LogToWatch
	var misses  []LogToWatch
	names := regexp.MustCompile(`((\w+)\-(\w+)\-log)|(basic\-log\-file\-\d+)`)

	for _, logfile := range list {
		switch {
		case names.MatchString(logfile.Name):
			matches = append(matches, logfile)
		default:
			misses  = append(misses,  logfile)
		}
	}

	if len(misses) != 0 {
		t.Errorf("Expected 0, got %d", len(misses))
		for _, logfile := range misses {
			t.Errorf("- LogFile#Name %s doesn't match regex %s", logfile.Name, names.String())
		}
	}
}
