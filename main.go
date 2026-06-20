package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/kssilveira/spacetraders-agent-go/game"
)

func main() {
	if err := all(); err != nil {
		log.Fatalf("err %v", err)
	}
}

func all() error {
	token, err := getToken()
	if err != nil {
		return err
	}
	game := game.Game{Token: token, Client: &http.Client{}}
	return game.All()
}

func getToken() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	tokenPath := filepath.Join(homeDir, ".spacetraders_token")
	tokenBytes, err := os.ReadFile(tokenPath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(tokenBytes)), nil
}
