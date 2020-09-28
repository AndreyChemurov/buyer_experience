package email

import (
	"fmt"
	"os"

	mailjet "github.com/mailjet/mailjet-apiv3-go"
)

// Send отправляет письмо на указанный email,
// если цена из объявления изменилась
func Send(email string, link string) {
	mailjetClient := mailjet.NewMailjetClient(os.Getenv("BX_PUBLIC_API_KEY"), os.Getenv("BX_PRIVATE_API_KEY"))

	messagesInfo := []mailjet.InfoMessagesV31{
		mailjet.InfoMessagesV31{
			From: &mailjet.RecipientV31{
				Email: "andrey.chemurov@gmail.com",
				Name:  "Avito BX Testing",
			},

			To: &mailjet.RecipientsV31{
				mailjet.RecipientV31{
					Email: email,
					Name:  "Subscriber",
				},
			},

			Subject:  "Avito offer subscription",
			TextPart: email,
			HTMLPart: fmt.Sprintf("<h3>Price for <a href='%s'>the offer</a> has just changed. Check it!</h3>", link),
			CustomID: "AvitoBXTest",
		},
	}

	messages := mailjet.MessagesV31{Info: messagesInfo}

	if _, err := mailjetClient.SendMailV31(&messages); err != nil {
		//
	}
}
