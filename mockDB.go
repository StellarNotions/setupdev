package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"os"
)

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

func findSetupPackage(packageFileName string) (setupPackage, bool, string) {
	for _, singleSetupPackage := range setupPackages {
		if singleSetupPackage.FileName == packageFileName {
			return singleSetupPackage, true, "found setup package"
		}
	}

	return setupPackage{}, false, "no setup package found"
}

func getSetupPackageFile(fileName string) (*os.File, string, string, bool, string) {
	fileExists, _ := checkFileExistsLocally(ENV.SetupPackageBasePath, fileName)
	if !fileExists {
		ok := getFileFromS3(ENV.SetupPackageBasePath, fileName)
		if !ok {
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
