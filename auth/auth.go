package auth

import (
	"errors"
	"net/http"
	"strings"
)

func GetApiKey(headers http.Header) (string, error) {
	keyValue := headers.Get("Authorization")
	if keyValue == "" {
		return "", errors.New("We did not found your API key in the request")
	}
	vals := strings.Split(keyValue, " ")
	if len(vals) != 2 {
		return "", errors.New("Malformed auth header")
	}
	return vals[1], nil
}
