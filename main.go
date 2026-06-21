package main

import (
	"log"
	"net/http"

	"github.com/kssilveira/spacetraders-agent-go/agent"
	"github.com/kssilveira/spacetraders-agent-go/client"
)

func main() {
	if err := all(); err != nil {
		log.Fatalf("err %v", err)
	}
}

func all() error {
	accountToken, err := agent.GetAccountToken()
	if err != nil {
		return err
	}
	agentToken, err := agent.GetAgentToken()
	if err != nil {
		return err
	}
	agent := agent.Agent{Client: client.Client{AccountToken: accountToken, AgentToken: agentToken, Client: &http.Client{}}}
	return agent.All()
}
