package net

import (
	"buyer_experience/internal/email"
	"buyer_experience/internal/parser"
	"log"
)

// RequestPrices ...
func RequestPrices(links []string, emails []string, prices []int) {
	var (
		currentPrice int
		err          error
	)

	for i := range links {
		if currentPrice, err = parser.ParsePage(links[i]); err != nil {
			log.Println(err)
		}

		if currentPrice != prices[i] {
			email.Send(emails[i], links[i])
		}
	}
}
