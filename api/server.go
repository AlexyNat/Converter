package api

import (
	"fmt"
	"github.com/AlexyNat/Converter/internal"
	"github.com/gin-gonic/gin"
	"math"
	"net/http"
	"strconv"
	"time"
)


var db internal.Connecter
var pr internal.Parser

/*
	Функция для связи с БД
*/
func startDb() {
	db = new(internal.DataBase)
	pr = new(internal.Parse)
	pr.GetValues()
	db.Connect("cursval","alex","alex1234")
	db.Update() /// обновление при запуске
	go update() /// обновляет существующие в БД записи раз в N минут
}

/*
	Функция запуска сервера
*/
func Start() {
	startDb()
	defer db.Disconnect()
	route := gin.Default()

	/// Создание записи в БД
	route.POST("/api/create", create)

	/// конвертация валюты
	route.GET("/api/convert", convert)

	err := route.Run()
	if err != nil {
		fmt.Println(err)
	}
}

/// обработчик записи
func create(context *gin.Context) {
	val1 := context.Request.PostFormValue("param1")
	val2 := context.Request.PostFormValue("param2")
	if val1 == val2 {
		context.JSON(http.StatusOK,"Валюта одна и та же")
	}
	str, err := db.Insert(val1, val2)
	if err != nil {
		context.JSON(http.StatusOK,err.Error())
	} else {
		context.JSON(http.StatusOK,str)
	}
}

/*
	Обработчик конвертации
*/
func convert(context *gin.Context) {
	data := context.Request.URL.Query()
	if len(data) == 0 {
		context.JSON(http.StatusOK, "Пусто")
	} else {
		name1 := data["param1"]
		name2 := data["param2"]
		value, _ := strconv.ParseFloat(data["value"][0],64)
		val, err := db.Read(name1[0],name2[0])
		if err != nil {
			context.JSON(http.StatusOK, err.Error())
		} else {
			val.Data *= value
			val.Data = math.Round(val.Data * 100) / 100
			context.JSON(http.StatusOK, val)
		}
	}
}

/*
	Функция обновления БД
*/
func update() {
	ticker := time.NewTicker(time.Hour * 6)
	for range ticker.C{
		db.Update()
	}
}