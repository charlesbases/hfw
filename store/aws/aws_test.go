package aws

import (
	"os"
	"testing"

	"github.com/charlesbases/hfw/store"
	"github.com/charlesbases/logger"
)

var (
	endpoint  = "s3.bcebos.cncq.icpc.changan.com"
	accessKey = "437e8bdc81b14da796789da67667dd52"
	secretKey = "9eb1e112d8a144a8ab125020cf6e7403"
)

func TestAWS(t *testing.T) {
	NewClient(endpoint, accessKey, secretKey, store.Timeout(3))
}

type TestDataMessage struct {
	A int
	B bool
	C float64
}

func TestPut(t *testing.T) {
	cli := NewClient(endpoint, accessKey, secretKey, store.Timeout(3))

	// Put file
	if err := cli.Put("testdata/data/aws.go", File("./aws.go"), store.PutBucket("mxdata")); err != nil {
		logger.Error(err)
	}

	// Put reader
	file, _ := os.Open("./object.go")
	stat, _ := file.Stat()
	if err := cli.Put("testdata/data/object.go", ReadSeeker(file, stat.Size()), store.PutBucket("mxdata")); err != nil {
		logger.Fatal(err)
	}

	// Put data
	if err := cli.Put("testdata/data/int", Number(1), store.PutBucket("mxdata")); err != nil {
		logger.Fatal(err)
	}
	if err := cli.Put("testdata/data/float", Number(0.1), store.PutBucket("mxdata")); err != nil {
		logger.Fatal(err)
	}
	if err := cli.Put("testdata/data/boolean", Boolean(true), store.PutBucket("mxdata")); err != nil {
		logger.Fatal(err)
	}
	if err := cli.Put("testdata/data/string", String("a"), store.PutBucket("mxdata")); err != nil {
		logger.Fatal(err)
	}
	if err := cli.Put("testdata/data/message", MarshalJson(&TestDataMessage{A: 1, B: true, C: 0.1}), store.PutBucket("mxdata")); err != nil {
		logger.Fatal(err)
	}
}

func TestGet(t *testing.T) {
	cli := NewClient(endpoint, accessKey, secretKey, store.Timeout(3))

	// get int
	{
		object, err := cli.Get("testdata/data/int", store.GetBucket("mxdata"))
		if err != nil {
			logger.Fatal(err)
		}
		var obj int
		if err := object.Decoding(&obj); err != nil {
			logger.Fatal(err)
		} else {
			logger.Debug(obj)
		}
	}

	// get float64
	{
		object, err := cli.Get("testdata/data/float", store.GetBucket("mxdata"))
		if err != nil {
			logger.Fatal(err)
		}
		var obj float64
		if err := object.Decoding(&obj); err != nil {
			logger.Fatal(err)
		} else {
			logger.Debug(obj)
		}
	}

	// get bool
	{
		object, err := cli.Get("testdata/data/boolean", store.GetBucket("mxdata"))
		if err != nil {
			logger.Fatal(err)
		}
		var obj bool
		if err := object.Decoding(&obj); err != nil {
			logger.Fatal(err)
		} else {
			logger.Debug(obj)
		}
	}

	// get string
	{
		object, err := cli.Get("testdata/data/string", store.GetBucket("mxdata"))
		if err != nil {
			logger.Fatal(err)
		}
		var obj string
		if err := object.Decoding(&obj); err != nil {
			logger.Fatal(err)
		} else {
			logger.Debug(obj)
		}
	}

	// get message
	{
		object, err := cli.Get("testdata/data/message", store.GetBucket("mxdata"))
		if err != nil {
			logger.Fatal(err)
		}
		var obj TestDataMessage
		if err := object.Decoding(&obj); err != nil {
			logger.Fatal(err)
		} else {
			logger.Debug(obj)
		}

	}
}

func TestDel(t *testing.T) {
	cli := NewClient(endpoint, accessKey, secretKey, store.Timeout(3))

	if err := cli.Del("testdata/data/", store.DelBucket("mxdata")); err != nil {
		logger.Fatal(err)
	}
}

func TestList(t *testing.T) {
	cli := NewClient(endpoint, accessKey, secretKey, store.Timeout(3))

	objs, err := cli.List("testdata", store.ListBucket("mxdata"), store.ListLimit(6))
	if err != nil {
		logger.Fatal(err)
	}
	for _, key := range objs.Keys() {
		logger.Debug(key)
	}
}
