package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// DumpDatabase produces a database dump of staging environment
func DumpDatabase(config Config) error {
	cmd := exec.Command(config.BinaryPaths.MySQLDump,
		"--defaults-file=mysql.cnf",
		"--no-create-db",
		"--skip-lock-tables",
		"--result-file=staging_dump.sql",
		config.Environments.Staging.Database.Name,
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

// DumpPersistentTables produces a database dump of persistent
// tables in the production environment
func DumpPersistentTables(config Config) error {
	persistentTables := config.Environments.Production.Database.PersistentTables
	if len(persistentTables) == 0 {
		return errors.New("could not find persistent tables in config")
	}
	for i, table := range persistentTables {
		persistentTables[i] = config.Environments.Production.Database.TablePrefix + table
	}

	args := append([]string{
		"--defaults-file=mysql.cnf",
		"--no-create-db",
		"--skip-lock-tables",
		"--result-file=persistent_tables_dump.sql",
		config.Environments.Production.Database.Name,
	}, persistentTables...)

	cmd := exec.Command(config.BinaryPaths.MySQLDump, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

// RestoreFromBackup restores the MySQL dump on production
func RestoreFromBackup(config Config, backupName string) error {
	err := createDatabase(config, backupName)
	if err != nil {
		return err
	}

	err = grantPrivilagesForHost(config, backupName)
	if err != nil {
		return err
	}

	cmd := exec.Command(config.BinaryPaths.MySQL,
		"--defaults-file=mysql.cnf",
		fmt.Sprintf("--host=%s", config.Environments.Production.Database.Host),
		fmt.Sprintf("--port=%d", config.Environments.Production.Database.Port),
		config.Environments.Production.Database.Name+"_"+backupName,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	stdin, _ := cmd.StdinPipe()
	go func() {
		defer stdin.Close()
		file, _ := os.Open("staging_dump.sql")
		io.Copy(stdin, file)
	}()

	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

// RestorePersistentTables restores the persistent tables to the MySQL backup
func RestorePersistentTables(config Config, backupName string) error {
	cmd := exec.Command(config.BinaryPaths.MySQL,
		"--defaults-file=mysql.cnf",
		fmt.Sprintf("--host=%s", config.Environments.Production.Database.Host),
		fmt.Sprintf("--port=%d", config.Environments.Production.Database.Port),
		config.Environments.Production.Database.Name+"_"+backupName,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	stdin, _ := cmd.StdinPipe()
	go func() {
		defer stdin.Close()
		file, _ := os.Open("persistent_tables_dump.sql")
		io.Copy(stdin, file)
	}()

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func createDatabase(config Config, backupName string) error {
	cmd := exec.Command(config.BinaryPaths.MySQLAdmin,
		"--defaults-file=mysql.cnf",
		fmt.Sprintf("--host=%s", config.Environments.Production.Database.Host),
		fmt.Sprintf("--port=%d", config.Environments.Production.Database.Port),
		"create",
		config.Environments.Production.Database.Name+"_"+backupName,
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func grantPrivilagesForHost(config Config, backupName string) error {
	args := []string{
		"--defaults-file=mysql.cnf",
		fmt.Sprintf("--host=%s", config.Environments.Production.Database.Host),
		fmt.Sprintf("--port=%d", config.Environments.Production.Database.Port),
		"--execute",
		fmt.Sprintf("GRANT SELECT, INSERT ON `%s`.* TO '%s'@'%%';",
			config.Environments.Production.Database.Name+"_"+backupName,
			config.Environments.Production.Database.Username,
		),
	}
	cmd := exec.Command(config.BinaryPaths.MySQL, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
