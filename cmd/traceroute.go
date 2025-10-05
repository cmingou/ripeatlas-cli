package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cmou/ripeatlas/pkg/analyzer"
	"github.com/cmou/ripeatlas/pkg/atlas"
	"github.com/cmou/ripeatlas/pkg/aws"
	"github.com/spf13/cobra"
)

var (
	asnsFlag      string
	targetFlag    string
	thresholdFlag float64
)

func init() {
	tracerouteCmd.Flags().StringVar(&asnsFlag, "asns", "", "Comma-separated list of ASNs (required)")
	tracerouteCmd.Flags().StringVar(&targetFlag, "target", "", "Target IP or AWS region (e.g., aws_us-west-2) (required)")
	tracerouteCmd.Flags().Float64Var(&thresholdFlag, "threshold", 0.8, "Threshold for common ASN (default: 0.8 = 80%)")

	tracerouteCmd.MarkFlagRequired("asns")
	tracerouteCmd.MarkFlagRequired("target")

	rootCmd.AddCommand(tracerouteCmd)
}

var tracerouteCmd = &cobra.Command{
	Use:   "traceroute",
	Short: "Run traceroute measurement and analyze common paths",
	Long: `Run ICMP traceroute measurements from specified ASNs to a target IP or AWS region.
Analyzes the results to find common ASN paths.

Example:
  ripeatlas traceroute --asns 5384,7713 --target 1.2.3.4
  ripeatlas traceroute --asns 5384,7713,9988 --target aws_us-west-2 --threshold 0.85`,
	RunE: runTraceroute,
}

