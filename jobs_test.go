package jobs

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

type printNumberJob struct {
	number int
}

//Id() string
//Do() JobResult
func (pj *printNumberJob) Do() JobResult {

	time.Sleep(500 * time.Millisecond)
	if rand.Intn(10) > 8 {
		println("success", pj.number)
		return Job_Result_Success
	} else {
		println("failure", pj.number)
		return Job_Result_Failure
	}

}

func (pj *printNumberJob) Id() string {
	return fmt.Sprintf("%v", pj.number)
}

func TestJobs(t *testing.T) {
	jm := NewJobManager(2, 2)
	go func() {
		for i := 1; i < 10; i++ {
			jm.ReceiveJob(&printNumberJob{i})
		}
		time.Sleep(1000 * time.Second)
		jm.Close()
	}()
	jm.HandleJob()
}
