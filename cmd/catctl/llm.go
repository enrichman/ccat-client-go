package main

import (
	"context"
	"fmt"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	cat "github.com/enrichman/ccat-client-go"
	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"
)

var ValidLLMs = []string{
	"LLMOpenAIChatConfig",
	"LLMOpenAIConfig",
	"LLMGeminiChatConfig",
	"LLMCohereConfig",
	"LLMAzureOpenAIConfig",
	"LLMAzureChatOpenAIConfig",
	"LLMHuggingFaceEndpointConfig",
	"LLMHuggingFaceTextGenInferenceConfig",
	"LLMOllamaConfig",
	"LLMOpenAICompatibleConfig",
	"LLMCustomConfig",
	"LLMDefaultConfig",
}

func NewLLMCmd(catclient *cat.Client) (*cobra.Command, error) {
	llmCmd := &cobra.Command{
		Use:   "llm",
		Short: "manage LLM settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	updateCmd, err := NewLLMUpdateCmd(catclient)
	if err != nil {
		return nil, err
	}

	llmCmd.AddCommand(
		NewLLMGetCmd(catclient),
		updateCmd,
	)

	return llmCmd, nil
}

func NewLLMGetCmd(catclient *cat.Client) *cobra.Command {
	llmGetCmd := &cobra.Command{
		Use:           "get",
		Short:         "get LLM settings",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.MaximumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 1 {
				return []string{}, cobra.ShellCompDirectiveNoFileComp
			}
			return ValidLLMs, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			llmSettings := []*cat.LLMSetting{}

			if len(args) == 1 {
				llmSetting, err := catclient.LLM.GetByID(context.Background(), args[0])
				if err != nil {
					return err
				}
				llmSettings = append(llmSettings, llmSetting)

			} else {
				setts, err := catclient.LLM.Get(context.Background())
				if err != nil {
					return err
				}
				llmSettings = setts
			}

			for _, setting := range llmSettings {
				y, err := yaml.Marshal(setting)
				if err != nil {
					return err
				}
				fmt.Println(string(y))
			}

			return nil
		},
	}

	return llmGetCmd
}

func NewLLMUpdateCmd(catclient *cat.Client) (*cobra.Command, error) {
	type updateCfg struct {
		keyValues []string
	}

	cfg := &updateCfg{}

	llmUpdateCmd := &cobra.Command{
		Use:           "update",
		Short:         "update LLM settings",
		SilenceUsage:  true,
		SilenceErrors: true,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 1 {
				return []string{}, cobra.ShellCompDirectiveNoFileComp
			}
			return ValidLLMs, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Usage()
			}

			updateRequest := map[string]any{}

			for _, keyValue := range cfg.keyValues {
				kVal := strings.Split(keyValue, "=")
				if len(kVal) == 1 {
					return fmt.Errorf("invalid set flag for '%s': missing value", keyValue)
				}
				updateRequest[kVal[0]] = kVal[1]
			}

			_, err := catclient.LLM.Update(context.Background(), args[0], updateRequest)
			if err != nil {
				return err
			}

			return nil
		},
	}

	llmUpdateCmd.Flags().StringArrayVar(&cfg.keyValues, "set", []string{}, "set key value")
	err := llmUpdateCmd.RegisterFlagCompletionFunc("set", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		llmSetting, err := catclient.LLM.GetByID(context.Background(), args[0])
		if err != nil {
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		}

		availableProps := maps.Keys(llmSetting.Schema.Properties)
		availablePropsSet := mapset.NewSet(availableProps...)

		existingValues := mapset.NewSet[string]()
		for _, kv := range cfg.keyValues {
			existingValues.Add(strings.Split(kv, "=")[0])
		}
		available := availablePropsSet.Difference(existingValues).ToSlice()

		return available, cobra.ShellCompDirectiveNoSpace
	})

	if err != nil {
		return nil, err
	}

	return llmUpdateCmd, nil
}
