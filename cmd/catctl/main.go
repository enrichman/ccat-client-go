package main

import (
	"fmt"
	"os"
	"strings"

	cat "github.com/enrichman/ccat-client-go"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	// The name of the config file (.ccat.yml, .ccat.json, ...)
	defaultConfigFilename = ".ccat"

	// The environment variable prefix of all environment variables bound to the command line flags.
	// For example, --apikey is bound to CCAT_APIKEY.
	envPrefix = "CCAT"
)

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

func NewRootCmd() (*cobra.Command, error) {
	catclient := &cat.Client{}

	type rootCfg struct {
		apiKey string
	}

	cfg := &rootCfg{}

	rootCmd := &cobra.Command{
		Use: "catctl",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			err := initializeConfig(cmd)
			if err != nil {
				return err
			}

			newCatclient, err := cat.NewClient(
				cat.WithAPIKey(cfg.apiKey),
			)
			if err != nil {
				return err
			}

			// swap underlying implementation after client setup
			*catclient = *newCatclient

			return nil
		},
	}

	rootCmd.AddCommand(
		NewChatCmd(catclient),
		NewSettingsCmd(catclient),
		NewVersionCmd(catclient),
	)

	rootCmd.PersistentFlags().StringVar(&cfg.apiKey, "apikey", "", "The apikey to use with authenticated server")

	return rootCmd, nil
}

func initializeConfig(cmd *cobra.Command) error {
	v := viper.New()

	// Set the base name of the config file, without the file extension.
	v.SetConfigName(defaultConfigFilename)

	// Set as many paths as you like where viper should look for the
	// config file. We are only looking in the current working directory.
	v.AddConfigPath(".")

	// Attempt to read the config file, gracefully ignoring errors
	// caused by a config file not being found. Return an error
	// if we cannot parse the config file.
	if err := v.ReadInConfig(); err != nil {
		// It's okay if there isn't a config file
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	// When we bind flags to environment variables expect that the
	// environment variables are prefixed, e.g. a flag like --number
	// binds to an environment variable STING_NUMBER. This helps
	// avoid conflicts.
	v.SetEnvPrefix(envPrefix)

	// Environment variables can't have dashes in them, so bind them to their equivalent
	// keys with underscores, e.g. --favorite-color to STING_FAVORITE_COLOR
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// Bind to environment variables
	// Works great for simple config names, but needs help for names
	// like --favorite-color which we fix in the bindFlags function
	v.AutomaticEnv()

	// Bind the current command's flags to viper
	return bindFlags(cmd, v)
}

// Bind each cobra flag to its associated viper configuration (config file and environment variable)
func bindFlags(cmd *cobra.Command, v *viper.Viper) error {
	var err error

	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		configName := strings.ReplaceAll(f.Name, "-", "")

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && v.IsSet(configName) {
			val := v.Get(configName)
			setErr := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
			if err == nil {
				err = setErr
			}
		}
	})

	return err
}
