package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/sync/errgroup"
)

func main() {
	tok := os.Getenv("GITHUB_ACCESS_TOKEN")
	_ = os.Remove("/root/.gitconfig")

	err := os.WriteFile(
		"/root/.gitconfig",
		[]byte(fmt.Sprintf("[url \"https://%s@github.com/\"]\n        insteadOf = https://github.com/", tok)),
		0644,
	)
	if err != nil {
		panic(err)
	}

	if err := exec.Command("go", "mod", "tidy").Run(); err != nil {
		panic(err)
	}

	goSourcesPath := "cmd"
	binPath := "bin"
	hashPath := ".hash"
	hashFileName := "data"
	applicationsFileName := "applications"

	if err := os.RemoveAll(binPath); err != nil {
		panic(err)
	}

	dirs, err := os.ReadDir(goSourcesPath)
	if err != nil {
		panic(err)
	}

	dirNames := make([]string, 0, len(dirs))
	for _, dir := range dirs {
		if dir.IsDir() {
			dirNames = append(dirNames, dir.Name())
		}
	}

	errG := errgroup.Group{}

	for i := range dirNames {
		errG.Go(func() error {
			cmd := exec.Command(
				"go",
				"build",
				"-o",
				filepath.Join(binPath, dirNames[i]),
				filepath.Join(goSourcesPath, dirNames[i], "main.go"),
			)
			if err := cmd.Run(); err != nil {
				return err
			}
			return nil
		})
	}

	if err := errG.Wait(); err != nil {
		panic(err)
	}

	bins, err := os.ReadDir(binPath)
	if err != nil {
		panic(err)
	}

	binToHash := make(map[string]string)

	for i := range bins {
		if !bins[i].IsDir() {
			binToHash[bins[i].Name()] = getFileHash(filepath.Join(binPath, bins[i].Name()))
		}
	}

	_, err = os.ReadDir(hashPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			panic(err)
		}
		_ = os.MkdirAll(hashPath, os.ModePerm)
	}

	cur := getCurrentHashes(hashPath, hashFileName)

	out := make([]string, 0)

	for i := range binToHash {
		if cur[i] != binToHash[i] {
			out = append(out, i)
		}
	}

	binToHashBytes, err := json.Marshal(binToHash)
	if err != nil {
		panic(err)
	}

	_ = os.WriteFile(filepath.Join(hashPath, applicationsFileName), []byte(strings.Join(out, "\n")), os.ModePerm)
	_ = os.WriteFile(filepath.Join(hashPath, hashFileName), binToHashBytes, os.ModePerm)

	fmt.Println(out)
}

func getFileHash(path string) string {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = file.Close()
	}()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		panic(err)
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func getCurrentHashes(path string, fileName string) map[string]string {
	_, err := os.ReadDir(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			_ = os.Mkdir(path, os.ModePerm)
			return make(map[string]string)
		}
		panic(err)
	}

	file, err := os.Open(filepath.Join(path, fileName))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return make(map[string]string)
		}
		panic(err)
	}

	buf := bytes.NewBuffer(nil)
	if _, err = io.Copy(buf, file); err != nil {
		panic(err)
	}

	if buf.Len() == 0 {
		return make(map[string]string)
	}

	out := make(map[string]string)
	err = json.Unmarshal(buf.Bytes(), &out)
	if err != nil {
		panic(err)
	}
	return out
}
