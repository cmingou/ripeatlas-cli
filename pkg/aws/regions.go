package aws

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

const (
	IPRangesURL = "https://ip-ranges.amazonaws.com/ip-ranges.json"
)

// IPRanges represents the AWS IP ranges JSON structure
type IPRanges struct {
	SyncToken  string   `json:"syncToken"`
	CreateDate string   `json:"createDate"`
	Prefixes   []Prefix `json:"prefixes"`
}

// Prefix represents an IP prefix in AWS
type Prefix struct {
	IPPrefix           string `json:"ip_prefix"`
	Region             string `json:"region"`
	Service            string `json:"service"`
	NetworkBorderGroup string `json:"network_border_group"`
}

// GetRegionIP fetches a random IP from the specified AWS region
func GetRegionIP(region string) (string, error) {
	// Parse region from format like "aws_us-west-2"
	if strings.HasPrefix(region, "aws_") {
		region = strings.TrimPrefix(region, "aws_")
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(IPRangesURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch AWS IP ranges: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch AWS IP ranges, status: %d", resp.StatusCode)
	}

	var ipRanges IPRanges
	if err := json.NewDecoder(resp.Body).Decode(&ipRanges); err != nil {
		return "", fmt.Errorf("failed to parse AWS IP ranges: %w", err)
	}

	// Filter prefixes for the specified region and EC2 service
	var matchingPrefixes []string
	for _, prefix := range ipRanges.Prefixes {
		if prefix.Region == region && (prefix.Service == "EC2" || prefix.Service == "AMAZON") {
			matchingPrefixes = append(matchingPrefixes, prefix.IPPrefix)
		}
	}

	if len(matchingPrefixes) == 0 {
		return "", fmt.Errorf("no IP ranges found for region: %s", region)
	}

	// Select a random prefix
	randomPrefix := matchingPrefixes[rand.Intn(len(matchingPrefixes))]

	// Extract the first IP from the CIDR notation
	// For simplicity, we'll just return the network address (first IP)
	// In a production environment, you might want to select a random IP within the range
	ip := strings.Split(randomPrefix, "/")[0]

	return ip, nil
}

// IsAWSRegion checks if the target is an AWS region identifier
func IsAWSRegion(target string) bool {
	return strings.HasPrefix(target, "aws_")
}

// ListAWSRegions returns a list of available AWS regions from the IP ranges
func ListAWSRegions() ([]string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(IPRangesURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch AWS IP ranges: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch AWS IP ranges, status: %d", resp.StatusCode)
	}

	var ipRanges IPRanges
	if err := json.NewDecoder(resp.Body).Decode(&ipRanges); err != nil {
		return nil, fmt.Errorf("failed to parse AWS IP ranges: %w", err)
	}

	// Collect unique regions
	regionMap := make(map[string]bool)
	for _, prefix := range ipRanges.Prefixes {
		if prefix.Region != "" && (prefix.Service == "EC2" || prefix.Service == "AMAZON") {
			regionMap[prefix.Region] = true
		}
	}

	regions := make([]string, 0, len(regionMap))
	for region := range regionMap {
		regions = append(regions, region)
	}

	return regions, nil
}
