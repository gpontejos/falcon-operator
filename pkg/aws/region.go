package aws

import (
	"fmt"
	"io"
	"net/http"
)

func closeAndCheckError(c io.Closer, err *error) {
	if cerr := c.Close(); cerr != nil && *err == nil {
		*err = fmt.Errorf("error closing: %v", cerr)
	}
}

func AvailabilityZone() (string, error) {
	// curl -s http://169.254.169.254/latest/meta-data/placement/availability-zone
	req, err := http.NewRequest("GET", "http://169.254.169.254/latest/meta-data/placement/availability-zone", nil)
	if err != nil {
		return "", err
	}
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return "", err
	}
	defer closeAndCheckError(resp.Body, &err)
	body, err := io.ReadAll(resp.Body)
	return string(body), err
}

func Region() (string, error) {
	// curl -s http://169.254.169.254/latest/meta-data/placement/region
	req, err := http.NewRequest("GET", "http://169.254.169.254/latest/meta-data/placement/region", nil)
	if err != nil {
		return "", err
	}
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return "", err
	}
	defer closeAndCheckError(resp.Body, &err)
	body, err := io.ReadAll(resp.Body)
	return string(body), err
}
