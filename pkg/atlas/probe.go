package atlas

import (
	"fmt"
	"math/rand"
)

const MaxProbesPerMeasurement = 1000

// AllocateProbes allocates probes from multiple ASNs
// Returns allocation details and any ASNs without probes
func AllocateProbes(probesByASN map[int][]Probe, requestedASNs []int) ([]ProbeAllocation, []int, error) {
	if len(requestedASNs) == 0 {
		return nil, nil, fmt.Errorf("no ASNs provided")
	}

	// Identify ASNs with and without probes
	var asnsWithProbes []int
	var asnsWithoutProbes []int

	for _, asn := range requestedASNs {
		probes, exists := probesByASN[asn]
		if !exists || len(probes) == 0 {
			asnsWithoutProbes = append(asnsWithoutProbes, asn)
		} else {
			asnsWithProbes = append(asnsWithProbes, asn)
		}
	}

	if len(asnsWithProbes) == 0 {
		return nil, asnsWithoutProbes, fmt.Errorf("no ASNs have available probes")
	}

	// Calculate initial quota per ASN
	quotaPerASN := MaxProbesPerMeasurement / len(asnsWithProbes)

	// First pass: allocate based on quota
	allocations := make([]ProbeAllocation, 0, len(asnsWithProbes))
	totalAllocated := 0

	for _, asn := range asnsWithProbes {
		probes := probesByASN[asn]
		available := len(probes)
		allocated := min(available, quotaPerASN)

		allocation := ProbeAllocation{
			ASN:       asn,
			Available: available,
			Allocated: allocated,
			ProbeIDs:  selectRandomProbes(probes, allocated),
		}

		allocations = append(allocations, allocation)
		totalAllocated += allocated
	}

	// Second pass: redistribute remaining quota (greedy allocation)
	remaining := MaxProbesPerMeasurement - totalAllocated

	for remaining > 0 {
		distributed := false

		for i := range allocations {
			if allocations[i].Allocated < allocations[i].Available {
				// This ASN has more probes available
				additionalProbes := min(allocations[i].Available-allocations[i].Allocated, remaining)

				// Get additional random probes
				allProbes := probesByASN[allocations[i].ASN]
				allocations[i].Allocated += additionalProbes
				allocations[i].ProbeIDs = selectRandomProbes(allProbes, allocations[i].Allocated)

				remaining -= additionalProbes
				distributed = true

				if remaining == 0 {
					break
				}
			}
		}

		// If we couldn't distribute any more, break to avoid infinite loop
		if !distributed {
			break
		}
	}

	return allocations, asnsWithoutProbes, nil
}

// selectRandomProbes randomly selects n probes from the list without replacement
func selectRandomProbes(probes []Probe, n int) []int {
	if n >= len(probes) {
		// Return all probe IDs
		ids := make([]int, len(probes))
		for i, p := range probes {
			ids[i] = p.ID
		}
		return ids
	}

	// Create a copy of indices
	indices := make([]int, len(probes))
	for i := range indices {
		indices[i] = i
	}

	// Shuffle using Fisher-Yates algorithm
	for i := len(indices) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		indices[i], indices[j] = indices[j], indices[i]
	}

	// Take first n probe IDs
	result := make([]int, n)
	for i := 0; i < n; i++ {
		result[i] = probes[indices[i]].ID
	}

	return result
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetTotalProbeCount returns the total number of allocated probes
func GetTotalProbeCount(allocations []ProbeAllocation) int {
	total := 0
	for _, alloc := range allocations {
		total += alloc.Allocated
	}
	return total
}
