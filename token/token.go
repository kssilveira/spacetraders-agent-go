package token

import (
	"os"
	"path/filepath"
	"strings"
)

func GetAccountToken() (string, error) {
	return getToken("account")
}

func GetAgentToken() (string, error) {
	return getToken("agent")
}

func getToken(name string) (string, error) {
	tokenPath, err := tokenPath(name)
	if err != nil {
		return "", err
	}
	tokenBytes, err := os.ReadFile(tokenPath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(tokenBytes)), nil
}

func tokenPath(name string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".spacetraders_"+name), nil
}

func SetAgentToken(token string) error {
	tokenPath, err := tokenPath("agent")
	if err != nil {
		return err
	}
	if err := os.WriteFile(tokenPath, []byte(token), 0644); err != nil {
		return err
	}
	return nil
}
