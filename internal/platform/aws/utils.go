// Copyright © 2018 Andrea Masi <eraclitux@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package aws

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

// nameFromARN returns resource name from an ARN in the form
// arn:partition:service:region:account-id:resourcetype/resource
func nameFromARN(ARN string) string {
	name := ""
	tokens := strings.Split(ARN, "/")
	if len(tokens) >= 2 {
		name = tokens[1]
	}
	return name
}

// getRegion tries to retrieve region from EC2 metadata.
func getRegion(metaDataEndpoints ...string) (string, error) {
	docEndpoint := "http://169.254.169.254/latest/dynamic/instance-identity/document"
	if len(metaDataEndpoints) > 0 {
		docEndpoint = metaDataEndpoints[0]
	}
	client := &http.Client{
		Timeout: 1 * time.Second,
	}
	res, err := client.Get(docEndpoint)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	resp := struct {
		Region string `json:"region"`
	}{}
	err = json.NewDecoder(res.Body).Decode(&resp)
	if err != nil && err != io.EOF {
		return "", err
	}
	return resp.Region, nil
}
