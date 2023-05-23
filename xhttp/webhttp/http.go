package webhttp

import (
	"bytes"
	"net/http"
	"strings"

	"github.com/charlesbases/logger"
)

// Get .
func Get(url string, options ...option) (*data, error) {
	opts := newOptions(options...)

	if len(opts.params) != 0 {
		var params = make([]string, 0, len(opts.params))
		for _, param := range opts.params {
			params = append(params, param.format())
		}
		url = url + "?" + strings.Join(params, "&")
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		logger.Errorf("Get | %s | http.NewRequest() error: %v", url, err)
		return nil, err
	}

	return opts.do(req)
}

// Post .
func Post(url string, vPointer interface{}, options ...option) (*data, error) {
	var opts = newOptions(options...)

	data, err := opts.marshaler.Marshal(vPointer)
	if err != nil {
		logger.Errorf("Post | %s | %s.Marshal() error: %v", url, opts.marshaler.Type(), err)
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		logger.Errorf("Post | %s | http.NewRequest() error: %v", url, err)
		return nil, err
	}

	return opts.do(req)
}

// Delete .
func Delete(url string, options ...option) (*data, error) {
	var opts = newOptions(options...)

	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		logger.Errorf("Delete | %s | http.NewRequest() error: %v", url, err)
		return nil, err
	}

	return opts.do(req)
}
