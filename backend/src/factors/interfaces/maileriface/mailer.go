package maileriface

type Mailer interface {
	SendMail(to, from, subject, html, text string) error
}
