package analyzer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cmingou/ripeatlas-cli/pkg/atlas"
)

// ASNLookupCache caches ASN lookups to avoid repeated API calls
var ASNLookupCache = make(map[string]int)

// ASNNameCache caches ASN names to avoid repeated API calls
var ASNNameCache = make(map[int]string)

// Semaphore for RIPEstat API rate limiting (max 8 concurrent requests)
var ripestatSemaphore = make(chan struct{}, 8)

// HTTP/2 client for RIPEstat API
var ripestatClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		ForceAttemptHTTP2: true, // Enable HTTP/2
	},
}

// AnalyzeCommonASNs analyzes traceroute results to find common ASNs
func AnalyzeCommonASNs(results []atlas.TracerouteResult, threshold float64) ([]atlas.ASNInfo, error) {
	totalProbes := len(results)
	if totalProbes == 0 {
		return nil, fmt.Errorf("no results to analyze")
	}

	// Track ASN occurrences and hop positions
	asnStats := make(map[int]*asnTracker)

	for _, result := range results {
		// Extract ASN path from this traceroute
		seenASNs := make(map[int]bool) // Track ASNs seen in this path to avoid double counting

		for _, hop := range result.Result {
			for _, reply := range hop.Result {
				if reply.From == "" || reply.X == "*" {
					continue
				}

				// Look up ASN for this IP
				asn, err := lookupASN(reply.From)
				if err != nil || asn == 0 {
					continue
				}

				if !seenASNs[asn] {
					if _, exists := asnStats[asn]; !exists {
						asnStats[asn] = &asnTracker{
							asn:          asn,
							occurrences:  0,
							hopPositions: make([]int, 0),
						}
					}

					asnStats[asn].occurrences++
					asnStats[asn].hopPositions = append(asnStats[asn].hopPositions, hop.Hop)
					seenASNs[asn] = true
				}
			}
		}
	}

	// Filter ASNs by threshold and prepare results
	minOccurrences := int(float64(totalProbes) * threshold)
	var commonASNs []atlas.ASNInfo

	for asn, stats := range asnStats {
		if stats.occurrences >= minOccurrences {
			percentage := float64(stats.occurrences) / float64(totalProbes) * 100

			// Calculate average hop position
			avgHop := 0
			if len(stats.hopPositions) > 0 {
				sum := 0
				minHop := stats.hopPositions[0]
				maxHop := stats.hopPositions[0]

				for _, hop := range stats.hopPositions {
					sum += hop
					if hop < minHop {
						minHop = hop
					}
					if hop > maxHop {
						maxHop = hop
					}
				}
				avgHop = sum / len(stats.hopPositions)
			}

			asnName, _ := lookupASNName(asn)

			commonASNs = append(commonASNs, atlas.ASNInfo{
				ASN:         asn,
				Name:        asnName,
				Occurrences: stats.occurrences,
				Percentage:  percentage,
				AvgHopStart: avgHop,
				AvgHopEnd:   avgHop + 2, // Approximate range
			})
		}
	}

	// Sort by percentage (descending)
	sortASNsByPercentage(commonASNs)

	return commonASNs, nil
}

// asnTracker tracks ASN statistics across traceroutes
type asnTracker struct {
	asn          int
	occurrences  int
	hopPositions []int
}

// RIPEstatResponse represents the response from RIPEstat API
type RIPEstatResponse struct {
	Data struct {
		ASNs []struct {
			ASN    int    `json:"asn"`
			Holder string `json:"holder"`
		} `json:"asns"`
	} `json:"data"`
}

// lookupASN looks up the ASN for a given IP address using RIPEstat API
func lookupASN(ip string) (int, error) {
	// Check cache first
	if asn, exists := ASNLookupCache[ip]; exists {
		return asn, nil
	}

	// Acquire semaphore to respect RIPEstat API rate limit (max 8 concurrent)
	ripestatSemaphore <- struct{}{}
	defer func() { <-ripestatSemaphore }()

	url := fmt.Sprintf("https://stat.ripe.net/data/prefix-overview/data.json?resource=%s", ip)
	resp, err := ripestatClient.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to lookup ASN: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("RIPEstat API returned status %d", resp.StatusCode)
	}

	var result RIPEstatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode RIPEstat response: %w", err)
	}

	// Extract ASN from response
	if len(result.Data.ASNs) == 0 {
		return 0, fmt.Errorf("no ASN found for IP %s", ip)
	}

	asn := result.Data.ASNs[0].ASN
	holder := result.Data.ASNs[0].Holder

	// Cache both ASN and name
	if asn > 0 {
		ASNLookupCache[ip] = asn
		if holder != "" {
			ASNNameCache[asn] = holder
		}
	}

	return asn, nil
}

// lookupASNName looks up the name/organization for an ASN
func lookupASNName(asn int) (string, error) {
	// Check cache first (may have been populated by lookupASN)
	if name, exists := ASNNameCache[asn]; exists {
		return name, nil
	}

	// If not in cache, query RIPEstat API
	// We use a dummy IP query with ASN notation
	ripestatSemaphore <- struct{}{}
	defer func() { <-ripestatSemaphore }()

	url := fmt.Sprintf("https://stat.ripe.net/data/as-overview/data.json?resource=AS%d", asn)
	resp, err := ripestatClient.Get(url)
	if err != nil {
		return fmt.Sprintf("AS%d", asn), nil
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			Holder string `json:"holder"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Sprintf("AS%d", asn), nil
	}

	if result.Data.Holder != "" {
		ASNNameCache[asn] = result.Data.Holder
		return result.Data.Holder, nil
	}

	return fmt.Sprintf("AS%d", asn), nil
}

// sortASNsByPercentage sorts ASN info by percentage in descending order
func sortASNsByPercentage(asns []atlas.ASNInfo) {
	// Simple bubble sort (good enough for small lists)
	n := len(asns)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if asns[j].Percentage < asns[j+1].Percentage {
				asns[j], asns[j+1] = asns[j+1], asns[j]
			}
		}
	}
}

// CalculatePathStats calculates statistics about the traceroute paths
func CalculatePathStats(results []atlas.TracerouteResult) (int, float64, int, int) {
	if len(results) == 0 {
		return 0, 0, 0, 0
	}

	uniquePaths := make(map[string]bool)
	totalHops := 0
	maxHops := 0
	incompletePaths := 0

	for _, result := range results {
		// Create a path signature
		pathSig := ""
		hopCount := 0

		for _, hop := range result.Result {
			hasReply := false
			for _, reply := range hop.Result {
				if reply.From != "" && reply.X != "*" {
					pathSig += reply.From + ","
					hasReply = true
					break
				}
			}

			if hasReply {
				hopCount++
			}
		}

		uniquePaths[pathSig] = true
		totalHops += hopCount

		if hopCount > maxHops {
			maxHops = hopCount
		}

		// Check if path is incomplete (ends with timeout)
		if len(result.Result) > 0 {
			lastHop := result.Result[len(result.Result)-1]
			allTimeout := true
			for _, reply := range lastHop.Result {
				if reply.From != "" && reply.X != "*" {
					allTimeout = false
					break
				}
			}
			if allTimeout {
				incompletePaths++
			}
		}
	}

	avgHops := float64(totalHops) / float64(len(results))

	return len(uniquePaths), avgHops, maxHops, incompletePaths
}
