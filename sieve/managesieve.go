package sieve

import (
	"crypto/tls"
	"fmt"
	"log"

	"go.guido-berhoerster.org/managesieve"
)

type Client struct {
	client *managesieve.Client
}

func (c *Client) Close() {
	err := c.client.Logout()
	if err != nil {
		return
	}
	err = c.client.Close()
	if err != nil {
		return
	}
}

func (c *Client) Get() (string, error) {
	_, name, err := c.client.ListScripts()
	if err != nil {
		return "", err
	}
	if name == "" {
		_, err2 := c.client.PutScript("sieve", "")
		if err2 != nil {
			return "", err2
		}
		err := c.client.ActivateScript("sieve")
		if err != nil {
			return "", err
		}
		name = "sieve"
	}
	script, err := c.client.GetScript(name)
	if err != nil {
		return "", err
	}
	log.Println("Get Script --> ", script)
	return script, nil
}

func (c *Client) Save(script string, force bool) error {
	log.Print("Save script --> ", script)
	_, name, err := c.client.ListScripts()
	if err != nil {
		return err
	}
	warn, err := c.client.CheckScript(script)
	if err != nil {
		return err
	}
	if warn != "" {
		if !force {
			return fmt.Errorf("warning save script  %s", warn)
		}
		log.Println("warning save managesieve script  " + warn)
	}
	_, err = c.client.PutScript(name, script)
	if err != nil {
		return err
	}
	return nil
}

func NewSieveClient(host, port, username, password string) (*Client, error) {
	// Connect to the ManageSieve server.
	c, err := managesieve.Dial(fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		return nil, fmt.Errorf("failed to connect %w", err)
	}

	// Establish a TLS connection.
	tlsConf := &tls.Config{ServerName: host, InsecureSkipVerify: true}
	if err := c.StartTLS(tlsConf); err != nil {
		return nil, fmt.Errorf("failed to start TLS connection %w", err)
	}

	// Authenticate the user using the PLAIN SASL mechanism.
	auth := managesieve.PlainAuth("", username, password, host)
	if err := c.Authenticate(auth); err != nil {
		return nil, fmt.Errorf("failed to authenticate user %s: %w", username, err)
	}
	return &Client{
		client: c,
	}, nil
}
