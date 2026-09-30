package mailer

import "github/OfrenDialsa/go-gin-starter/internal/dto"

type MailerService interface {
	Send(req dto.MailerRequest) (string, error)
}
