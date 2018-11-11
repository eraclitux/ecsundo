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

import "testing"

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
