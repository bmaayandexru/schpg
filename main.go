package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/bmaayandexru/scheduler/handlers"
	"github.com/bmaayandexru/scheduler/service"
	"github.com/bmaayandexru/scheduler/storage"
	"github.com/bmaayandexru/scheduler/tests"

	"github.com/go-pg/pg/v10"
)

var (
	mux *http.ServeMux
	db  *pg.DB
	err error
)

func main() {
	// инициализация сервиса и хранилища
	// открытие БД
	if db, err = storage.InitDBase(tests.ConnStr); err != nil {
		fmt.Printf("Ошибка открытия базы %v\n", err)
		panic(err)
	}
	defer db.Close()
	store := storage.NewTaskStore(db)
	service.Service = service.NewTaskService(store)

	// поднимаем сервер
	mux = http.NewServeMux()
	// вешаем обработчики
	mux.HandleFunc("/api/nextdate", handlers.NextDateHandle)
	mux.HandleFunc("/api/task", auth(handlers.TaskHandle))
	mux.HandleFunc("/api/task/done", auth(handlers.TaskDoneHandle))
	mux.HandleFunc("/api/tasks", auth(handlers.TasksHandle))
	mux.HandleFunc("/api/signin", handlers.SignInHandle)
	// обеспечиваем интерфейс
	mux.Handle("/", http.FileServer(http.Dir("web/")))
	strPort := defStrPort()
	// пускаем сервер
	err = http.ListenAndServe(strPort, mux)
	if err != nil {
		panic(err)
	}
}

// auth(...) - проверяет аутентификацию
func auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// смотрим наличие пароля
		pass := os.Getenv("TODO_PASSWORD")
		// авторизация будет проверяться только при наличии TODO_PASSWORD
		if len(pass) == 0 {
			// это чтоб работало без переменной окружения
			pass = handlers.CPassword
		}
		if len(pass) > 0 {
			var jwt string // JWT-токен из куки
			// получаем куку
			cookie, err := r.Cookie("token")
			if err == nil {
				jwt = cookie.Value
			}
			var valid bool
			jwtp := handlers.JwtFromPass(pass)
			// здесь код для валидации и проверки JWT-токена
			valid = (jwt == jwtp)
			if !valid {
				// возвращаем ошибку авторизации 401
				http.Error(w, "Authentification required", http.StatusUnauthorized)
				return
			}
		}
		next(w, r)
	})
}

// defStrPort() - определение номера порта. возвращает строку
func defStrPort() string {
	defPort := "7540"
	// переменая tests.Port из settings.go
	settingsStrPort := fmt.Sprintf("%d", tests.Port)
	if settingsStrPort != "" {
		defPort = settingsStrPort
	}
	envStrPort := os.Getenv("TODO_PORT")
	if envStrPort != "" {
		defPort = envStrPort
	}
	// defPort = "7540"
	fmt.Printf("Set port %s \n", defPort)
	return ":" + defPort
}
