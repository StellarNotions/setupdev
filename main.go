package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gorilla/mux"
)

const ENVFilePath string = "./"
const ENVFileName string = ".env.json"

var ENV ENVStruct

type ENVStruct struct {
	AWS struct {
		Credentials struct {
			AccessKeyId  string `json:"ACCESS_KEY_ID"`
			SecretKey    string `json:"SECRET_ACCESS_KEY"`
		}
		S3 struct {
			BucketName   string `json:"BUCKET_NAME"`
			Region       string `json:"REGION"`
		}
	}
	SetupPackageBasePath string `json:"SETUP_PACKAGE_BASE_PATH"`
}

type notFoundHttpResponse struct {
	Msg string `json:"msg"`
}

type setupPackage struct {
	FileName string `json:"fileName"`
}

type allSetupPackages []setupPackage

var setupPackages = allSetupPackages{

	{
		FileName: "nodejs_12-14-0_ubuntu_v1.zip",
	},
}

type supportedApplication struct {
	Name                 string `json:"name"`
	Version              string `json:"version"`
	OperatingSystem      string `json:"operatingSystem"`
	SetupPackageFileName string `json:"setupPackageFileName"`
}

type allSupportedApplications []supportedApplication

var supportedApplications = allSupportedApplications{
	{
		Name:                 "nodejs",
		Version:              "12.14.0",
		OperatingSystem:      "ubuntu",
		SetupPackageFileName: "nodejs_12-14-0_ubuntu_v1.zip",
	},
}

func loadENVs(filePath, fileName string) bool {
	ENVFile, err := os.Open(filePath + fileName)
	defer ENVFile.Close()
	if err != nil {
		return false
	}
	jsonParser := json.NewDecoder(ENVFile)
	_ = jsonParser.Decode(&ENV)
	return true
}

func homeLink(w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprintf(w, "Welcome home!")
}

func findSetupPackage(packageFileName string) (setupPackage, bool, string) {
	for _, singleSetupPackage := range setupPackages {
		if singleSetupPackage.FileName == packageFileName {
			return singleSetupPackage, true, "found setup package"
		}
	}

	return setupPackage{}, false, "no setup package found"
}

func findSupportedApplication(packageName, packageVersion, packageOperatingSystem string) (supportedApplication, bool, string) {
	for _, singleSupportedApplication := range supportedApplications {
		if singleSupportedApplication.Name == packageName &&
			singleSupportedApplication.Version == packageVersion &&
			singleSupportedApplication.OperatingSystem == packageOperatingSystem {
			return singleSupportedApplication, true, "found application"
		}
	}

	return supportedApplication{}, false, "no application found"
}

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
	FileStat, _ := OpenFile.Stat()                     // Get info from file
	FileSize := strconv.FormatInt(FileStat.Size(), 10) // Get file size as a string

	return FileContentType, FileSize
}

func getFileFromS3(filePath, fileName string) bool {
	// The session the S3 Downloader will use
	sess := session.Must(session.NewSession(&aws.Config{
		//Credentials: credentials.NewStaticCredentials(ENV.AWS.Credentials.AccessKeyId, ENV.AWS.Credentials.SecretKey, ""),
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
		//log.Fatal(fmt.Errorf("failed to download file, %v", err))
		log.Fatal(err)
		return false
	}
	fmt.Printf("file downloaded, %d bytes\n", n)
	return true
}

func getSetupPackageFile(fileName string) (*os.File, string, string, bool, string) {
	fileExists, _ := checkFileExistsLocally(ENV.SetupPackageBasePath, fileName)
	if !fileExists {
		ok := getFileFromS3(ENV.SetupPackageBasePath, fileName)
		if !ok {
			log.Fatal(ok)
			panic("unable to get file from S3")
		}
	}

	OpenFile, err := os.Open(ENV.SetupPackageBasePath + "/" + fileName)
	if err != nil {
		return nil, "", "", false, "setup package file does not exist"
	}

	FileContentType, FileSize := getFileInfo(OpenFile)

	return OpenFile, FileContentType, FileSize, true, "found setup package file"
}

func getSetupPackage(w http.ResponseWriter, r *http.Request) {
	packageName := mux.Vars(r)["name"]
	packageVersion := mux.Vars(r)["version"]
	packageOperatingSystem := mux.Vars(r)["operatingSystem"]

	foundSupportedApplication, ok, msg := findSupportedApplication(packageName, packageVersion, packageOperatingSystem)

	if !ok {
		_ = json.NewEncoder(w).Encode(notFoundHttpResponse{msg})
		return
	}

	foundSetupPackage, ok, msg := findSetupPackage(foundSupportedApplication.SetupPackageFileName)

	if !ok {
		_ = json.NewEncoder(w).Encode(notFoundHttpResponse{msg})
		return
	}

	OpenFile, FileContentType, FileSize, ok, msg := getSetupPackageFile(foundSetupPackage.FileName)

	if !ok {
		_ = json.NewEncoder(w).Encode(notFoundHttpResponse{msg})
		return
	}

	// Send the headers
	w.Header().Set("Content-Disposition", "attachment; filename="+foundSetupPackage.FileName)
	w.Header().Set("Content-Type", FileContentType)
	w.Header().Set("Content-Length", FileSize)

	// Send the file
	// 512 bytes read from the file already, so reset the offset back to 0
	_, _ = OpenFile.Seek(0, 0)
	_, _ = io.Copy(w, OpenFile)
}

func main() {
	loadedSuccessfully := loadENVs(ENVFilePath, ENVFileName)
	if !loadedSuccessfully {
		log.Fatal("Unable to load ENV file")
	} else {
		router := mux.NewRouter()
		router.HandleFunc("/", homeLink)
		router.HandleFunc("/{name}/{version}/{operatingSystem}", getSetupPackage)

		fmt.Print("Running on port 8080...")
		log.Fatal(http.ListenAndServe(":8080", router))
	}
}
