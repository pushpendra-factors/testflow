package util

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type RequestBuilder struct {
	methodType  string
	url         string
	headers     map[string]string
	postParams  interface{}
	cookies     []*http.Cookie
	queryParams map[string]string
	// timeout *int
}

func NewRequestBuilder(methodType, URL string) *RequestBuilder {
	return &RequestBuilder{
		methodType: methodType,
		url:        URL,
		headers:    make(map[string]string),
		cookies:    make([]*http.Cookie, 0, 0),
		postParams: nil,
	}
}

func (rb *RequestBuilder) WithHeader(key, value string) *RequestBuilder {
	rb.headers[key] = value
	return rb
}

func (rb *RequestBuilder) WithCookie(cookie *http.Cookie) *RequestBuilder {
	c := rb.cookies
	c = append(c, cookie)
	rb.cookies = c
	return rb
}

func (rb *RequestBuilder) WithPostParams(data interface{}) *RequestBuilder {
	rb.postParams = data
	return rb
}

func (rb *RequestBuilder) WithQueryParams(params map[string]string) *RequestBuilder {
	rb.queryParams = params
	return rb
}

func (rb *RequestBuilder) Build() (*http.Request, error) {

	// create post params is present
	var r io.Reader
	if rb.postParams != nil {
		jsonBytes, err := json.Marshal(rb.postParams)
		if err != nil {
			return nil, err
		}
		r = bytes.NewBuffer(jsonBytes)
	}

	// make request object
	req, err := http.NewRequest(rb.methodType, rb.url, r)
	if err != nil {
		return nil, err
	}

	// add headers
	for k, v := range rb.headers {
		req.Header.Set(k, v)
	}

	// add cookies
	for _, cookie := range rb.cookies {
		req.AddCookie(cookie)
	}

	// Add query params
	q := req.URL.Query()
	for k, v := range rb.queryParams {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	return req, nil
}
