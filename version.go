package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version set at build time
const (
	Version = "0.0.1"
)

//NewVersionCmd command for emitting version info.
func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version info",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version %s\n", Version)
		},
	}
}
