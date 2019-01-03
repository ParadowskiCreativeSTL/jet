package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

// GetWorkingDirectory returns the current working directory
func GetWorkingDirectory() string {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get the current working directory.")
		os.Exit(1)
	}

	return wd
}

// LoadConfigFile loads the config file and returns the JSON
func LoadConfigFile() (Config, error) {
	var config Config
	configFile, err := os.Open(path.Join(GetWorkingDirectory(), "config.json"))
	defer configFile.Close()
	if err != nil {
		return config, errors.New("failed to load the configuration file: " + err.Error())
	}

	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)

	return config, nil
}

// TransferFile moves a file from one place to another using scp
func TransferFile(localFile string, config Config) error {
	var transferString string
	if config.Environments.Production.Host == "" {
		transferString = config.Environments.Production.RootDirectory
	} else {
		transferString = fmt.Sprintf("%s@%s:%s",
			config.Environments.Production.User,
			config.Environments.Production.Host,
			config.Environments.Production.RootDirectory)
	}
	args := []string{
		"-Cp",
		localFile,
		transferString,
	}

	cmd := exec.Command(config.BinaryPaths.SCP, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

// UpdateEnvFile updates the .env file to point to the new database
func UpdateEnvFile(backupName string) error {
	if backupName == "" {
		return errors.New("backupName string cannot be blank")
	}

	filePermissions, err := os.Stat(path.Join(GetWorkingDirectory(), ".env"))
	if err != nil {
		return errors.New("cannot read file permissions for .env file")
	}

	envFile, err := ioutil.ReadFile(path.Join(GetWorkingDirectory(), ".env"))
	if err != nil {
		return err
	}

	lines := strings.Split(string(envFile), "\n")

	for i := 0; i < len(lines); i++ {
		if strings.Contains(lines[i], "DB_NAME") {
			findString := strings.Split(lines[i], "=")[1]
			updatedLine := strings.Replace(lines[i], findString, backupName, 1)
			lines[i] = updatedLine
		}
	}

	updatedFile := strings.Join(lines, "\n")

	err = ioutil.WriteFile(path.Join(GetWorkingDirectory(), ".env"), []byte(updatedFile), filePermissions.Mode())
	if err != nil {
		return err
	}

	return nil
}
