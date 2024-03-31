package main

import (
	"fmt"

	cat "github.com/enrichman/ccat-client-go"
	"github.com/spf13/cobra"
)

var Version = "0.0.0-dev"

func NewVersionCmd(catclient *cat.Client) *cobra.Command {
	versionCmd := &cobra.Command{
		Use:           "version",
		Short:         "Show client and Server version",
		SilenceUsage:  true,
		SilenceErrors: true,
		Run: func(cmd *cobra.Command, args []string) {
			version, err := catclient.Server.Version(cmd.Context())
			if err != nil {
				// TODO add log for failure
				fmt.Println("Client Version:", Version)
				fmt.Println("Server Version: unknown")
				return
			}

			fmt.Println("😸", version.Status)
			fmt.Println("Client Version:", Version)
			fmt.Println("Server Version:", version.Version)
		},
	}

	return versionCmd
}
