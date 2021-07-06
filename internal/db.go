package internal

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"math"
	"sync"
	"time"
)

/*
	Структура БД
*/
type DataBase struct {
	db  *sqlx.DB
	err error
	mx  sync.Mutex
	pr  Parse /// парсер
}

/*
	Структура записи в БД
*/
type curs struct {
	Name1,Name2	string
	Data		float64
	Date		time.Time
}

/*
	Интерфейс для взаимодейстия с БД курса валют
*/
type Connecter interface {
	Connect(dbname, user, password string)
	Disconnect()
	Check()
	Update()
	Insert(name1, name2 string) (string, error)
	Read(name1, name2 string) (*curs, error)
}

/*
	Функция проверки ошибок
*/
func (db *DataBase) Check()   {
	if db.err != nil {
		fmt.Println(db.err)
	}
}

/*
	Функция соеденения к базе данных
	dbname user password - данные, для подключения к БД
*/
func (db *DataBase) Connect(dbname, user, password string)  {
	if db.db != nil {
		fmt.Println("БД уже подключена ")
		return
	}
	conn := fmt.Sprintf("host=localhost port=5432 sslmode=disable user=%s password=%s dbname=%s",
					user,password,dbname)
	db.mx.Lock()
	db.db, db.err = sqlx.Connect("postgres",conn)
	db.mx.Unlock()
}

/*
	Функция закрытия БД
 */
func (db *DataBase) Disconnect() {
	if db.db == nil {
		fmt.Println("БД не подключена")
		return
	}
	db.mx.Lock()
	db.db.Close()
	db.mx.Unlock()
}


/*
	Метод проверки существования записи в БД
*/
func (db *DataBase) checkExist(name1, name2 string) bool {
	query := "select exists(select * from cursdata where name1=$1 AND name2=$2)"
	var check bool
	db.mx.Lock()
	defer db.mx.Unlock()
	exec := db.db.QueryRow(query,name1,name2)
	db.err = exec.Scan(&check)
	return check
}

/*
	Метод создания записи в БД
*/
func (db *DataBase) Insert(name1, name2 string) (string, error) {
	check := db.checkExist(name1,name2)
	val := db.pr.GetValues()
	var data float64
	var val1 = math.Round(val.Val[name1].Value / float64(val.Val[name1].Nominal) * 100) / 100
	var val2 = math.Round(val.Val[name2].Value / float64(val.Val[name2].Nominal) * 100) / 100
	/// Парсим сайт с валютами соотношения к рублю
	if name1 == "RUB" {
		data = 1 / val2
	} else if name2 == "RUB" {
		data = val1
	} else {
		data = val1 / val2
	}
	db.mx.Lock()
	defer db.mx.Unlock()
	if !check { /// добавить запись
		_, db.err = db.db.Exec("INSERT INTO cursdata VALUES ($1, $2, $3, $4) ", name1, name2, data, val.Timestamp)
		return fmt.Sprintf("1 %s = %g %s",name1,data,name2) ,nil
	} else {
		return "", errors.New("Запись уже существует")
	}
}

/*
	Метод обновления в БД записи
	вызывается в горутине для обновления
*/
func (db *DataBase) Update() {
	val := db.pr.GetValues()
	query := "SELECT * FROM cursdata"
	var rows *sql.Rows
	var elem curs
	var val1 float64
	var val2 float64
	db.mx.Lock()
	defer db.mx.Unlock()
	rows, db.err = db.db.Query(query)
	for rows.Next() { /// итерация по всем записям
		db.err = rows.Scan(&elem.Name1,&elem.Name2,&elem.Data,&elem.Date)
		val1 = val.Val[elem.Name1].Value / float64(val.Val[elem.Name1].Nominal)
		val2 = val.Val[elem.Name2].Value / float64(val.Val[elem.Name2].Nominal)
		elem.Date = val.Timestamp
		if elem.Name1 == "RUB" {
			elem.Data = 1 / val2
		} else if elem.Name2 == "RUB" {
			elem.Data = val1
		} else {
			elem.Data = val1 / val2
		}
		elem.Data = math.Round(elem.Data * 100) / 100
		/// Запись в БД обновленных данных
		query = "UPDATE cursdata SET data = $3,date = $4 WHERE name1 = $1 AND name2 = $2"
		_, db.err = db.db.Exec(query,elem.Name1,elem.Name2, elem.Data, elem.Date)
	}
}

/*
	Метод получения значения курса
*/
func (db *DataBase) Read(name1, name2 string) (*curs, error) {
	var data = new(curs)
	db.mx.Lock()
	db.err = db.db.Get(data,"SELECT * FROM cursdata WHERE name1 = $1 AND name2 = $2",name1,name2)
	db.mx.Unlock()
	if data.Name1 != "" {
		db.err = errors.New("NO data")
		return data, nil
	} else {
		return nil, errors.New("Запись не найдена")
	}
}




