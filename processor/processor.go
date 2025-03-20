package processor

import (
	"os"
	"path/filepath"
	_ "github.com/glebarez/go-sqlite"
)

func ReadMP3Files(directory string) ([]string, map[string][]byte, error) {
	mp3Blobs := make(map[string][]byte)
	var mp3Files []string

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".mp3" {

			data, err := os.ReadFile(path)
			mp3Files = append(mp3Files, path)
			if err != nil {
				return err
			}

			mp3Blobs[path] = data
		}

		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	return mp3Files, mp3Blobs, nil
}
