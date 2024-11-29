package service

import (
	"time"

	"github.com/bmaayandexru/scheduler/nextdate"
	"github.com/bmaayandexru/scheduler/storage"
)

type TaskService struct {
	store storage.TaskStore
}

const limit = 50

var Service TaskService

func NewTaskService(store storage.TaskStore) TaskService {
	return TaskService{store: store}
}

// сервис не должен выдавать sql.Result. только error
// func (ts TaskService) Add(task storage.Task) (sql.Result, error) {
func (ts TaskService) Add(task storage.Task) error {
	return ts.store.Add(task)
}

//	func (ts TaskService) Delete(id string) (sql.Result, error) {
//		func (ts TaskService) Delete(id string) error {
func (ts TaskService) Delete(id int) error {
	return ts.store.Delete(id)
}

// func (ts TaskService) Find(search string) (*sql.Rows, error) {
func (ts TaskService) Find(search string) ([]storage.Task, error) {
	return ts.store.Find(search)
}

// func (ts TaskService) Get(id string) (storage.Task, error) {
func (ts TaskService) Get(id int) (storage.Task, error) {
	return ts.store.Get(id)
}

// func (ts TaskService) Update(task storage.Task) (sql.Result, error) {
func (ts TaskService) Update(task storage.Task) error {
	return ts.store.Update(task)
}

// func (ts TaskService) Done(id string) error {
func (ts TaskService) Done(id int) error {
	// выполнение задачи - это перенос либо удаление
	var task storage.Task
	var err error
	// запросить задаче по id
	if task, err = ts.Get(id); err != nil {
		return err
	}
	// если правила повторения нет удалить
	if len(task.Repeat) == 0 {
		err = ts.Delete(id)
		return err
	} else {
		// если есть вызвать nextdate и перенести дату (update)
		if task.Date, err = nextdate.NextDate(time.Now(), task.Date, task.Repeat); err != nil {
			return err
		}
		if err = ts.Update(task); err != nil {
			return err
		}
		return nil
	}
}