func runTraceroute(cmd *cobra.Command, args []string) error {
	startTime := time.Now()

	// Parse ASNs
	asns, err := parseASNs(asnsFlag)
	if err != nil {
		return fmt.Errorf("invalid ASNs: %w", err)
	}

	fmt.Printf("🔍 Initializing RIPE Atlas traceroute measurement...\n\n")

	// Resolve target
	target := targetFlag
	if aws.IsAWSRegion(targetFlag) {
		fmt.Printf("📍 Resolving AWS region: %s\n", targetFlag)
		ip, err := aws.GetRegionIP(targetFlag)
		if err != nil {
			return fmt.Errorf("failed to resolve AWS region: %w", err)
		}
		fmt.Printf("   Selected IP: %s\n\n", ip)
		target = ip
	}

	// Create Atlas client
	client := atlas.NewClient(cfg.APIKey)

	// Get probes for ASNs
	fmt.Printf("🔎 Fetching probes for ASNs: %s\n", asnsFlag)
	probesByASN, err := client.GetProbesByASN(asns)
	if err != nil {
		return fmt.Errorf("failed to fetch probes: %w", err)
	}

	// Allocate probes
	allocations, asnsWithoutProbes, err := atlas.AllocateProbes(probesByASN, asns)
	if err != nil {
		return fmt.Errorf("probe allocation failed: %w", err)
	}

	// Display allocation summary
	var asnsWithProbes []int
	for _, alloc := range allocations {
		asnsWithProbes = append(asnsWithProbes, alloc.ASN)
	}

	fmt.Printf("   ASNs with probes: %v\n", asnsWithProbes)
	if len(asnsWithoutProbes) > 0 {
		fmt.Printf("   ⚠️  ASNs without probes: %v\n", asnsWithoutProbes)

		// Ask user if they want to continue
		fmt.Printf("\n❓ Some ASNs have no available probes. Continue? (y/n): ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "y" && response != "yes" {
			return fmt.Errorf("operation cancelled by user")
		}
	}
	fmt.Println()

	// Prepare probe list
	var probeIDs []int
	for _, alloc := range allocations {
		probeIDs = append(probeIDs, alloc.ProbeIDs...)
	}

	// Create measurement
	fmt.Printf("🚀 Creating traceroute measurement...\n")
	fmt.Printf("   Target: %s\n", target)
	fmt.Printf("   Probes: %d\n", len(probeIDs))

	measurementReq := atlas.MeasurementRequest{
		Definitions: []atlas.MeasurementDefinition{
			{
				Type:            "traceroute",
				AF:              4,
				Target:          target,
				Description:     fmt.Sprintf("Traceroute to %s from ASNs %s", targetFlag, asnsFlag),
				Protocol:        "ICMP",
				Packets:         3,
				Size:            48,
				MaxHops:         40,
				Paris:           16,
				ResponseTimeout: 4000,
			},
		},
		Probes: []atlas.ProbeSet{
			{
				Type:      "probes",
				Value:     strings.Trim(strings.Join(strings.Fields(fmt.Sprint(probeIDs)), ","), "[]"),
				Requested: len(probeIDs),
			},
		},
		IsOneoff: true,
	}

	measurementID, err := client.CreateMeasurement(measurementReq)
	if err != nil {
		return fmt.Errorf("failed to create measurement: %w", err)
	}

	fmt.Printf("   ✅ Measurement created: ID %d\n", measurementID)
	fmt.Printf("   🔗 https://atlas.ripe.net/measurements/%d\n\n", measurementID)

	// Wait for measurement to complete (with 5-minute timeout)
	fmt.Printf("⏳ Waiting for measurement to complete...\n")

	waitStartTime := time.Now()
	timeout := 5 * time.Minute

	for {
		elapsed := time.Since(waitStartTime)

		if elapsed >= timeout {
			fmt.Printf("\n⏱️  Measurement has been running for 5 minutes.\n")
			fmt.Printf("   Measurement URL: https://atlas.ripe.net/measurements/%d\n", measurementID)
			fmt.Printf("   Please check the URL manually.\n\n")
			fmt.Printf("❓ Wait for another 5 minutes? (y/n): ")

			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))

			if response != "y" && response != "yes" {
				return fmt.Errorf("measurement still running, check URL manually")
			}

			// Reset timeout
			waitStartTime = time.Now()
		}

		err := client.WaitForMeasurement(measurementID, 3*time.Second)
		if err == nil {
			break
		}

		// Check if it's a timeout error (expected during polling)
		if !strings.Contains(err.Error(), "timeout") {
			return fmt.Errorf("error waiting for measurement: %w", err)
		}

		time.Sleep(100 * time.Millisecond) // Small delay before checking user timeout
	}

	fmt.Printf("   ✅ Measurement completed!\n\n")

	// Fetch results
	fmt.Printf("📥 Fetching measurement results...\n")
	results, err := client.GetMeasurementResults(measurementID)
	if err != nil {
		return fmt.Errorf("failed to fetch results: %w", err)
	}

	fmt.Printf("   Retrieved %d traceroute results\n\n", len(results))

	// Analyze common ASNs
	fmt.Printf("🔬 Analyzing common ASN paths...\n\n")
	commonASNs, err := analyzer.AnalyzeCommonASNs(results, thresholdFlag)
	if err != nil {
		return fmt.Errorf("failed to analyze results: %w", err)
	}

	// Calculate path statistics
	uniquePaths, avgHops, maxHops, incompletePaths := analyzer.CalculatePathStats(results)

	// Generate report
	report := atlas.Report{
		MeasurementID:     measurementID,
		Target:            targetFlag,
		CreatedAt:         startTime,
		Duration:          time.Since(startTime),
		RequestedASNs:     asns,
		ASNsWithProbes:    asnsWithProbes,
		ASNsWithoutProbes: asnsWithoutProbes,
		Allocations:       allocations,
		CommonASNs:        commonASNs,
		Threshold:         thresholdFlag,
		TotalProbes:       len(probeIDs),
		UniquePaths:       uniquePaths,
		AvgHops:           avgHops,
		MaxHops:           maxHops,
		IncompletePaths:   incompletePaths,
	}

	// Display report
	fmt.Println(atlas.GenerateReport(report))

	return nil
}

// parseASNs parses a comma-separated list of ASNs
func parseASNs(s string) ([]int, error) {
	parts := strings.Split(s, ",")
	asns := make([]int, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		asn, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid ASN: %s", part)
		}

		if asn <= 0 {
			return nil, fmt.Errorf("ASN must be positive: %d", asn)
		}

		asns = append(asns, asn)
	}

	if len(asns) == 0 {
		return nil, fmt.Errorf("no valid ASNs provided")
	}

	return asns, nil
}
