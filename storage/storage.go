package storage

import (
	"fmt"
	"log"

	//_ "modernc.org/sqlite"
	// _ "github.com/lib/pq" // Импорт драйвера
	//"github.com/bmaayandexru/go_final_project/tests"
	"github.com/go-pg/pg/v10"
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"` // omitempty
	Title   string `json:"title"`
	Comment string `json:"comment"` // omitempty
	Repeat  string `json:"repeat"`  // omitempty
}

const (
	templ = "20060102"
	limit = 50
)

type TaskStore struct {
	DB *pg.DB
}

func NewTaskStore(db *pg.DB) TaskStore {
	return TaskStore{DB: db}
}

// storage не должен выдавать sql.Result только error
func (ts TaskStore) Add(task Task) error {
	/*
		_, err := ts.DB.Exec("INSERT INTO scheduler(date, title, comment, repeat) VALUES ($1, $2, $3, $4) ",
			task.Date, task.Title, task.Comment, task.Repeat)
	*/
	return nil
}

// func (ts TaskStore) Delete(id string) (sql.Result, error) {
func (ts TaskStore) Delete(id string) error {
	/*
		_, err := ts.DB.Exec("DELETE FROM scheduler WHERE id = $1", id)
	*/
	return nil
}

// func (ts TaskStore) Find(search string) (*sql.Rows, error) {
func (ts TaskStore) Find(search string) ([]Task, error) {
	var (
		//		err   error
		//		rows  *sql.Rows
		tasks []Task
	)
	// пустой слайс задач
	tasks = make([]Task, 0)
	// возвращаем всё если строка пустая
	/*
		if len(search) == 0 {
			rows, err = ts.DB.Query("SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT $1", limit)
		}
		// парсим строку на дату
		if date, err := time.Parse("02-01-2006", search); err == nil {
			// дата есть
			rows, err = ts.DB.Query("SELECT id, date, title, comment, repeat FROM scheduler WHERE date = $1 LIMIT $2",
				date.Format(templ),
				limit)
		} else {
			// даты нет
			search = "%" + search + "%"
			rows, err = ts.DB.Query("SELECT id, date, title, comment, repeat FROM scheduler WHERE UPPER(title) LIKE UPPER($1) OR UPPER(comment) LIKE UPPER($1) ORDER BY date LIMIT $2",
				search,
				limit)
		}
		if err != nil {
			return tasks, err
		}
		for rows.Next() {
			task := Task{}
			err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
			if err != nil {
				return tasks, err
			}
			tasks = append(tasks, task)
		}
		if err = rows.Err(); err != nil {
			return tasks, err
		}
	*/
	return tasks, nil
}

func (ts TaskStore) Get(id string) (Task, error) {
	task := Task{}
	/*
		row := ts.DB.QueryRow("SELECT * FROM scheduler WHERE id = $1", id)
		err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	*/
	return task, nil
}

// func (ts TaskStore) Update(task Task) (sql.Result, error) {
func (ts TaskStore) Update(task Task) error {
	/*
		_, err := ts.DB.Exec("UPDATE scheduler SET  date = $2, title = $3, comment = $4, repeat = $5 WHERE id = $1",
			task.ID,
			task.Date,
			task.Title,
			task.Comment,
			task.Repeat)
	*/
	return nil
}

var schemaSQL string = `
CREATE TABLE IF NOT EXISTS scheduler (
    id SERIAL PRIMARY KEY,
    date CHAR(8) NOT NULL, 
    title VARCHAR(32) NOT NULL,
    comment TEXT NOT NULL,
    repeat VARCHAR(128) NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_date ON scheduler (date); 
CREATE INDEX IF NOT EXISTS idx_title ON scheduler (title); 
`

func InitDBase(connStr string) (*pg.DB, error) {
	fmt.Println("Init Data Base...")
	// Создание конфигурации для подключения к базе данных
	db := pg.Connect(&pg.Options{
		Addr:     "localhost:5432", // Адрес PostgreSQL сервера
		User:     "postgres",       // Имя пользователя
		Password: "password",       // Пароль
		Database: "dbscheduler",    // Имя базы данных
	})

	// Проверка соединения
	// err := db.Ping()
	var tasks []Task
	_, err := db.Query(&tasks, `SELECT * FROM scheduler`)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	fmt.Println("Успешное подключение к базе данных!")
	// создание таблицы
	if err := createTable(db, schemaSQL); err != nil {
		return nil, err
	}
	return db, err
}

func createTable(db *pg.DB, query string) error {
	_, err := db.Exec(query)
	return err
}
