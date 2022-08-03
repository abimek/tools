package main

import (
	"archive/zip"
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/auth"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println("towel <address:port>")
		return
	}
	fmt.Println("Running Towel v1.0.0...")
	dialer := minecraft.Dialer{
		TokenSource: auth.TokenSource,
	}

	address := os.Args[1]

	conn, err := dialer.Dial("raknet", address)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	if err := conn.DoSpawn(); err != nil {
		panic(err)
	}
	for _, pack := range conn.ResourcePacks() {
		fmt.Printf("Getting Resource Pack: %s", pack.Name())
		fmt.Println("...")

		buff, err := ioutil.ReadAll(pack.Content)
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
			_, err = pack.Content.WriteTo(file)
			if err != nil {
				fmt.Println("error writing pack")
				return
			}
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
