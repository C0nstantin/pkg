package types

import (
	`time`
)

type DovecotEvent struct {
	Event     string    `json:"event"`
	Hostname  string    `json:"hostname"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Fields    struct {
		User           string `json:"user"`
		Mailbox        string `json:"mailbox,omitempty"`
		CmdName        string `json:"cmd_name,omitempty"`
		CmdInputName   string `json:"cmd_input_name,omitempty"`
		CmdArgs        string `json:"cmd_args,omitempty"`
		CmdHumanArgs   string `json:"cmd_human_args,omitempty"`
		MailFrom       string `json:"mail_from,omitempty"`
		RcptTo         string `json:"rcpt_to,omitempty"`
		MessageSubject string `json:"message_subject,omitempty"`
		MessageId      string `json:"message_id,omitempty"`
		MessageFrom    string `json:"message_from,omitempty"`
		Success        string `json:"success,omitempty"`
	} `json:"fields"`
}

type Message struct {
	Action   string       `json:"action"`
	User     string       `json:"user"`
	Mailbox  string       `json:"mailbox"`
	Category string       `json:"category"`
	Source   DovecotEvent `json:"source"`
}

type SmtpMessage struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Wmid    string `json:"wmid"`
	Message string `json:"message"`
}

type DomainMessage struct {
	HostName string  `json:"host_name"`
	User     string  `json:"user"`
	Domain   *Domain `json:"domain,omitempty"`
	Action   string  `json:"action"`
	Result   []byte  `json:"result"`
}

type Domain struct {
	Id            int    `json:"id"`
	ConfirmStatus string `json:"confirm_status"`
	ConfirmMethod string `json:"confirm_method"`
	Guid          string `json:"guid"`
}
