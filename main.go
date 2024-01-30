// приложение эмулирует получение и обработку тасков, пытается и получать и обрабатывать в многопоточном режиме
// в конце должно выводить успешные таски и ошибки выполнения остальных тасков
package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// A Task represents a meaninglessness of our life
type Task struct {
	Id          string
	CreatedAt   time.Time
	CompletedAt time.Time
}

func initTask() Task {
	now := time.Now()
	id := uuid.New().String()

	return Task{Id: id, CreatedAt: now}
}

func processTask(task Task) Task {
	task.CompletedAt = time.Now()

	return task
}

func hasErrors(task Task) bool {
	return task.CreatedAt.Nanosecond()%2 > 0 // вот такое условие появления ошибочных тасков
}

func sortTask(task Task, completedChan chan<- Task, errorChan chan<- error) {
	if hasErrors(task) {
		errorChan <- fmt.Errorf("Task id %s time %s, error: something went wrong", task.Id, task.CreatedAt.Format(time.RFC3339Nano))
	} else {
		completedChan <- task
	}
}

func readAll[T any](channel <-chan T, action func(T)) {
	for item := range channel {
		action(item)
	}
}

func printResult(tasks map[string]Task, errors []error) {
	println("Errors:")
	for _, err := range errors {
		println(err.Error())
	}

	println("Done tasks:")
	for _, task := range tasks {
		println(fmt.Sprintf("Task id %s time %s completed %s", task.Id, task.CreatedAt.Format(time.RFC3339Nano), task.CompletedAt.Format(time.RFC3339Nano)))
	}
}

func main() {
	taskChan := make(chan Task, 10)

	completedTaskChan := make(chan Task)
	completedTasks := map[string]Task{}

	errorChan := make(chan error)
	errors := []error{}

	var taskSync sync.WaitGroup
	isAllowedToInit := true

	go func() {
		for isAllowedToInit {
			taskSync.Add(1)
			taskChan <- initTask()
		}
		close(taskChan)
	}()

	go readAll(taskChan, func(task Task) { task = processTask(task); sortTask(task, completedTaskChan, errorChan) })

	go readAll(completedTaskChan, func(task Task) { completedTasks[task.Id] = task; taskSync.Done() })
	go readAll(errorChan, func(err error) { errors = append(errors, err); taskSync.Done() })

	// ожидание накопления тасков
	time.Sleep(time.Second * 3)
	isAllowedToInit = false
	// ожидание завершения обработки накопленных тасков
	taskSync.Wait()

	close(completedTaskChan)
	close(errorChan)

	printResult(completedTasks, errors)
	// P.S. спасибо за нестандартный подход! не грех было и с golang познакомиться :)
}
