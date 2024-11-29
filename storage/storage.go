package storage

import (
	"fmt"
	"log"
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
	/*
		pti := &Task{
			Date:    task.Date,
			Title:   task.Title,
			Comment: task.Comment,
			Repeat:  task.Repeat,
		}
	*/
	_, err := ts.DB.Model(&task).Insert()
	return err
}

//	func (ts TaskStore) Delete(id string) (sql.Result, error) {
//		func (ts TaskStore) Delete(id string) error {
func (ts TaskStore) Delete(id int) error {
	/*
		_, err := ts.DB.Exec("DELETE FROM scheduler WHERE id = $1", id)
	*/
	var tasks []Task
	_, err := ts.DB.Model(&tasks).Where("id = ?", id).Delete()
	return err
}

// func (ts TaskStore) Find(search string) (*sql.Rows, error) {
func (ts TaskStore) Find(search string) ([]Task, error) {
	var (
		err   error
		tasks []Task
	)
	// пустой слайс задач
	// возвращаем всё если строка пустая

	if len(search) == 0 {
		//			rows, err = ts.DB.Query("SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT $1", limit)
		err = ts.DB.Model(&tasks).Limit(limit).Order("date").Order("title").Select()
		return tasks, err
	}
	// парсим строку на дату
	if date, err := time.Parse("02-01-2006", search); err == nil {
		// дата есть
		/*
			rows, err = ts.DB.Query("SELECT id, date, title, comment, repeat FROM scheduler WHERE date = $1 LIMIT $2",
				date.Format(templ),
				limit)
		*/
		err = ts.DB.Model(&tasks).Where("date = ?", date.Format(templ)).Limit(limit).Order("title").Select()
		return tasks, err

	} else {
		// даты нет
		search = "%" + search + "%"
		/*
			rows, err = ts.DB.Query("SELECT id, date, title, comment, repeat FROM scheduler WHERE UPPER(title) LIKE UPPER($1) OR UPPER(comment) LIKE UPPER($1) ORDER BY date LIMIT $2",
				search,
				limit)
		*/
		// тут по ходу надо query
		_, err = ts.DB.Query(&tasks, "select * from tasks where upper(title) like upper(?) or upper(comment) like upper(?)", search, search)
		// #42601 ошибка синтаксиса (примерное положение: ")")
		// err = ts.DB.Model(&tasks).Where("upper(title) like upper(?) or upper(comment) like upper(?)", search).Limit(limit).Order("date").Order("title").Select()
		return tasks, err
	}
	/*
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

// func (ts TaskStore) Get(id string) (Task, error) {
func (ts TaskStore) Get(id int) (Task, error) {
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

/*
var indexQuery string = `
CREATE INDEX IF NOT EXISTS idx_date ON tasks (date);
CREATE INDEX IF NOT EXISTS idx_title ON tasks (title);
`
*/
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
	// err := db.Ping()
	/*
		var tasks []Task
		_, err := db.Query(&tasks, "SELECT * FROM scheduler")
		if err != nil {
			//log.Fatalf("Ошибка подключения к базе данных: %v", err)
			return nil, err
		}
		for _, task := range tasks {
			fmt.Println(task)
		}
		fmt.Println("Успешное подключение к базе данных!")
		return db, err
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
	// Получение пользователей
	var tasks []Task
	err := db.Model(&tasks).Select()
	if err != nil {
		log.Fatalf("Ошибка получения данных: %v", err)
		return nil, err
	}
	fmt.Println(tasks)
	fmt.Println("База создана")
	return db, err

}

func createTable(db *pg.DB, query string) error {
	_, err := db.Exec(query)
	return err
}
