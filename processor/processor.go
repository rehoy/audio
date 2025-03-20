package processor

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

func EncodeToMP3(inputAudio []byte, outputFile string) error {
    // Create a temporary file to store the input audio data
    tempInputFile, err := os.CreateTemp("", "input-*")
    if err != nil {
        return fmt.Errorf("failed to create temporary input file: %w", err)
    }
    defer os.Remove(tempInputFile.Name())

    // Write the input audio data to the temporary file
    if _, err := tempInputFile.Write(inputAudio); err != nil {
        return fmt.Errorf("failed to write to temporary input file: %w", err)
    }
    tempInputFile.Close()

    // Use ffmpeg to encode the audio to MP3
    cmd := exec.Command("ffmpeg", "-y", "-i", tempInputFile.Name(), "-vn", "-ar", "44100", "-ac", "2", "-b:a", "192k", outputFile)
    var stderr bytes.Buffer
    cmd.Stderr = &stderr

    if err := cmd.Run(); err != nil {
        return fmt.Errorf("ffmpeg error: %s, %w", stderr.String(), err)
    }

    return nil
}
