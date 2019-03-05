package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/smtp"
	"strings"
	"time"

	fdk "github.com/fnproject/fdk-go"
)

func main() {
	fdk.Handle(fdk.HandlerFunc(atpDBEventsEmailNotificationHandler))
}

func atpDBEventsEmailNotificationHandler(ctx context.Context, in io.Reader, out io.Writer) {
	log.Println("ATP DB events notification handler invoked on", time.Now())

	var evt OCIEvent
	json.NewDecoder(in).Decode(&evt)
	log.Println("Got OCI cloud event payload")

	eventDetails := evt.Data
	log.Println("Event data", eventDetails)

	fnCtx := fdk.GetContext(ctx)

	config := fnCtx.Config()
	username := config["OCI_EMAIL_DELIVERY_USER_OCID"]
	password := config["OCI_EMAIL_DELIVERY_USER_PASSWORD"]
	ociSMTPServer := config["OCI_EMAIL_DELIVERY_SMTP_SERVER"]
	approvedOCIEmailDeliverySender := config["OCI_EMAIL_DELIVERY_APPROVED_SENDER"]
	emailRecepientAddress := config["EMAIL_NOTIFICAITON_RECEPIENT_ADDRESS"]

	log.Println("OCI_EMAIL_DELIVERY_USER_OCID", username)
	log.Println("OCI_EMAIL_DELIVERY_USER_PASSWORD", password)
	log.Println("OCI_EMAIL_DELIVERY_SMTP_SERVER", ociSMTPServer)
	log.Println("OCI_EMAIL_DELIVERY_APPROVED_SENDER", approvedOCIEmailDeliverySender)
	log.Println("EMAIL_NOTIFICAITON_RECEPIENT_ADDRESS", emailRecepientAddress)

	response := sendEmailNotification(username, password, ociSMTPServer, approvedOCIEmailDeliverySender, emailRecepientAddress, eventDetails)
	log.Println("Response", response)
	out.Write([]byte(response))
}

func sendEmailNotification(username, password, ociSMTPServer, approvedOCIEmailDeliverySender, emailRecepientAddress string, eventDetails Data) string {
	log.Println("sending email notification")

	auth := smtp.PlainAuth("", username, password, ociSMTPServer)
	to := strings.Split(emailRecepientAddress, ",")
	//to := []string{emailRecepientAddress}
	subject := "ATP Database instance " + eventDetails.DisplayName + " in status " + eventDetails.LifecycleState
	body := subject + "\n" + "Instance OCID: " + eventDetails.ID
	msg := []byte("To: " + emailRecepientAddress + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body + "\r\n")

	log.Println("Message ", string(msg))
	err := smtp.SendMail(ociSMTPServer+":25", auth, approvedOCIEmailDeliverySender, to, msg)
	if err != nil {
		log.Println("Error sending notification email", err.Error())
		return "Error sending notification email " + err.Error()
	}

	log.Println("Notification email sent successfully!")
	return "Notification email sent successfully!"
}

//OCIEvent ...
type OCIEvent struct {
	CloudEventsVersion string `json:"cloudEventsVersion"`
	EventID            string `json:"eventID"`
	EventType          string `json:"eventType"`
	Source             string `json:"source"`
	EventTypeVersion   string `json:"eventTypeVersion"`
	EventTime          string `json:"eventTime"`
	SchemaURL          string `json:"schemaURL"`
	ContentType        string `json:"contentType"`
	Extensions         `json:"extensions"`
	Data               Data `json:"data"`
}

//Extensions - "extension" attribute in events JSON payload
type Extensions struct {
	CompartmentId string `json:"compartmentId"`
}

//Data - represents (part of) the event data
type Data struct {
	ID             string `json:"id"`
	LifecycleState string `json:"lifecycleState"`
	DisplayName    string `json:"displayName"`
}
