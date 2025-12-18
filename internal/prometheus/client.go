package prometheus

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type PromClient struct {
	baseUrl string
	http    *http.Client
}

// Usage: prometheus.New("http://localhost:9090")
func New(baseUrl string) *PromClient {
	return &PromClient{
		baseUrl: baseUrl,
		http: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

/*
	@ayuspoduel
	So far we have created struct to store a prometheus httpClient
	We have also created a constructor to create such client with 5
	seconds timeout. But now, we need to decode responses from Prom
	which returns in JSON. So we will create a struct to store respones.
	Since we care only about values.
	Explanation in ./queryResponseStruct.md
*/

type queryResponse struct {
	Data struct {
		Result []struct {
			Value []interface{} `json:"value"`
		} `json:"result"`
	} `json:"data`
}

/*
	@ayuspoudel
	Now we need actual query functions.
*/

func (c *PromClient) Query(ctx context.Context, query string) (float64, error) {
	/*
		@ayuspoudel
		So to prometheus instances we can query with
		https://localhost:9000/api/v1/query?query=<PROMQL_QUERY>. So we
		need to build the url. The lines below do that.
		Why not string concatenate?
		-----------------------------
		PromQL expressions often contain characters like: rate(http_requests_total[5m])
		If we did string concatenation this would break. So we use Encode() method
		so it becomes safely encoded as:
		http://localhost:9000/api/v1/query?query=rate%28http_requests_total%5B5m%5D%29
		Also q.Set() will always add ? first then the "string" = value.
	*/
	u, err := url.Parse(c.baseUrl + "/api/v1/query")
	if err != nil {
		return 0, err
	}
	q := u.Query()
	q.Set("query", query)
	u.RawQuery = q.Encode()

	/*
		Sending the request now
	*/

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return 0, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	/*
		We have recieved the response. Now we can parse the response.
		Great thing is we already have a struct ready to do that. :)
	*/
	var pr queryResponse
	err = json.NewDecoder(resp.Body).Decode(&pr)
	if err != nil {
		return 0, err
	}
	if len(pr.Data.Result) == 0 || len(pr.Data.Result[0].Value) < 2 {
		return 0, nil // no data
	}

	/*
		Prometheus returns numbers as strings, but we need float.
	*/
	valueStr, ok := pr.Data.Result[0].Value[1].(string)
	if !ok {
		return 0, nil // unexpected type
	}
	var value float64
	_, err = fmt.Sscan(valueStr, &value)
	if err != nil {
		return 0, err
	}
	return value, nil
}
