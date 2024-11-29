package storage

import (
	"context"
	"fmt"
	"time"

	//_ "modernc.org/sqlite"
	// _ "github.com/lib/pq" // Импорт драйвера
	//"github.com/bmaayandexru/go_final_project/tests"
	"github.com/go-pg/pg/v10"
)

/*
type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"` // omitempty
	Title   string `json:"title"`
	Comment string `json:"comment"` // omitempty
	Repeat  string `json:"repeat"`  // omitempty
}
*/
// добавляем итрибуты для pg
type Task struct {
	// ID      string `pg:"id,pk" json:"id"`
	// меняем string на int
	ID      int    `pg:"id,pk" json:"id"`
	Date    string `pg:"date,notnull" json:"date"` // omitempty
	Title   string `pg:"title,notnull" json:"title"`
	Comment string `pg:"comment" json:"comment"` // omitempty
	Repeat  string `pg:"repeat" json:"repeat"`   // omitempty
}

const (
	templ = "20060102"
	limit = 50
)

var (
	err   error
	tasks []Task
)

type TaskStore struct {
	DB *pg.DB
}

func NewTaskStore(db *pg.DB) TaskStore {
	return TaskStore{DB: db}
}

func (ts TaskStore) Add(task Task) error {
	_, err = ts.DB.Model(&task).Insert()
	return err
}

func (ts TaskStore) Delete(id int) error {
	_, err = ts.DB.Model(&tasks).Where("id = ?", id).Delete()
	return err
}

func (ts TaskStore) Find(search string) ([]Task, error) {
	if len(search) == 0 {
		// возвращаем всё если строка пустая
		err = ts.DB.Model(&tasks).Limit(limit).Order("date").Order("title").Select()
		return tasks, err
	}
	// парсим строку на дату
	if date, err := time.Parse("02-01-2006", search); err == nil {
		// дата есть
		err = ts.DB.Model(&tasks).Where("date = ?", date.Format(templ)).Limit(limit).Order("title").Select()
		return tasks, err

	} else {
		// даты нет
		search = "%" + search + "%"
		// тут по ходу надо query *** подтвердилось
		_, err = ts.DB.Query(&tasks, "select * from tasks where upper(title) like upper(?) or upper(comment) like upper(?)", search, search)
		return tasks, err
	}
}

func (ts TaskStore) Get(id int) (Task, error) {
	if err = ts.DB.Model(&tasks).Where("id = ?", id).Select(); err != nil {
		return Task{}, err
	}
	return tasks[0], err
}

func (ts TaskStore) Update(task Task) error {
	// так работает
	_, err = ts.DB.Model(&task).WherePK().Update()
	return err
}

// scheduler заменен на tasks
var schemaSQL string = `
CREATE TABLE IF NOT EXISTS tasks (
    id SERIAL PRIMARY KEY,
    date CHAR(8) NOT NULL, 
    title VARCHAR(32) NOT NULL,
    comment TEXT,
    repeat VARCHAR(128)
);
CREATE INDEX IF NOT EXISTS idx_date ON tasks (date); 
CREATE INDEX IF NOT EXISTS idx_title ON tasks (title); 
`

func InitDBase() (*pg.DB, error) {
	fmt.Println("Init Data Base...")
	// Создание конфигурации для подключения к базе данных
	db := pg.Connect(&pg.Options{
		Addr:     "localhost:5432", // Адрес PostgreSQL сервера
		User:     "postgres",       // Имя пользователя
		Password: "password",       // Пароль
		Database: "dbscheduler",    // Имя базы данных
	})
	// оставляем старый способ, который подключает индексы
	// создание таблицы
	if err := createTable(db, schemaSQL); err != nil {
		return nil, err
	}
	// Проверка соединения
	ctx := context.Background()
	if err := db.Ping(ctx); err != nil {
		// ошибка подключения к базе
		return nil, err
	}
	fmt.Println("База подключена (Ping Ok)")
	return db, err
}

func createTable(db *pg.DB, query string) error {
	_, err := db.Exec(query)
	return err
}

// Получение пользователей вместо пинга
/*
	var tasks []Task
	err := db.Model(&tasks).Select()
	if err != nil {
		log.Fatalf("Ошибка получения данных: %v", err)
		return nil, err
	}
	fmt.Println(tasks)
	fmt.Println("База создана")
*/

// Создание таблицы
/*
	// новый способ. не подключает индексы
	err := db.Model((*Task)(nil)).CreateTable(&orm.CreateTableOptions{
		IfNotExists:   true,
		Temp:          false, // Постоянная таблица
		FKConstraints: true,  // Включить ограничения (индексы, FK)
	})
	if err != nil {
		// log.Fatalf("Ошибка создания таблицы: %v", err)
		fmt.Printf("Ошибка создания таблицы: %v", err)
		return nil, err
	}
	// добавление индекса
	if _, err = db.Exec(indexQuery); err != nil {
		fmt.Printf("Ошибка добавления индекса: %v", err)
		return nil, err
	}
*/
