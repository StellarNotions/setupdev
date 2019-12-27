package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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
