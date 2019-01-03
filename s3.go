package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// SyncUploads syncs the local filesystem with S3
func SyncUploads(config Config) error {
	err := loadAwsConfigFile()
	if err != nil {
		return err
	}

	local := loadLocalFiles(path.Join(GetWorkingDirectory(), config.Environments.Staging.UploadsLocation))

	s3URL, err := url.Parse(config.S3.URL)
	if err != nil {
		return errors.New("could not parse the s3uri")
	}
	if s3URL.Scheme != "s3" {
		return errors.New("s3uri argument does not have valid protocol, should be 's3'")
	}
	if s3URL.Host == "" {
		return errors.New("s3uri is missing bucket name")
	}

	sess, err := getSession(config.S3.Region)
	if err != nil {
		return err
	}

	s3config := &S3Config{
		S3Service:    s3.New(sess),
		Bucket:       s3URL.Host,
		BucketPrefix: config.S3.BucketPrefix + "/",
	}

	remote := loadS3Files(s3config, 50000)

	files := compare(local, remote)

	syncFiles(s3config, files)

	return nil
}

// SyncDatabaseBackup syncs the database backup to S3
func SyncDatabaseBackup(config Config, backupName string) error {
	err := loadAwsConfigFile()
	if err != nil {
		return err
	}

	local := loadLocalFiles(path.Join(GetWorkingDirectory(), "staging_dump.sql"))

	s3URL, err := url.Parse(config.S3.URL)
	if err != nil {
		return errors.New("could not parse the s3uri")
	}
	if s3URL.Scheme != "s3" {
		return errors.New("s3uri argument does not have valid protocol, should be 's3'")
	}
	if s3URL.Host == "" {
		return errors.New("s3uri is missing bucket name")
	}

	sess, err := getSession(config.S3.Region)
	if err != nil {
		return err
	}

	s3config := &S3Config{
		S3Service:    s3.New(sess),
		Bucket:       s3URL.Host,
		BucketPrefix: "database_backups/" + backupName + "/",
	}

	remote := loadS3Files(s3config, 50000)

	files := compare(local, remote)

	syncFiles(s3config, files)

	return nil
}

func loadAwsConfigFile() error {
	if _, err := os.Stat(os.Getenv("HOME") + "/.aws/credentials"); os.IsNotExist(err) {
		return errors.New("aws credentials file does not exist")
	}

	return nil
}

func loadLocalFiles(basePath string) chan *FileStat {
	out := make(chan *FileStat)
	basePath = filepath.ToSlash(basePath)

	go func() {
		defer close(out)

		stat, err := os.Stat(basePath)
		if err != nil {
			out <- &FileStat{Err: err}
			return
		}

		absPath, err := filepath.Abs(basePath)
		if err != nil {
			out <- &FileStat{Err: err}
			return
		}

		if !stat.IsDir() {
			out <- &FileStat{
				Name:    filepath.Base(basePath),
				Path:    absPath,
				ModTime: stat.ModTime(),
				Size:    stat.Size(),
			}
			return
		}

		err = filepath.Walk(basePath, func(filePath string, stat os.FileInfo, err error) error {
			relativePath := relativePath(basePath, filepath.ToSlash(filePath))

			if stat == nil || stat.IsDir() {
				return nil
			}

			absPath, err := filepath.Abs(filePath)
			if err != nil {
				out <- &FileStat{
					Err: err,
				}
			}

			out <- &FileStat{
				Name:    relativePath,
				Path:    absPath,
				ModTime: stat.ModTime(),
				Size:    stat.Size(),
			}

			return nil
		})
		if err != nil {
			out <- &FileStat{Err: err}
			return
		}
	}()

	return out
}

func relativePath(path string, filePath string) string {
	if path == "." {
		return strings.TrimPrefix(filePath, "/")
	}
	path = strings.TrimPrefix(path, "./")
	a := strings.TrimPrefix(filePath, path)

	return strings.TrimPrefix(a, "/")
}

