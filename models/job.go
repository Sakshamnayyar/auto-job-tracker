package models

import "time"

type Job struct {
    Company      string
    Position     string
    JobURL       string
    ApplyDate    time.Time
    Referral     bool
    ResponseDate *time.Time
    Stage        string // Applied, Rejected, Interview
}

type FailedJob struct {
	Subject string
	Body    string
	Email   string
	Date    time.Time
	Reason  string
}

var FailedJobs []FailedJob

func AddFailedJobWithReason(job Job, reason string) {
	FailedJobs = append(FailedJobs, FailedJob{
		Subject: job.Position,
		Body:    job.JobURL,
		Email:   job.Company,
		Date:    job.ApplyDate,
		Reason:  reason,
	})
}
