package coprhd

import (
	"errors"
	"fmt"
	"time"
)

const (
	TaskPollDelay              = time.Millisecond * 250
	TaskStatePending TaskState = "pending"
	TaskStateError   TaskState = "error"
	TaskStateReady   TaskState = "ready"

	queryTaskUriTpl = "vdc/tasks/%s.json"
)

var (
	ErrTaskWaitTimeout = errors.New("WaitDone timeout")
)

type (
	TaskService struct {
		*Client
	}

	Task struct {
		Name        string        `json:"name"`
		Id          string        `json:"id"`
		State       TaskState     `json:"state"`
		Message     string        `json:"message"`
		Description string        `json:"description"`
		Progress    int           `json:"progress"`
		Resource    NamedResource `json:"resource"`
		Workflow    Resource      `json:"workflow"`
	}

	TaskState string
)

func (this *Client) Task() *TaskService {
	return &TaskService{this.Copy()}
}

// Query returns the task object
func (this *TaskService) Query(id string) (Task, error) {
	path := fmt.Sprintf(queryTaskUriTpl, id)
	task := Task{}

	err := this.get(path, nil, &task)

	return task, err
}

// WaitDone does a busy poll to wait for a task to reach the specified state with the timeout
func (this *TaskService) WaitDone(id string, state TaskState, to time.Duration) error {
	timeout := time.After(to)
	timer := time.Tick(TaskPollDelay)

	for {
		task, err := this.Query(id)
		if err != nil {
			return err
		}

		if task.State == TaskStateError {
			return errors.New(task.Message + ":" + task.Description)
		}

		if task.State == state && task.Progress == 100 {
			if task.Workflow.Id != "" {
				return this.Workflow().
					WaitDone(task.Workflow.Id, WorkflowStateSuccess, to)
			}
			return nil
		}

		select {
		case <-timer:
		case <-timeout:
			return ErrTaskWaitTimeout
		}
	}
}
