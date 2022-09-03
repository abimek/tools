package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/oauth2"
	"io/ioutil"
	"os"
	"path/filepath"
)

type TokensConfig struct {
	File   *os.File
	Tokens map[string]string `json:"tokens"`
}

func (t *TokensConfig) Contains(identifier string) bool {
	if _, ok := t.Tokens[identifier]; ok {
		return true
	}
	return false
}

func (t *TokensConfig) Get(identifier string) (oauth2.TokenSource, error) {
	if !t.Contains(identifier) {
		return nil, fmt.Errorf("error: identifier does not eixst")
	}
	if t.Tokens[identifier] == "" {
		t.Remove(identifier)
		return nil, fmt.Errorf("error: token variable does not eixst")
	}
	token := new(oauth2.Token)
	_ = json.Unmarshal([]byte(t.Tokens[identifier]), token)
	return &TokenSource{token: token}, nil
}

func (t *TokensConfig) Add(identifier string, token *oauth2.Token) error {
	b, _ := json.Marshal(token)
	t.Tokens[identifier] = string(b)
	return nil
}

func (t *TokensConfig) Remove(identifier string) {
	if _, ok := t.Tokens[identifier]; ok {
		delete(t.Tokens, identifier)
	}
}

func NewTokensConfig() (*TokensConfig, error) {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	if err != nil {
		return nil, fmt.Errorf("unable to get current directory")
	}
	path := filepath.Dir(ex) + "/mctoken/tokens.json"
	var file *os.File
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		file, err = os.Create(path)
	} else {
		file, err = os.OpenFile(path, os.O_RDWR, 0666)
	}
	if err != nil {
		return nil, err
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("unable to read token config")
	}
	var config TokensConfig
	err = json.Unmarshal(content, &config)
	if err != nil {
		config.Tokens = map[string]string{}
	}
	config.File = file
	return &config, nil
}

func (t *TokensConfig) Close() error {
	defer t.File.Close()
	data, err := json.Marshal(*t)
	if err != nil {
		return fmt.Errorf("unable to marhshal token config")
	}
	_, err = t.File.WriteAt(data, 0)
	if err != nil {
		return fmt.Errorf("unable to write data to token config")
	}
	return nil
}

func remove[T any](slice []T, s int) []T {
	return append(slice[:s], slice[s+1:]...)
}
