package main

import (
	"bytes"
	"fmt"
	"github.com/mhale/smtpd"
	"github.com/spf13/viper"
	"io"
	"log"
	"net"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"regexp"
	"strings"
)

var targetAddress string
var targetPort string
var plusAddressing string

func mailHandler(_ net.Addr, from string, to []string, data []byte) error {
	newSender := ""
	var enveloppeTo []string
	log.Printf("from: %s", from)
	msg, _ := mail.ReadMessage(bytes.NewReader(data))
	// loop through headers
	for h, v := range msg.Header {
		for _, value := range v {
			if strings.Compare("received", strings.ToLower(h)) == 0 {
				re := regexp.MustCompile(`Authenticated sender: ([^\)[:space:]]*)`)
				sender := re.FindStringSubmatch(value)
				if sender != nil {
					if len(sender) > 1 {
						log.Printf("sender ==> %s", sender[1])
						newSender = sender[1]
					}
				}
			} else if strings.Compare("To", strings.ToLower(h)) == 0 {
				re := regexp.MustCompile(`^(.*)(` + plusAddressing + `)?@(.*)$`)
				rcpt := re.FindStringSubmatch(value)
				if rcpt != nil {
					if len(rcpt) > 1 {
						log.Printf("sender ==> %s", rcpt[1]+"@"+rcpt[3])
						enveloppeTo = append(enveloppeTo, value)
						value = rcpt[1] + "@" + rcpt[3]
					}
				}
			}
		}
	}
	if newSender != "" {
		log.Printf("changing sender From: %s", newSender)
		// overwrite From header
		textproto.MIMEHeader(msg.Header).Set("From", newSender)
		textproto.MIMEHeader(msg.Header).Set("X-Sender-From", from)
	} else {
		newSender = from
	}
	if len(enveloppeTo) > 0 {
		textproto.MIMEHeader(msg.Header).Del("X-Enveloppe-To")
		for _, r := range enveloppeTo {
			textproto.MIMEHeader(msg.Header).Add("X-Enveloppe-To", r)
		}
	}
	var mailContent string
	for h, v := range msg.Header {
		for _, value := range v {
			mailContent += h + ": " + value + "\r\n"
		}
	}
	log.Printf("%s", mailContent)
	mailContent += "\r\n"
	b, _ := io.ReadAll(msg.Body)
	mailContent += string(b)
	log.Printf("Trying to send mail to %s:%s, %s => %v", targetAddress, targetPort, newSender, to)
	return smtp.SendMail(targetAddress+":"+targetPort, nil, newSender, to, []byte(mailContent))

}

func main() {
	viper.SetConfigName("smtp-gw.yml")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc")
	viper.AddConfigPath("$HOME/.config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	address := "127.0.0.1"
	if viper.IsSet("address") {
		address = viper.GetString("address")
	}
	port := "2525"
	if viper.IsSet("port") {
		port = viper.GetString("port")
	}
	targetAddress = "127.0.0.1"
	if viper.IsSet("target") {
		targetAddress = viper.GetString("target")
	}
	targetPort = "25"
	if viper.IsSet("targetPort") {
		targetPort = viper.GetString("targetPort")
	}
	plusAddressing = "+"
	if viper.IsSet("plus") {
		plusAddressing = viper.GetString("plus")
	}
	fmt.Println("go-simple-smtp-rewriter proxy")
	fmt.Printf("Listening on %s:%s\n", address, port)
	fmt.Printf("Forwarding to %s:%s\n", targetAddress, targetPort)
	err = smtpd.ListenAndServe(address+":"+port, mailHandler, "smtp-gw", "smtp-gw.b2n.fr")
	if err != nil {
		fmt.Println(err)
		return
	}
}
