package client

import "net/http"

type Client struct {
	Token  string
	Client *http.Client
}
