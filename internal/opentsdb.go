package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
)

const opentsdbRestApiSuffix string = "/api/put?summary&details"

// Opentsdb is an Opentsdb connector
type Opentsdb struct {
	opentsdbURL string
}

type opentsbResponse struct {
	Failed  uint `json:"failed"`
	Success uint `json:"success"`
}

// NewOpentsdb instanciates an Opentsdb connector
func NewOpentsdb(c ExporterConf) (Opentsdb, error) {
	return Opentsdb{opentsdbURL: c.OpentsdbURL + opentsdbRestApiSuffix}, nil
}

// Push pushes metrics to Opentsdb
func (o Opentsdb) Push(m []OpentsdbMetric) error {
	// TODO : stream
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
