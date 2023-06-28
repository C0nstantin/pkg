module gitlab.wm.local/mail/pkg/ws

go 1.19

replace gitlab.wm.local/mail/mail_backend/pkg/transport/rabbitmq => ../transport/rabbitmq

require gitlab.wm.local/mail/mail_backend/pkg/transport/rabbitmq v0.0.0-00010101000000-000000000000

require github.com/rabbitmq/amqp091-go v1.8.1 // indirect
