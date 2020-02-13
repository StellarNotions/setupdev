package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"log"
	"net/http"
	"os"
	"strconv"
)

func checkFileExistsLocally(filePath, fileName string) (bool, string) {
	if _, err := os.Stat(filePath + fileName); os.IsNotExist(err) {
		return false, "file does not exist"
	}

	return true, "file found"
}

func getFileInfo(OpenFile *os.File) (string, string) {
	// Create a buffer to store the header of the file in
	FileHeader := make([]byte, 512)
	// Copy the headers into the FileHeader buffer
	_, _ = OpenFile.Read(FileHeader)
	// Get content type of file
	FileContentType := http.DetectContentType(FileHeader)
	// Get the file size
	FileStat, _ := OpenFile.Stat()                           // Get info from file
	FileSize := strconv.FormatInt(FileStat.Size(), 10) // Get file size as a string

	return FileContentType, FileSize
}

func getFileFromS3(filePath, fileName string) bool {
	// The session the S3 Downloader will use
	sess := session.Must(session.NewSession(&aws.Config{
		Credentials: credentials.NewSharedCredentials("", "default"),
		Region: aws.String(ENV.AWS.S3.Region),
	}))

	// Create a downloader with the session and default options
	downloader := s3manager.NewDownloader(sess)

	// Create a file to write the S3 Object contents to.
	f, err := os.Create(filePath + "/" + fileName)
	if err != nil {
		//return fmt.Errorf("failed to create file %q, %v", filename, err)
		return false
	}

	// Write the contents of S3 Object to the file
	n, err := downloader.Download(f, &s3.GetObjectInput{
		Bucket: aws.String(ENV.AWS.S3.BucketName),
		Key:    aws.String(fileName),
	})
	if err != nil {
		//return fmt.Errorf("failed to download file, %v", err)
		log.Fatal(fmt.Errorf("failed to download file, %v", err))
		return false
	}
	fmt.Printf("file downloaded, %d bytes\n", n)
	return true
}
