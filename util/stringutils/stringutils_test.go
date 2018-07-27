package stringutils

import "testing"

func TestStringUtils(t *testing.T) {
	t.Log(ParseNamePattern("test-###"))
	t.Logf("%s", UUID4())
	t.Logf("%s", Interface2String(nil))
	t.Logf("%s", Interface2String(2))
	t.Logf("%s", Interface2String("test string"))
	type TestStruct struct {
		Name   string
		Age    int
		Gender string
	}
	t.Logf("%s", Interface2String(TestStruct{Name: "micheal", Age: 24, Gender: "Male"}))
}
