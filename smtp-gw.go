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

func mailHandler(remoteAddr net.Addr, from string, to []string, data []byte) error {
	newSender := ""
	log.Printf("from: %s", from)
	msg, _ := mail.ReadMessage(bytes.NewReader(data))
	// loop through headers
	for h, v := range msg.Header {
		// log.Printf("header: %s", h)
		for _, value := range v {
			// log.Printf("\tvalue: %s", value)
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
				
			}
		}
	}
	if newSender != "" {
		log.Printf("changing sender From: %s", newSender)
		// overwrite From header
		textproto.MIMEHeader(msg.Header).Set("From", newSender)
		textproto.MIMEHeader(msg.Header).Set("X-Sender-From", from)
		var mailContent string
		for h, v := range msg.Header {
			for _, value := range v {
				mailContent += h + ": " + value + "\r\n"
			}
		}
		log.Printf("%s", mailContent)
		mailContent += "\r\n"
		b, _ := io.ReadAll(msg.Body)
		log.Printf("adding body: %q", b)
		mailContent += string(b)
		log.Printf("Trying to send mail: %s", mailContent)
		return smtp.SendMail(targetAddress+":"+targetPort, nil, newSender, to, []byte(mailContent))
	} else {
		// send as is
		log.Printf("mail sent as-is: %s => %s", from, to[0])
		return smtp.SendMail(targetAddress+":"+targetPort, nil, from, to, data)
	}
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
	fmt.Println("SMTP-GW from rewriter proxy")
	fmt.Printf("Listening on %s:%s\n", address, port)
	fmt.Printf("Forwarding to %s:%s\n", targetAddress, targetPort)
	smtpd.ListenAndServe(address+":"+port, mailHandler, "smtp-gw", "smtp-gw.b2n.fr")
}
