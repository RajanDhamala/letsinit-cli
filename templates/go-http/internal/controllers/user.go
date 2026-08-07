package controller

import (
	"encoding/json"
	"net/http"
)

func (c *Controller) LoginHandler(w http.ResponseWriter, r *http.Request) {
	// login logic
	w.Header().Set("Content-Type", "application/json")

	reponse := map[string]string{
		"message": "server is up and running",
	}
	json.NewEncoder(w).Encode(reponse)
}
