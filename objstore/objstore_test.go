package objstore

import (
	"flag"
	"os"
	"testing"

	"time"

	"github.com/go-redis/redis"
)

var rdurl string

func init() {
	flag.StringVar(&rdurl, "RD_URL", "", "redis url for testing")
	flag.Parse()

	if rdurl == "" {
		rdurl = os.Getenv("RD_URL")
	}
}

func TestService_Save(t *testing.T) {
	tests := []struct {
		key     string
		value   string
		timeout int
		err     error
	}{
		{"test", `"{hello: "world"}`, 0, nil},
		{"test2", `"{hello: "world"}`, 2, redis.Nil},
	}

	s, err := InitService(rdurl)

	if err != nil {
		t.Fatalf("Failed to create objstore service: %v", err)
	}

	for _, test := range tests {
		err := s.Save(test.key, []byte(test.value), test.timeout)

		if err != nil {
			t.Errorf("%v - Failed to save value: %v", test.key, err)
		}

		time.Sleep(3 * time.Second)

		v := s.rd.Get(test.key)

		if v.Err() != test.err {
			t.Errorf("%v - Errors dont match: Expected %v, got %v", test.key, test.err, v.Err())
		}

		// If we are expecting a real value back then continue testing
		if test.err != redis.Nil {
			value, err := v.Result()

			if err != nil {
				t.Errorf("%v - Failed to retrieve key: %v", test.key, err)
			}

			if value != test.value {
				t.Errorf("%v - Values differ from saved: Expected: %v, got %v", test.key, test.value, value)
			}
		}
	}
}
