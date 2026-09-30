package mailer

import (
	"fmt"
	"github/OfrenDialsa/go-gin-starter/config"
	"github/OfrenDialsa/go-gin-starter/internal/dto"

	"github.com/rs/zerolog/log"
	"gopkg.in/gomail.v2"
)

type MailerServiceImpl struct {
	Config       *config.EnvironmentVariable
	MailFrom     string
	MailFromName string
}

func NewMailerService(cfg *config.EnvironmentVariable, mailFrom, mailFromName string) MailerService {
	return &MailerServiceImpl{
		Config:       cfg,
		MailFrom:     mailFrom,
		MailFromName: mailFromName,
	}
}

func (s *MailerServiceImpl) Send(req dto.MailerRequest) (string, error) {
	from := fmt.Sprintf("%s <%s>", s.MailFromName, s.MailFrom)

	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", req.To...)
	m.SetHeader("Subject", req.Subject)
	m.SetBody("text/html", req.Body)

	for _, val := range req.Attachments {
		m.Attach(val)
	}

	d := gomail.NewDialer(
		s.Config.Mail.SMTP.Host,
		s.Config.Mail.SMTP.Port,
		s.Config.Mail.SMTP.User,
		s.Config.Mail.SMTP.Password,
	)

	if err := d.DialAndSend(m); err != nil {
		log.Error().Err(err).Msg("failed to send email via SMTP")
		return "failed", err
	}

	return "success", nil
}