func loadS3Files(conf *S3Config, buffer int) chan *FileStat {
	out := make(chan *FileStat, buffer)

	go func() {
		continuationToken := listS3Files(conf, out, nil)
		for continuationToken != nil {
			continuationToken = listS3Files(conf, out, continuationToken)
		}
		close(out)
	}()

	return out
}

func listS3Files(config *S3Config, out chan *FileStat, token *string) *string {
	list, err := config.S3Service.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket:            aws.String(config.Bucket),
		Prefix:            aws.String(config.BucketPrefix),
		ContinuationToken: token,
	})
	if err != nil {
		out <- &FileStat{Err: err}
		return nil
	}

	for _, object := range list.Contents {
		out <- &FileStat{
			Name:    strings.TrimPrefix(*object.Key, config.BucketPrefix+"/"),
			Path:    *object.Key,
			Size:    *object.Size,
			ModTime: *object.LastModified,
		}
	}

	return list.NextContinuationToken
}

func getSession(region string) (*session.Session, error) {
	options := session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}

	sess, err := session.NewSessionWithOptions(options)
	if err != nil {
		return sess, err
	}

	sess.Config.Region = aws.String(region)

	return sess, nil
}

func compare(foundLocal, foundRemote chan *FileStat) chan *FileStat {
	update := make(chan *FileStat, 8)

	// first we sink the local files into a lookup map so its quick and easy to compare that to the remote
	localFiles := make(map[string]*FileStat)
	for r := range foundLocal {
		if r.Err != nil {
			log.Fatal(r.Err)
			continue
		}
		localFiles[r.Name] = r
	}

	numRemoteFiles := 0

	go func() {
		defer close(update)

		for remote := range foundRemote {
			if remote.Err != nil {
				return
			}
			numRemoteFiles++
			if local, ok := localFiles[remote.Name]; ok {
				if local.Size != remote.Size {
					update <- local
				} else if local.ModTime.After(remote.ModTime) {
					update <- local
				}
				delete(localFiles, remote.Name)
			}
		}

		for _, local := range localFiles {
			update <- local
		}
	}()

	return update
}

func syncFiles(config *S3Config, in chan *FileStat) {
	concurrency := 5
	sem := make(chan bool, concurrency)
	var numSyncedFiles int

	for file := range in {
		// add one
		sem <- true
		go func(config *S3Config, file *FileStat) {
			err := upload(config, file)
			if err != nil {
				fmt.Println(err)
			} else {
				numSyncedFiles++
			}
			// remove one
			<-sem
		}(config, file)
	}

	// After the last goroutine is fired, there are still concurrency amount of goroutines running. In order to make
	// sure we wait for all of them to finish, we attempt to fill the semaphore back up to its capacity. Once that
	// succeeds, we know that the last goroutine has read from the semaphore, as we've done len(files) + cap(sem) writes
	// and len(files) reads off the channel.
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}
}

func upload(config *S3Config, fileStat *FileStat) error {
	file, err := os.Open(fileStat.Path)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("Problem closing file %s: %v", fileStat.Path, err)
		}
	}()

	contentType := "application/octet-stream"
	// Don't try to detect content types on empty files
	if fileStat.Size != 0 {
		// detect the ContentType in the first 512 bytes of the file
		magicBytes := make([]byte, 512)
		if _, err := file.Read(magicBytes); err != nil {
			return err
		}
		if _, err := file.Seek(0, 0); err != nil {
			return err
		}
		contentType = http.DetectContentType(magicBytes)
	}

	key := filepath.Join(config.BucketPrefix, fileStat.Name)
	key = strings.TrimPrefix(key, "/")

	// Create an uploader (can do multipart) with S3 client and default options
	uploader := s3manager.NewUploaderWithClient(config.S3Service)
	params := &s3manager.UploadInput{
		Bucket:      aws.String(config.Bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
	}

	if _, err = uploader.Upload(params); err != nil {
		return err
	}

	return nil
}
