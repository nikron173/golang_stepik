package router

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"taskbot/internal/models"
	"taskbot/internal/service"
)

type Router struct {
	ts *service.TaskService
}

var (
	newTaskReg      = regexp.MustCompile(`^/new\s.*`)
	assignTaskReg   = regexp.MustCompile(`^/assign_\d+`)
	unassignTaskReg = regexp.MustCompile(`^/unassign_\d+`)
	resolveTaskReg  = regexp.MustCompile(`^/resolve_\d+`)
)

func NewRouter() *Router {
	return &Router{
		ts: service.NewTaskService(),
	}
}

func (r *Router) Route(cmd string, user *models.User) (map[int64]string, error) {
	switch {
	case newTaskReg.MatchString(cmd):
		{
			taskName := strings.Join(strings.Split(cmd, " ")[1:], " ")
			resp, err := r.ts.Create(taskName, user)
			if err != nil {
				return nil, err
			}
			return resp, nil
		}
	case assignTaskReg.MatchString(cmd):
		{
			taskId, err := strconv.Atoi(strings.Split(cmd, "_")[1])
			if err != nil {
				return nil, fmt.Errorf("Error command \"%s\", not valid task id", cmd)
			}
			resp, err := r.ts.Assign(taskId, user)
			if err != nil {
				return nil, err
			}
			return resp, nil
		}
	case unassignTaskReg.MatchString(cmd):
		{
			taskId, err := strconv.Atoi(strings.Split(cmd, "_")[1])
			if err != nil {
				return nil, fmt.Errorf("Error command \"%s\", not valid task id", cmd)
			}
			resp, err := r.ts.Unassign(taskId, user)
			if err != nil {
				return nil, err
			}
			return resp, nil
		}
	case resolveTaskReg.MatchString(cmd):
		{
			taskId, err := strconv.Atoi(strings.Split(cmd, "_")[1])
			if err != nil {
				return nil, fmt.Errorf("Error command \"%s\", not valid task id", cmd)
			}
			resp, err := r.ts.Resolve(taskId, user)
			if err != nil {
				return nil, err
			}
			return resp, err
		}
	case strings.EqualFold("/tasks", cmd):
		{
			resp, err := r.ts.GetAll(user)
			if err != nil {
				return nil, err
			}
			return resp, nil
		}
	case strings.EqualFold("/my", cmd):
		{
			resp, err := r.ts.ExecutorTask(user)
			if err != nil {
				return nil, err
			}
			return resp, nil
		}
	case strings.EqualFold("/owner", cmd):
		{
			resp, err := r.ts.AuthorTask(user)
			if err != nil {
				return nil, err
			}
			return resp, nil
		}
	case strings.EqualFold("/start", cmd):
		{
			return map[int64]string{
				user.ChatID: "Добро пожаловать в планировщик задач",
			}, nil
		}
	default:
		{
			return nil, fmt.Errorf("Command `%s` not found", cmd)
		}

	}
}
