package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"text/tabwriter"

	mapset "github.com/deckarep/golang-set/v2"
	cat "github.com/enrichman/ccat-client-go"
	"github.com/spf13/cobra"
)

func NewRootCmd() (*cobra.Command, error) {
	catclient, err := cat.NewClient()
	if err != nil {
		return nil, err
	}

	rootCmd := &cobra.Command{
		Use: "catctl",
		Run: func(cmd *cobra.Command, args []string) {
			// Do Stuff Here
		},
	}

	rootCmd.AddCommand(
		NewChatCmd(catclient),
		NewSettingsCmd(catclient),
	)

	return rootCmd, nil
}

func NewSettingsCmd(catclient *cat.Client) *cobra.Command {
	settingsCmd := &cobra.Command{
		Use:   "settings",
		Short: "manage settings",
		Run:   func(cmd *cobra.Command, args []string) {},
	}

	settingsCmd.AddCommand(
		NewSettingsGetCmd(catclient),
		NewSettingsCreateCmd(catclient),
	)

	return settingsCmd
}

func NewSettingsGetCmd(catclient *cat.Client) *cobra.Command {
	settingsCmd := &cobra.Command{
		Use:   "get",
		Short: "get settings",
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			existingArgs := mapset.NewSet(args...)

			settings, _ := catclient.Settings.Get(context.Background(), cat.SettingsGetOpts{})

			allSettings := mapset.NewSet(args...)
			for _, setting := range settings {
				allSettings.Add(setting.ID)
			}

			validArgs := allSettings.
				Difference(existingArgs).
				ToSlice()

			return validArgs, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			setting, err := catclient.Settings.GetByID(context.Background(), args[0])
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME\tCATEGORY")
			//for _, setting := range settings {
			fmt.Fprintf(w, "%s\t%s\t%s\n", setting.ID, setting.Name, setting.Category)
			//}
			w.Flush()

			return nil
		},
	}

	return settingsCmd
}

func NewSettingsCreateCmd(catclient *cat.Client) *cobra.Command {
	type createCfg struct {
		name     string
		category string
	}

	cfg := &createCfg{}

	settingsCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "create settings",
		RunE: func(cmd *cobra.Command, args []string) error {

			m := map[string]any{}
			err := json.Unmarshal([]byte(args[0]), &m)
			if err != nil {
				return err
			}

			req := cat.SettingCreateRequest{
				Name:     cfg.name,
				Category: cfg.category,
				Value:    m,
			}

			settings, err := catclient.Settings.Create(cmd.Context(), req)

			fmt.Println("created", settings, err)
			return err
		},
	}

	settingsCreateCmd.Flags().StringVar(&cfg.name, "name", "", "name of the setting")

	return settingsCreateCmd
}

func NewChatCmd(catclient *cat.Client) *cobra.Command {
	chatCmd := &cobra.Command{
		Use: "chat",
		Run: func(cmd *cobra.Command, args []string) {
			in, out := make(chan string), make(chan string)

			go func() {
				fmt.Println("Say hi!")
				for {
					reader := bufio.NewReader(os.Stdin)
					line, _ := reader.ReadString('\n')
					line = strings.TrimSpace(line)
					in <- line
				}
			}()

			go func() {
				for {
					fmt.Println(<-out)
				}
			}()

			interrupt := make(chan os.Signal, 1)
			signal.Notify(interrupt, os.Interrupt)

			ctx, cancel := context.WithCancel(context.Background())
			go func() {
				<-interrupt
				log.Println("interrupt. Canceling")
				cancel()
			}()

			err := catclient.Chat.Chat(ctx, in, out)
			if err != nil {
				log.Println(err)
			}
		},
	}

	return chatCmd
}

func main() {
	rootCmd, err := NewRootCmd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
