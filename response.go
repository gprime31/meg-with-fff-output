package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

// a response is a wrapper around an HTTP response;
// it contains the request value for context.
type response struct {
	request request

	status     string
	statusCode int
	headers    []string
	body       []byte
	err        error
}

// String returns a string representation of the request and response
func (r response) String() string {
	b := &bytes.Buffer{}

	b.WriteString(r.request.URL())
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("> %s %s HTTP/1.1\n", r.request.method, r.request.path))

	// request headers
	for _, h := range r.request.headers {
		b.WriteString(fmt.Sprintf("> %s\n", h))
	}
	b.WriteString("\n")

	// status line
	b.WriteString(fmt.Sprintf("< HTTP/1.1 %s\n", r.status))

	// response headers
	for _, h := range r.headers {
		b.WriteString(fmt.Sprintf("< %s\n", h))
	}
	b.WriteString("\n")

	// body
	b.Write(r.body)

	return b.String()
}

func (r response) StringNoHeaders() string {
	b := &bytes.Buffer{}

	b.Write(r.body)

	return b.String()
}

func (r response) save(pathPrefix string, noHeaders bool) (string, error) {
    var headersContent, bodyContent []byte

    // Convert headers slice to a single byte slice
    if !noHeaders {
        for _, header := range r.headers {
            headersContent = append(headersContent, header...)
            headersContent = append(headersContent, '\n')
        }
    }

    // Body content is already in the correct format
    bodyContent = r.body

    // Generate checksum for uniqueness
    checksum := sha1.Sum(append(headersContent, bodyContent...))
    parts := []string{pathPrefix, r.request.Hostname(), fmt.Sprintf("%x", checksum)}
    basePath := path.Join(parts...)

    // Create directory if it doesn't exist
    if _, err := os.Stat(path.Dir(basePath)); os.IsNotExist(err) {
        if err := os.MkdirAll(path.Dir(basePath), 0750); err != nil {
            return basePath, err
        }
    }

    // Save headers to a separate file if they exist
    headersPath := basePath + ".headers"
    if !noHeaders {
        if err := ioutil.WriteFile(headersPath, headersContent, 0640); err != nil {
            return headersPath, err
        }
    }

    // Save body to its file
    bodyPath := basePath + ".body"
    if err := ioutil.WriteFile(bodyPath, bodyContent, 0640); err != nil {
        return bodyPath, err
    }

    return basePath, nil

}
