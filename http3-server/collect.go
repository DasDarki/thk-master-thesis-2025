package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

var httpClient = &http.Client{}

func collectMetrics(runID int, data map[string]any) {
	url := fmt.Sprintf("https://thkm25_collect.nauri.io/%d/update", runID)
	j, err := json.Marshal(data)
	if err != nil {
		fmt.Println("[COLLECTOR] Error marshalling data:", err)
		return
	}

	req, err := http.NewRequest("PUT", url, bytes.NewReader(j))
	if err != nil {
		fmt.Println("[COLLECTOR] Error creating request:", err)
		return
	}

	req.Header.Set("X-API-KEY", "thk_masterthesis_2025_hwtwswrtc")
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Println("[COLLECTOR] Error sending request:", err)
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		fmt.Println("[COLLECTOR] Error response from server:", resp.Status)
		return
	}

	fmt.Println("[COLLECTOR] Metrics collected successfully")
}
