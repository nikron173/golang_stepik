package service

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"taskbot/internal/models"
	"taskbot/internal/repository"
)

type TaskService struct {
	st *repository.TaskRepository
}

func NewTaskService() *TaskService {
	return &TaskService{
		st: repository.NewTaskRepository(),
	}
}

func (ts *TaskService) Create(taskName string, user *models.User) (map[int64]string, error) {
	task := &models.Task{
		Author:   user,
		Name:     taskName,
		Executor: nil,
	}
	log.Printf("TaskService: method Create: task - %#v\n", task)
	taskId := ts.st.Add(task)
	// resp := &models.Responce{
	// 	ChatId:  user.ChatID,
	// 	Message: fmt.Sprintf("Задача \"%s\" создана, id=%d", taskName, taskId),
	// }
	resp := map[int64]string{
		user.ChatID: fmt.Sprintf("Задача \"%s\" создана, id=%d", taskName, taskId),
	}
	return resp, nil
}

func (ts *TaskService) GetAll(user *models.User) (map[int64]string, error) {
	builder := &strings.Builder{}
	tasks := ts.st.GetAll()
	if len(tasks) == 0 {
		resp := map[int64]string{
			user.ChatID: "Нет задач",
		}
		return resp, nil
	}
	for taskId, task := range tasks {
		if task.Executor != nil && reflect.DeepEqual(task.Executor, user) {
			builder.WriteString(fmt.Sprintf("%d. %s by @%s\nassignee: я\n/unassign_%d /resolve_%d\n\n",
				taskId, task.Name, task.Author.Username, taskId, taskId))
		} else if task.Executor != nil {
			builder.WriteString(fmt.Sprintf("%d. %s by @%s\nassignee: @%s\n\n",
				taskId, task.Name, task.Author.Username, task.Executor.Username))
		} else {
			builder.WriteString(fmt.Sprintf("%d. %s by @%s\n/assign_%d\n\n",
				taskId, task.Name, task.Author.Username, taskId))
		}
	}
	respStr := builder.String()
	resp := map[int64]string{
		user.ChatID: respStr[:len(respStr)-2],
	}
	log.Printf("TaskService: method GetAll: get all tasks\n")
	return resp, nil
}

func (ts *TaskService) Assign(taskId int, user *models.User) (map[int64]string, error) {
	task, prevExecutor, err := ts.st.Assign(taskId, user)
	if err != nil {
		return nil, fmt.Errorf("Task %d not found", taskId)
	}

	if prevExecutor != nil {
		return map[int64]string{
			user.ChatID:         fmt.Sprintf("Задача \"%s\" назначена на вас", task.Name),
			prevExecutor.ChatID: fmt.Sprintf("Задача \"%s\" назначена на @%s", task.Name, user.Username),
		}, nil
	}

	if reflect.DeepEqual(user, task.Author) {
		return map[int64]string{
			user.ChatID: fmt.Sprintf("Задача \"%s\" назначена на вас", task.Name),
		}, nil
	}
	// return fmt.Sprintf("Задача \"%s\" назначена на вас", task.Name), nil
	return map[int64]string{
		user.ChatID:        fmt.Sprintf("Задача \"%s\" назначена на вас", task.Name),
		task.Author.ChatID: fmt.Sprintf("Задача \"%s\" назначена на @%s", task.Name, user.Username),
	}, nil
}

func (ts *TaskService) Unassign(taskId int, user *models.User) (map[int64]string, error) {
	task, err := ts.st.Unassign(taskId, user)
	if err != nil && task != nil {
		return map[int64]string{
			user.ChatID: "Задача не на вас",
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("Task %d not found", taskId)
	}
	return map[int64]string{
		user.ChatID:        "Принято",
		task.Author.ChatID: fmt.Sprintf("Задача \"%s\" осталась без исполнителя", task.Name),
	}, nil
}

func (ts *TaskService) Resolve(taskId int, user *models.User) (map[int64]string, error) {
	task, err := ts.st.Delete(taskId, user)
	if err != nil && task != nil {
		return map[int64]string{
			user.ChatID: "Задача не на вас",
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("Task %d not found", taskId)
	}

	return map[int64]string{
		user.ChatID:        fmt.Sprintf("Задача \"%s\" выполнена", task.Name),
		task.Author.ChatID: fmt.Sprintf("Задача \"%s\" выполнена @%s", task.Name, task.Executor.Username),
	}, nil
}

func (ts *TaskService) ExecutorTask(user *models.User) (map[int64]string, error) {
	builder := &strings.Builder{}
	tasks := ts.st.Executor(user)
	if len(tasks) == 0 {
		return map[int64]string{
			user.ChatID: "Нет задач",
		}, nil
	}
	for taskId, task := range tasks {
		builder.WriteString(fmt.Sprintf("%d. %s by @%s\n/unassign_%d /resolve_%d\n\n",
			taskId, task.Name, task.Author.Username, taskId, taskId))
	}
	respStr := builder.String()
	return map[int64]string{
		user.ChatID: respStr[:len(respStr)-2],
	}, nil
}

func (ts *TaskService) AuthorTask(user *models.User) (map[int64]string, error) {
	builder := &strings.Builder{}
	tasks := ts.st.Author(user)
	if len(tasks) == 0 {
		return map[int64]string{
			user.ChatID: "Нет задач",
		}, nil
	}
	for taskId, task := range tasks {
		builder.WriteString(fmt.Sprintf("%d. %s by @%s\n/assign_%d\n\n",
			taskId, task.Name, task.Author.Username, taskId))
	}
	respStr := builder.String()
	return map[int64]string{
		user.ChatID: respStr[:len(respStr)-2],
	}, nil
}
