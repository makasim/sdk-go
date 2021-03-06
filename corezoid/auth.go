package corezoid

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type Auth interface {
	Sign(req *http.Request) error
}

type ApiKeyAuth struct {
	Login  int
	Secret string
}

type SATokenAuth struct {
	Token        string
	encodedToken string
}

func (a *ApiKeyAuth) Sign(req *http.Request) error {
	payload, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}

	timestamp := time.Now().Unix()

	var signature string
	if strings.Contains(req.Header.Get("Content-Type"), "multipart/form-data") {
		signature = a.genMultipartSignature(payload, timestamp)
	} else {
		signature = a.genSignature(payload, timestamp)
	}

	req.URL.Path = fmt.Sprintf(
		"%s/%d/%d/%s",
		req.URL.Path,
		a.Login,
		timestamp,
		signature,
	)

	req.Body = ioutil.NopCloser(bytes.NewReader(payload))

	return nil
}

func (a *SATokenAuth) Sign(req *http.Request) error {
	if a.encodedToken == "" {
		a.encodedToken = base64.StdEncoding.EncodeToString([]byte(a.Token))
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.encodedToken))

	return nil
}

func (a *ApiKeyAuth) genMultipartSignature(payload []byte, timestamp int64) string {
	payload = regexp.MustCompile(`-+[\d\w]+-{0,}\r\n`).ReplaceAll(payload, []byte(""))
	payload = regexp.MustCompile(`\r\n\r\n`).ReplaceAll(payload, []byte("\r\n"))
	payload = []byte(strings.TrimSpace(string(payload)))

	chunks := strings.Split(string(payload), "\r\n")

	reg := regexp.MustCompile(`^Content-`)

	result := ""
	for _, chunk := range chunks {
		result = result + chunk

		if reg.Match([]byte(chunk)) {
			result = result + "\r\n"
		}
	}

	return a.genSignature([]byte(result), timestamp)
}

func (a *ApiKeyAuth) genSignature(payload []byte, timestamp int64) string {
	sha := sha1.Sum([]byte(fmt.Sprintf("%d%s%s%s", timestamp, a.Secret, payload, a.Secret)))

	return hex.EncodeToString(sha[:])
}
