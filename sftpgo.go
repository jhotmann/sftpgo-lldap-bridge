package main

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sftpgo/sdk"
)

var (
	sftpgoClient          *resty.Client
	sftpgoTokenExpiration time.Time
)

type SftpgoTokenResponse struct {
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
}

func initSftpGoClient() {
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
		log.Printf("Received " + resp.Status() + " getting SFTPGO token\n")
		return
	}

	sftpgoClient.SetAuthToken(tokenResponse.AccessToken)
	log.Printf("SFTPGo token will expire at %t", tokenResponse.ExpiresAt)
	sftpgoTokenExpiration = tokenResponse.ExpiresAt
}

func getSftpGroups(retry int) ([]sdk.Group, error) {
	if time.Now().After(sftpgoTokenExpiration) {
		initSftpGoClient()
	}
	var groups []sdk.Group
	resp, err := sftpgoClient.R().
		SetResult(&groups).
		Get("/api/v2/groups")

	if err != nil {
		return groups, err
	}

	if resp.StatusCode() < 200 || resp.StatusCode() > 299 {
		if resp.StatusCode() == http.StatusUnauthorized && retry < 3 {
			initSftpGoClient()
			return getSftpGroups(retry + 1)
		} else {
			return groups, errors.New("Received " + resp.Status() + " getting SFTPGo groups")
		}
	}

	return groups, nil
}
