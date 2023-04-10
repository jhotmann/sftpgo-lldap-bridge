package main

import (
	"errors"
	"log"

	"github.com/go-resty/resty/v2"
	"github.com/sftpgo/sdk"
)

var (
	sftpgoClient *resty.Client
)

type SftpgoTokenResponse struct {
	AccessToken string `json:"access_token"`
}

func init() {
	sftpgoClient = resty.New().SetBaseURL(config.Sftpgo.BaseURL)

	getSftpToken()
}

func getSftpToken() {
	var tokenResponse SftpgoTokenResponse
	resp, err := sftpgoClient.R().
		SetBasicAuth(config.Sftpgo.User, config.Sftpgo.Password).
		SetResult(&tokenResponse).
		Get("/api/v2/token")

	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode() < 200 || resp.StatusCode() > 299 {
		log.Fatal("Received " + resp.Status() + " getting SFTPGO token")
	}

	sftpgoClient.SetAuthToken(tokenResponse.AccessToken)
}

func getSftpGroups() ([]sdk.Group, error) {
	var groups []sdk.Group
	resp, err := sftpgoClient.R().
		SetResult(&groups).
		Get("/api/v2/groups")

	if err != nil {
		return groups, err
	}

	if resp.StatusCode() < 200 || resp.StatusCode() > 299 {
		return groups, errors.New("Received " + resp.Status() + " getting SFTPGo groups")
	}

	return groups, nil
}
