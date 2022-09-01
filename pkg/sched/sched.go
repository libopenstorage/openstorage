package sched

import (
	"container/list"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/libopenstorage/openstorage/pkg/dbg"
	"github.com/sirupsen/logrus"
)

type TaskID uint64

const (
	TaskNone        = TaskID(0)
	workerBatchSize = 10
	maxWorkers      = 100
	// Need to see how this value of idleTimeout works out. We may need to increase it if
	// we see too much churn in the workers. But on the other hand, increasing the value by a lot could
	// cause the excess workers to never exit since they might pick up tasks in round-robin fashion.
	idleTimeout      = 30 * time.Second
	statsLogDuration = 10 * time.Minute
)

func ValidTaskID(t TaskID) bool { return t != TaskNone }

type ScheduleTask func(Interval)

type histogram interface {
	record(time.Duration)
	getHistogram() string
}

type histImpl struct {
	// lock for this struct
	sync.Mutex
	// name of the histogram
	name string
	// each bucket holds the count of values that fall between the limits of the previous bucket and this bucket
	buckets []uint64
	// limits for each bucket (except for the last one which is a catchall bucket);
	// these are calcuated using min, max and multiplier
	limits []time.Duration
	// limit for the first bucket; values less than min are counted in the first bucket
	min time.Duration
	// limit for the second-to-last bucket; values greater than max are counted in the last bucket
	max time.Duration
	// each bucket's limit is calculated by multiplying the previous bucket's limit by this value
	multiplier int
}

func newHistogram(name string, min, max time.Duration) histogram {
	hist := &histImpl{
		name:       name,
		min:        min,
		max:        max,
		multiplier: 2,
	}
	for i := min; i < max; i = i * time.Duration(hist.multiplier) {
		hist.buckets = append(hist.buckets, 0)
		hist.limits = append(hist.limits, i)
	}
	// last bucket is a catchall for values greater than the max (approx)
	hist.buckets = append(hist.buckets, 0)
	return hist
}

func (h *histImpl) record(val time.Duration) {
	h.Lock()
	defer h.Unlock()
	for i := range h.limits {
		if val < h.limits[i] {
			h.buckets[i]++
			return
		}
	}
	h.buckets[len(h.buckets)-1]++
}

func (h *histImpl) getHistogram() string {
	h.Lock()
	defer h.Unlock()

	var total uint64
	for i := range h.buckets {
		total = total + h.buckets[i]
	}
	if total == 0 {
		return ""
	}
	var out strings.Builder
	out.WriteString(fmt.Sprintf("%s: ", h.name))
	sep := ""
	for i := range h.limits {
		if h.buckets[i] > 0 {
			out.WriteString(fmt.Sprintf("%s%v < %v (%v%%)", sep, h.buckets[i], h.limits[i], h.buckets[i]*100/total))
			sep = ", "
		}
	}
	lastBucket := len(h.buckets) - 1
	if h.buckets[lastBucket] > 0 {
		out.WriteString(fmt.Sprintf(
			"%s%v > %v (%v%%)", sep, h.buckets[lastBucket], h.limits[lastBucket-1], h.buckets[lastBucket]*100/total))
	}
	return out.String()
}

type Scheduler interface {
	// Schedule given task at given interval.
	// Returns associated task id if scheduled successfully,
	// or a non-nil error in case of error.
	Schedule(task ScheduleTask, interval Interval,
		runAt time.Time, onlyOnce bool) (TaskID, error)

	// Cancel given task.
	Cancel(taskID TaskID) error

	// Restart scheduling.
	Start()

	// Stop scheduling.
	Stop()
}

var instance Scheduler

type taskInfo struct {
	// ID unique task identifier
	ID TaskID
	// task function to run
	task ScheduleTask
	// interval at which task is scheduled
	interval Interval
	// runtAt is next time at which task is going to be scheduled
	runAt time.Time
	// onlyOnce one time execution only
	onlyOnce bool
	// valid is true until task is not cancelled
	valid bool
	// lock for the enqueued member
	lock sync.Mutex
	// enqueued is true if task is scheduled to run
	enqueued bool
}

