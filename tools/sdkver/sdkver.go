/*
Package sdk is the gRPC implementation of the SDK gRPC server
Copyright 2018 Portworx

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
// This file is used to read the SDK version generated from api.proto
// and update any other files with that information
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/libopenstorage/openstorage/api"
)

type optionTypes struct {
	swaggerFile string
}

var (
	options optionTypes
)

func init() {
	flag.StringVar(&options.swaggerFile, "swagger", "", "api.swagger.json file")
	flag.Parse()
}

func main() {

	// Set version
	version := fmt.Sprintf("%d.%d.%d",
		api.SdkVersion_Major,
		api.SdkVersion_Minor,
		api.SdkVersion_Patch)

	// Set swagger file
	if len(options.swaggerFile) != 0 {
		if err := setSwaggerVersion(options.swaggerFile, version); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Printf("%s\n", version)
}

func setSwaggerVersion(file, version string) error {

	swaggerFile, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	jsonFile := make(map[string]interface{})
	err = json.Unmarshal(swaggerFile, &jsonFile)
	if err != nil {
		return err
	}

	// Set SDK information
	jsonFile["info"].(map[string]interface{})["title"] = "OpenStorage SDK"
	jsonFile["info"].(map[string]interface{})["version"] = version

	bytes, err := json.MarshalIndent(jsonFile, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(file, bytes, 0644)
	if err != nil {
		return err
	}

	return nil
}
