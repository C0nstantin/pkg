package ws

import (
	"encoding/json"
	"errors"
	"time"

	"gitlab.wm.local/mail/pkg/transport/rabbitmq"
)

type Pusher interface {
	Push(key, eType string, sender string, body interface{}) error
	Close()
}

type PusherImpl struct {
	Sender rabbitmq.Sender
}

func (e *PusherImpl) Close() {
	e.Sender.Close()
}

// Push send message type wsType with body, for set user use key
func (e *PusherImpl) Push(key, wsType, sender string, body interface{}) error {
	if key == "" {
		return errors.New("Key can't be empty")
	}
	if body == nil || wsType == "" || sender == "" {
		return errors.New("message format not supported")
	}

	event := &struct {
		Type   string
		Sender string
		Date   time.Time
		Body   interface{}
	}{
		Type:   wsType,
		Sender: sender,
		Date:   time.Now(),
		Body:   body,
	}
	evn, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return e.Sender.Send(key, evn)
}
