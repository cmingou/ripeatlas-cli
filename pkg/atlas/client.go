package atlas

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	BaseURL = "https://atlas.ripe.net/api/v2"
)

// Client is the RIPE Atlas API client
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new RIPE Atlas API client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetProbesByASN retrieves probes for given ASNs
func (c *Client) GetProbesByASN(asns []int) (map[int][]Probe, error) {
	// Build ASN query parameter
	asnParam := ""
	for i, asn := range asns {
		if i > 0 {
			asnParam += ","
		}
		asnParam += fmt.Sprintf("%d", asn)
	}

	url := fmt.Sprintf("%s/probes/?status=1&asn_v4__in=%s", BaseURL, asnParam)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var probeResp ProbeResponse
	if err := json.NewDecoder(resp.Body).Decode(&probeResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Group probes by ASN
	probesByASN := make(map[int][]Probe)
	for _, probe := range probeResp.Results {
		asn := probe.ASNV4
		if asn > 0 {
			probesByASN[asn] = append(probesByASN[asn], probe)
		}
	}

	return probesByASN, nil
}

// CreateMeasurement creates a new traceroute measurement
func (c *Client) CreateMeasurement(req MeasurementRequest) (int, error) {
	url := fmt.Sprintf("%s/measurements/", BaseURL)

	jsonData, err := json.Marshal(req)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Key %s", c.apiKey))

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return 0, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		return 0, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var msmResp MeasurementResponse
	if err := json.Unmarshal(body, &msmResp); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(msmResp.Measurements) == 0 {
		return 0, fmt.Errorf("no measurement ID returned")
	}

	return msmResp.Measurements[0], nil
}

// GetMeasurementStatus retrieves the status of a measurement
func (c *Client) GetMeasurementStatus(measurementID int) (*MeasurementStatus, error) {
	url := fmt.Sprintf("%s/measurements/%d/", BaseURL, measurementID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var status MeasurementStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &status, nil
}

// GetMeasurementResults retrieves the results of a measurement
func (c *Client) GetMeasurementResults(measurementID int) ([]TracerouteResult, error) {
	url := fmt.Sprintf("%s/measurements/%d/results/", BaseURL, measurementID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var results []TracerouteResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return results, nil
}

// WaitForMeasurement polls the measurement status until it completes or times out
// It uses a hybrid approach: checks both result completeness and status changes
func (c *Client) WaitForMeasurement(measurementID int, expectedProbes int, timeout time.Duration) error {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	timeoutCh := time.After(timeout)

	for {
		select {
		case <-ticker.C:
			status, err := c.GetMeasurementStatus(measurementID)
			if err != nil {
				return fmt.Errorf("failed to get measurement status: %w", err)
			}

			// Fast path: Check if all probes have reported results
			// This allows early completion without waiting for status to change from Ongoing to Stopped
			results, err := c.GetMeasurementResults(measurementID)
			if err == nil && len(results) == expectedProbes {
				// All probes have reported, measurement is effectively complete
				return nil
			}

			// Slow path: Check official status changes
			// Status ID: 0=Specified, 1=Scheduled, 2=Ongoing, 4=Stopped, 5=Forced to stop, 6=No suitable probes, 7=Failed, 8=Archived
			if status.Status.ID == 4 || status.Status.ID == 5 {
				return nil
			}

			if status.Status.ID == 6 {
				return fmt.Errorf("measurement failed: no suitable probes")
			}

			if status.Status.ID == 7 {
				return fmt.Errorf("measurement failed")
			}

		case <-timeoutCh:
			return fmt.Errorf("timeout waiting for measurement to complete")
		}
	}
}
