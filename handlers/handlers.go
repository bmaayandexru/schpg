package handlers

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	nd "github.com/bmaayandexru/scheduler/nextdate"
	"github.com/bmaayandexru/scheduler/service"
	"github.com/bmaayandexru/scheduler/storage"
)

// для формирования ошибки
type strcErr struct {
	Error string `json:"error"`
}

// для формирования идентификатора
type strcId struct {
	Id string `json:"id"`
}

// для формирования строки
type strcPwd struct {
	Password string `json:"password"`
}

const (
	CPassword = "12111"
	template  = "20060102"
)

type strcToken struct {
	Token string `json:"token"`
}

// для формирования "пустышки"
type strcEmpty struct{}

// для формирования слайса задач
type strcTasks struct {
	Tasks []storage.Task `json:"tasks"`
}

var sTasks strcTasks

func NextDateHandle(res http.ResponseWriter, req *http.Request) {
	strNow := req.FormValue("now")
	strDate := req.FormValue("date")
	strRepeat := req.FormValue("repeat")
	now, err := time.Parse(template, strNow)
	if err != nil {
		return
	}
	retStr, err := nd.NextDate(now, strDate, strRepeat)
	if err != nil {
		return
	}
	_, _ = res.Write([]byte(retStr))
}

func retError(res http.ResponseWriter, sErr string, statusCode int) {
	var bE strcErr
	bE.Error = sErr
	aBytes, _ := json.Marshal(bE)
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(statusCode)
	_, _ = res.Write(aBytes)
}

