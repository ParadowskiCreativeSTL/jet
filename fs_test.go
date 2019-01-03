package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"strings"
	"testing"
)

func TestGetWorkingDirectory(t *testing.T) {
	wd := GetWorkingDirectory()
	if wd == "" {
		t.Error("could not print the working directory")
	} else {
		fmt.Println("Working Directory: " + wd)
	}
}

func TestLoadConfigFile(t *testing.T) {
	f, err := os.Create("config.json")
	if err != nil {
		t.Error("could not create a test config file")
	}
	defer f.Close()

	_, err = LoadConfigFile()
	if err != nil {
		t.Error("could not load the config file")
	}

	err = os.Remove("config.json")
	if err != nil {
		t.Error("could not destroy the test config file")
	}
}

func TestTransferFile(t *testing.T) {
	usr, err := user.Current()
	if err != nil {
		t.Error("unable to obtain the current user: " + err.Error())
	}
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
			SCP: "/usr/bin/scp",
		},
		Environments: struct {
			Staging      Environment `json:"staging"`
			Production   Environment `json:"production"`
			LoadBalancer Environment `json:"load_balancer"`
		}{
			Production: Environment{
				User:          "",
				Host:          "",
				RootDirectory: usr.HomeDir,
			},
		},
	}

	err = TransferFile("testdata/dummy.pdf", *config)
	if err != nil {
		t.Error("there was a problem trying to use scp: ", err.Error())
	}
}

func TestUpdateEnvFile(t *testing.T) {
	sampleEnv := []byte(`DB_NAME=test_database_00-00-0000
DB_USER=test_user
DB_PASSWORD=test_password
DB_HOST=localhost
DB_PREFIX=testprefix_`)
	filePermissions, err := os.Stat(path.Join(GetWorkingDirectory(), ".env"))
	if err != nil {
		t.Error("unable to determine file permissions: ", err.Error())
	}

	err = ioutil.WriteFile(".env", sampleEnv, filePermissions.Mode())
	if err != nil {
		t.Error("unable to write sample .env file: ", err.Error())
	}

	backupName := GenerateBackupString()

	err = UpdateEnvFile(backupName)
	if err != nil {
		t.Error("unable to update env file: ", err.Error())
	}

	envFile, err := ioutil.ReadFile(path.Join(GetWorkingDirectory(), ".env"))
	if err != nil {
		t.Error("unable to open file sample .env file for comparison")
	}

	lines := strings.Split(string(envFile), "\n")

	for i := 0; i < len(lines); i++ {
		if strings.Contains(lines[i], "DB_NAME") {
			compareString := strings.Split(lines[i], "=")[1]
			if !strings.Contains(lines[i], compareString) {
				t.Error("DB_NAME was not successfully replaced")
			}
		}
	}

	err = os.Remove(".env")
	if err != nil {
		t.Error("could not destroy the test .env file")
	}
}
