package main

import (
	"testing"
)

func TestRenameUrls(t *testing.T) {
	config := &Config{
		BinaryPaths: struct {
			SSH        string `json:"ssh"`
			MySQLAdmin string `json:"mysql_admin"`
			MySQLDump  string `json:"mysql_dump"`
			MySQL      string `json:"mysql"`
			SCP        string `json:"scp"`
			PHP        string `json:"php"`
			WP         string `json:"wp"`
		}{
			PHP: "/usr/local/bin/php",
		},
		Environments: struct {
			Staging      Environment `json:"staging"`
			Production   Environment `json:"production"`
			LoadBalancer Environment `json:"load_balancer"`
		}{
			Production: Environment{
				Database: Database{
					Name:     "test",
					Host:     "127.0.0.1",
					Port:     4336,
					Username: "test",
					Password: "test",
				},
				TargetURLPatterns: []string{
					"badexample.com",
					"notexample.com",
				},
				ReplacementURL: "example.com",
			},
		},
	}

	backupName := GenerateBackupString()

	err := RenameUrls(*config, backupName)
	if err != nil {
		t.Error("there was an issue renaming URLs", err.Error())
	}
}

func TestFlushWordPressCache(t *testing.T) {
	config := &Config{
		BinaryPaths: struct {
			SSH        string `json:"ssh"`
			MySQLAdmin string `json:"mysql_admin"`
			MySQLDump  string `json:"mysql_dump"`
			MySQL      string `json:"mysql"`
			SCP        string `json:"scp"`
			PHP        string `json:"php"`
			WP         string `json:"wp"`
		}{
			WP: "/usr/local/bin/wp",
		},
	}

	err := FlushWordPressCache(*config)
	if err != nil {
		t.Error("there was an issue flushing the WordPress cache", err.Error())
	}
}
