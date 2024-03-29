package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	mapset "github.com/deckarep/golang-set/v2"
	cat "github.com/enrichman/ccat-client-go"
	"github.com/spf13/cobra"
)

func NewSettingsCmd(catclient *cat.Client) *cobra.Command {
	settingsCmd := &cobra.Command{
		Use:   "settings",
		Short: "manage settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	settingsCmd.AddCommand(
		NewSettingsGetCmd(catclient),
		NewSettingsCreateCmd(catclient),
	)

	return settingsCmd
}

func NewSettingsGetCmd(catclient *cat.Client) *cobra.Command {
	type getCfg struct {
		search string
	}

	cfg := &getCfg{}

	settingsGetCmd := &cobra.Command{
		Use:   "get",
		Short: "get settings",
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			existingArgs := mapset.NewSet(args...)

			settings, _ := catclient.Settings.Get(context.Background(), cat.SettingsGetOpts{Search: cfg.search})

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
			filteredSettings := []*cat.Setting{}

			if cfg.search == "" && len(args) == 1 {
				setting, err := catclient.Settings.GetByID(context.Background(), args[0])
				if err != nil {
					return err
				}
				filteredSettings = append(filteredSettings, setting)
			} else {
				// if a query was specified we want to filter with it
				// and if any arguments were also specified we would like to filter for them
				settings, err := catclient.Settings.Get(context.Background(), cat.SettingsGetOpts{Search: cfg.search})
				if err != nil {
					return err
				}

				// then we want to filter the ID with the provided args, if any
				if len(args) == 0 {
					filteredSettings = settings
				} else {
					idsToFilter := mapset.NewSet(args...)
					for _, setting := range settings {
						if idsToFilter.Contains(setting.ID) {
							filteredSettings = append(filteredSettings, setting)
						}
					}
				}
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME\tCATEGORY")
			for _, setting := range filteredSettings {
				fmt.Fprintf(w, "%s\t%s\t%s\n", setting.ID, setting.Name, setting.Category)
			}
			w.Flush()

			return nil
		},
	}

	settingsGetCmd.Flags().StringVar(&cfg.search, "search", "", "The search query used to filter the settings by name")

	return settingsGetCmd
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
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			m := map[string]any{}

			if len(args) > 0 {
				err := json.Unmarshal([]byte(args[0]), &m)
				if err != nil {
					return fmt.Errorf("invalid JSON setting: %w", err)
				}
			}

			req := cat.SettingCreateRequest{
				Name:     cfg.name,
				Category: cfg.category,
				Value:    m,
			}

			setting, err := catclient.Settings.Create(cmd.Context(), req)
			if err != nil {
				return fmt.Errorf("creating setting: %w", err)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME\tCATEGORY")
			fmt.Fprintf(w, "%s\t%s\t%s\n", setting.ID, setting.Name, setting.Category)
			w.Flush()

			return nil
		},
	}

	settingsCreateCmd.Flags().StringVar(&cfg.name, "name", "", "The name of the setting")
	settingsCreateCmd.Flags().StringVar(&cfg.category, "category", "", "The category of the setting")

	return settingsCreateCmd
}
