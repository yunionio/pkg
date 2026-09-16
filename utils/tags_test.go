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
		name string
		in   string
		want []string
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
			// A colon is one of the separators, so a bare timestamp splits
			// further; quoting it keeps it whole, as above.
			name: "space separated",
			in:   `2018-08-31 15:20:33`,
			want: []string{"2018-08-31", "15", "20", "33"},
		},
		{
			name: "comma separated",
			in:   `a,b`,
			want: []string{"a", "b"},
		},
		{
			name: "mixed separators",
			in:   `a, b:c	d`,
			want: []string{"a", "b", "c", "d"},
		},
		{
			name: "addresses",
			in:   `10.0.0.1 10.0.0.2`,
			want: []string{"10.0.0.1", "10.0.0.2"},
		},
		{
			name: "closing bracket terminates",
			in:   `a]b`,
			want: []string{"a", "b"},
		},
		{
			name: "empty",
			in:   ``,
			want: []string{},
		},
		{
			name: "trailing separator",
			in:   `a,b,`,
			want: []string{"a", "b"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := FindWords([]byte(c.in), 0)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("FindWords(%q) = %#v, want %#v", c.in, got, c.want)
			}
		})
	}
}

func TestSplitWords(t *testing.T) {
	cases := []struct {
		input string
		want  []string
	}{
		{
			input: `'file:///opt/test test.txt'   '/data/test test.txt'`,
			want:  []string{"file:///opt/test test.txt", "/data/test test.txt"},
		},
		{
			input: `'file:///opt/test test.txt'   /data/testtest.txt`,
			want:  []string{"file:///opt/test test.txt", "/data/testtest.txt"},
		},
	}
	for _, c := range cases {
		got, err := SplitWords(c.input)
		if err != nil {
			t.Errorf("input %s got error %s", c.input, err)
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("input %s got %#v want %#v", c.input, got, c.want)
		}
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

func TestTagMapMalformed(t *testing.T) {
	cases := []struct {
		Tag  string
		Want map[string]string
	}{
		{
			Tag:  `json:"name"`,
			Want: map[string]string{"json": "name"},
		},
		{
			Tag:  `json:"name" nullable:"false"`,
			Want: map[string]string{"json": "name", "nullable": "false"},
		},
		{
			// a fragment without a colon takes no value, the scan carries on
			// with the fragments that follow it
			Tag:  `json:"name" charset="ascii" nullable:"false"`,
			Want: map[string]string{"json": "name", `charset="ascii"`: "", "nullable": "false"},
		},
		{
			Tag:  `json:"name" charset="ascii"`,
			Want: map[string]string{"json": "name", `charset="ascii"`: ""},
		},
		{
			Tag:  ``,
			Want: map[string]string{},
		},
	}
	for _, c := range cases {
		got := TagMap(reflect.StructTag(c.Tag))
		if !reflect.DeepEqual(got, c.Want) {
			t.Errorf("TagMap(%q) got %v want %v", c.Tag, got, c.Want)
		}
	}
}
