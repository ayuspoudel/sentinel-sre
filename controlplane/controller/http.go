package controller

import (
	"encoding/json"
	"net/http"
)

func (c *Controller) StatusHandler(w http.ResponseWriter, r *http.Request) {
	decision := c.LatestDecision()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(decision)

}
