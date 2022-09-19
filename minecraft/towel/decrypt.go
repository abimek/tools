package main

import (
	"archive/zip"
	"bytes"
	"crypto/aes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

// decrypt using cfb with segmentsize = 1
func cfb_decrypt(data []byte, key []byte) ([]byte, error) {
	b, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}

	shift_register := append(key[:16], data...) // prefill with iv + cipherdata
	_tmp := make([]byte, 16)
	off := 0
	for off < len(data) {
		b.Encrypt(_tmp, shift_register)
		data[off] ^= _tmp[0]
		shift_register = shift_register[1:]
		off++
	}
	return data, nil
}

type ContentEntry struct {
	Path string `json:"path"`
	Key  string `json:"key"`
}

type ContentJson struct {
	Content []ContentEntry `json:"content"`
}

func decrypt_pack(pack_zip []byte, filename, key string) error {
	// open reader and writers
	r := bytes.NewReader(pack_zip)
	z, err := zip.NewReader(r, r.Size())
	if err != nil {
		fmt.Errorf("ERROR CREATING PACK ZIP READER %s", err)
		return err
	}
	f, err := os.Create(filename)
	if err != nil {
		fmt.Errorf("ERROR CREATING ZIP FILE %s", err)
		return err
	}
	zw := zip.NewWriter(f)
	defer f.Close()
	defer zw.Close()

	written := make(map[string]interface{})

	// read content json file
	var content ContentJson
	{
		ff, err := z.Open("contents.json")
		if err != nil {
			if os.IsNotExist(err) {
				content = ContentJson{}
			} else {
				fmt.Errorf("ERROR WHILE FINDING CONTENT FILE %s", err)
				return err
			}
		} else {
			buf, err := io.ReadAll(ff)
			if err != nil {
				fmt.Errorf("ERROR READING CONTENT FILE %s", err)
				return err
			}
			dec, err := cfb_decrypt(buf[:], []byte(key))
			if err != nil {
				fmt.Errorf("ERROR DECRYPTING CONTENT FILE %s", err)
				return err
			}
			dec = bytes.Split(dec, []byte("\x00"))[0] // remove trailing \x00 (example: play.galaxite.net)
			fw, err := zw.Create("contents.json")
			if err != nil {
				fmt.Errorf("ERROR CREATING JSON CONTENT %s", err)
				return err
			}
			_, err = fw.Write(dec)
			if err != nil {
				fmt.Errorf("ERROR WRITING CONTENT DEC %s", err)
				return err
			}
			if err := json.Unmarshal(dec, &content); err != nil {
				fmt.Errorf("ERROR UNMARSHING CONTENT JSON %s", err)
				return err
			}
			written["contents.json"] = true
		}
	}

	for _, entry := range content.Content {
		ff, err := z.Open(entry.Path)
		if err != nil {
			log.Print(err.Error())
			continue
		}
		buf, _ := io.ReadAll(ff)
		if entry.Key != "" {
			buf, err = cfb_decrypt(buf, []byte(entry.Key))
			if err != nil {
				fmt.Errorf("ERROR DECRYTING FILE %s, ERROR: %s", entry.Path, err)
				return err
			}
		}
		fw, _ := zw.Create(entry.Path)
		fw.Write(buf)
		written[entry.Path] = true
	}

	// copy everything not in the contents file
	for _, src_file := range z.File {
		if written[src_file.Name] == nil {
			zw.Copy(src_file)
		}
	}

	return nil
}
