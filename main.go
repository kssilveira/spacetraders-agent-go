package main

import (
	"log"
	"net/http"

	"github.com/kssilveira/spacetraders-agent-go/agent"
	"github.com/kssilveira/spacetraders-agent-go/client"
	"github.com/kssilveira/spacetraders-agent-go/token"
)

func main() {
	if err := all(); err != nil {
		log.Fatalf("err %v", err)
	}
}

func all() error {
	accountToken, err := token.GetAccountToken()
	if err != nil {
		return err
	}
	agentToken, err := token.GetAgentToken()
	if err != nil {
		return err
	}
	agent := agent.Agent{Client: client.Client{AccountToken: accountToken, AgentToken: agentToken, Client: &http.Client{}}}
	return agent.All()
}
