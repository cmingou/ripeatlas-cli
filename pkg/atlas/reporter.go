package atlas

import (
	"fmt"
	"strings"
	"time"
)

const (
	BoxTop    = "╔══════════════════════════════════════════════════════════════╗"
	BoxBottom = "╚══════════════════════════════════════════════════════════════╝"
	Separator = "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
)

// Report represents a complete analysis report
type Report struct {
	MeasurementID     int
	Target            string
	CreatedAt         time.Time
	Duration          time.Duration
	RequestedASNs     []int
	ASNsWithProbes    []int
	ASNsWithoutProbes []int
	Allocations       []ProbeAllocation
	CommonASNs        []ASNInfo
	Threshold         float64
	TotalProbes       int
	UniquePaths       int
	AvgHops           float64
	MaxHops           int
	IncompletePaths   int
}

// GenerateReport creates a formatted text report
func GenerateReport(report Report) string {
	var sb strings.Builder

	// Header
	sb.WriteString(BoxTop + "\n")
	sb.WriteString(centerText("RIPE Atlas Traceroute Analysis Report", 62) + "\n")
	sb.WriteString(BoxBottom + "\n\n")

	// Measurement Information
	sb.WriteString("Measurement Information:\n")
	sb.WriteString(fmt.Sprintf("  • Measurement ID: %d\n", report.MeasurementID))
	sb.WriteString(fmt.Sprintf("  • Target: %s\n", report.Target))
	sb.WriteString(fmt.Sprintf("  • Status: Completed ✓\n"))
	sb.WriteString(fmt.Sprintf("  • Created: %s\n", report.CreatedAt.Format("2006-01-02 15:04:05 MST")))
	sb.WriteString(fmt.Sprintf("  • Duration: %s\n", formatDuration(report.Duration)))
	sb.WriteString(fmt.Sprintf("  • View online: https://atlas.ripe.net/measurements/%d\n\n", report.MeasurementID))

	sb.WriteString(Separator + "\n\n")

	// Probe Distribution
	sb.WriteString("Probe Distribution:\n")
	sb.WriteString(fmt.Sprintf("  Requested ASNs: %d\n", len(report.RequestedASNs)))

	if len(report.ASNsWithProbes) > 0 {
		asnList := formatIntList(report.ASNsWithProbes)
		sb.WriteString(fmt.Sprintf("    ✓ %s (%d ASNs with probes)\n", asnList, len(report.ASNsWithProbes)))
	}

	if len(report.ASNsWithoutProbes) > 0 {
		asnList := formatIntList(report.ASNsWithoutProbes)
		sb.WriteString(fmt.Sprintf("    ✗ %s (%d ASNs without probes)\n", asnList, len(report.ASNsWithoutProbes)))
	}

	sb.WriteString("\n  Probe Allocation:\n")
	for _, alloc := range report.Allocations {
		percentage := float64(alloc.Allocated) / float64(report.TotalProbes) * 100
		bar := createProgressBar(percentage, 20)
		sb.WriteString(fmt.Sprintf("    AS%-6d %s %4d probes (%5.1f%%)\n",
			alloc.ASN, bar, alloc.Allocated, percentage))
	}

	sb.WriteString(fmt.Sprintf("    %s\n", strings.Repeat("─", 45)))
	sb.WriteString(fmt.Sprintf("    Total:  %32d probes\n\n", report.TotalProbes))

	sb.WriteString(Separator + "\n\n")

	// Common Path Analysis
	sb.WriteString(fmt.Sprintf("Common Path Analysis (Threshold: %.1f%% = %d/%d probes):\n\n",
		report.Threshold*100,
		int(report.Threshold*float64(report.TotalProbes)),
		report.TotalProbes))

	if len(report.CommonASNs) == 0 {
		sb.WriteString("  No common ASNs found meeting the threshold.\n\n")
	} else {
		for i, asn := range report.CommonASNs {
			sb.WriteString(fmt.Sprintf("  %d. AS%d - %s\n", i+1, asn.ASN, asn.Name))
			sb.WriteString(fmt.Sprintf("     Frequency: %.1f%% (%d/%d probes)\n",
				asn.Percentage, asn.Occurrences, report.TotalProbes))
			sb.WriteString(fmt.Sprintf("     Average position: Hop %d-%d\n\n",
				asn.AvgHopStart, asn.AvgHopEnd))
		}
	}

	sb.WriteString(Separator + "\n\n")

	// Path Diversity Summary
	sb.WriteString("Path Diversity Summary:\n")
	sb.WriteString(fmt.Sprintf("  • Unique paths: %d\n", report.UniquePaths))
	sb.WriteString(fmt.Sprintf("  • Average hops: %.1f\n", report.AvgHops))
	sb.WriteString(fmt.Sprintf("  • Max hops reached: %d\n", report.MaxHops))
	sb.WriteString(fmt.Sprintf("  • Incomplete paths: %d (%.1f%%)\n\n",
		report.IncompletePaths,
		float64(report.IncompletePaths)/float64(report.TotalProbes)*100))

	sb.WriteString(Separator + "\n")

	return sb.String()
}

// centerText centers text within a given width
func centerText(text string, width int) string {
	if len(text) >= width {
		return text
	}

	padding := (width - len(text)) / 2
	return fmt.Sprintf("║%s%s%s║",
		strings.Repeat(" ", padding),
		text,
		strings.Repeat(" ", width-padding-len(text)))
}

// createProgressBar creates a Unicode progress bar
func createProgressBar(percentage float64, width int) string {
	filled := int(percentage / 100.0 * float64(width))

	if filled > width {
		filled = width
	}

	return strings.Repeat("█", filled) + strings.Repeat(" ", width-filled)
}

// formatIntList formats a list of integers as a comma-separated string
func formatIntList(nums []int) string {
	strs := make([]string, len(nums))
	for i, n := range nums {
		strs[i] = fmt.Sprintf("AS%d", n)
	}
	return strings.Join(strs, ", ")
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0f seconds", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.1f minutes", d.Minutes())
	}
	return fmt.Sprintf("%.1f hours", d.Hours())
}
