package signup

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/common-go/mail"
)

type VerifiedEmailSender struct {
	MailSender     mail.SimpleMailSender
	Config         UserVerifiedConfig
	From           mail.Email
	TemplateLoader mail.TemplateLoader
}

func NewVerifiedEmailSender(mailSender mail.SimpleMailSender, config UserVerifiedConfig, from mail.Email, templateLoader mail.TemplateLoader) *VerifiedEmailSender {
	return &VerifiedEmailSender{mailSender, config, from, templateLoader}
}

func truncatingSprintf(str string, args ...interface{}) string {
	n := strings.Count(str, "%s")
	if n > len(args) {
		n = len(args)
	}
	return fmt.Sprintf(str, args[0:n]...)
}

func (s *VerifiedEmailSender) Send(ctx context.Context, to string, code string, expireAt time.Time, params interface{}) error {
	confirmUrl := s.buildVerifiedUrl(to, code)
	diff := expireAt.Sub(time.Now())
	strDiffMinutes := fmt.Sprintf("%.f", diff.Minutes())
	subject, template, err := s.TemplateLoader.Load(ctx, to)
	if err != nil {
		return err
	}

	content := truncatingSprintf(template,
		confirmUrl, confirmUrl, confirmUrl, strDiffMinutes,
		confirmUrl, confirmUrl, confirmUrl, strDiffMinutes,
		confirmUrl, confirmUrl, confirmUrl, strDiffMinutes,
		confirmUrl, confirmUrl, confirmUrl, strDiffMinutes,
		confirmUrl, confirmUrl, confirmUrl, strDiffMinutes,
		confirmUrl, confirmUrl, confirmUrl, strDiffMinutes,
		confirmUrl, confirmUrl, confirmUrl, strDiffMinutes,
		confirmUrl, confirmUrl, confirmUrl, strDiffMinutes,
		confirmUrl, confirmUrl, confirmUrl, strDiffMinutes,
		confirmUrl, confirmUrl, confirmUrl, strDiffMinutes)

	toMail := params.(string)
	mailTo := []mail.Email{{Address: toMail}}
	mailData := mail.NewSimpleHtmlMail(s.From, subject, mailTo, nil, content)
	return s.MailSender.Send(*mailData)
}

func (s *VerifiedEmailSender) buildVerifiedUrl(userId string, code string) string {
	var strPort string
	if s.Config.Port == 80 || s.Config.Port == 443 {
		strPort = ""
	} else {
		strPort = ":" + fmt.Sprint(s.Config.Port)
	}

	var strHttp string
	if s.Config.Secure {
		strHttp = "https"
	} else {
		strHttp = "http"
	}

	return strHttp + "://" + s.Config.Domain + strPort + s.Config.Action + "/" + userId + "/" + code
}
