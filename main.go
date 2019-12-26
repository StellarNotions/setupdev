package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
)

const SetupPackageBasePath string = "./setupPackages"

type notFoundHttpResponse struct {
	Msg string `json:"msg"`
}

type foundHttpResponse struct {
	Msg                  string `json:"msg"`
	SetupPackageFileName string `json:"setupPackageFileName"`
}

type setupPackage struct {
	FileName string `json:"fileName"`
}

type allSetupPackages []setupPackage

var setupPackages = allSetupPackages{

	{
		FileName: "nodejs_12-14-0_ubuntu.zip",
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
		SetupPackageFileName: "nodejs_12-14-0_ubuntu.zip",
	},
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

func checkFileExists(filePath, fileName string) (bool, string) {
	if _, err := os.Stat(filePath + fileName); os.IsNotExist(err) {
		return false, "file does not exist"
	}

	return true, "found file"
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

func getSetupPackageFile(fileName string) (*os.File, string, string, bool, string) {
	fileExists, msg := checkFileExists(SetupPackageBasePath, fileName)
	if !fileExists {
		return nil, "", "", false, msg
	}

	OpenFile, err := os.Open(SetupPackageBasePath + "/" + fileName)
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
	router := mux.NewRouter()
	router.HandleFunc("/", homeLink)
	router.HandleFunc("/{name}/{version}/{operatingSystem}", getSetupPackage)

	fmt.Print("Running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", router))
}
