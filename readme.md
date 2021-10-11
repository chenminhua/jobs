**一个简单易用的并发任务执行lib**

```go
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

jm := NewJobManager().
    SetConcurrency(5).       // 5个worker
    SetExpectedJobCnt(2000).  // 最多2000个job
    SetReTryTimes(3).        // 如果有失败，最多做3次
    SetTimeout(5)           // 5秒没完成，强制结束

// make a new goroutine to take jobs
go func() {
    for i := 1; i <= 2000; i++ {
        jm.ReceiveJob(&printNumberJob{i})
    }
}()

// 开始take job 
jm.Start()
```

