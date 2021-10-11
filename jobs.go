package jobs

import (
	"fmt"
	"sync"
)

type JobManager struct {
	ReceivedJobs chan Job
	CloseSignal  chan struct{}
	FailedJobs   []Job
	Concurrency  int
	tryTimes     int
	wg           sync.WaitGroup
	metas        sync.Map
}

type JobMeta struct {
	runTimes int
}

func NewJobManager(concurrency int, tryTimes int) *JobManager {
	return &JobManager{
		ReceivedJobs: make(chan Job, 100),
		CloseSignal:  make(chan struct{}),
		FailedJobs:   make([]Job, 0),
		tryTimes:     tryTimes,
		Concurrency:  concurrency,
	}
}

type JobResult int

const (
	Job_Result_Success JobResult = iota
	Job_Result_Failure
)

type Job interface {
	Id() string
	Do() JobResult
}

func (jm *JobManager) Close() {
	close(jm.CloseSignal)
}

func (jm *JobManager) ReceiveJobs(jobs []Job) {
	for _, job := range jobs {
		jm.ReceiveJob(job)
	}
}

func (jm *JobManager) ReceiveJob(job Job) {
	meta, ok := jm.metas.Load(job.Id())
	if !ok {
		// new job
		jm.metas.Store(job.Id(), &JobMeta{runTimes: 1})
	} else {
		count := meta.(*JobMeta).runTimes
		jm.metas.Store(job.Id(), &JobMeta{runTimes: count + 1})
	}
	jm.ReceivedJobs <- job
}

func (jm *JobManager) HandleJob() {
	for i := 0; i < jm.Concurrency; i++ {
		jm.wg.Add(1)
		go jm.handleJob()
	}
	jm.wg.Wait()
}

func (jm *JobManager) handleJob() {
	defer func() {
		jm.wg.Done()
	}()

	for {
		select {
		case <-jm.CloseSignal:
			fmt.Println("jobrunner closed")
			return
		case job := <-jm.ReceivedJobs:
			res := job.Do()
			if res == Job_Result_Success {
				jm.metas.Delete(job.Id())
			} else {
				meta, _ := jm.metas.Load(job.Id())
				count := meta.(*JobMeta).runTimes
				if count < jm.tryTimes {
					jm.ReceiveJob(job)
				} else {
					jm.FailedJobs = append(jm.FailedJobs, job) // 并发安全？
				}
			}
		}

	}

}
