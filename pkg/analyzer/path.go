package analyzer

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cmou/ripeatlas/pkg/atlas"
)

// ASNLookupCache caches ASN lookups to avoid repeated API calls
var ASNLookupCache = make(map[string]int)

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
							asn:         asn,
							occurrences: 0,
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

// lookupASN looks up the ASN for a given IP address using a WHOIS-like service
func lookupASN(ip string) (int, error) {
	// Check cache first
	if asn, exists := ASNLookupCache[ip]; exists {
		return asn, nil
	}

	// Use Team Cymru's IP to ASN lookup service
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	url := fmt.Sprintf("https://api.hackertarget.com/aslookup/?q=%s", ip)
	resp, err := client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to lookup ASN: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	// Parse response (format: "AS#### Organization Name")
	result := string(body)
	if strings.Contains(result, "error") || strings.Contains(result, "API count exceeded") {
		return 0, fmt.Errorf("ASN lookup failed")
	}

	var asn int
	fmt.Sscanf(result, "AS%d", &asn)

	// Cache the result
	if asn > 0 {
		ASNLookupCache[ip] = asn
	}

	return asn, nil
}

// lookupASNName looks up the name/organization for an ASN
func lookupASNName(asn int) (string, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	url := fmt.Sprintf("https://api.bgpview.io/asn/%d", asn)
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Sprintf("AS%d", asn), nil
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			Name        string `json:"name"`
			Description string `json:"description_short"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Sprintf("AS%d", asn), nil
	}

	if result.Data.Name != "" {
		return result.Data.Name, nil
	}

	if result.Data.Description != "" {
		return result.Data.Description, nil
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
