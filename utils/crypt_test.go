package utils

import "testing"

func TestEncryptAESBase64(t *testing.T) {
	key := "123"
	lMsg := "This is the secret!"

	cypher, err := EncryptAESBase64(key, lMsg)
	if err != nil {
		t.Errorf("EncryptAESBase64 %s", err)
	} else {
		t.Logf("EncryptAESBase64 %s", cypher)
	}

	msg, err := DescryptAESBase64(key, cypher)
	if err != nil {
		t.Errorf("DescryptAESBase64 %s", err)
	} else {
		t.Logf("DescryptAESBase64 %s", msg)
	}

	if msg != lMsg {
		t.Errorf("%s != %s", msg, lMsg)
	}
}

func TestDescryptAESBase64(t *testing.T) {
	msg, err := DescryptAESBase64("123", "zMuWP5HwnC+zqNayjZGSouZCHA==")
	if err != nil {
		t.Errorf("DescryptAESBase64 %s", err)
	} else {
		t.Logf("DescryptAESBase64 %s", msg)
	}
}

func TestEncryptAESBase64Url(t *testing.T) {
	key := "mysecret"
	lMsg := "This is a long long cypher text msg!!!"

	cypher, err := EncryptAESBase64Url(key, lMsg)
	if err != nil {
		t.Errorf("EncryptAESBase64Url %s", err)
	} else {
		t.Logf("EncryptAESBase64Url: %s", cypher)
	}

	msg, err := DescryptAESBase64Url(key, cypher)
	if err != nil {
		t.Errorf("DescryptAESBase64Url %s", err)
	} else {
		t.Logf("DescryptAESBase64Url %s", msg)
	}

	if msg != lMsg {
		t.Errorf("%s != %s", msg, lMsg)
	}
}
