package doveadm

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type CommandRequest struct {
	Command    string `json:"command"`
	Parameters []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"parameters"`
}

type Request [][]CommandRequest

type Response struct {
}
type FolderRequestParams struct {
	User        string   `json:"user"`
	Field       []string `json:"field"`
	MailboxMask []string `json:"mailboxMask"`
}

type Client struct {
	DoveadmServer string
	DoveadmApiKey string
}

func (d *Client) Request(body []byte) ([]byte, error) {
	req, err := http.NewRequest(
		http.MethodPost, urlRequest(d.DoveadmServer, "/doveadm/v1"), bytes.NewBuffer(body),
	)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("Accept", "application/json, text/plain, */*")

	encode := base64.StdEncoding.EncodeToString([]byte(d.DoveadmApiKey))
	req.Header.Add("Authorization", "X-Dovecot-API "+encode)

	c := &http.Client{}
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	resp, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return []byte{}, fmt.Errorf(string(resp))
	}

	return resp, nil
}

func urlRequest(host string, path string) string {
	base, _ := url.Parse(host)
	part, _ := url.Parse(path)
	return base.ResolveReference(part).String()
}

type MailBox struct {
	Guid     string `json:"guid"`
	Mailbox  string `json:"mailbox"`
	Messages string `json:"messages"`
	Unseen   string `json:"unseen"`
	Vsize    string `json:"vsize"`
}

func (d *Client) MailBoxStatus(params FolderRequestParams) ([]MailBox, error) {

	request := [][]interface{}{{
		"mailboxStatus",
		params,
		"tag",
	}}
	b, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("faild parse params folder: %w", err)
	}

	result, err := d.Request(b)
	if err != nil {
		return nil, fmt.Errorf("faild request folder: %w", err)
	}

	var response [][]interface{}
	if err := json.Unmarshal(result, &response); err != nil {
		return nil, err
	}
	var mailboxes []MailBox
	res := response[0]
	if status, ok := res[0].(string); ok {
		switch status {
		case "doveadmResponse":
			if mbox, ok := res[1].([]interface{}); ok {
				for _, val := range mbox {
					if box, ok := val.(map[string]interface{}); ok {
						var folder MailBox
						for tagName, tagVal := range box {
							switch tagName {
							case "mailbox":
								folder.Mailbox = tagVal.(string)
							case "unseen":
								folder.Mailbox = tagVal.(string)
							case "vsize":
								folder.Vsize = tagVal.(string)
							case "messages":
								folder.Messages = tagVal.(string)
							case "guid":
								folder.Guid = tagVal.(string)
							}
						}
						mailboxes = append(mailboxes, folder)
					}
				}
			} else {
				return nil, fmt.Errorf("parse folders error")
			}
		case "error":
			if e, ok := res[1].(map[string]interface{}); ok {
				for key, val := range e {
					switch key {
					case "exitCode":
						exitCode := val.(float64)
						if exitCode == 68 {
							return nil, fmt.Errorf("folder_not_found")
						} else {
							return nil, fmt.Errorf("folder exit code %g error", exitCode)
						}
					}
				}
			}
		default:
			return nil, fmt.Errorf("unknown response error")
		}
	}
	return mailboxes, nil
}
