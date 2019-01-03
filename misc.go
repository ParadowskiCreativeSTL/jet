package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// GenerateBackupString generates a backup with the time format YYYY-MM-DD_HH-mm-ss
func GenerateBackupString() string {
	t := time.Now()

	return fmt.Sprintf("%d-%02d-%d_%d-%d-%d", t.Year(), t.Month(), t.Day(), t.Hour(),
		t.Minute(), t.Second())
}

// RenameUrls uses the PHP binary to rename URL's
func RenameUrls(config Config, backupName string) error {
	replacePattern := strings.Join(config.Environments.Production.TargetURLPatterns, "|")
	cmd := exec.Command(config.BinaryPaths.PHP,
		fmt.Sprintf("%s/vendor/bin/srdb.cli.php", GetWorkingDirectory()),
		fmt.Sprintf("--host=%s", config.Environments.Production.Database.Host),
		fmt.Sprintf("--port=%d", config.Environments.Production.Database.Port),
		fmt.Sprintf("--user=%s", config.Environments.Production.Database.Username),
		fmt.Sprintf("--pass=%s", config.Environments.Production.Database.Password),
		"--regex",
		fmt.Sprintf("--search=%s", replacePattern),
		fmt.Sprintf("--replace=%s", config.Environments.Production.ReplacementURL),
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

// FlushWordPressCache flushes the WP cache
func FlushWordPressCache(config Config) error {
	cmd := exec.Command(config.BinaryPaths.WP,
		"cache",
		"flush",
		fmt.Sprintf("--path=%s", GetWorkingDirectory()+"htdocs/wp/"),
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

// CallProductionScript calls the production script and passes in the backup name
func CallProductionScript(config Config, backupName string) error {
	cmd := exec.Command(config.BinaryPaths.SSH,
		fmt.Sprintf("%s@%s",
			config.Environments.Production.User,
			config.Environments.Production.Host,
		),
		fmt.Sprintf("cd %s && /usr/local/bin/jet --environment=production %s",
			config.Environments.Production.RootDirectory,
			backupName,
		),
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
