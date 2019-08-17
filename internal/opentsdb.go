package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
)

const opentsdbRestApiSuffix string = "/api/put?summary&details"
const defaultBulkSize uint = 50

// Opentsdb is an Opentsdb connector
type Opentsdb struct {
	opentsdbURL string
	bulkSize    uint
}

type opentsbResponse struct {
	Failed  uint `json:"failed"`
	Success uint `json:"success"`
}

// NewOpentsdb instanciates an Opentsdb connector
func NewOpentsdb(c ExporterConf) (Opentsdb, error) {
	o := Opentsdb{
		opentsdbURL: c.OpentsdbURL + opentsdbRestApiSuffix,
		bulkSize:    c.BulkSize,
	}
	if c.BulkSize == 0 {
		logrus.Infof("Default bulksize will be used: %v", defaultBulkSize)
		o.bulkSize = defaultBulkSize
	}
	return o, nil
}

// Push pushes metrics to Opentsdb
func (o Opentsdb) Push(ctx context.Context, m []OpentsdbMetric) error {
	start, end := uint(0), uint(0)
	length := uint(len(m))
	errOccured := false
	for end < length {
		end += o.bulkSize
		if end > length {
			end = length
		}
		logrus.Debugf("Pushing values %v to %v (total: %v)", start+1, end, length)
		if err := o.doPush(ctx, m[start:end]); err != nil {
			errOccured = true
			logrus.Errorf("error while pushing to Opentsdb: %v", err)
		}
		start = end
	}
	if errOccured {
		return fmt.Errorf("Some errors occured while pushing to opentsdb")
	}
	return nil
}

func (o Opentsdb) doPush(ctx context.Context, m []OpentsdbMetric) error {
	data, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("error while marshaling data: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, o.opentsdbURL, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error while pushing data: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return nil
	}

	logrus.Warnf("Opentsdb HTTP status: %v", resp.Status)
	respCont, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error while reading response: %v", err)
	}
	fmt.Fprintf(os.Stderr, "%v", string(respCont))

	oResp := opentsbResponse{}
	err = json.Unmarshal(respCont, &oResp)
	if err != nil {
		return fmt.Errorf("error while parsing response: %v", err)
	}
	logrus.Warnf("Opentsdb rejected metrics: %v", oResp.Failed)
	return fmt.Errorf("Some metrics have been rejected (%v)", oResp.Failed)
}
