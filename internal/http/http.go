package http

import (
	"buyer_experience/internal/database"
	"buyer_experience/internal/errors"
	"buyer_experience/internal/parser"
	"buyer_experience/internal/types"
	"encoding/json"
	"net"
	"net/http"
	"regexp"
	"strings"
)

func checkTransportInfo(r *http.Request) (status int, message string, response *types.SubscriberResponse) {
	// Проверить метод
	if r.Method != "POST" {
		return http.StatusMethodNotAllowed, "Method not allowed: use POST", response
	}

	var (
		s         types.Subscriber
		shortLink string = ""
		err       error
	)

	// Проверить валидность json'а
	if err = json.NewDecoder(r.Body).Decode(&s); err != nil {
		return http.StatusInternalServerError, "Invalid JSON format.", response
	}

	email := s.Mail
	link := s.Link

	// Проверить корректность введенных параметров
	if email == "" || link == "" {
		return http.StatusInternalServerError, "Wrong parameter(s)", response
	}

	// Сократить ссылку до "https://www.avito.ru/{id-объявления}"
	if shortLink, err = parser.LinkSimplifier(link); err != nil {
		return http.StatusInternalServerError, err.Error(), response
	}

	if !isEmailValid(email) {
		return http.StatusInternalServerError, "Invalid email", response
	}

	if !isLinkValid(shortLink) {
		return http.StatusInternalServerError, "Invalid link", response
	}

	response = &types.SubscriberResponse{
		Mail: email,
		Link: shortLink,
	}

	return http.StatusOK, "", response
}

// Проверить мейл по регулярке
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

// Проверить, что до объявления можно "достучаться"
func isLinkValid(l string) bool {
	client := &http.Client{}

	_, err := client.Get(l)

	if err != nil {
		return false
	}

	return true
}

// Subscribe оформляет подписку на объявление
// Аргументы:
//	mail: email, на который присылать уведомления
//	link: ссылка на объявление Авито
// Возвращаемые значения:
//	статус OK 200, если все в порядке
func Subscribe(w http.ResponseWriter, r *http.Request) {
	var (
		price int
		err   error

		userInfo *types.SubscriberResponse

		responseJSON []byte
		status       int
		message      string
	)

	// Проверить валидность данных
	if status, message, userInfo = checkTransportInfo(r); status != http.StatusOK {
		responseJSON = errors.ErrorType(status, message)

		w.WriteHeader(status)
		w.Write(responseJSON)

		return
	}

	// Проверить есть ли хотя бы одна подписка на объявление
	price, err = database.CheckSubscriptionExists(userInfo.Link)

	if price == -1 { // Если подписки нет

		// Распарсить страницу и получить цену за объявление
		if price, err = parser.ParsePage(userInfo.Link); err != nil {
			responseJSON = errors.ErrorType(http.StatusInternalServerError, err.Error())

			w.WriteHeader(http.StatusInternalServerError)
			w.Write(responseJSON)

			return
		}

	} else if err != nil { // Другая ошибка
		responseJSON = errors.ErrorType(http.StatusInternalServerError, err.Error())

		w.WriteHeader(http.StatusInternalServerError)
		w.Write(responseJSON)

		return
	}

	// Проверить, что пользователь не пытается оформить подписку на одно объявление дважды
	if err = database.DoubleSubscription(userInfo); err != nil {
		responseJSON = errors.ErrorType(http.StatusInternalServerError, err.Error())

		w.WriteHeader(http.StatusInternalServerError)
		w.Write(responseJSON)

		return
	}

	// Создать новую подписку с первым пользователем
	if err = database.CreateNewSubscription(userInfo, price); err != nil {
		responseJSON = errors.ErrorType(http.StatusInternalServerError, err.Error())

		w.WriteHeader(http.StatusInternalServerError)
		w.Write(responseJSON)

		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(types.ResponseOK)

	return
}

// NotFound вызывается, если путь не существует
func NotFound(w http.ResponseWriter, r *http.Request) {
	responseJSON := errors.ErrorType(http.StatusNotFound, "Not found")

	w.WriteHeader(http.StatusNotFound)
	w.Write(responseJSON)

	return
}
