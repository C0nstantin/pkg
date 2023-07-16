package smtpd

import log "github.com/sirupsen/logrus"

type Envelope interface {
	AddRecipient(rcpt MailAddress) error
	BeginData() error
	Write(line []byte) error
	Close() error
}

type BasicEnvelope struct {
	rcpts []MailAddress
}

func (e *BasicEnvelope) AddRecipient(rcpt MailAddress) error {
	e.rcpts = append(e.rcpts, rcpt)
	return nil
}

func (e *BasicEnvelope) BeginData() error {
	if len(e.rcpts) == 0 {
		return SMTPError("554 5.5.1 Error: no valid recipients")
	}
	return nil
}

func (e *BasicEnvelope) Write(line []byte) error {
	log.Debugf("Line: %q", string(line))
	return nil
}

func (e *BasicEnvelope) Close() error {
	return nil
}
