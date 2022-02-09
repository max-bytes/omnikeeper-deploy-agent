package omnikeeper

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/shurcooL/graphql"
	"golang.org/x/oauth2"
)

func BuildGraphQLClient(ctx context.Context, omnikeeperURL string, keycloakClientID string, username string, password string) (*graphql.Client, error) {

	oAuthEndpoint, err := fetchOAuthInfo(omnikeeperURL)
	if err != nil {
		return nil, fmt.Errorf("Error fetching oauth info: %w", err)
	}

	oauth2cfg := &oauth2.Config{
		ClientID: keycloakClientID,
		Endpoint: *oAuthEndpoint,
	}

	token, err := oauth2cfg.PasswordCredentialsToken(ctx, username, password)
	if err != nil {
		return nil, fmt.Errorf("Error getting token: %w", err)
	}

	httpClient := oauth2cfg.Client(ctx, token)
	client := graphql.NewClient(fmt.Sprintf("%s/graphql", omnikeeperURL), httpClient)

	return client, nil
}

func fetchOAuthInfo(omnikeeperURL string) (*oauth2.Endpoint, error) {

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/.well-known/openid-configuration", omnikeeperURL), nil)
	if err != nil {
		return nil, fmt.Errorf("Could not create get request to openid-configuration: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	var httpClient = &http.Client{
		Timeout: time.Second * 20,
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Could not fetch openid-configuration from omnikeeper instance at %s: %w", omnikeeperURL, err)
	}
	defer resp.Body.Close()

	type Resp struct {
		TokenEndpoint         string `json:"token_endpoint"`
		AuthorizationEndpoint string `json:"authorization_endpoint"`
	}

	var r Resp
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return nil, fmt.Errorf("Could not decode openid-configuration from omnikeeper instance at %s: %w", omnikeeperURL, err)
	}

	ret := oauth2.Endpoint{
		AuthURL:  r.AuthorizationEndpoint,
		TokenURL: r.TokenEndpoint,
	}

	return &ret, nil
}
