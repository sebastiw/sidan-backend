package router

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/smtp"

	ru "github.com/sebastiw/sidan-backend/src/router_util"
)

type MailHandler struct {
	Host     string
	Port     int
	Username string
	Password string
}

type Mail struct {
	FromEmail string   `json:"from_email"`
	ToEmails  []string `json:"to_emails"`
	Message   string   `json:"message"`
	Title     string   `json:"title"`
}

type ReturnMail struct {
	Result string
}

func (m Mail) Fmt() string {
	return fmt.Sprintf("Mail{Title: %s, Message: %s, From: %s, To: %s}", m.Title, m.Message, m.FromEmail, m.ToEmails)
}

func (mh MailHandler) createMailHandler(w http.ResponseWriter, r *http.Request) {
	var m Mail
	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		http.Error(w, "Failed to parse JSON request body", http.StatusBadRequest)
		return
	}

	msg := []byte(fmt.Sprintf(
		"To: %s\r\n"+
			"Subject: %s\r\n"+
			"\r\n"+
			"%s\r\n", m.ToEmails, m.Title, m.Message))

	slog.Debug(ru.GetRequestId(r), m.Fmt())

	auth := smtp.PlainAuth("", mh.Username, mh.Password, mh.Host)
	err = smtp.SendMail(fmt.Sprintf("%s:%d", mh.Host, mh.Port), auth, m.FromEmail, m.ToEmails, msg)
	if err != nil {
		slog.Error(ru.GetRequestId(r), "Send mail error:", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ReturnMail{Result: "ok"})
}
