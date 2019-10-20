package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const (
	envKeyAwsProfile       = "AWS_PROFILE"
	envKeyAwsSDKLoadConfig = "AWS_SDK_LOAD_CONFIG"
	appName                = "prometheus-aws-tags-exporter"
)

func setSDKLoadConfig() error {
	if _, isset := os.LookupEnv(envKeyAwsProfile); !isset {
		return nil
	}
	if _, isset := os.LookupEnv(envKeyAwsSDKLoadConfig); !isset {
		if err := os.Setenv(envKeyAwsSDKLoadConfig, "1"); err != nil {
			return err
		}
	}
	return nil
}

func setupCmd(cmds []*cobra.Command) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   appName,
		Short: "A Prometheus exporter for AWS Tags",
	}

	// silence usage on Error
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		rootCmd.SilenceUsage = true
		return nil
	}

	if err := setSDKLoadConfig(); err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	for _, c := range cmds {
		rootCmd.AddCommand(c)
	}

	return rootCmd
}
func main() {
	cmds := []*cobra.Command{
		NewExporterCmd(&ExporterCmd{}),
		NewVersionCmd(),
	}
	rootCmd := setupCmd(cmds)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}

	//This section will start the HTTP server and expose
	//any metrics on the /metrics endpoint.

}
