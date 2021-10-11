package jobs

func DoWithRetry(job Job, times int) JobResult {
	count := 0
	for {
		count++
		res := job.Do()
		if res == Job_Result_Success {
			return Job_Result_Success
		} else if count >= times {
			return Job_Result_Failure
		}
	}
}