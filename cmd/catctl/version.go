package main

import (
	"fmt"

	cat "github.com/enrichman/ccat-client-go"
	"github.com/spf13/cobra"
)

var Version = "0.0.0-dev"

func NewVersionCmd(catclient *cat.Client) *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show client and Server version",
		RunE: func(cmd *cobra.Command, args []string) error {
			version, err := catclient.Server.Version(cmd.Context())
			if err != nil {
				return err
			}

			fmt.Println("😸", version.Status)
			fmt.Println("Client Version:", Version)
			fmt.Println("Server Version:", version.Version)

			return nil
		},
	}

	return versionCmd
}
