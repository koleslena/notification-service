package services

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
)

type Task struct {
	sync.Mutex
	id         int
	StartAfter time.Time
	Interval   time.Duration
	StopTime   time.Time
	TaskFunc   func() error
	ErrFunc    func(error)

	ctx    context.Context
	cancel context.CancelFunc
}

func (t *Task) UpdateTask(f func()) {
	t.Lock()
	defer t.Unlock()

	f()
}

func (t *Task) Clone() *Task {
	task := &Task{}
	task.TaskFunc = t.TaskFunc
	task.ErrFunc = t.ErrFunc
	task.StartAfter = t.StartAfter
	task.id = t.id
	task.ctx = t.ctx
	task.cancel = t.cancel
	return task
}

type Scheduler struct {
	sync.RWMutex

	tasks map[int]*Task
}

var (
	ErrIDNotFound = errors.New("could not find task within the task list")
)

func NewScheduler() *Scheduler {
	s := &Scheduler{}
	s.tasks = make(map[int]*Task)
	return s
}

func (schd *Scheduler) AddTask(id int, t *Task) error {
	if t.TaskFunc == nil {
		return errors.New("task function cannot be nil")
	}
	if t.Interval <= time.Duration(0) {
		return errors.New("task interval must be defined")
	}

	t.ctx, t.cancel = context.WithCancel(context.Background())

	schd.Lock()
	defer schd.Unlock()

	schd.tasks[id] = t
	schd.scheduleTask(t)
	_ = time.AfterFunc(time.Until(t.StopTime), func() {
		go func() {
			schd.Del(t.id)
		}()
	})

	return nil
}

func (schd *Scheduler) Del(id int) {
	t, err := schd.Lookup(id)
	if err == ErrIDNotFound {
		return
	}

	defer t.cancel()

	t.Lock()
	defer t.Unlock()

	schd.Lock()
	defer schd.Unlock()
	delete(schd.tasks, t.id)
}

func (schd *Scheduler) Lookup(id int) (*Task, error) {
	schd.RLock()
	defer schd.RUnlock()
	t, ok := schd.tasks[id]
	if ok {
		return t, nil
	}
	return nil, ErrIDNotFound
}

func (schd *Scheduler) Exists(id int) (bool, error) {
	schd.RLock()
	defer schd.RUnlock()
	_, ok := schd.tasks[id]
	if ok {
		return true, nil
	}
	return false, ErrIDNotFound
}

func (schd *Scheduler) Tasks() map[int]*Task {
	schd.RLock()
	defer schd.RUnlock()
	m := make(map[int]*Task)
	for k, v := range schd.tasks {
		m[k] = v.Clone()
	}
	return m
}

func (schd *Scheduler) Stop() {
	tt := schd.Tasks()
	for n := range tt {
		schd.Del(n)
	}
}

func (schd *Scheduler) scheduleTask(t *Task) {
	_ = time.AfterFunc(time.Until(t.StartAfter), func() {
		go func() {
			schd.execTask(t)
		}()
	})
}

func (schd *Scheduler) execTask(t *Task) {
	exists, err := schd.Exists(t.id)
	if exists && err != ErrIDNotFound {
		err = t.TaskFunc()
		if err != nil && t.ErrFunc != nil {
			t.ErrFunc(err)
		}
		if t.Interval == 0 {
			schd.Del(t.id)
		} else {
			time.AfterFunc(t.Interval, func() { schd.execTask(t) })
		}
	}
}
