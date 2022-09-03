package main

import (
	"fmt"
	"github.com/abimek/tools/minecraft/mctoken/api"
	"github.com/akamensky/argparse"
	"github.com/sandertv/gophertunnel/minecraft/auth"
	"os"
)

func main() {
	parser := argparse.NewParser("mctoken", "Set's a minecraft token inside of you're environment variables")
	config, err := api.NewTokensConfig()
	defer config.Close()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	identifier := parser.String("i", "identifier", &argparse.Options{
		Required: false,
		Help:     "The token identifier used within the environment variables",
		Default:  api.DefaultToken,
	})
	prints := parser.Flag("p", "print", &argparse.Options{
		Required: false,
		Help:     "Print out a token",
		Default:  false,
	})
	list := parser.Flag("l", "list", &argparse.Options{
		Required: false,
		Help:     "List all available minecraft token identifiers",
		Default:  false,
	})
	s := parser.Flag("s", "set", &argparse.Options{
		Required: false,
		Help:     "Set or replace a minecraft token in env variables",
		Default:  false,
	})
	expired := parser.Flag("e", "expired", &argparse.Options{
		Required: false,
		Help:     "checks to see whether a token has expired",
		Default:  false,
	})

	err = parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		return
	}
	if *s != false {
		SetToken(identifier, config)
		return
	}
	if *expired != false {
		CheckExpired(identifier, config)
		return
	}
	if *list != false {
		ListIdentifiers(config)
		return
	}
	if *prints != false {
		PrintToken(identifier, config)
		return
	}
}

// SetToken handles the logic of setting the token
func SetToken(identifier *string, config *api.TokensConfig) {
	token, err := auth.RequestLiveToken()
	if err != nil {
		fmt.Println("Unable to set token: %s", *identifier)
		return
	}
	src := auth.RefreshTokenSource(token)
	tok, _ := src.Token()
	err = config.Add(*identifier, tok)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("Set token %s\n", *identifier)
}

// CheckExpired checks to see if a token has expired
func CheckExpired(identifier *string, config *api.TokensConfig) {
	if config.Contains(*identifier) {
		token, err := config.Get(*identifier)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		tok, err := token.Token()
		if err != nil {
			fmt.Println("Unable to get token")
			return
		}
		src := auth.RefreshTokenSource(tok)
		_, err = src.Token()
		if err != nil {
			fmt.Printf("Token: %s has expired", *identifier)
			return
		}
		fmt.Printf("Token: %s has not expired", *identifier)
		return
	}
	fmt.Printf("Token: %s does not eixst", *identifier)
}

func PrintToken(identifier *string, config *api.TokensConfig) {
	if config.Contains(*identifier) {
		t := os.Getenv(*identifier)
		if t != "" {
			fmt.Printf("%v", t)
			return
		}
		fmt.Printf("Token: %s is not set", *identifier)
		config.Remove(*identifier)
	}
	fmt.Printf("Token: %s does not eixst", *identifier)
}

func ListIdentifiers(config *api.TokensConfig) {
	lens := len(config.Tokens)
	if lens == 0 {
		fmt.Println("You have 0 Identifiers")
		return
	}
	fmt.Printf("Total Identifiers: %v\n", lens)
	fmt.Println("----------------------------------------")
	for id, _ := range config.Tokens {
		fmt.Println(id)
	}
	fmt.Println("----------------------------------------")
}