type manager struct {
	sync.Mutex
	// minimumInterval defines minimum task scheduling interval
	minimumInterval time.Duration
	// tasks is list of scheduled tasks
	tasks *list.List
	// currTaskID grows monotonically and gives next taskID
	currTaskID TaskID
	// ticker ticks every minimumInterval
	ticker *time.Ticker
	// started is true if schedular is not stopped
	started bool
	// enqueuedTasksLock protects enqueuedTasks
	enqueuedTasksLock sync.Mutex
	// cv signals when there are enqueuedTasks
	cv *sync.Cond
	// enqueuedTasks is list of tasks that must be run now
	enqueuedTasks *list.List
	// total number of worker goroutines that were started
	workersStarted uint64
	// total number of worker goroutines that exited
	workersExited uint64
	// histogram for interval between the ticks
	tickIntervalHist histogram
	// histogram for tick processing time
	tickProcessHist histogram
	// histogram for how late we were in enqueuing a ready task for handoff to the worker
	taskScheduleHist histogram
	// histogram for how late the worker was when starting the task
	taskStartHist histogram
	// histogram for how long a task was running
	taskDurationHist histogram
}

func (s *manager) Schedule(
	task ScheduleTask,
	interval Interval,
	runAt time.Time,
	onlyOnce bool,
) (TaskID, error) {
	s.Lock()
	defer s.Unlock()

	if task == nil {
		return TaskNone, fmt.Errorf("invalid task specified")
	}
	now := time.Now()
	if interval.nextAfter(now).Sub(now) < time.Second {
		return TaskNone, fmt.Errorf("minimum interval is a second")
	}

	s.currTaskID++
	t := &taskInfo{ID: s.currTaskID,
		task:     task,
		interval: interval,
		runAt:    interval.nextAfter(runAt),
		valid:    true,
		onlyOnce: onlyOnce,
		lock:     sync.Mutex{},
		enqueued: false}

	s.tasks.PushBack(t)
	return t.ID, nil
}

func (s *manager) Cancel(
	taskID TaskID,
) error {
	s.Lock()
	defer s.Unlock()

	for e := s.tasks.Front(); e != nil; e = e.Next() {
		t := e.Value.(*taskInfo)
		if t.ID == taskID {
			t.valid = false
			s.tasks.Remove(e)
			return nil
		}
	}
	return fmt.Errorf("invalid task ID: %v", taskID)
}

func (s *manager) Stop() {
	s.Lock()
	s.ticker.Stop()
	s.started = false
	s.Unlock()

	// Stop running any scheduled tasks.
	s.enqueuedTasksLock.Lock()
	s.enqueuedTasks.Init()
	s.enqueuedTasksLock.Unlock()
}

func (s *manager) Start() {
	s.Lock()
	defer s.Unlock()

	if !s.started {
		s.ticker = time.NewTicker(s.minimumInterval)
		s.started = true
	}
}

func (s *manager) scheduleTasks() {
	histLastPrinted := time.Now()
	prevTick := time.Now()
	for range s.ticker.C {
		now := time.Now()
		s.tickIntervalHist.record(now.Sub(prevTick))
		s.Lock()
		tasksReady := make([]*taskInfo, 0)
		elementsToRemove := make([]*list.Element, 0)
		for e := s.tasks.Front(); e != nil; e = e.Next() {
			t := e.Value.(*taskInfo)
			t.lock.Lock()
			if !t.enqueued && (now.Equal(t.runAt) || now.After(t.runAt)) {
				tasksReady = append(tasksReady, t)
				t.enqueued = true
				if t.onlyOnce {
					elementsToRemove = append(elementsToRemove, e)
				}
			}
			t.lock.Unlock()
		}
		for _, e := range elementsToRemove {
			s.tasks.Remove(e)
		}
		s.Unlock()

		var numEnqueued int
		s.enqueuedTasksLock.Lock()
		for _, t := range tasksReady {
			s.taskScheduleHist.record(time.Since(t.runAt))
			s.enqueuedTasks.PushBack(t)
		}
		s.cv.Broadcast()
		numEnqueued = s.enqueuedTasks.Len()
		s.enqueuedTasksLock.Unlock()
		workersStarted, workersExited := s.addWorkersIfNeeded(numEnqueued)

		// print stats
		if time.Since(histLastPrinted) > statsLogDuration {
			logrus.Infof("sched stats: workers: current=%v, started=%v, exited=%v",
				workersStarted-workersExited, workersStarted, workersExited)
			for _, hist := range []histogram{
				s.tickIntervalHist, s.tickProcessHist, s.taskScheduleHist, s.taskStartHist, s.taskDurationHist} {

				histStr := hist.getHistogram()
				if histStr != "" {
					logrus.Infof("sched stats: %s", hist.getHistogram())
				}
			}
			histLastPrinted = time.Now()
		}
		s.tickProcessHist.record(time.Since(now))
		prevTick = time.Now()
	}
}

