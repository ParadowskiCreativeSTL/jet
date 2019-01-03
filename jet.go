// Copyright 2018 Paradowski Creative

// Permission is hereby granted, free of charge, to any person obtaining a copy of this
// software and associated documentation files (the "Software"), to deal in the Software
// without restriction, including without limitation the rights to use, copy, modify, merge,
// publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons
// to whom the Software is furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all copies
// or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED,
// INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
// HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
// OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
// OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package main

import (
	"flag"
	"log"
	"os"
	"time"

	"go.uber.org/zap"
)

var (
	backupName         string
	currentEnvironment string
)

func main() {
	flag.StringVar(&currentEnvironment, "environment", "", "contains the environment in which the tool is currently running")
	flag.Parse()

	if currentEnvironment == "" {
		log.Fatalf("Please pass in the --environment flag.")
		os.Exit(1)
	}
	start := time.Now()

	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logger.Info("Production Deployment Started")

	// Load configuration file
	config, err := LoadConfigFile()
	if err != nil {
		logger.Fatal(err.Error())
	}

	/**
	 * Staging Server
	 */
	if currentEnvironment == "staging" {
		// Generate a backup name
		backupName := GenerateBackupString()
		logger.Info("Generated Backup Name",
			zap.String("name", backupName),
		)

		// Push wp-uploads to S3
		err = SyncUploads(config)
		if err != nil {
			logger.Fatal("There was an error syncing uploads with S3",
				zap.Error(err),
			)
		}
		logger.Info("Pushed Uploads to S3")

		// Dump MySQL database
		err = DumpDatabase(config)
		if err != nil {
			logger.Fatal("There was an error dumping the MySQL database",
				zap.Error(err),
			)
		}
		logger.Info("Staging Database Backup Created")

		// Push MySQL Dump to Production
		err = TransferFile("staging_dump.sql", config)
		if err != nil {
			logger.Fatal("There was an error transfering the MySQL dump to production",
				zap.Error(err),
			)
		}
		logger.Info("Pushed MySQL Dump to Production")
	}

	/**
	 * Production Server
	 */
	if currentEnvironment == "production" {
		// Back up persistent tables
		err = DumpPersistentTables(config)
		if err != nil {
			logger.Fatal("There was an error dumping the MySQL database",
				zap.Error(err),
			)
		}
		logger.Info("Persistent Tables Backup Created")

		// Restore the database backup
		err = RestoreFromBackup(config, backupName)
		if err != nil {
			logger.Fatal("There was an error restoring the database",
				zap.Error(err),
			)
		}

		// Restore persistent tables
		err = RestorePersistentTables(config, backupName)
		if err != nil {
			logger.Fatal("There was an error restoring the persistent tables",
				zap.Error(err),
			)
		}

		// Rename URLs
		err = RenameUrls(config, backupName)
		if err != nil {
			logger.Fatal("There was an error renaming URLs in the production database",
				zap.Error(err),
			)
		}

		// Flush WordPress cache
		err = FlushWordPressCache(config)
		if err != nil {
			logger.Fatal("There was an error flushing the WordPress cache",
				zap.Error(err),
			)
		}

		// Sync MySQL backup to S3
		err = SyncDatabaseBackup(config, backupName)
		if err != nil {
			logger.Fatal("There was an error syncing the database backup to S3",
				zap.Error(err),
			)
		}

		// Update production .env file
		err = UpdateEnvFile(backupName)
		if err != nil {
			logger.Fatal("There was an error updating the .env file",
				zap.Error(err),
			)
		}
	}

	logger.Info("Production Deployment Completed Successfully!",
		zap.Duration("execution time", time.Since(start)),
	)
}
