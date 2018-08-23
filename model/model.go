package model

import (
	"fmt"
	"time"
)

type State int

const (
	IDLE State = iota
	STARTING
	RUNNING
	FAILED
)

type JobDocker struct {
	Image string
}

type JobContainer struct {
	Kind   ContainerKind
	Docker JobDocker
}

type ContainerKind int

const (
	Docker ContainerKind = iota
	Mesos
)

type JobSchedule struct {
	Kind ScheduleKind
	Cron string
}

type ScheduleKind int

const (
	Cron ScheduleKind = iota
)

// TODO Support for force pull time in Docker
// TODO Support for custom args like Docker ENTRYPOINT
// TODO Support for secrets
type Job struct {
	Group       string
	Project     string
	ID          string
	Schedule    JobSchedule
	CreatedAt   time.Time
	LastStartAt time.Time
	Env         map[string]string
	Container   JobContainer
	State       State
	CPUs        float64
	Mem         float64
	Cmd         string
	User        string
}

func (j *Job) String() string {
	return fmt.Sprintf("%s:%s:%s", j.Group, j.Project, j.ID)
}