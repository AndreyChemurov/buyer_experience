# Сервис подписки на объявление: Авито Buyer-Experience

## Информация
Приложение работает на порту ```8000```. </br>
Запросы делаются через POST-метод с использованием JSON. </br>
Уведомления отправляются через сервис Mailjet. Некоторые почтовые клиенты могут воспринимать сообщения как спам. </br>
</br>
Я использую Linux Debian.

## Запуск приложения
```bash
git clone https://github.com/AndreyChemurov/buyer_experience.git
cd buyer_experience/
[sudo] docker-compose up
```

## Архитектура
На сервер отправляется POST-запрос с json'ом, в котором содержатся ссылка на объявление и email. На серверной части: 1) Проверить, существует ли уже подписка на данное объявление. Если нет, то парсится страница с объявлением. Если да, то цена берется из бд. 2) Проверить, что юзер не пытается оформить подписку с одного и того же мейла на одно и то же объявление. 3) Занести данные в бд. В параллельной горутине из бд селектятся данные для проверки на изменение цены. По каждой уникальной ссылке парсится страница и сравнивается цена с той, что хранится в бд. Если цена изменилась, то на email отправляется уведомление через сервис Mailjet. </br>
![Untitled Diagram](https://user-images.githubusercontent.com/58785926/94446925-ebb0be00-01b1-11eb-8191-13384ae6425b.jpg)

### Подписка на изменение цены
```go
// CreateNewSubscription создает новую подписку
func CreateNewSubscription(user *types.SubscriberResponse, price int) (err error) {
	if db, err = openDB(); err != nil {
		return err
	}

	defer db.Close()

	var (
		SQLStmt string = `INSERT INTO "subscribers" VALUES (DEFAULT, $1, $2, $3);`
	)

	if _, err = db.Exec(SQLStmt, user.Link, user.Mail, price); err != nil {
		return err
	}

	return nil
}
```
### Отслеживание изменений цены
```go
// CheckPriceChanged селектит инфу из бд для дальнейшего сравнения с объявлением на Авито
func CheckPriceChanged() {
	var (
		SQLStmt string = `SELECT DISTINCT link, user_email, price FROM "subscribers";`
		links   []string
		prices  []int
		emails  []string

		link  string
		price int
		email string
	)

	for {
		if db, err = openDB(); err != nil {
			log.Println(err)
		}

		defer db.Close()

		rows, err := db.Query(SQLStmt)

		if err != nil {
			log.Println(err)
		}

		defer rows.Close()

		for rows.Next() {
			err = rows.Scan(&link, &email, &price)

			if err != nil {
				log.Println(err)

				continue
			}
			links = append(links, link)
			emails = append(emails, email)
			prices = append(prices, price)
		}

		go net.RequestPrices(links, emails, prices)

		time.Sleep(time.Minute * 2)

		links = nil
	}
}
```
### Отправка уведомления на почту
```go
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
		log.Println(err)
	}
}
```
### Работа с БД
```go
// CreateTables создает таблицу БД "subscribers",
// если она еще не была создана.
func CreateTables() (err error) {
	if db, err = openDB(); err != nil {
		return err
	}

	defer db.Close()

	_, err = db.Exec(types.CreateTables)

	if err != nil {
		return err
	}

	return nil
}

// CheckSubscriptionExists проверяет, что подписка на объявление уже существует.
//	Если да, то возвращается цена, чтобы не парсить ее заново
func CheckSubscriptionExists(link string) (price int, err error) {
	if db, err = openDB(); err != nil {
		return -1, err
	}

	defer db.Close()

	var (
		SQLStmt string   = `SELECT price FROM "subscribers" WHERE link=$1;`
		row     *sql.Row = db.QueryRow(SQLStmt, link)
	)

	if err = row.Scan(&price); err != nil {
		return -1, err
	}

	return price, nil
}

// DoubleSubscription проверяет, что пользователь не пытается
//	оформить подписку на одно и то же объявление
func DoubleSubscription(user *types.SubscriberResponse) (err error) {
	if db, err = openDB(); err != nil {
		return err
	}

	defer db.Close()

	var (
		SQLStmt string   = `SELECT id FROM "subscribers" WHERE user_email=$1 AND link=$2;`
		row     *sql.Row = db.QueryRow(SQLStmt, user.Mail, user.Link)
		id      int      = -1
	)

	err = row.Scan(&id)

	if id != -1 { // Пользователь пытается подписаться дважды на одно объявление
		return errors.New("User already has subscription for this offer")
	} else if err == sql.ErrNoRows { // Нет ни одного пользователя
		return nil // -> можно подписаться
	} else {
		return err
	}
}
```
