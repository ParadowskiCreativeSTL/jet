package main

import (
	"time"

	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

// FileStat describes a local and remote file
type FileStat struct {
	Err     error
	Name    string
	Path    string
	Size    int64
	ModTime time.Time
}

// S3Config contains common paths and configuration
type S3Config struct {
	S3Service    s3iface.S3API
	Bucket       string
	BucketPrefix string
}

// Database describes what a database config looks like
type Database struct {
	Name             string   `json:"name"`
	Host             string   `json:"host"`
	Port             int16    `json:"port"`
	Username         string   `json:"username"`
	Password         string   `json:"password"`
	PersistentTables []string `json:"persistent_tables"`
	TablePrefix      string   `json:"table_prefix"`
}

// Environment describes the structure of an environment
type Environment struct {
	User              string   `json:"user"`
	Host              string   `json:"host"`
	RootDirectory     string   `json:"root_directory"`
	UploadsLocation   string   `json:"uploads_location"`
	Database          Database `json:"database"`
	TargetURLPatterns []string `json:"target_url_patterns"`
	ReplacementURL    string   `json:"replacement_url"`
}

// Config contains the jet config file
type Config struct {
	S3 struct {
		URL          string `json:"url"`
		Region       string `json:"region"`
		BucketPrefix string `json:"bucket_prefix"`
	} `json:"s3"`
	BinaryPaths struct {
		SSH        string `json:"ssh"`
		MySQLAdmin string `json:"mysql_admin"`
		MySQLDump  string `json:"mysql_dump"`
		MySQL      string `json:"mysql"`
		SCP        string `json:"scp"`
		PHP        string `json:"php"`
		WP         string `json:"wp"`
	} `json:"binary_paths"`
	Environments struct {
		Staging      Environment `json:"staging"`
		Production   Environment `json:"production"`
		LoadBalancer Environment `json:"load_balancer"`
	} `json:"environments"`
}
