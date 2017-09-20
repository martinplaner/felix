package felix

import (
	"bytes"
	"io"
	"io/ioutil"
	"testing"

	felix "github.com/martinplaner/felix2"
)

func TestLogger(t *testing.T) {
	b := &bytes.Buffer{}
	log := &DefaultLogger{
		out: b,
	}
	testLogger(t, log, b)
	log.Info("uneven args", "justakey")
}

// TestLogger tests the passed logger and expects the output to be written to the given reader.
func testLogger(t *testing.T, log felix.Logger, r io.Reader) {
	testCases := []struct {
		msg     string
		keyvals []interface{}
		f       func(string, ...interface{})
	}{
		{
			msg:     "debug level",
			f:       log.Debug,
			keyvals: []interface{}{"debugkey", "debugvalue"},
		},
		{
			msg:     "info level",
			f:       log.Info,
			keyvals: []interface{}{"infokey", "infovalue"},
		},
		{
			msg:     "warn level",
			f:       log.Warn,
			keyvals: []interface{}{"warnkey", "warnvalue"},
		},
		{
			msg:     "error level",
			f:       log.Error,
			keyvals: []interface{}{"errorkey", "errorvalue"},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.msg, func(t *testing.T) {
			tC.f(tC.msg, tC.keyvals...)

			b, err := ioutil.ReadAll(r)

			if err != nil {
				t.Error("error reading from logger output:", err)
			}

			if ok := bytes.Contains(b, []byte(tC.msg)); !ok {
				t.Errorf("msg '%s' not found in log output", tC.msg)
			}

			for _, kv := range tC.keyvals {
				if ok := bytes.Contains(b, []byte(kv.(string))); !ok {
					t.Errorf("key/val '%v' not found in log output", kv)
				}
			}
		})
	}
}
