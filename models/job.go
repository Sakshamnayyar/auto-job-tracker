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
