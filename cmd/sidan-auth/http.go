package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// httpPost sends a POST request with JSON body
func httpPost(url string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}
	
	req, err := http.NewRequest("POST", url, reqBody)
	if err != nil {
		return nil, err
	}
	
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	
	client := &http.Client{}
	return client.Do(req)
}
