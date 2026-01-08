package prometheus

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestQuery_ReturnsValue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
		{
			"status": "success",
			"data": {
				"result": [
					{
						"value": [1234567890, "42.5"]
					}
				]
			}
		}
		`))
	}))

	defer server.Close()
	client := New(server.URL)
	value, err := client.Query(context.Background(), "sum(rate(http_requests_total[5m]))")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if value != 42.5 {
		t.Fatalf("expected value 42.5, got %v", value)
	}
}
