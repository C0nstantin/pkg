package smtp

import (
	"fmt"
	"net/smtp"
)

type Sender interface {
	SendMessage(from, to, data string) error
}

type SenderImpl struct {
	Host string
}

func (s *SenderImpl) SendMessage(from, to, data string) error {
	err := s.sendMail(s.Host, from, to, data)
	if err != nil {
		return err
	}
	return nil
}

func (s *SenderImpl) sendMail(host, from, to, data string) error {
	c, err := smtp.Dial(host)
	
	if err != nil {
		return fmt.Errorf("connection dial error: %w", err)
	}
	// Set the sender and recipient first
	if err := c.Mail(from); err != nil {
		return fmt.Errorf("send command from error: %w", err)
	}
	if err := c.Rcpt(to); err != nil {
		return fmt.Errorf("send command rcpt error: %w", err)
	}
	
	// Send the email body.
	wc, err := c.Data()
	if err != nil {
		return fmt.Errorf("create writer error: %w", err)
	}
	_, err = fmt.Fprintf(wc, "%s", data)
	if err != nil {
		return fmt.Errorf("send message data error: %w", err)
	}
	err = wc.Close()
	if err != nil {
		return fmt.Errorf("close writer error: %w", err)
	}
	
	// Send the QUIT command and close the connection.
	err = c.Quit()
	if err != nil {
		return fmt.Errorf("send command quit error: %w", err)
	}
	return nil
}
