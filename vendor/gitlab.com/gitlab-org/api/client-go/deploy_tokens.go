//
// Copyright 2021, Sander van Harmelen
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package gitlab

import (
	"fmt"
	"net/http"
	"time"
)

type (
	DeployTokensServiceInterface interface {
		ListAllDeployTokens(options ...RequestOptionFunc) ([]*DeployToken, *Response, error)
		ListProjectDeployTokens(pid interface{}, opt *ListProjectDeployTokensOptions, options ...RequestOptionFunc) ([]*DeployToken, *Response, error)
		GetProjectDeployToken(pid interface{}, deployToken int, options ...RequestOptionFunc) (*DeployToken, *Response, error)
		CreateProjectDeployToken(pid interface{}, opt *CreateProjectDeployTokenOptions, options ...RequestOptionFunc) (*DeployToken, *Response, error)
		DeleteProjectDeployToken(pid interface{}, deployToken int, options ...RequestOptionFunc) (*Response, error)
		ListGroupDeployTokens(gid interface{}, opt *ListGroupDeployTokensOptions, options ...RequestOptionFunc) ([]*DeployToken, *Response, error)
		GetGroupDeployToken(gid interface{}, deployToken int, options ...RequestOptionFunc) (*DeployToken, *Response, error)
		CreateGroupDeployToken(gid interface{}, opt *CreateGroupDeployTokenOptions, options ...RequestOptionFunc) (*DeployToken, *Response, error)
		DeleteGroupDeployToken(gid interface{}, deployToken int, options ...RequestOptionFunc) (*Response, error)
	}

	// DeployTokensService handles communication with the deploy tokens related methods
	// of the GitLab API.
	//
	// GitLab API docs: https://docs.gitlab.com/api/deploy_tokens/
	DeployTokensService struct {
		client *Client
	}
)

// DeployToken represents a GitLab deploy token.
type DeployToken struct {
	ID        int        `json:"id"`
	Name      string     `json:"name"`
	Username  string     `json:"username"`
	ExpiresAt *time.Time `json:"expires_at"`
	Revoked   bool       `json:"revoked"`
	Expired   bool       `json:"expired"`
	Token     string     `json:"token,omitempty"`
	Scopes    []string   `json:"scopes"`
}

func (k DeployToken) String() string {
	return Stringify(k)
}

// ListAllDeployTokens gets a list of all deploy tokens.
//
// GitLab API docs:
// https://docs.gitlab.com/api/deploy_tokens/#list-all-deploy-tokens
func (s *DeployTokensService) ListAllDeployTokens(options ...RequestOptionFunc) ([]*DeployToken, *Response, error) {
	req, err := s.client.NewRequest(http.MethodGet, "deploy_tokens", nil, options)
	if err != nil {
		return nil, nil, err
	}

	var ts []*DeployToken
	resp, err := s.client.Do(req, &ts)
	if err != nil {
		return nil, resp, err
	}

	return ts, resp, nil
}

// ListProjectDeployTokensOptions represents the available ListProjectDeployTokens()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/api/deploy_tokens/#list-project-deploy-tokens
type ListProjectDeployTokensOptions ListOptions

// ListProjectDeployTokens gets a list of a project's deploy tokens.
//
// GitLab API docs:
// https://docs.gitlab.com/api/deploy_tokens/#list-project-deploy-tokens
func (s *DeployTokensService) ListProjectDeployTokens(pid interface{}, opt *ListProjectDeployTokensOptions, options ...RequestOptionFunc) ([]*DeployToken, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/deploy_tokens", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var ts []*DeployToken
	resp, err := s.client.Do(req, &ts)
	if err != nil {
		return nil, resp, err
	}

	return ts, resp, nil
}

// GetProjectDeployToken gets a single deploy token.
//
// GitLab API docs:
// https://docs.gitlab.com/api/deploy_tokens/#get-a-project-deploy-token
func (s *DeployTokensService) GetProjectDeployToken(pid interface{}, deployToken int, options ...RequestOptionFunc) (*DeployToken, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/deploy_tokens/%d", PathEscape(project), deployToken)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	t := new(DeployToken)
	resp, err := s.client.Do(req, t)
	if err != nil {
		return nil, resp, err
	}

	return t, resp, nil
}

