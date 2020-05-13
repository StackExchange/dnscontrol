package dnsimple

import (
	"context"
	"fmt"
)

// TemplatesService handles communication with the template related
// methods of the DNSimple API.
//
// See https://developer.dnsimple.com/v2/templates/
type TemplatesService struct {
	client *Client
}

// Template represents a Template in DNSimple.
type Template struct {
	ID          int64  `json:"id,omitempty"`
	SID         string `json:"sid,omitempty"`
	AccountID   int64  `json:"account_id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

func templatePath(accountID string, templateIdentifier string) (path string) {
	path = fmt.Sprintf("/%v/templates", accountID)
	if templateIdentifier != "" {
		path += fmt.Sprintf("/%v", templateIdentifier)
	}
	return
}

// TemplateResponse represents a response from an API method that returns a Template struct.
type TemplateResponse struct {
	Response
	Data *Template `json:"data"`
}

// TemplatesResponse represents a response from an API method that returns a collection of Template struct.
type TemplatesResponse struct {
	Response
	Data []Template `json:"data"`
}

// ListTemplates list the templates for an account.
//
// See https://developer.dnsimple.com/v2/templates/#list
func (s *TemplatesService) ListTemplates(ctx context.Context, accountID string, options *ListOptions) (*TemplatesResponse, error) {
	path := versioned(templatePath(accountID, ""))
	templatesResponse := &TemplatesResponse{}

	path, err := addURLQueryOptions(path, options)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.get(ctx, path, templatesResponse)
	if err != nil {
		return templatesResponse, err
	}

	templatesResponse.HTTPResponse = resp
	return templatesResponse, nil
}

// CreateTemplate creates a new template.
//
// See https://developer.dnsimple.com/v2/templates/#create
func (s *TemplatesService) CreateTemplate(ctx context.Context, accountID string, templateAttributes Template) (*TemplateResponse, error) {
	path := versioned(templatePath(accountID, ""))
	templateResponse := &TemplateResponse{}

	resp, err := s.client.post(ctx, path, templateAttributes, templateResponse)
	if err != nil {
		return nil, err
	}

	templateResponse.HTTPResponse = resp
	return templateResponse, nil
}

// GetTemplate fetches a template.
//
// See https://developer.dnsimple.com/v2/templates/#get
func (s *TemplatesService) GetTemplate(ctx context.Context, accountID string, templateIdentifier string) (*TemplateResponse, error) {
	path := versioned(templatePath(accountID, templateIdentifier))
	templateResponse := &TemplateResponse{}

	resp, err := s.client.get(ctx, path, templateResponse)
	if err != nil {
		return nil, err
	}

	templateResponse.HTTPResponse = resp
	return templateResponse, nil
}

// UpdateTemplate updates a template.
//
// See https://developer.dnsimple.com/v2/templates/#update
func (s *TemplatesService) UpdateTemplate(ctx context.Context, accountID string, templateIdentifier string, templateAttributes Template) (*TemplateResponse, error) {
	path := versioned(templatePath(accountID, templateIdentifier))
	templateResponse := &TemplateResponse{}

	resp, err := s.client.patch(ctx, path, templateAttributes, templateResponse)
	if err != nil {
		return nil, err
	}

	templateResponse.HTTPResponse = resp
	return templateResponse, nil
}

// DeleteTemplate deletes a template.
//
// See https://developer.dnsimple.com/v2/templates/#delete
func (s *TemplatesService) DeleteTemplate(ctx context.Context, accountID string, templateIdentifier string) (*TemplateResponse, error) {
	path := versioned(templatePath(accountID, templateIdentifier))
	templateResponse := &TemplateResponse{}

	resp, err := s.client.delete(ctx, path, nil, nil)
	if err != nil {
		return nil, err
	}

	templateResponse.HTTPResponse = resp
	return templateResponse, nil
}
