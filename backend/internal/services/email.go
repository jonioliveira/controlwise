package services

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"

	"github.com/controlwise/backend/internal/config"
)

type EmailService struct {
	cfg config.EmailConfig
}

func NewEmailService(cfg config.EmailConfig) *EmailService {
	return &EmailService{cfg: cfg}
}

func (s *EmailService) SendNotification(to, subject, body string) error {
	return s.send(to, subject, body)
}

func (s *EmailService) SendBudgetSent(to, clientName, budgetNumber string) error {
	subject := "Novo Orçamento Disponível"
	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>Olá %s,</h2>
			<p>Foi criado um novo orçamento para si: <strong>%s</strong></p>
			<p>Por favor, reveja o orçamento e informe-nos se tem alguma questão.</p>
			<br>
			<p>Obrigado,<br>A equipa controlwise</p>
		</body>
		</html>
	`, clientName, budgetNumber)

	return s.send(to, subject, body)
}

func (s *EmailService) SendBudgetApproved(to, managerName, budgetNumber string) error {
	subject := "Orçamento Aprovado"
	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>Olá %s,</h2>
			<p>O orçamento <strong>%s</strong> foi aprovado pelo cliente!</p>
			<p>Pode agora proceder com o início da obra.</p>
			<br>
			<p>Obrigado,<br>A equipa controlwise</p>
		</body>
		</html>
	`, managerName, budgetNumber)

	return s.send(to, subject, body)
}

func (s *EmailService) SendTaskAssigned(to, userName, taskTitle string) error {
	subject := "Nova Tarefa Atribuída"
	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>Olá %s,</h2>
			<p>Foi-lhe atribuída uma nova tarefa: <strong>%s</strong></p>
			<p>Por favor, aceda à plataforma para mais detalhes.</p>
			<br>
			<p>Obrigado,<br>A equipa controlwise</p>
		</body>
		</html>
	`, userName, taskTitle)

	return s.send(to, subject, body)
}

func (s *EmailService) SendPaymentDue(to, clientName, amount, dueDate string) error {
	subject := "Pagamento Pendente"
	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>Olá %s,</h2>
			<p>Este é um lembrete de que tem um pagamento pendente no valor de <strong>€%s</strong>.</p>
			<p>Data de vencimento: <strong>%s</strong></p>
			<br>
			<p>Obrigado,<br>A equipa controlwise</p>
		</body>
		</html>
	`, clientName, amount, dueDate)

	return s.send(to, subject, body)
}

func (s *EmailService) send(to, subject, body string) error {
	// Skip if SMTP not configured
	if s.cfg.SMTPHost == "" || s.cfg.SMTPUser == "" {
		return nil
	}

	auth := smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPassword, s.cfg.SMTPHost)

	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"From: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", to, s.cfg.SMTPFrom, subject, body))

	addr := fmt.Sprintf("%s:%s", s.cfg.SMTPHost, s.cfg.SMTPPort)
	return smtp.SendMail(addr, auth, s.cfg.SMTPFrom, []string{to}, msg)
}

func (s *EmailService) renderTemplate(templateName string, data interface{}) (string, error) {
	tmpl, err := template.New(templateName).Parse(getTemplate(templateName))
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func getTemplate(name string) string {
	// Default templates - could be loaded from files
	templates := map[string]string{
		"welcome": `
			<h2>Bem-vindo ao controlwise!</h2>
			<p>Obrigado por se registar.</p>
		`,
	}
	return templates[name]
}