// CreateProjectDeployTokenOptions represents the available CreateProjectDeployToken() options.
//
// GitLab API docs:
// https://docs.gitlab.com/api/deploy_tokens/#create-a-project-deploy-token
type CreateProjectDeployTokenOptions struct {
	Name      *string    `url:"name,omitempty" json:"name,omitempty"`
	ExpiresAt *time.Time `url:"expires_at,omitempty" json:"expires_at,omitempty"`
	Username  *string    `url:"username,omitempty" json:"username,omitempty"`
	Scopes    *[]string  `url:"scopes,omitempty" json:"scopes,omitempty"`
}

// CreateProjectDeployToken creates a new deploy token for a project.
//
// GitLab API docs:
// https://docs.gitlab.com/api/deploy_tokens/#create-a-project-deploy-token
func (s *DeployTokensService) CreateProjectDeployToken(pid interface{}, opt *CreateProjectDeployTokenOptions, options ...RequestOptionFunc) (*DeployToken, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/deploy_tokens", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	t := new(DeployToken)
	resp, err := s.client.Do(req, t)
	if err != nil {
		return nil, resp, err
	}

	return t, resp, nil
}

// DeleteProjectDeployToken removes a deploy token from the project.
//
// GitLab API docs:
// https://docs.gitlab.com/api/deploy_tokens/#delete-a-project-deploy-token
func (s *DeployTokensService) DeleteProjectDeployToken(pid interface{}, deployToken int, options ...RequestOptionFunc) (*Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("projects/%s/deploy_tokens/%d", PathEscape(project), deployToken)

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// ListGroupDeployTokensOptions represents the available ListGroupDeployTokens()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/api/deploy_tokens/#list-group-deploy-tokens
type ListGroupDeployTokensOptions ListOptions

// ListGroupDeployTokens gets a list of a group’s deploy tokens.
//
// GitLab API docs:
// https://docs.gitlab.com/api/deploy_tokens/#list-group-deploy-tokens
func (s *DeployTokensService) ListGroupDeployTokens(gid interface{}, opt *ListGroupDeployTokensOptions, options ...RequestOptionFunc) ([]*DeployToken, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/deploy_tokens", PathEscape(group))

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var ts []*DeployToken
	resp, err := s.client.Do(req, &ts)
	if err != nil {
		return nil, resp, err
	}

	return ts, resp, nil
}

// GetGroupDeployToken gets a single deploy token.
//
// GitLab API docs:
// https://docs.gitlab.com/api/deploy_tokens/#get-a-group-deploy-token
func (s *DeployTokensService) GetGroupDeployToken(gid interface{}, deployToken int, options ...RequestOptionFunc) (*DeployToken, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/deploy_tokens/%d", PathEscape(group), deployToken)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	t := new(DeployToken)
	resp, err := s.client.Do(req, t)
	if err != nil {
		return nil, resp, err
	}

	return t, resp, nil
}

// CreateGroupDeployTokenOptions represents the available CreateGroupDeployToken() options.
//
// GitLab API docs:
// https://docs.gitlab.com/api/deploy_tokens/#create-a-group-deploy-token
type CreateGroupDeployTokenOptions struct {
	Name      *string    `url:"name,omitempty" json:"name,omitempty"`
	ExpiresAt *time.Time `url:"expires_at,omitempty" json:"expires_at,omitempty"`
	Username  *string    `url:"username,omitempty" json:"username,omitempty"`
	Scopes    *[]string  `url:"scopes,omitempty" json:"scopes,omitempty"`
}

// CreateGroupDeployToken creates a new deploy token for a group.
//
// GitLab API docs:
// https://docs.gitlab.com/api/deploy_tokens/#create-a-group-deploy-token
func (s *DeployTokensService) CreateGroupDeployToken(gid interface{}, opt *CreateGroupDeployTokenOptions, options ...RequestOptionFunc) (*DeployToken, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/deploy_tokens", PathEscape(group))

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	t := new(DeployToken)
	resp, err := s.client.Do(req, t)
	if err != nil {
		return nil, resp, err
	}

	return t, resp, nil
}

// DeleteGroupDeployToken removes a deploy token from the group.
//
// GitLab API docs:
// https://docs.gitlab.com/api/deploy_tokens/#delete-a-group-deploy-token
func (s *DeployTokensService) DeleteGroupDeployToken(gid interface{}, deployToken int, options ...RequestOptionFunc) (*Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("groups/%s/deploy_tokens/%d", PathEscape(group), deployToken)

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}
