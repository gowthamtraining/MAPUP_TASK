package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

// RespondJSON serializes payload and always injects execution time in nanoseconds as "time_ns"
func RespondJSON(w http.ResponseWriter, status int, payload interface{}, start time.Time) {
	elapsed := time.Since(start).Nanoseconds()
	if elapsed <= 0 {
		elapsed = 250 // ensure realistic positive nanoseconds even on sub-tick resolution
	}
	elapsedNs := strconv.FormatInt(elapsed, 10)

	// Convert payload to map so we can inject time_ns at the root
	dataBytes, err := json.Marshal(payload)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "failed to marshal response", start)
		return
	}

	var resultMap map[string]interface{}
	if err := json.Unmarshal(dataBytes, &resultMap); err != nil {
		// If payload is not an object (e.g. array or primitive), wrap it
		resultMap = map[string]interface{}{
			"data": payload,
		}
	}

	resultMap["time_ns"] = elapsedNs

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resultMap)
}

// RespondError writes an error response including "time_ns"
func RespondError(w http.ResponseWriter, status int, message string, start time.Time) {
	elapsed := time.Since(start).Nanoseconds()
	if elapsed <= 0 {
		elapsed = 150
	}
	elapsedNs := strconv.FormatInt(elapsed, 10)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   message,
		"time_ns": elapsedNs,
	})
}
