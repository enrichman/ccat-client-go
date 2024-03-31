package client

import (
	"context"
)

type LLMService struct {
	client *Client
}

type llmSettingsResponse struct {
	Settings []*LLMSetting `json:"settings"`
}

type LLMSetting struct {
	Name   string         `json:"name,omitempty"`
	Value  map[string]any `json:"value,omitempty"`
	Schema *LLMSchema     `json:"schema,omitempty"`
}

type LLMSchema struct {
	HumanReadableName string                      `json:"humanReadableName"`
	Description       string                      `json:"description"`
	Properties        map[string]SchemaProperties `json:"properties"`
}

type SchemaProperties struct {
	Title   string `json:"title"`
	Type    string `json:"type"`
	Default any    `json:"default"`
}

func (s *LLMService) Get(ctx context.Context) ([]*LLMSetting, error) {
	resp, err := get(ctx, s.client, "/llm/settings", llmSettingsResponse{})
	if err != nil {
		return nil, err
	}
	return resp.Value.Settings, nil
}

func (s *LLMService) GetByID(ctx context.Context, ID string) (*LLMSetting, error) {
	resp, err := get(ctx, s.client, "/llm/settings/"+ID, &LLMSetting{})
	if err != nil {
		return nil, err
	}
	return resp.Value, nil
}

func (s *LLMService) Update(ctx context.Context, ID string, req map[string]any) (*Setting, error) {
	resp, err := put(ctx, s.client, "/llm/settings/"+ID, req, settingResponse{})
	if err != nil {
		return nil, err
	}
	return resp.Value.Setting, nil
}
