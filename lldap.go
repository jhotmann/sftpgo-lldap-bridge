package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/hasura/go-graphql-client"
)

var (
	lldapRestClient *resty.Client
	// lldapGraphqlClient *graphql.Client
	// lldapRefreshToken  string
)

type LldapTokenBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LldapTokenResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
}

type LldapUser struct {
	Id     string       `graphql:"id"`
	Groups []LldapGroup `graphql:"groups"`
}

type LldapGroup struct {
	DisplayName string `graphql:"displayName"`
}

type LldapUserResponse struct {
	User LldapUser `graphql:"user(userId: $username)"`
}

func init() {
	lldapRestClient = resty.New()
}

func getLldapToken(username string, password string) (string, error) {
	var tokenResponse LldapTokenResponse
	resp, err := lldapRestClient.R().
		SetHeader("Content-Type", "application/json").
		SetBody(LldapTokenBody{Username: username, Password: password}).
		SetResult(&tokenResponse).
		Post(fmt.Sprintf("%s/auth/simple/login", config.Lldap.BaseURL))

	if err != nil {
		return "", err
	}

	if resp.StatusCode() < 200 || resp.StatusCode() > 299 {
		return "", errors.New("Received " + resp.Status() + " getting LLDAP token")
	}

	return tokenResponse.Token, nil
}

func getLldapUser(token string, username string) (LldapUserResponse, error) {
	variables := map[string]interface{}{
		"username": username,
	}
	var query LldapUserResponse
	lldapGraphqlClient := graphql.NewClient(fmt.Sprintf("%s/api/graphql", config.Lldap.BaseURL), nil).
		WithRequestModifier(func(r *http.Request) {
			r.Header.Add("Authorization", "Bearer "+token)
		})
	err := lldapGraphqlClient.Query(context.Background(), &query, variables)
	return query, err
}
