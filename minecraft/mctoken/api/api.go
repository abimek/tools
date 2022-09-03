package api

import (
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft/auth"
	"golang.org/x/oauth2"
)

const DefaultToken = "minecraft_default_token"

func GetTokenSource(identifier string) (*oauth2.TokenSource, error) {
	config, err := NewTokensConfig()
	if err != nil {
		return nil, err
	}
	token, err := config.Get(identifier)
	if err != nil {
		return nil, err
	}
	tok, err := token.Token()
	if err != nil {
		return nil, fmt.Errorf("unable to get token")
	}
	src := auth.RefreshTokenSource(tok)
	_, err = src.Token()
	if err != nil {
		tk, err := auth.RequestLiveToken()
		if err != nil {
			return nil, fmt.Errorf("unable to get new token")
		}
		src = auth.RefreshTokenSource(tk)
		t, err := src.Token()
		if err != nil {
			return nil, fmt.Errorf("unable to get token")
		}
		err = config.Add(identifier, t)
		if err != nil {
			return nil, err
		}
	}
	return &src, nil
}
