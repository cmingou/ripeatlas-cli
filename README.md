# RIPE Atlas Traceroute Analysis Tool

A CLI tool to analyze network paths using RIPE Atlas probes, helping you find common ASN paths across multiple traceroutes.

## Features

- 🔍 Automatically selects and allocates probes from multiple ASNs
- 🌐 Supports both direct IP targets and AWS regions
- 📊 Analyzes traceroute results to identify common ASN paths
- 📝 Generates detailed, human-readable reports
- ⚖️ Intelligent probe distribution with greedy allocation
- ⏱️ Interactive timeout handling for long-running measurements

## Installation

### Prerequisites

- Go 1.24.2 or higher
- RIPE Atlas API key

### Build

```bash
git clone <repository-url>
cd ripeatlas
go build -o ripeatlas
```

## Configuration

Create an `env.key` file in the project directory:

```
RIPE_ATLAS_KEY="your-api-key-here"
```

## Usage

### Basic Command

```bash
./ripeatlas traceroute --asns <ASN_LIST> --target <TARGET>
```

### Examples

#### Traceroute to a specific IP

```bash
./ripeatlas traceroute --asns 5384,7713 --target 1.2.3.4
```

#### Traceroute to an AWS region

```bash
./ripeatlas traceroute --asns 5384,7713,9988 --target aws_us-west-2
```

#### Custom threshold for common ASN detection

```bash
./ripeatlas traceroute --asns 5384,7713 --target 8.8.8.8 --threshold 0.85
```

This will identify ASNs that appear in at least 85% of the traceroute paths.

### Available Flags

- `--asns`: Comma-separated list of ASNs (required)
- `--target`: Target IP address or AWS region (e.g., `aws_us-west-2`) (required)
- `--threshold`: Percentage threshold for common ASN detection (default: 0.8 = 80%)
- `--config`: Path to configuration file (default: `env.key`)

## How It Works

1. **Probe Discovery**: Queries RIPE Atlas API for available probes in specified ASNs
2. **Probe Allocation**: Distributes up to 1000 probes across ASNs using greedy allocation
3. **Measurement Creation**: Creates a one-off ICMP traceroute measurement
4. **Monitoring**: Polls measurement status every 3 seconds with 5-minute timeout windows
5. **Result Analysis**: Analyzes traceroute results to identify common ASN paths
6. **Report Generation**: Produces a detailed report with visualizations

## Probe Allocation Strategy

The tool uses an intelligent greedy allocation algorithm:

1. Calculates initial quota: `1000 / number_of_ASNs_with_probes`
2. First pass: Allocates probes up to quota for each ASN
3. Second pass: Redistributes remaining slots to ASNs with available probes
4. Randomly selects probes without replacement

This ensures optimal probe distribution even when ASNs have varying numbers of available probes.

## Common ASN Detection

The tool identifies ASNs that appear frequently across all traceroute paths:

- Default threshold: 80% (appears in at least 80% of paths)
- Configurable via `--threshold` flag
- Excludes duplicate ASN counting per path
- Calculates average hop position for each common ASN

## Output Report

The tool generates a comprehensive report including:

- ✅ Measurement information and URL
- 📊 Probe distribution across ASNs
- 🔍 Common ASN analysis with frequencies
- 📈 Path diversity statistics
- ⏱️ Execution time and duration

Example output:

```
╔══════════════════════════════════════════════════════════════╗
║        RIPE Atlas Traceroute Analysis Report                ║
╚══════════════════════════════════════════════════════════════╝

Measurement Information:
  • Measurement ID: 12345678
  • Target: aws_us-west-2
  • Status: Completed ✓
  • Created: 2025-10-05 14:30:15 UTC
  • Duration: 142 seconds
  • View online: https://atlas.ripe.net/measurements/12345678

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Probe Distribution:
  ...
```

## RIPE Atlas Quotas

Please be aware of RIPE Atlas quotas:

- Up to 100 simultaneous measurements
- Up to 1000 probes per measurement
- Up to 100,000 results per day
- Up to 50 measurement results per second per measurement
- Up to 1,000,000 credits per day

## Troubleshooting

### No probes found for ASN

Some ASNs may not have any active RIPE Atlas probes. The tool will:
1. Identify ASNs without probes
2. Ask if you want to continue with remaining ASNs
3. Redistribute probe quotas accordingly

### Measurement timeout

If a measurement takes longer than 5 minutes:
1. The tool will prompt you to continue waiting
2. Provides the measurement URL for manual checking
3. You can choose to wait another 5 minutes or exit

### API errors

- Verify your API key in `env.key`
- Check RIPE Atlas service status
- Ensure you haven't exceeded quotas

## AWS Region Support

The tool automatically resolves AWS regions to IP addresses using the official AWS IP ranges:

- Format: `aws_<region-code>` (e.g., `aws_us-west-2`)
- Fetches from: `https://ip-ranges.amazonaws.com/ip-ranges.json`
- Randomly selects an EC2 IP from the region
- Supports all AWS regions

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## Acknowledgments

- RIPE NCC for the Atlas platform
- spf13/cobra for the CLI framework
- jedib0t/go-pretty for table rendering
