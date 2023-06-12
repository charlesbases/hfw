package aws

import (
	"fmt"
	"math"
	"os"
	"testing"
	"time"

	"github.com/charlesbases/hfw/storage"
	"github.com/charlesbases/logger"
)

var (
	bucket = "mxdata"

	endpoint  = "s3.bcebos.cncq.icpc.changan.com"
	accessKey = "437e8bdc81b14da796789da67667dd52"
	secretKey = "9eb1e112d8a144a8ab125020cf6e7403"
)

type Message struct {
	Date string
	Pi   float64
}

// now .
func now() string {
	return time.Now().Format("2006-01-02T15:04:05")
}

func Test(t *testing.T) {
	cli := NewClient(endpoint, accessKey, secretKey, storage.Timeout(3))

	start := time.Now()
	total := 1000000
	for i := 0; i < total; i++ {
		cli.PutObject(bucket, fmt.Sprintf("testdata/data/a/%d", i), storage.String(now()))
	}
	logger.Info(time.Since(start))
}

func TestAWS(t *testing.T) {
	NewClient(endpoint, accessKey, secretKey, storage.Timeout(3))
}

func TestPut(t *testing.T) {
	cli := NewClient(endpoint, accessKey, secretKey, storage.Timeout(3))

	// Put data
	{
		if err := cli.PutObject(bucket, "testdata/data/int", storage.Number(time.Now().UnixMilli())); err != nil {
			logger.Fatal(err)
		}
		if err := cli.PutObject(bucket, "testdata/data/float", storage.Number(math.Pi)); err != nil {
			logger.Fatal(err)
		}
		if err := cli.PutObject(bucket, "testdata/data/string", storage.String(now())); err != nil {
			logger.Fatal(err)
		}
		if err := cli.PutObject(bucket, "testdata/data/boolean", storage.Boolean(true)); err != nil {
			logger.Fatal(err)
		}
	}

	// Put message
	if err := cli.PutObject(bucket, "testdata/data/mess", storage.MarshalJson(&Message{Pi: math.Pi, Date: now()})); err != nil {
		logger.Fatal(err)
	}

	// Put file
	if err := cli.PutObject(bucket, "testdata/data/file", storage.File("s3.go")); err != nil {
		logger.Fatal(err)
	}

	// Put io.ReadSeeker
	file, _ := os.Open("s3_test.go")
	defer file.Close()
	stat, _ := file.Stat()
	if err := cli.PutObject(bucket, "testdata/data/file", storage.ReadSeeker(file, stat.Size())); err != nil {
		logger.Fatal(err)
	}
}

func TestGet(t *testing.T) {
	cli := NewClient(endpoint, accessKey, secretKey, storage.Timeout(3))

	{
		key := "testdata/data/int"
		if output, err := cli.GetObject(bucket, key); err != nil {
			logger.Fatal(err)
		} else {
			var obj int
			if err := output.Decoding(&obj); err != nil {
				logger.Fatal(err)
			} else {
				logger.Debugf("%s >> %v", key, obj)
			}
		}
	}
	{
		key := "testdata/data/float"
		if output, err := cli.GetObject(bucket, key); err != nil {
			logger.Fatal(err)
		} else {
			var obj float64
			if err := output.Decoding(&obj); err != nil {
				logger.Fatal(err)
			} else {
				logger.Debugf("%s >> %v", key, obj)
			}
		}
	}
	{
		key := "testdata/data/string"
		if output, err := cli.GetObject(bucket, key); err != nil {
			logger.Fatal(err)
		} else {
			var obj string
			if err := output.Decoding(&obj); err != nil {
				logger.Fatal(err)
			} else {
				logger.Debugf("%s >> %v", key, obj)
			}
		}
	}
	{
		key := "testdata/data/boolean"
		if output, err := cli.GetObject(bucket, key); err != nil {
			logger.Fatal(err)
		} else {
			var obj bool
			if err := output.Decoding(&obj); err != nil {
				logger.Fatal(err)
			} else {
				logger.Debugf("%s >> %v", key, obj)
			}
		}
	}
	{
		key := "testdata/data/mess"
		if output, err := cli.GetObject(bucket, key); err != nil {
			logger.Fatal(err)
		} else {
			var obj = new(Message)
			if err := output.Decoding(&obj); err != nil {
				logger.Fatal(err)
			} else {
				logger.Debugf("%s >> %v", key, obj)
			}
		}
	}
}

func TestDelKey(t *testing.T) {
	cli := NewClient(endpoint, accessKey, secretKey, storage.Timeout(3))

	key := "testdata/data/"
	if err := cli.DelObject(bucket, key); err != nil {
		logger.Fatal(err)
	}
}

func TestDelPrefix(t *testing.T) {
	cli := NewClient(endpoint, accessKey, secretKey, storage.Timeout(3))

	key := "testdata/auth"
	if err := cli.DelPrefix(bucket, key); err != nil {
		logger.Fatal(err)
	}

	time.Sleep(time.Minute)
}

func TestList(t *testing.T) {
	cli := NewClient(endpoint, accessKey, secretKey, storage.Timeout(3))

	key := "testdata/data/"
	objs, err := cli.ListObjects(bucket, key, storage.ListMaxKeys(1001) /*, storage.ListDisableRecursive()*/)
	if err != nil {
		logger.Fatal(err)
	}
	keys := objs.Keys()
	logger.Debugf("==> %d", len(keys))
	for _, key := range keys {
		logger.Debug(*key)
	}
}

func TestIsExist(t *testing.T) {
	cli := NewClient(endpoint, accessKey, secretKey, storage.Timeout(3))

	key := "testdata/auth/"
	isExist, err := cli.IsExist(bucket, key)
	if err != nil {
		logger.Fatal(err)
	}
	fmt.Println(isExist)
}

func TestPresign(t *testing.T) {
	cli := NewClient(endpoint, accessKey, secretKey, storage.Timeout(3))

	key := "testdata/data/string"
	url, err := cli.Presign(bucket, key)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Debug(url)
}

func TestListCompress(t *testing.T) {
	cli := NewClient(endpoint, accessKey, secretKey, storage.Timeout(3))

	key := "testdata/data/"
	objs, err := cli.ListObjects(bucket, key)
	if err != nil {
		logger.Fatal(err)
	}

	f, _ := os.OpenFile("testdata.tar.gz", os.O_CREATE|os.O_RDWR, 0755)
	defer f.Close()
	if err := objs.Compress(f); err != nil {
		logger.Fatal()
	}
}
