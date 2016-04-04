package coprhd

import (
	"errors"
	"fmt"
	"time"
)

const (
	WorkflowPollDelay                  = time.Millisecond * 250
	WorkflowStateError   WorkflowState = "ERROR"
	WorkflowStateSuccess WorkflowState = "SUCCESS"

	queryWorkflowUriTpl = "vdc/workflows/%s.json"
)

var (
	ErrWorkflowWaitTimeout = errors.New("WaitDone timeout")
)

type (
	WorkflowService struct {
		*Client
	}

	Workflow struct {
		StorageObject `json:",inline"`
		Completed     bool          `json:"completed"`
		State         WorkflowState `json:"completion_state"`
		Message       string        `json:"completion_message"`
	}

	WorkflowState string
)

func (this *Client) Workflow() *WorkflowService {
	return &WorkflowService{this.Copy()}
}

// Query returns the task object
func (this *WorkflowService) Query(id string) (Workflow, error) {
	path := fmt.Sprintf(queryWorkflowUriTpl, id)
	task := Workflow{}

	err := this.get(path, nil, &task)

	return task, err
}

// WaitDone does a busy poll to wait for a task to reach the specified state with the timeout
func (this *WorkflowService) WaitDone(id string, state WorkflowState, to time.Duration) error {
	timeout := time.After(to)
	timer := time.Tick(WorkflowPollDelay)

	for {
		work, err := this.Query(id)
		if err != nil {
			return err
		}

		if work.State == WorkflowStateError {
			return errors.New(work.Message)
		}

		if work.State == state && work.Completed {
			return nil
		}

		select {
		case <-timer:
		case <-timeout:
			return ErrWorkflowWaitTimeout
		}
	}
}
