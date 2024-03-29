package client

import (
	"context"
	"net/url"
)

type SettingsService struct {
	client *Client
}

type Setting struct {
	ID        string         `json:"setting_id"`
	Name      string         `json:"name,omitempty"`
	Value     map[string]any `json:"value,omitempty"`
	Category  string         `json:"category,omitempty"`
	UpdatedAt int64          `json:"updated_at,omitempty"`
}

type settingsResponse struct {
	Settings []*Setting `json:"settings"`
}

type settingResponse struct {
	Setting *Setting `json:"setting"`
}

type SettingCreateRequest createUpdateRequest

type SettingUpdateRequest createUpdateRequest

type createUpdateRequest struct {
	Name     string         `json:"name"`
	Category string         `json:"category"`
	Value    map[string]any `json:"value"`
}

type SettingsGetOpts struct {
	Search string
}

func (s *SettingsService) Get(ctx context.Context, opts SettingsGetOpts) ([]*Setting, error) {
	endpoint := "/settings"

	values := url.Values{}
	if opts.Search != "" {
		values.Set("search", opts.Search)
	}

	if len(values) > 0 {
		endpoint += "?" + values.Encode()
	}

	resp, err := get(ctx, s.client, endpoint, settingsResponse{})
	if err != nil {
		return nil, err
	}
	return resp.Value.Settings, nil
}

func (s *SettingsService) GetByID(ctx context.Context, ID string) (*Setting, error) {
	resp, err := get(ctx, s.client, "/settings/"+ID, settingResponse{})
	if err != nil {
		return nil, err
	}
	return resp.Value.Setting, nil
}

func (s *SettingsService) Create(ctx context.Context, req SettingCreateRequest) (*Setting, error) {
	resp, err := post(ctx, s.client, "/settings", req, settingResponse{})
	if err != nil {
		return nil, err
	}
	return resp.Value.Setting, nil
}

func (s *SettingsService) Update(ctx context.Context, ID string, req SettingUpdateRequest) (*Setting, error) {
	resp, err := put(ctx, s.client, "/settings/"+ID, req, settingResponse{})
	if err != nil {
		return nil, err
	}
	return resp.Value.Setting, nil
}

func (s *SettingsService) Delete(ctx context.Context, ID string) (*Setting, error) {
	resp, err := del(ctx, s.client, "/settings/"+ID, settingResponse{})
	if err != nil {
		return nil, err
	}
	return resp.Value.Setting, nil
}
