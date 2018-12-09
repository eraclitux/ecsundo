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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_nameFromARN(t *testing.T) {
	tests := []struct {
		arn  string
		want string
	}{
		{
			arn:  "arn:partition:service:region:account-id:resourcetype/resource-name",
			want: "resource-name",
		},
		{
			arn:  "arn:partition:service:region:account-id:resourcetype/resource-name/qualifier",
			want: "resource-name",
		},
		{
			arn:  "arn:partition:service:region:account-id:resource",
			want: "",
		},
		{
			arn:  "",
			want: "",
		},
	}
	for _, tt := range tests {
		if got := nameFromARN(tt.arn); got != tt.want {
			t.Errorf("nameFromARN() = %v, want %v", got, tt.want)
		}
	}
}

func Test_getRegion(t *testing.T) {
	validRegion := "us-east-1"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w,
			`{
			"availabilityZone" : "us-east-1c",
			"devpayProductCodes" : null,
			"marketplaceProductCodes" : null,
			"version" : "2017-09-30",
			"instanceId" : "i-01d91b64bee66602b",
			"billingProducts" : null,
			"instanceType" : "t2.micro",
			"imageId" : "ami-40d28157",
			"privateIp" : "172.31.60.249",
			"accountId" : "787961527100",
			"architecture" : "x86_64",
			"kernelId" : null,
			"ramdiskId" : null,
			"pendingTime" : "2016-11-07T14:30:32Z",
			"region" : "us-east-1"
		  }`,
		)
	}))
	defer server.Close()
	region, err := getRegion(server.URL)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	if region != validRegion {
		t.Fatalf("got: %s, expected: %s\n", region, validRegion)
	}
}
