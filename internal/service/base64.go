package service

import "encoding/base64"

func base64Encode(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}

func base64Decode(s string) ([]byte, error) {
	if b, err := base64.RawURLEncoding.DecodeString(s); err == nil {
		return b, nil
	}
	return base64.StdEncoding.DecodeString(s)
}
