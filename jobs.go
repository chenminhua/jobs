package jobs

import (
	"sync"
	"sync/atomic"
	"time"
)

type JobManager struct {
	ReceivedJobs   chan Job
	CloseSignal    chan struct{}
	FailedJobs     []Job
	concurrency    int
	retryTimes     int
	wg             sync.WaitGroup
	succeedCnt     int32
	expectedJobCnt int32 // 仅在知道有多少Job的情况下才会set，当成功job数达到totalJob时，主动结束。为0表示没有预期Job数
	timeout        int   // 过多少秒JobManager自动关闭，默认为0表示不自动关闭
}

type JobMeta struct {
	runTimes int
}

func NewJobManager() *JobManager {
	return &JobManager{
		ReceivedJobs:   make(chan Job, 100),
		CloseSignal:    make(chan struct{}),
		FailedJobs:     make([]Job, 0),
		retryTimes:     1,
		concurrency:    1,
		timeout:        0,
		expectedJobCnt: 0,
	}
}

func (jm *JobManager) SetConcurrency(concurrency int) *JobManager {
	jm.concurrency = concurrency
	return jm
}

func (jm *JobManager) SetReTryTimes(retryTimes int) *JobManager {
	jm.retryTimes = retryTimes
	return jm
}

func (jm *JobManager) SetTimeout(timeout int) *JobManager {
	jm.timeout = timeout
	return jm
}

func (jm *JobManager) SetExpectedJobCnt(total int32) *JobManager {
	jm.expectedJobCnt = total
	return jm
}


type JobResult int

const (
	Job_Result_Success JobResult = iota
	Job_Result_Failure
)

type Job interface {
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
	jm.ReceivedJobs <- job
}

func (jm *JobManager) Start() {
	// 处理JobManager超时
	if jm.timeout > 0 {
		go func() {
			time.Sleep(time.Duration(jm.timeout) * time.Second)
			jm.Close()
		}()
	}

	// FailedCollector

	for i := 0; i < jm.concurrency; i++ {
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
			//fmt.Println("jobrunner closed", jm.succeedCnt, len(jm.FailedJobs))
			return
		case job := <-jm.ReceivedJobs:
			res := DoWithRetry(job, jm.retryTimes)
			if res != Job_Result_Success {
				jm.FailedJobs = append(jm.FailedJobs, job) // 并发安全？
			} else {
				newV := atomic.AddInt32(&jm.succeedCnt, 1)
				if jm.expectedJobCnt > 0 && newV == jm.expectedJobCnt {
					jm.Close()
				}
			}
		}
	}
}
