package main

import (
	"os"
	"testing"
)

func TestDumpDatabase(t *testing.T) {
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
			MySQLDump: "/usr/local/bin/mysqldump",
		},
		Environments: struct {
			Staging      Environment `json:"staging"`
			Production   Environment `json:"production"`
			LoadBalancer Environment `json:"load_balancer"`
		}{
			Staging: Environment{
				Database: Database{
					Name:     "test",
					Host:     "127.0.0.1",
					Port:     4336,
					Username: "test",
					Password: "test",
				},
			},
		},
	}

	err := DumpDatabase(*config)
	if err != nil {
		t.Error("there was a problem dumping the database: " + err.Error())
	}

	err = os.Remove("staging_dump.sql")
	if err != nil {
		t.Error("could not destroy the test sql dump file")
	}
}

func TestDumpPersistentTables(t *testing.T) {
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
			MySQLDump: "/usr/local/bin/mysqldump",
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
					PersistentTables: []string{
						"test",
						"another_test",
					},
					TablePrefix: "test_",
				},
			},
		},
	}

	err := DumpPersistentTables(*config)
	if err != nil {
		t.Error("there was a problem dumping the persistent tables: ", err.Error())
	}

	err = os.Remove("persistent_tables_dump.sql")
	if err != nil {
		t.Error("could not destroy the test sql dump file")
	}
}

func TestRestoreFromBackup(t *testing.T) {
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
			MySQL:      "/usr/local/bin/mysql",
			MySQLAdmin: "/usr/local/bin/mysqladmin",
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
					PersistentTables: []string{
						"test",
						"another_test",
					},
					TablePrefix: "test_",
				},
			},
		},
	}

	backupName := GenerateBackupString()

	err := RestoreFromBackup(*config, backupName)
	if err != nil {
		t.Error("there was an issue restoring the sql backup: ", err.Error())
	}
}

func TestRestorePersistentTables(t *testing.T) {
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
			MySQL:      "/usr/local/bin/mysql",
			MySQLAdmin: "/usr/local/bin/mysqladmin",
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
					PersistentTables: []string{
						"test",
						"another_test",
					},
					TablePrefix: "test_",
				},
			},
		},
	}

	backupName := GenerateBackupString()

	err := RestoreFromBackup(*config, backupName)
	if err != nil {
		t.Error("there was an issue restoring the sql backup: ", err.Error())
	}

	err = RestorePersistentTables(*config, backupName)
	if err != nil {
		t.Error("there was an issue restoring the sql backup: ", err.Error())
	}
}
