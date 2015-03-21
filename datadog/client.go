package datadog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

var APIURL = "https://app.datadoghq.com"

type Series []Metric

type Metric struct {
	Name   string   `json:"metric"`
	Points []Point  `json:"points"`
	Host   string   `json:"host"`
	Tags   []string `json:"tags"`
}

type Point struct {
	Timestamp time.Time
	Value     float32
}

func (p Point) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`[%d, %f]`, p.Timestamp.Unix(), p.Value)), nil
}

func (p *Point) UnmarshalJSON(data []byte) error {
	var tuple []float64

	if err := json.Unmarshal(data, &tuple); err != nil {
		return err
	}

	p.Timestamp = time.Unix(int64(tuple[0]), 0)
	p.Value = float32(tuple[1])

	return nil
}

type request struct {
	Series Series `json:"series"`
}

type client struct {
	apiKey string
	client *http.Client
}

func NewClient(apiKey string) *client {
	return &client{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

func (c *client) PublishSeries(series Series) error {
	request := request{
		Series: series,
	}

	buffer := &bytes.Buffer{}
	if err := json.NewEncoder(buffer).Encode(request); err != nil {
		return fmt.Errorf("encoding request: %s", err)
	}

	req, err := http.NewRequest("POST", APIURL+"/api/v1/series", buffer)
	if err != nil {
		return fmt.Errorf("building request: %s", err)
	}

	auth := url.Values{}
	auth.Set("api_key", c.apiKey)
	req.URL.RawQuery = auth.Encode()

	req.Header.Set("Content-type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("response: %s", err)
	}

	return resp.Body.Close()
}
