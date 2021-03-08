package router

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
)

type MailHandler struct {
	Host string
	Port string
	Username string
	Password string
	SMTPServer string
}


type Mail struct {
	FromEmail string `json:"from_email"`
	ToEmails []string `json:"to_emails"`
	Message string `json:"message"`
	Title string `json:"title"`
}

type ReturnMail struct {
	Result string
}

func (m Mail) Fmt() string {
	return fmt.Sprintf("Mail{Title: %s, Message: %s, From: %s, To: %s}", m.Title, m.Message, m.FromEmail, m.ToEmails)
}

func (mh MailHandler) createMailHandler(w http.ResponseWriter, r *http.Request) {
	var m Mail
	_ = json.NewDecoder(r.Body).Decode(&m)

	msg := []byte(fmt.Sprintf(
		"To: %s\r\n" +
		"Subject: %s\r\n" +
		"\r\n" +
		"%s\r\n", m.ToEmails, m.Title, m.Message))

	log.Println(get_request_id(r), m.Fmt())

	auth := smtp.PlainAuth("", mh.Username, mh.Password, mh.SMTPServer)
	err := smtp.SendMail(fmt.Sprintf("%s:%s", mh.Host, mh.Port), auth, m.FromEmail, m.ToEmails, msg)
	if err != nil {
		log.Println(get_request_id(r), err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ReturnMail{Result: "ok"})
}
