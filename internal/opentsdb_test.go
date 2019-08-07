package internal

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"io"
)

func TestPush(t *testing.T) {
	var tcs = []struct {
		tcID     string
		inStatus int
		inResponse string
		expOK bool
	}{
		{"nominal", http.StatusOK, "{ \"failed\":0, \"success\":2 }", true},
		{"errorParsable", http.StatusBadRequest, "{ \"failed\":1, \"success\":0 }", false},
		{"errorUnparsable", http.StatusBadRequest, "{", false},
	}
	for _, tc := range tcs {
		t.Run(tc.tcID, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.inStatus)
				io.WriteString(w, tc.inResponse)
			}))
			defer ts.Close()
		
			c := ExporterConf{OpentsdbURL: ts.URL}
			o, err := NewOpentsdb(c)
			assert.Nil(t, err)
		
			m := []OpentsdbMetric{
				OpentsdbMetric{
					Metric: "blabla",
					Timestamp: 42,
					Value: 1.3,
					Tags: map[string]string {
						"k1":"v1",
						"k2":"v2",
					},
				},
			}
		
			err = o.Push(m)
			assert.Equal(t, tc.expOK, err==nil)
		})
	}
}

