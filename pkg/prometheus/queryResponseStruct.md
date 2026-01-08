## Query Response Struct Used in Client.go explanation

> Author : Ayush Poudel

A typical prometheus api (http) returns JSON, not go types. I needed a way to convert it to go type so I can use the values returned by /api/v1/query nicely.

It looks like this:

```json
{
  "status": "success",
  "data": {
    "resultType": "vector",
    "result": [
      {
        "metric": { "__name__": "up" },
        "value": [ 1712167200.123, "1" ]
      }
    ]
  }
}

```

Now, since go struct do not need to care about everything json returns, it can only care about what it needs, I created a struct which contains everything I need form the response.

- Top Level: data:{} mapped to Data struct{} `json:"data"`
- Inside Data struct: result mapped to R Result `json:"result"`. Each element inside result represents a single time-series returned by Prom.
- Inside each result entry: Json value is mapped to Value[]interfave{}. Prometheus returns index 0 as timestamp, index 1 as metric value (string). Since it contains mixed type interface is needed.

My final go struct looks like:

```go
type queryResponse struct {
	Data struct {
		Result []struct {
			Value []interface{} `json:"value"`
		} `json:"result"`
	} `json:"data`
}
```
