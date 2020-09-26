package parser

import (
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ParsePage парсит страницу с объявлением
// Возвращает цену из объявления
func ParsePage(url string) (intPrice int, err error) {
	response, err := http.Get(url)

	if err != nil {
		return -1, errors.New("Internal server error: link is unreachable")
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return -1, errors.New("Parse page error: status code " + strconv.Itoa(response.StatusCode))
	}

	document, err := goquery.NewDocumentFromReader(response.Body)

	if err != nil {
		return -1, errors.New("Internal server error: cannot parse page")
	}

	price := ""

	document.Find(".js-item-price").Each(func(i int, s *goquery.Selection) {
		price, _ = s.Attr("content")
	})

	intPrice, err = strconv.Atoi(price)

	if err != nil {
		return -1, errors.New("Internal server error: cannot convert price to int")
	}

	return intPrice, nil
}

// LinkSimplifier сокращает ссылку объявления до "https://www.avito.ru/{id-объявления}"
// Это нужно для того, чтобы, если продавец изменит название объявления,
//	до него все равно можно было бы "достучаться"
func LinkSimplifier(longURL string) (shortURL string, err error) {
	// Проверить, что передается ссылка с Авито
	if avitoLink := strings.Contains(longURL, "avito.ru"); !avitoLink {
		return "", errors.New("Link does not relate to Avito")
	}

	re := regexp.MustCompile("[0-9]+") // Создать паттерн поиска

	regexNumbers := re.FindAllString(longURL, -1) // Найти все последовательности цифр в ссылке

	// Проверить, что в ссылке есть цифровая последовательность
	if len(regexNumbers) == 0 {
		return "", errors.New("Link does not contain any id")
	}

	itemID := regexNumbers[len(regexNumbers)-1] // Вычленить id объявления (последний элемент)

	shortURL = "https://www.avito.ru/" + itemID // Создать укороченную ссылку

	return shortURL, nil
}
