package main

import (
	"testing"
)

func TestSyncUploads(t *testing.T) {
	config := &Config{
		S3: struct {
			URL          string `json:"url"`
			Region       string `json:"region"`
			BucketPrefix string `json:"bucket_prefix"`
		}{
			URL:          "s3://jet-content-deployment-test",
			Region:       "us-east-2",
			BucketPrefix: "htdocs/app/uploads",
		},
		Environments: struct {
			Staging      Environment `json:"staging"`
			Production   Environment `json:"production"`
			LoadBalancer Environment `json:"load_balancer"`
		}{
			Staging: Environment{
				UploadsLocation: "testdata",
			},
		},
	}

	err := SyncUploads(*config)
	if err != nil {
		t.Error("there was a problem syncing uploads with s3: " + err.Error())
	}
}

func TestSyncDatabaseBackup(t *testing.T) {
	config := &Config{
		S3: struct {
			URL          string `json:"url"`
			Region       string `json:"region"`
			BucketPrefix string `json:"bucket_prefix"`
		}{
			URL:          "s3://jet-content-deployment-test",
			Region:       "us-east-2",
			BucketPrefix: "htdocs/app/uploads",
		},
		Environments: struct {
			Staging      Environment `json:"staging"`
			Production   Environment `json:"production"`
			LoadBalancer Environment `json:"load_balancer"`
		}{
			Staging: Environment{
				UploadsLocation: "testdata",
			},
		},
	}

	backupName := GenerateBackupString()

	err := SyncDatabaseBackup(*config, backupName)
	if err != nil {
		t.Error("there was a problem syncing uploads with s3: " + err.Error())
	}
}
