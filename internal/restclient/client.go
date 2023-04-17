/*
 * Copyright (c) 2021 Siemens AG
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package restclient

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	v1 "onboardservice/api/siemens_iedge_dmapi_v1"

	"google.golang.org/protobuf/encoding/protojson"
)

// EdgeCoreRuntimeRest  API interface
type EdgeCoreRuntimeRest interface {
	Activate(configuration *v1.DeviceConfiguration) (bool, error)
	Onboarded() (bool, error)
}

// EdgeCoreClient implements EdgeCoreRuntimeRest interface.
type EdgeCoreClient struct {
	activationURL   string
	onboardCheckURL string
	client          *http.Client
}

// NewClient Creates a new EdgeCoreClient with given IEM base url.
func NewClient(baseurl string) *EdgeCoreClient {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	cli := &http.Client{Transport: tr}
	return &EdgeCoreClient{
		activationURL:   baseurl + "/v1/activate",
		onboardCheckURL: baseurl + "/v1",
		client:          cli,
	}

}

// Activate allows the device activation to EdgeCoreRuntime by delivering onboarding configuration.
func (hc *EdgeCoreClient) Activate(configuration *v1.DeviceConfiguration) (bool, error) {
	jsonContent := protojson.MarshalOptions{Multiline: true, UseProtoNames: true}.Format(configuration)
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	//Multipart message , "files" section as documented in EdgeCoreRuntime API
	part1, _ := writer.CreateFormFile("files", "deviceconfig.json")
	part1.Write([]byte(jsonContent + "\n"))

	err := writer.Close()
	if err != nil {
		fmt.Println(err)
		return false, err

	}

	req, err := http.NewRequest("POST", hc.activationURL, payload)

	if err != nil {
		fmt.Println(err)
		return false, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	//Make the Request to API
	res, err := hc.client.Do(req)
	if err == nil {
		defer res.Body.Close()
	} else {
		log.Println(req.URL, "failed", err)
		return false, err
	}
	//Parse response.
	body, err := ioutil.ReadAll(res.Body)
	fmt.Println(string(body))
	if res.StatusCode >= 200 && res.StatusCode < 400 {
		log.Println("onboard succeed")
		return true, nil
	}

	if err == nil {
		// Prepare a error message containing the response from /activate
		err = errors.New(string(body))
	}

	log.Println(req.URL, ":onboard failed:", err)

	return false, err
}

// Onboarded function answers : Is the device  onboarded or not?
func (hc *EdgeCoreClient) Onboarded() (bool, error) {
	response, err := hc.client.Get(hc.onboardCheckURL)

	retVal := false
	if err != nil {
		log.Println("Onboard Status Check call failed:", err)

	} else if response.StatusCode >= 200 && response.StatusCode < 400 {
		retVal = true
	} else if response.StatusCode == 531 {
		retVal = false
	} else {
		err = errors.New("Service not available")
	}
	if response != nil {
		log.Println("HTTP Response :", response.StatusCode)
	}

	return retVal, err
}
