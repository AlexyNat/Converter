package internal

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

/// Получение данных с сайта, заполнение структуры с данными

/*
	Тип Парсер содержит данные курса валют
	curs - содержит все курсы валюты к рублю
	err - ошибки
 */
type Parse struct {
	err  error
	curs Valutes
}

/*
	Интерфейс для взаимодействия с парсером
	получить карту значений валют
 */
type Parser interface {
	GetValues() *Valutes
	Check()
}

// Valutes JSON
type Valutes struct {
	Timestamp time.Time         `json:"Timestamp"`
	Val       map[string]Valute `json:"Valute"`
}

type Valute struct {
	Id			string 	`json:"Id"`
	CharCode	string	`json:"CharCode"`
	Nominal		int		`json:"Nominal"`
	Name		string 	`json:"Name"`
	Value		float64	`json:"Value"`
}

/*
	Функция проверки ошибок
*/
func (p *Parse) Check() {
	if p.err != nil {
		fmt.Println(p.err)
	}
}

/*
	Метод получения данных валют со станицы
*/
func (p *Parse) parse() {
	var resp *http.Response
	resp, p.err = http.Get("https://www.cbr-xml-daily.ru/daily_json.js")
	defer resp.Body.Close()
	var data []byte
	data, p.err = ioutil.ReadAll(resp.Body)
	p.err = json.Unmarshal(data,&p.curs)
}


/*
	Метод получения данных
	для обновления БД
	[Карта значений]
 */
func (p *Parse) GetValues() *Valutes {
	p.parse()
	return &p.curs
}
