package jobs

import (
	"math/rand"
	"testing"
	"time"
)

type printNumberJob struct {
	number int
}

//Do() JobResult
func (pj *printNumberJob) Do() JobResult {
	time.Sleep(5 * time.Millisecond)
	if rand.Intn(100) > 1 {
		println("job_success", pj.number)
		return Job_Result_Success
	} else {
		return Job_Result_Failure
	}
}

func TestJobs(t *testing.T) {
	jm := NewJobManager().
		SetConcurrency(5).
		SetExpectedJobCnt(2000).
		SetReTryTimes(3).
		SetTimeout(5)

	// make a new goroutine to take jobs
	go func() {
		for i := 1; i <= 2000; i++ {
			jm.ReceiveJob(&printNumberJob{i})
		}
	}()
	jm.Start()
}
