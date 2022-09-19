package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/abimek/tools/minecraft/mctoken/api"
	"github.com/akamensky/argparse"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/auth"
	"golang.org/x/oauth2"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"unsafe"
)

const MINECRAFT_TOKEN = "minecraft_token"

func main() {
	if len(os.Args) == 1 {
		fmt.Println("towel <address:port>")
		return
	}

	parser := argparse.NewParser("towel", "Minecraft pack decryption")
	identifier := parser.String("i", "identifier", &argparse.Options{
		Required: false,
		Help:     "The token identifier used within the environment variables",
		Default:  api.DefaultToken,
	})
	address := parser.String("a", "address", &argparse.Options{
		Required: true,
		Help:     "The minecraft server to get packs from",
		Default:  nil,
	})
	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("Running Towel v1.0.0...")

	src, err := api.GetTokenSource(*identifier)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	dialer := minecraft.Dialer{
		TokenSource: *src,
	}

	conn, err := dialer.Dial("raknet", *address)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	if err := conn.DoSpawn(); err != nil {
		panic(err)
	}
	for _, pack := range conn.ResourcePacks() {
		fmt.Println("pack content key: ", pack.ContentKey())
		fmt.Printf("Getting Resource Pack: %s", pack.Name())
		fmt.Println("...")
		temp := reflect.ValueOf(pack).Elem()
		rf := temp.FieldByName("content")
		rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
		contentVal := rf.Interface()
		content, ok := contentVal.(*bytes.Reader)
		if !ok {
			fmt.Println("unable to reflect pack.Content")
			return
		}
		buff, err := ioutil.ReadAll(content)
		if err != nil {
			panic("error reading pack content")
		}
		if pack.Encrypted() {
			fmt.Println("Decoding...")
			err = decrypt_pack(buff, pack.Name()+".zip", pack.ContentKey())
			if err != nil {
				panic("error converting pack")
			}
			fmt.Println("Decoded")
			fmt.Println("Unzipping")
			err = unzipSource(pack.Name()+".zip", pack.Name())
			if err != nil {
				fmt.Println("unable to unzip pack")
				return
			}
			err = os.Remove(pack.Name() + ".zip")
			if err != nil {
				fmt.Println("unable to remove zip file")
				return
			}
			fmt.Println("Unzipped")
		} else {
			file, err := os.Create(pack.Name() + ".zip")
			if err != nil {
				fmt.Println("Error getting pack")
				return
			}
			_, err = (content).WriteTo(file)
			if err != nil {
				fmt.Println("error writing pack")
				return
			}
			err = unzipSource(pack.Name()+".zip", pack.Name())
			if err != nil {
				log.Print(err.Error())
				fmt.Println("unable to unzip pack")
				continue
			}
			err = os.Remove(pack.Name() + ".zip")
			if err != nil {
				fmt.Println("unable to remove zip file")
				return
			}
		}
	}
	fmt.Println("Towel has run successfully!")
}

func unzipSource(source, destination string) error {
	reader, err := zip.OpenReader(source)
	if err != nil {
		return err
	}
	defer reader.Close()

	destination, err = filepath.Abs(destination)
	if err != nil {
		return err
	}

	for _, f := range reader.File {
		err := unzipFile(f, destination)
		if err != nil {
			return err
		}
	}

	return nil
}

func unzipFile(f *zip.File, destination string) error {
	filePath := filepath.Join(destination, f.Name)
	if !strings.HasPrefix(filePath, filepath.Clean(destination)+string(os.PathSeparator)) {
		return fmt.Errorf("invalid file path: %s", filePath)
	}
	if f.FileInfo().IsDir() {
		if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
			return err
		}
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return err
	}
	destinationFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer destinationFile.Close()
	zippedFile, err := f.Open()
	if err != nil {
		return err
	}
	defer zippedFile.Close()
	if _, err := io.Copy(destinationFile, zippedFile); err != nil {
		return err
	}
	return nil
}

func tokenSource() oauth2.TokenSource {
	check := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	token := new(oauth2.Token)
	tokenData := []byte(os.Getenv(MINECRAFT_TOKEN))
	if len(tokenData) != 0 {
		_ = json.Unmarshal(tokenData, token)
	} else {
		tokens, err := auth.RequestLiveToken()
		check(err)
		token = tokens
	}
	src := auth.RefreshTokenSource(token)
	_, err := src.Token()
	if err != nil {
		// The cached refresh token expired and can no longer be used to obtain a new token. We require the
		// user to log in again and use that token instead.
		token, err = auth.RequestLiveToken()
		check(err)
		src = auth.RefreshTokenSource(token)
	}
	go func() {
		c := make(chan os.Signal, 3)
		signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
		<-c

		tok, _ := src.Token()
		b, _ := json.Marshal(tok)
		err = os.Setenv(MINECRAFT_TOKEN, string(b))
		if err != nil {
			fmt.Printf("Error setting env variable: %s", MINECRAFT_TOKEN)
			return
		}
		os.Exit(0)
	}()
	return src
}