func TasksGETSearch(res http.ResponseWriter, req *http.Request) {
	var err error
	search := req.URL.Query().Get("search")
	sTasks.Tasks, err = service.Service.Find(search)
	if err != nil {
		retError(res, fmt.Sprintf("TH GET AT: Ошибка поиска задач: %v\n", err), http.StatusOK)
		return
	}

	arrBytes, err := json.Marshal(sTasks)
	if err != nil {
		retError(res, fmt.Sprintf("TH GET SS: Ошибка json.Marshal(sTsks): %v\n", err), http.StatusOK)
		return
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	_, _ = res.Write(arrBytes)
}

func TasksGETAllTasks(res http.ResponseWriter, req *http.Request) {
	var err error
	sTasks.Tasks, err = service.Service.Find("")
	if err != nil {
		retError(res, fmt.Sprintf("TH GET AT: Ошибка поиска задач: %v\n", err), http.StatusOK)
		return
	}
	arrBytes, err := json.Marshal(sTasks)
	if err != nil {
		retError(res, fmt.Sprintf("TH GET AT: Ошибка json.Marshal(sTasks): %v\n", err), http.StatusOK)
		return
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	_, _ = res.Write(arrBytes)
}

func TaskGETHandle(res http.ResponseWriter, req *http.Request) {
	str_id := req.URL.Query().Get("id")
	id, err := strconv.Atoi(str_id)
	if err != nil {
		fmt.Println(err)
		retError(res, fmt.Sprintf("Tk GET id: Ошибка Atoi(): %s\n", err.Error()), http.StatusOK)
		return
	}
	task, err := service.Service.Get(id)
	if err != nil {
		fmt.Println(err)
		retError(res, fmt.Sprintf("Tk GET id: Ошибка row.Scan(): %s\n", err.Error()), http.StatusOK)
		return
	}
	arrBytes, err := json.Marshal(task)
	if err != nil {
		retError(res, fmt.Sprintf("Tk GET id: Ошибка json.Marshal(sTsks): %v\n", err), http.StatusOK)
		return
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	_, _ = res.Write(arrBytes)
}

func TasksGETHandle(res http.ResponseWriter, req *http.Request) {
	// получаем значение GET-параметра с именем search
	search := req.URL.Query().Get("search")
	if len(search) == 0 { // вывести все задачи
		TasksGETAllTasks(res, req)
		return
	} else { // поиск по дате или строке
		// fmt.Printf("Search *%s*\n", search)
		TasksGETSearch(res, req)
		return
	}
}

func TaskPUTHandle(res http.ResponseWriter, req *http.Request) {
	var task storage.Task
	var buf bytes.Buffer
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		retError(res, fmt.Sprintf("Tk PUT: Ошибка чтения тела запроса: %s\n", err.Error()), http.StatusOK)
		return
	}
	if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
		retError(res, fmt.Sprintf("Tk PUT: Ошибка десериализации: %s\n", err.Error()), http.StatusOK)
		return
	}
	/*
		// анализ task.ID не пустой, это число, есть в базе,
		if len(task.ID) == 0 { // пустой id
			retError(res, fmt.Sprintln("Tk PUT: Пустой ID"), http.StatusOK)
			return
		}
		_, err = strconv.Atoi(task.ID)
		if err != nil { // id не число
			retError(res, fmt.Sprintln("Tk PUT: ID не число"), http.StatusOK)
			return
		}
		// тут ID строка
		смена string на int
	*/
	_, err = service.Service.Get(task.ID)
	if err != nil { // запрос SELECT * WHERE id = :id не должен вернуть ошибку
		retError(res, fmt.Sprintf("Tk PUT: ID нет в базе. Ошибка: %v\n", err), http.StatusOK)
		return
	}
	// ID корректный и в базе есть
	if len(task.Title) == 0 { // Поле Title обязательное
		retError(res, "Tk PUT: Поле `Задача*` пустое", http.StatusOK)
		return
	}
	if len(task.Date) == 0 { // Если поле date не указано или содержит пустую строку,
		task.Date = time.Now().Format(template) // берётся сегодняшнее число.
	} else {
		//  task.Date не пустое. пробуем распарсить
		_, err = time.Parse(template, task.Date)
		if err != nil {
			retError(res, fmt.Sprintf("Tk PUT: Ошибка разбора даты: %v\n", err), http.StatusOK)
			return
		}
	}
	// тут валидная строка в task.Date
	if task.Date < time.Now().Format(template) {
		if len(task.Repeat) == 0 {
			task.Date = time.Now().Format(template)
		} else {
			task.Date, err = nd.NextDate(time.Now(), task.Date, task.Repeat)
			if err != nil {
				retError(res, fmt.Sprintf("Tk PUT: Ошибка NextDate: %v", err), http.StatusOK)
				return
			}
		}
	} else {
		if len(task.Repeat) > 0 {
			task.Date, err = nd.NextDate(time.Now(), task.Date, task.Repeat)
			if err != nil {
				retError(res, fmt.Sprintf("Tk PUT: Ошибка NextDate: %v", err), http.StatusOK)
				return
			}
		}
	}
	// Task перезаписать в базе
	err = service.Service.Update(task)
	if err != nil {
		retError(res, fmt.Sprintf("Tк PUT: Ошибка при изменении в БД: %v\n", err), http.StatusOK)
		return
	}
	// всё отлично. возвращем пустышку
	var bE strcEmpty
	arrBytes, err := json.Marshal(bE)
	if err != nil {
		retError(res, fmt.Sprintf("Tk PUT: Ошибка json.Marshal(bE): %v\n", err), http.StatusOK)
		return
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	_, _ = res.Write(arrBytes)
}

func TaskPOSTHandle(res http.ResponseWriter, req *http.Request) {
	// добавление задачи
	var task storage.Task
	var buf bytes.Buffer
	var err error
	var bId strcId
	// читаем тело запроса
	if _, err := buf.ReadFrom(req.Body); err != nil {
		retError(res, fmt.Sprintf("Ts POST: Ошибка чтения тела запроса: %s\n", err.Error()), http.StatusOK)
		return
	}
	// десериализуем JSON в task
	if err := json.Unmarshal(buf.Bytes(), &task); err != nil {
		retError(res, fmt.Sprintf("Tr POST: Ошибка десериализации: %s\n", err.Error()), http.StatusOK)
		return
	}
	if len(task.Title) == 0 { // Поле Title обязательное
		retError(res, "Ts POST: Поле `Задача*` пустое", http.StatusOK)
		return
	}
	// Если поле date содержит пустую строку,
	if len(task.Date) == 0 { // берётся сегодняшнее число.
		task.Date = time.Now().Format(template)
	} else { //  task.Date не пустое. пробуем распарсить
		if _, err := time.Parse(template, task.Date); err != nil {
			retError(res, fmt.Sprintf("Ts POST: Ошибка разбора даты: %v\n", err), http.StatusOK)
			return
		}
	}
	// тут валидная строка в task.Date
	// это либо строка из текущей даты либо корректная строка
	nows := time.Now().Format(template)
	if len(task.Repeat) > 0 {
		if task.Date < nows { // правило есть и дата меньше сегодняшней
			tn, _ := time.Parse(template, nows)
			if task.Date, err = nd.NextDate(tn, task.Date, task.Repeat); err != nil {
				retError(res, fmt.Sprintf("Ts POST: Ошибка NextDate: %v", err), http.StatusOK)
				return
			}
		}
	} else { // правила повторения нет
		if task.Date < nows { // дата меньше сегодняшней
			task.Date = nows
		}
	}
	//	resSql, err := service.Service.Add(task)
	err = service.Service.Add(task)
	if err != nil {
		retError(res, fmt.Sprintf("Ts POST: Ошибка при добавлении в БД: %v\n", err), http.StatusOK)
		return
	}
	/*
		id, err := resSql.LastInsertId()
		if err != nil {
			retError(res, fmt.Sprintf("Ts POST: Ошибка LastInsetId(): %v\n", err), http.StatusOK)
			return
		}
	*/
	bId.Id = strconv.Itoa(int(0))
	arrBytes, err := json.Marshal(bId)
	if err != nil {
		retError(res, fmt.Sprintf("Ts POST: Ошибка json.Marshal(id): %v\n", err), http.StatusOK)
		return
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	_, _ = res.Write(arrBytes)
}

func TaskDELETEHandle(res http.ResponseWriter, req *http.Request) {
	var (
		id  int
		err error
	)
	// получить id
	str_id := req.URL.Query().Get("id")
	if id, err = strconv.Atoi(str_id); err != nil {
		retError(res, "Tk DELETE. id не число", http.StatusOK)
		return
	}
	if _, err = service.Service.Get(id); err != nil {
		retError(res, fmt.Sprintf("Tk DELETE: id нет в базе. %s", err.Error()), http.StatusOK)
		return
	}
	// удалить по id
	err = service.Service.Delete(id)
	if err != nil {
		retError(res, fmt.Sprintf("Tk DELETE: id: Ошибка удаления из базы: %s\n", err.Error()), http.StatusOK)
		return
	}
	var bE strcEmpty
	arrBytes, err := json.Marshal(bE)
	if err != nil {
		retError(res, fmt.Sprintf("Tk DELETE: Ошибка json.Marshal(): %v\n", err), http.StatusOK)
		return
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	_, _ = res.Write(arrBytes)
}

func TaskHandle(res http.ResponseWriter, req *http.Request) {
	// одна задача
	switch req.Method {
	case "POST": // добавление задачи
		TaskPOSTHandle(res, req)
	case "GET": // запрос для редактирование
		TaskGETHandle(res, req)
	case "PUT": // запрос на изменение
		TaskPUTHandle(res, req)
	case "DELETE": // удаление задачи
		TaskDELETEHandle(res, req)
	}
}

func TaskDonePOSTHandle(res http.ResponseWriter, req *http.Request) {
	// задача выполнена
	str_id := req.URL.Query().Get("id")
	// id string -> int
	id, err := strconv.Atoi(str_id)
	if err != nil {
		fmt.Println(err)
		retError(res, fmt.Sprintf("Tkd POST id: Ошибка Atoi(): %s\n", err.Error()), http.StatusOK)
		return
	}
	task, err := service.Service.Get(id)
	if err != nil {
		fmt.Println(err)
		retError(res, fmt.Sprintf("Tkd POST id: Ошибка row.Scan(): %s\n", err.Error()), http.StatusOK)
		return
	}
	if len(task.Repeat) == 0 {
		err = service.Service.Delete(task.ID)
		if err != nil {
			retError(res, fmt.Sprintf("Tkd POST id: Ошибка удаления из базы: %s\n", err.Error()), http.StatusOK)
			return
		}
	} else { // при наличии правила повторения переназначение даты и UPDATE
		dnow := time.Now()
		dnow = dnow.AddDate(0, 0, 1)
		if dnow.Format(template) < task.Date {
			dnow, _ = time.Parse(template, task.Date)
			dnow = dnow.AddDate(0, 0, 1)
		}
		newDate, err := nd.NextDate(dnow, task.Date, task.Repeat)
		if err != nil {
			retError(res, fmt.Sprintf("Tkd POST: Ошибка NextDate(): %s\n", err.Error()), http.StatusOK)
			return
		}
		task.Date = newDate
		if err = service.Service.Update(task); err != nil {
			retError(res, fmt.Sprintf("Tkd POST: Ошибка UpdateTask(): %s\n", err.Error()), http.StatusOK)
			return
		}
	}
	var bE strcEmpty
	arrBytes, err := json.Marshal(bE)
	if err != nil {
		retError(res, fmt.Sprintf("Tkd POST: Ошибка json.Marshal(): %v\n", err), http.StatusOK)
		return
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	_, _ = res.Write(arrBytes)
}

func TaskDoneHandle(res http.ResponseWriter, req *http.Request) {
	// одна задача (api/task/done)
	if req.Method == "POST" {
		TaskDonePOSTHandle(res, req)
		return
	}
	retError(res, "Нужен только POST запрос", http.StatusOK)
}

func TasksHandle(res http.ResponseWriter, req *http.Request) {
	// много задач (api/tasks)
	if req.Method == "GET" {
		TasksGETHandle(res, req)
		return
	}
	retError(res, "Ts Нужен только GET запрос", http.StatusOK)
}

func SignInPOSTHandle(res http.ResponseWriter, req *http.Request) {
	fmt.Println("Запрос на авторизацию")
	var buf bytes.Buffer
	var pwds strcPwd
	// читаем тело запроса
	if _, err := buf.ReadFrom(req.Body); err != nil {
		retError(res, fmt.Sprintf("Si POST: Ошибка чтения тела запроса: %s\n", err.Error()), http.StatusOK)
		return
	}
	// десериализуем JSON в task
	if err := json.Unmarshal(buf.Bytes(), &pwds); err != nil {
		retError(res, fmt.Sprintf("Si POST: Ошибка десериализации: %s\n", err.Error()), http.StatusOK)
		return
	}
	// Функция должна сверять указанный пароль с хранимым в переменной окружения TODO_PASSWORD.
	// Если они совпадают, нужно сформировать JWT-токен и возвратить его в поле token JSON-объекта.
	envPassword := os.Getenv("TODO_PASSWORD")
	if envPassword == "" {
		// пароля в окружении нет. приваиваем свой
		envPassword = CPassword
	}
	if pwds.Password == envPassword {
		// при совпадении паролей SignInPOSHHandler возвращает в res token,
		// который frondend пишет в куки и который потом из куки используется для авторизации.
		// settings.Token нужно указывать только для тестирования алгоритма авторизации
		// Процесс смены пароля:
		// 1. Меняем CPassword
		// 2. Заходим из браузера с паролем из CPassword
		// 3. Из ответа сервера хэш копипастим в settings.Token для тетирования авторизации
		var tkn strcToken
		tkn.Token = JwtFromPass(envPassword)
		aBytes, _ := json.Marshal(tkn)
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		_, _ = res.Write(aBytes)
		return
	}
	// Если пароль неверный или произошла ошибка, возвращается JSON c текстом ошибки в поле error.
	retError(res, "Пароль не верный", http.StatusUnauthorized)
}

func JwtFromPass(pass string) string {
	result := sha256.Sum256([]byte(pass))
	return hex.EncodeToString(result[:])
}

func SignInHandle(res http.ResponseWriter, req *http.Request) {
	// авторизация (api/signin)
	if req.Method == "POST" {
		SignInPOSTHandle(res, req)
		return
	}
	retError(res, "Нужен только POST запрос", http.StatusOK)
}
