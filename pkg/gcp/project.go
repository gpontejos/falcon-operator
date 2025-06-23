package gcp

import (
	"fmt"
	"io"
	"net/http"
)

// Get project-id of the GCP project within which this workload is running in
func GetProjectID() (string, error) {
	// curl -s "http://metadata.google.internal/computeMetadata/v1/project/project-id" -H "Metadata-Flavor: Google"
	req, err := http.NewRequest("GET", "http://metadata.google.internal/computeMetadata/v1/project/project-id", nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Metadata-Flavor", "Google")
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing pod logs: %v", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	return string(body), err
}
