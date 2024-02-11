package main

import (
	"net/smtp"
	"os"
	"strings"
	"testing"
)

var smtpSendMail func(addr string, a smtp.Auth, from string, to []string, msg []byte) error

type mockSMTP struct {
	sentMail chan struct {
		from, to, subject, body string
	}
}

func (m *mockSMTP) SendMail(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
	headers, body := parseEmailMessage(msg)
	m.sentMail <- struct {
		from, to, subject, body string
	}{from, to[0], headers["Subject"], body}
	return nil
}

func parseEmailMessage(msg []byte) (map[string]string, string) {
	headers := make(map[string]string)
	body := ""

	lines := strings.Split(string(msg), "\n")
	for i, line := range lines {
		if i == 0 || len(line) == 0 {
			continue // Skip the first line and empty lines
		}

		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		} else {
			body = strings.Join(lines[i:], "\n")
			break
		}
	}

	return headers, body
}

func TestSendEmail(t *testing.T) {
	mock := &mockSMTP{sentMail: make(chan struct {
		from, to, subject, body string
	}, 1)}

	smtpSendMail = mock.SendMail
	defer func() { smtpSendMail = smtp.SendMail }()

	subject := "Test Subject"
	body := "Test Body"

	err := sendEmail(subject, body)
	if err != nil {
		t.Fatalf("Error sending email: %v", err)
	}

	sentEmail := <-mock.sentMail
	if sentEmail.subject != subject {
		t.Errorf("Expected subject %q, got %q", subject, sentEmail.subject)
	}

	if sentEmail.body != body {
		t.Errorf("Expected body %q, got %q", body, sentEmail.body)
	}
}

func TestMain(m *testing.M) {

	exitCode := m.Run()

	os.Exit(exitCode)
}
