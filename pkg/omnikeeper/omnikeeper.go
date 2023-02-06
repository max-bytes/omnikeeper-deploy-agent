package omnikeeper

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hasura/go-graphql-client"
	"golang.org/x/oauth2"
)

func BuildGraphQLClient(ctx context.Context, omnikeeperURL string, keycloakClientID string, username string, password string, insecureSkipVerify bool) (*graphql.Client, error) {

	oAuthEndpoint, err := fetchOAuthInfo(omnikeeperURL, insecureSkipVerify)
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

	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: insecureSkipVerify}
	var baseHttpClient = &http.Client{
		Timeout:   time.Second * 30,
		Transport: customTransport,
	}
	modifiedCtx := context.WithValue(ctx, oauth2.HTTPClient, baseHttpClient)
	httpClient := oauth2cfg.Client(modifiedCtx, token)
	client := graphql.NewClient(fmt.Sprintf("%s/graphql", omnikeeperURL), httpClient)

	return client, nil
}

func fetchOAuthInfo(omnikeeperURL string, insecureSkipVerify bool) (*oauth2.Endpoint, error) {

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/.well-known/openid-configuration", omnikeeperURL), nil)
	if err != nil {
		return nil, fmt.Errorf("Could not create get request to openid-configuration: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: insecureSkipVerify}
	var httpClient = &http.Client{
		Timeout:   time.Second * 20,
		Transport: customTransport,
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
