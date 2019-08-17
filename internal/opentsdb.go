package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	opentsdbRestApiSuffix string        = "/api/put?summary&details"
	defaultBulkSize       uint          = 50
	defaultThreadCount    uint          = 1
	defaultPushTimeout    time.Duration = time.Minute
	routierIdKey          string        = "routineId"
)

// Opentsdb is an Opentsdb connector
type Opentsdb struct {
	opentsdbURL string
	bulkSize    uint
	threadCount uint
	pushTimeout time.Duration
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
		threadCount: c.ThreadCount,
	}
	if c.BulkSize == 0 {
		logrus.Infof("Default bulk size will be used: %v", defaultBulkSize)
		o.bulkSize = defaultBulkSize
	}
	if c.ThreadCount == 0 {
		logrus.Infof("Default thread count will be used: %v", defaultThreadCount)
		o.threadCount = defaultThreadCount
	}
	var err error
	switch c.PushTimeout {
	case "":
		logrus.Infof("Default push timeout will be used: %v", defaultPushTimeout)
		o.pushTimeout = defaultPushTimeout
	default:
		if o.pushTimeout, err = time.ParseDuration(c.PushTimeout); err != nil {
			return o, fmt.Errorf("error while parsing push timeout duration (%v): %v", c.PushTimeout, err)
		}
	}
	return o, nil
}

// Push pushes metrics to Opentsdb
func (o Opentsdb) Push(ctx context.Context, m []OpentsdbMetric) error {
	logrus.SetLevel(logrus.DebugLevel)
	tasks := make(chan []OpentsdbMetric, o.threadCount)
	wg := sync.WaitGroup{}
	wg.Add(int(o.threadCount))
	errOccured := false

	// consumer
	for i := uint(0); i < o.threadCount; i++ {
		thId := i
		go func() {
			thIdLocal := thId
			subCtx := context.WithValue(ctx, routierIdKey, thIdLocal)
			defer wg.Done()
			for curTask := range tasks {
				logrus.Debugf("pusher %v, curTask: %v", thIdLocal, curTask)
				if err := o.doPush(subCtx, curTask); err != nil {
					errOccured = true
					logrus.Errorf("error while pushing to Opentsdb: %v", err)
				}
			}
		}()
	}

	// provider
	start, end := uint(0), uint(0)
	length := uint(len(m))
	for end < length {
		end += o.bulkSize
		if end > length {
			end = length
		}
		logrus.Debugf("new task, %v to %v, total: %v", start+1, end, length)
		tasks <- m[start:end]
		start = end
	}
	close(tasks)

	wg.Wait()
	logrus.Debugf("Push finished")

	if errOccured {
		return fmt.Errorf("Some errors occured while pushing to opentsdb")
	}
	return nil
}

func (o Opentsdb) doPush(ctx context.Context, m []OpentsdbMetric) error {
	data, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("pusher %v, error while marshaling data: %v", ctx.Value(routierIdKey), err)
	}
	req, err := http.NewRequest(http.MethodPost, o.opentsdbURL, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: o.pushTimeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("pusher %v, error while pushing data: %v", ctx.Value(routierIdKey), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		logrus.Debugf("pusher %v, pushed %v points with success", ctx.Value(routierIdKey), len(m))
		return nil
	}

	logrus.Warnf("pusher %v, opentsdb HTTP status: %v", ctx.Value(routierIdKey), resp.Status)
	respCont, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("pusher %v, error while reading response: %v", ctx.Value(routierIdKey), err)
	}
	fmt.Fprintf(os.Stderr, "%v", string(respCont))

	oResp := opentsbResponse{}
	err = json.Unmarshal(respCont, &oResp)
	if err != nil {
		return fmt.Errorf("pusher %v, error while parsing response: %v", ctx.Value(routierIdKey), err)
	}
	logrus.Warnf("pusher %v, Opentsdb rejected metrics: %v", ctx.Value(routierIdKey), oResp.Failed)
	return fmt.Errorf("pusher %v, some metrics have been rejected (%v)", ctx.Value(routierIdKey), oResp.Failed)
}
