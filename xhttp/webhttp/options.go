package webhttp

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/charlesbases/hfw/codec"
	"github.com/charlesbases/hfw/codec/json"
	"github.com/charlesbases/hfw/codec/proto"
	"github.com/charlesbases/hfw/content"
	"github.com/charlesbases/hfw/xhttp/webcode"
	"github.com/charlesbases/logger"
)

const contentType = "content-type"

// defaultHeader .
var defaultHeader = map[string]string{contentType: content.Json.String()}

// defaultMarshaler json.DefaultMarshaler
var defaultMarshaler = json.DefaultMarshaler

// options .
type options struct {
	client *http.Client

	// params param for http get
	params []*param
	// header .
	header map[string]string
	// marshaler .
	marshaler codec.Marshaler
}

// param .
type param struct {
	key string
	val string
}

// format .
func (p *param) format() string {
	return strings.Join([]string{p.key, p.val}, "=")
}

type option func(o *options)

// defaultClient .
func defaultClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}

// newOptions .
func newOptions(opts ...option) *options {
	var options = &options{client: defaultClient(), header: defaultHeader, marshaler: defaultMarshaler}
	for _, o := range opts {
		o(options)
	}

	if options.header[contentType] == content.Proto.String() {
		options.marshaler = proto.DefaultMarshaler
	}

	return options
}

// Header .
func Header(key, val string) option {
	return func(o *options) {
		o.header[key] = val
	}
}

// Param .
func Param(key, val string) option {
	return func(o *options) {
		if len(o.params) == 0 {
			o.params = make([]*param, 0, 4)
		}
		o.params = append(o.params, &param{key: key, val: val})
	}
}

// ContentType default content.Json
func ContentType(ct content.Type) option {
	return func(o *options) {
		o.header[contentType] = ct.String()
	}
}

type data struct {
	opts *options
	data []byte
}

// Bytes .
func (d *data) Bytes() []byte {
	return d.data
}

// Unmarshal .
func (d *data) Unmarshal(pointer interface{}) error {
	return d.opts.marshaler.Unmarshal(d.data, pointer)
}

// do .
func (opts *options) do(req *http.Request) (*data, error) {
	// header
	for key, val := range opts.header {
		req.Header.Set(key, val)
	}

	// do
	rsp, err := opts.client.Do(req)
	if err != nil {
		logger.Errorf("%s | %s | client.Do() error: %v", req.Method, req.URL, err)
		return nil, err
	}

	switch rsp.StatusCode {
	case 200:
		defer rsp.Body.Close()

		if body, err := ioutil.ReadAll(rsp.Body); err != nil {
			logger.Errorf("%s | %s | ioutil.ReadAll() error: %v", req.Method, req.URL, err)
			return nil, err
		} else {
			var response = new(Response)
			if err := opts.marshaler.Unmarshal(body, response); err != nil {
				logger.Errorf("%s | %s | %s.Unmarshal() error: %v", req.Method, req.URL, opts.marshaler.Type(), err)
				return nil, err
			}
			// rsponse message
			if response.Code != webcode.StatusOK.Int32() {
				logger.Errorf(`%s | %s | {"code": %d, "message": "%s"}`, req.Method, req.URL, response.Code, response.Message)
				return nil, webcode.InternalErr
			}
			return &data{opts: opts, data: response.Data}, nil
		}
	default:
		logger.Errorf("%s | %s | failed: %s", req.Method, req.URL, rsp.Status)
		return nil, webcode.InternalErr
	}
}
