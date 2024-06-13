package repository

import (
	"fmt"
	"log"
	"reflect"
	"sync"
	"taskbot/internal/models"
)

type TaskRepository struct {
	mu       *sync.RWMutex
	tasks    map[int]*models.Task
	sequence int
}

func NewTaskRepository() *TaskRepository {
	return &TaskRepository{
		mu:       &sync.RWMutex{},
		tasks:    map[int]*models.Task{},
		sequence: 0,
	}
}

func (tr *TaskRepository) Add(task *models.Task) int {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	log.Printf("TaskRepository: method Add: tr.sequence = %d\n", tr.sequence)
	tr.sequence++
	log.Printf("TaskRepository: method Add: tr.sequence = %d\n", tr.sequence)
	tr.tasks[tr.sequence] = task
	log.Printf("TaskRepository: method Add: save task - %#v with ID: %d\n", task, tr.sequence)
	return tr.sequence
}

func (tr *TaskRepository) Get(taskId int) (*models.Task, bool) {
	tr.mu.RLock()
	defer tr.mu.RUnlock()
	task, ok := tr.tasks[taskId]
	log.Printf("TaskRepository: method Get: get task with ID: %d - task: %#v\n", taskId, task)
	return task, ok
}

func (tr *TaskRepository) GetAll() map[int]*models.Task {
	log.Printf("TaskRepository: method GetAll: get tasks\n")
	return tr.tasks
}

func (tr *TaskRepository) Delete(taskId int, user *models.User) (*models.Task, error) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	task, ok := tr.tasks[taskId]
	if !ok {
		return nil, fmt.Errorf("Task %d not found\n", taskId)
	}
	if !reflect.DeepEqual(task.Executor, user) {
		return task, fmt.Errorf("Task %d unassign executor %#v not equal\n", taskId, user)
	}
	delete(tr.tasks, taskId)
	log.Printf("TaskRepository: method Deleted: deleted task with ID: %d\n", taskId)
	return task, nil
}

func (tr *TaskRepository) Assign(taskId int, user *models.User) (*models.Task, *models.User, error) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	log.Printf("TaskRepository: method Assign: tasks: %#v and taskId - %d\n", tr.tasks, taskId)
	task, ok := tr.tasks[taskId]
	log.Printf("TaskRepository: method Assign: task: %#v\n", task)
	if !ok {
		log.Println("TaskRepository: method Assign: not found task")
		return nil, nil, fmt.Errorf("Task %d not found\n", taskId)
	}
	if task.Executor != nil {
		prevExecutor := task.Executor
		task.Executor = user
		return task, prevExecutor, nil
	}
	task.Executor = user
	log.Printf("TaskRepository: method Assign: assign task with ID: %d - executor: %#v\n", taskId, user)
	return task, nil, nil
}

func (tr *TaskRepository) Unassign(taskId int, user *models.User) (*models.Task, error) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	task, ok := tr.tasks[taskId]
	if !ok {
		return nil, fmt.Errorf("Task %d not found\n", taskId)
	}
	if !reflect.DeepEqual(task.Executor, user) {
		return task, fmt.Errorf("Task %d unassign executor %#v not equal\n", taskId, user)
	}
	task.Executor = nil
	log.Printf("TaskRepository: method Unassign: unassign task with ID: %d\n", taskId)
	return task, nil
}

func (tr *TaskRepository) Executor(user *models.User) map[int]*models.Task {
	tr.mu.RLock()
	defer tr.mu.RUnlock()
	tasks := make(map[int]*models.Task)

	for taskId, task := range tr.tasks {
		if task.Executor != nil && reflect.DeepEqual(task.Executor, user) {
			tasks[taskId] = task
		}
	}

	log.Printf("TaskRepository: method Execotur: execotur id %d - tasks\n", user.ID)
	return tasks
}

func (tr *TaskRepository) Author(user *models.User) map[int]*models.Task {
	tr.mu.RLock()
	defer tr.mu.RUnlock()
	tasks := make(map[int]*models.Task)

	for taskId, task := range tr.tasks {
		if reflect.DeepEqual(task.Author, user) {
			tasks[taskId] = task
		}
	}

	log.Printf("TaskRepository: method Author: author id %d - tasks\n", user.ID)
	return tasks
}
