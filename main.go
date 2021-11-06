package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/dinumathai/admission-webhook-sample/injector"
)

func main() {
	// https://github.com/kubernetes/kubernetes/tree/release-1.9/test/images/webhook

	// RUN ./admission-webhook-sample -stderrthreshold=INFO -v=3
	// JSON to Test - https://github.com/snowdrop/kubernetes-info-webhook/blob/master/src/test/resources/admission-review.json

	// Dummy Default Value
	k8sPatch := getDefaultPatch()
	if os.Getenv("PATCH_FILE_NAME") != "" {
		dat, err := ioutil.ReadFile(os.Getenv("PATCH_FILE_NAME"))
		if err != nil {
			fmt.Printf("ERROR : Unable to read patch file : %v", err)
			os.Exit(1)
		}
		k8sPatch = string(dat)
	}

	// Set the environment variable SSL_CRT_FILE_NAME and SSL_KEY_FILE_NAME to start in https mode
	// This programs checks for the annotation "inject-init-container": "true" and injects returns the above patch if annotation matches

	// TODO : Multiple call like this must give multiple servers
	injector.StartServer(k8sPatch, ":8080", "/inject-init-container")
}

func getDefaultPatch() string {
	return `[
		{
		  "op": "add",
		  "path": "/spec/initContainers",
		  "value": [
			{
			  "image": "busybox:1.28",
			  "name": "init-myservice",
			  "command": [
				"sh",
				"-c",
				"echo -------------------Executed_The_InIt_Container-----------------"
			  ],
			  "resources": {
				
			  }
			}
		  ]
		}
	  ]`
}