// for testing
func (s *manager) getWorkerCount() int {
	s.Lock()
	defer s.Unlock()
	return int(s.workersStarted - s.workersExited)
}

// Checks if the worker must exit because it has been idle for too long. The worker
// must exit if the returned value is true.
func (s *manager) workerMustExit(lastTaskRunAt time.Time) bool {
	exit := false
	if time.Since(lastTaskRunAt) > idleTimeout {
		s.Lock()
		if (s.workersStarted - s.workersExited) > workerBatchSize {
			s.workersExited++
			exit = true
		}
		s.Unlock()
	}
	return exit
}

func (s *manager) runTasks(workerID uuid.UUID) {
	logrus.Debugf("sched worker %s starting", workerID)
	lastTaskRunAt := time.Now()
	for {
		if s.workerMustExit(lastTaskRunAt) {
			logrus.Debugf("sched worker %s exiting", workerID)
			return
		}
		s.cv.L.Lock()
		if s.enqueuedTasks.Len() == 0 {
			s.cv.Wait()
		}
		var t *taskInfo
		if s.enqueuedTasks.Len() > 0 {
			t = s.enqueuedTasks.Front().Value.(*taskInfo)
			s.enqueuedTasks.Remove(s.enqueuedTasks.Front())
		}
		s.cv.L.Unlock()
		if t != nil && t.valid {
			s.taskStartHist.record(time.Since(t.runAt))
			// run the task
			startTime := time.Now()
			t.task(t.interval)
			s.taskDurationHist.record(time.Since(startTime))

			t.lock.Lock()
			t.runAt = t.interval.nextAfter(time.Now())
			t.enqueued = false
			t.lock.Unlock()
			lastTaskRunAt = time.Now()
		}
	}
}

// Starts a new batch of worker goroutines if needed
func (s *manager) addWorkersIfNeeded(enqueuedTasks int) (uint64, uint64) {
	s.Lock()
	defer s.Unlock()
	numWorkers := s.workersStarted - s.workersExited
	if numWorkers < workerBatchSize || uint64(enqueuedTasks) > numWorkers+workerBatchSize {
		for i := 0; i < workerBatchSize && numWorkers < maxWorkers; i++ {
			go s.runTasks(uuid.New())
			numWorkers++
			s.workersStarted++
		}
	}
	return s.workersStarted, s.workersExited
}

func New(minimumInterval time.Duration) Scheduler {
	m := &manager{
		tasks:             list.New(),
		currTaskID:        0,
		minimumInterval:   minimumInterval,
		ticker:            time.NewTicker(minimumInterval),
		enqueuedTasksLock: sync.Mutex{},
		enqueuedTasks:     list.New(),
		tickIntervalHist:  newHistogram("tick interval", 1200*time.Millisecond, 10*time.Minute),
		tickProcessHist:   newHistogram("tick process durations", 500*time.Millisecond, 20*time.Minute),
		taskScheduleHist:  newHistogram("task scheduling delay", 1*time.Second, 30*time.Minute),
		taskStartHist:     newHistogram("task start delay", 1*time.Second, 30*time.Minute),
		taskDurationHist:  newHistogram("task runtime", 500*time.Millisecond, 20*time.Minute),
	}
	m.cv = sync.NewCond(&m.enqueuedTasksLock)
	m.addWorkersIfNeeded(0)
	m.Start()
	go m.scheduleTasks()
	return m
}

func Init(minimumInterval time.Duration) {
	dbg.Assert(instance == nil, "Scheduler already initialized")
	instance = New(minimumInterval)
}

func Instance() Scheduler {
	return instance
}
