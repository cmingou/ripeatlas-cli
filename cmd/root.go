package cmd

import (
	"fmt"
	"os"

	"github.com/cmou/ripeatlas/internal/config"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	cfg     *config.Config
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "ripeatlas",
	Short: "RIPE Atlas traceroute analysis tool",
	Long: `A CLI tool to analyze network paths using RIPE Atlas probes.

This tool helps you:
- Select probes from multiple ASNs
- Run traceroute measurements to target IPs or AWS regions
- Analyze common ASN paths across multiple traceroutes`,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "env.key", "config file (default is ./env.key)")
}

func initConfig() {
	var err error
	cfg, err = config.Load(cfgFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}
}

// GetConfig returns the loaded configuration
func GetConfig() *config.Config {
	return cfg
}
