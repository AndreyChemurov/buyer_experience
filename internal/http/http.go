package http

import (
	"buyer_experience/internal/errors"
	"buyer_experience/internal/parser"
	"buyer_experience/internal/types"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"regexp"
	"strings"
)

// PathHandler - функция обарботчик путей
func PathHandler() {
	http.HandleFunc("/subscribe", subscribe)

	http.HandleFunc("/", notFound)
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func checkTransportInfo(r *http.Request) (status int, message string) {
	// Проверить метод
	if r.Method != "POST" {
		return 405, "Method not allowed: use POST."
	}

	// Проверить валидность json'а
	decoder := json.NewDecoder(r.Body)

	var s types.Subscriber

	if err := decoder.Decode(&s); err != nil {
		return 500, "Invalid JSON format."
	}

	email := s.Mail
	link := s.Link

	// Проверить корректность введенных параметров
	if email == "" || link == "" {
		return 500, "Wrong parameters."
	}

	if !isEmailValid(email) {
		return 500, "Invalid email."
	}

	if !isLinkValid(link) {
		return 500, "Invalid link."
	}

	types.Mail = email
	types.Link = link

	return 200, ""
}

func isEmailValid(e string) bool {
	var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

	if len(e) < 3 && len(e) > 254 {
		return false
	}

	if !emailRegex.MatchString(e) {
		return false
	}

	parts := strings.Split(e, "@")

	mx, err := net.LookupMX(parts[1])

	if err != nil || len(mx) == 0 {
		return false
	}

	return true
}

func isLinkValid(l string) bool {
	client := &http.Client{}

	_, err := client.Get(l) // https://www.avito.ru тоже работает

	if err != nil {
		return false
	}

	return true
}

func subscribe(w http.ResponseWriter, r *http.Request) {
	// Проверить валидность данных
	if status, message := checkTransportInfo(r); status != 200 {
		responseJSON := errors.ErrorType(status, message)

		w.WriteHeader(status)
		w.Write(responseJSON)

		return
	}

	// Верифицировать мейл
	// if message, err := email.Verification(types.Mail); err != nil {
	// 	responseJSON := errors.ErrorType(500, message)

	// 	w.WriteHeader(500)
	// 	w.Write(responseJSON)

	// 	return
	// }

	// Распарсить страницу
	if err := parser.ParsePage(types.Link); err != nil {
		responseJSON := errors.ErrorType(500, err.Error())

		w.WriteHeader(500)
		w.Write(responseJSON)

		return
	}

	// Оформить подписку (бд)
}

func notFound(w http.ResponseWriter, r *http.Request) {
	responseJSON := errors.ErrorType(404, "Not found.")

	w.WriteHeader(http.StatusNotFound)
	w.Write(responseJSON)

	return
}
