package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"ride-sharing/shared/contracts"
)

func handleTripPreview(w http.ResponseWriter, r *http.Request) {
	var reqBody previewTripRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "failed to parse json data", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	if reqBody.UserID == "" {
		http.Error(w, "user ID is required", http.StatusBadRequest)
		return
	}

	reqEncoded, err := json.Marshal(reqBody)
	if err != nil {
		http.Error(w, "failed to marshal request", http.StatusBadRequest)
		return
	}

	resp, err := http.Post("http://trip-service:8083/preview", "application/json", bytes.NewReader(reqEncoded))
	if err != nil {
		http.Error(w, "failed to post to trip service", http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	var respBody any
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		http.Error(w, "failed to parse JSON data from trip service", http.StatusBadRequest)
		return
	}

	response := contracts.APIResponse{Data: respBody}

	writeJSON(w, http.StatusCreated, response)
}
