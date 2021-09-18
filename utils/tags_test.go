// Copyright 2019 Yunion
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"reflect"
	"testing"
)

func TestFindWord(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{
		{`'abc'`, `abc`},
		{`"abc"`, `abc`},
		{`'id.in(123-123,456-456)'`, `id.in(123-123,456-456)`},
		{`--config`, `--config`},
	}
	for _, c := range cases {
		o := Unquote(c.in)
		t.Logf("in: %s out: %s expect: %s", c.in, o, c.out)
	}
}

func TestFindWords(t *testing.T) {
	cases := []struct {
		name  string
		in    string
		want  []string
		panic bool
	}{
		{
			name: "double quoted",
			in:   `"2018-08-31 15:20:33"`,
			want: []string{`2018-08-31 15:20:33`},
		},
		{
			name: "single quoted",
			in:   `'2018-08-31 15:20:33'`,
			want: []string{`2018-08-31 15:20:33`},
		},
		{
			name:  "panic",
			in:    `2018-08-31 15:20:33`,
			panic: true,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			defer func() {
				v := recover()
				if v != nil {
					if !c.panic {
						t.Fatalf("panic: %s", v)
					}
				} else {
					if c.panic {
						t.Fatalf("want panic, but did not happen")
					}
				}
			}()
			got := FindWords([]byte(c.in), 0)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("want %#v, got %#v", c.want, got)
			}
		})
	}
}

func TestSplitCSV(t *testing.T) {
	cases := []struct {
		input string
		want  []string
	}{
		{
			input: "",
			want:  []string{},
		},
		{
			input: ",",
			want:  []string{"", ""},
		},
		{
			input: ",,",
			want:  []string{"", "", ""},
		},
		{
			input: ",\",\"",
			want:  []string{"", ","},
		},
		{
			input: ",\",\",,',,',",
			want:  []string{"", ",", "", ",,", ""},
		},
		{
			input: "53purt2e22zjn6efbmh4ph5fzugiainet5undyz4rqk3uy7n6esa,2021-08-31T00:00:00Z/2021-09-01T00:00:00Z,842274245,AWS,Anniversary,056683719894,2021-08-01T00:00:00Z,2021-09-01T00:00:00Z,056683719894,DiscountedUsage,2021-08-31T00:00:00Z,2021-09-01T00:00:00Z,AmazonEC2,EUC1-BoxUsage:c5.large,RunInstances,eu-central-1a,i-00ccff41477b8ce36,24.0000000000,4.0,96.0000000000,USD,0.0000000000,0.0000000000,0.0808333333,1.9399999992,\"USD 0.0 per Linux/UNIX (Amazon VPC), c5.large reserved instance applied\",,0.0000000000,0.0000000000,\"Amazon Web Services, Inc.\",Amazon Elastic Compute Cloud,,,,NA,,Used,false,3 GHz,,Yes,,Up to 2250 Mbps,,,10,,Yes,,,,,,,Compute optimized,c5.large,c5,Yes,Yes,Yes,No License required,EU (Frankfurt),AWS Region,,OnDemand,,,,,,4 GiB,,,,,Up to 10 Gigabit,4,Linux,RunInstances,Intel Xeon Platinum 8124M,,,NA,64-bit,Intel AVX; Intel AVX2; Intel AVX512; Intel Turbo,Compute Instance,,,eu-central-1,,,AmazonEC2,Amazon Elastic Compute Cloud,S3BME23KN52QCQ5Q,,,EBS only,,,Shared,,,,EUC1-BoxUsage:c5.large,2,,,,true,1yr,standard,All Upfront,S3BME23KN52QCQ5Q.6QCMYABX3D.6YS6EN2CT7,1663976301,USD,2.3280000000,0.0970000000,Reserved,Hrs,1.3698631200,,1.3698631200,,,1.2328766400,,1.2328766400,0.0000000000,,,,,,0.0000000000,arn:aws:ec2:eu-central-1:056683719894:reserved-instances/adc5ca41-fa0c-4824-bf3d-50cc01debbaf,,6598909739,,,,,,,,,,,,,,,,,,,,,,hwdatacenter,",
			want: []string{
				"53purt2e22zjn6efbmh4ph5fzugiainet5undyz4rqk3uy7n6esa", "2021-08-31T00:00:00Z/2021-09-01T00:00:00Z", "842274245", "AWS", "Anniversary", "056683719894", "2021-08-01T00:00:00Z", "2021-09-01T00:00:00Z", "056683719894", "DiscountedUsage", "2021-08-31T00:00:00Z", "2021-09-01T00:00:00Z", "AmazonEC2", "EUC1-BoxUsage:c5.large", "RunInstances", "eu-central-1a", "i-00ccff41477b8ce36", "24.0000000000", "4.0", "96.0000000000", "USD", "0.0000000000", "0.0000000000", "0.0808333333", "1.9399999992", "USD 0.0 per Linux/UNIX (Amazon VPC), c5.large reserved instance applied", "", "0.0000000000", "0.0000000000", "Amazon Web Services, Inc.", "Amazon Elastic Compute Cloud", "", "", "", "NA", "", "Used", "false", "3 GHz", "", "Yes", "", "Up to 2250 Mbps", "", "", "10", "", "Yes", "", "", "", "", "", "", "Compute optimized", "c5.large", "c5", "Yes", "Yes", "Yes", "No License required", "EU (Frankfurt)", "AWS Region", "", "OnDemand", "", "", "", "", "", "4 GiB", "", "", "", "", "Up to 10 Gigabit", "4", "Linux", "RunInstances", "Intel Xeon Platinum 8124M", "", "", "NA", "64-bit", "Intel AVX; Intel AVX2; Intel AVX512; Intel Turbo", "Compute Instance", "", "", "eu-central-1", "", "", "AmazonEC2", "Amazon Elastic Compute Cloud", "S3BME23KN52QCQ5Q", "", "", "EBS only", "", "", "Shared", "", "", "", "EUC1-BoxUsage:c5.large", "2", "", "", "", "true", "1yr", "standard", "All Upfront", "S3BME23KN52QCQ5Q.6QCMYABX3D.6YS6EN2CT7", "1663976301", "USD", "2.3280000000", "0.0970000000", "Reserved", "Hrs", "1.3698631200", "", "1.3698631200", "", "", "1.2328766400", "", "1.2328766400", "0.0000000000", "", "", "", "", "", "0.0000000000", "arn:aws:ec2:eu-central-1:056683719894:reserved-instances/adc5ca41-fa0c-4824-bf3d-50cc01debbaf", "", "6598909739", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "hwdatacenter", "",
			},
		},
	}
	for _, c := range cases {
		got := SplitCSV(c.input)
		if len(got) != len(c.want) {
			t.Errorf("input %s got: %d want: %d", c.input, len(got), len(c.want))
		} else {
			for i := range got {
				if got[i] != c.want[i] {
					t.Errorf("input %s got %s want %s at %d", c.input, got[i], c.want[i], i)
				}
			}
		}
	}
}
