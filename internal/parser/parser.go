package parser

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
)

// ParsePage парсит страницу с объявлением
// Вытаскивает цену и id объявления
func ParsePage(url string) (err error) {
	response, err := http.Get(url)

	if err != nil {
		return errors.New("")
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		return errors.New("")
	}

	bodyText, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return errors.New("")
	}

	log.Println(bodyText)

	//

	return nil
}
