package api

import (
	"golang.org/x/oauth2"
)

type TokenSource struct {
	token *oauth2.Token
}

func (t *TokenSource) Token() (*oauth2.Token, error) {
	return t.token, nil
}
