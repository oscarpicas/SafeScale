/*
 * Copyright 2018-2020, CS Systemes d'Information, http://www.c-s.fr
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	uuidpkg "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"

	"github.com/CS-SI/SafeScale/lib/server/iaas"
	"github.com/CS-SI/SafeScale/lib/utils/concurrency"
	"github.com/CS-SI/SafeScale/lib/utils/scerr"
)

// Job is the interface of a daemon job
type Job interface {
	SafeGetID() string
	SafeGetName() string
	SafeGetTask() concurrency.Task
	SafeGetService() iaas.Service
	SafeGetDuration() time.Duration
	String() string
	Abort() error
	Aborted() bool
	Close()
}

// job contains the information needed by safescaled to execute a request
type job struct {
	description string
	uuid        string
	tenant      string
	task        concurrency.Task
	cancel      context.CancelFunc
	service     iaas.Service
	startTime   time.Time
}

// NewJob creates a new instance of struct Job
func NewJob(ctx context.Context, cancel context.CancelFunc, svc iaas.Service, description string) (Job, error) {
	if ctx == nil {
		return nil, scerr.InvalidParameterError("ctx", "cannot be nil")
	}
	if cancel == nil {
		return nil, scerr.InvalidParameterError("cancel", "cannot be nil")
	}
	if svc == nil {
		return nil, scerr.InvalidParameterError("svc", "cannot be nil")
	}

	var (
		md  metadata.MD
		id  string
		err error
		ok  bool
	)

	md, ok = metadata.FromIncomingContext(ctx)
	if !ok {
		logrus.Warn("context does not contain a grpc uuid, generating one")
		uuid, err := uuidpkg.NewV4()
		if err != nil {
			return nil, scerr.Wrap(err, "failed to generate uuid for job")
		}
		id = uuid.String()
	} else {
		u := md.Get("uuid")
		if len(u) == 0 {
			return nil, scerr.InvalidParameterError("ctx", "does not contain a grpc uuid")
		}

		if id = u[0]; id == "" {
			return nil, scerr.InvalidParameterError("ctx", "does not contain a valid grpc uuid")
		}
	}

	task, err := concurrency.NewTaskWithContext(ctx, nil)
	if err != nil {
		return nil, err
	}
	err = task.SetID("job-task:" + id)
	if err != nil {
		return nil, err
	}

	nj := job{
		description: description,
		uuid:        id,
		task:        task,
		cancel:      cancel,
		service:     svc,
		tenant:      svc.GetName(),
		startTime:   time.Now(),
	}
	err = register(&nj)
	if err != nil {
		return nil, err
	}
	return &nj, nil
}

// SafeGetID returns the id of the job (ie the uuid of gRPC message)
func (j *job) SafeGetID() string {
	return j.uuid
}

// SafeGetName returns the name (== id) of the job
func (j *job) SafeGetName() string {
	return j.uuid
}

// SafeGetTask returns the task instance
func (j *job) SafeGetTask() concurrency.Task {
	return j.task
}

// SafeGetService returns the service instance
func (j *job) SafeGetService() iaas.Service {
	return j.service
}

// SafeGetDuration returns the duration of the job
func (j *job) SafeGetDuration() time.Duration {
	return time.Since(j.startTime)
}

// Abort tells the job it has to abort operations
func (j *job) Abort() error {
	if j == nil {
		return scerr.InvalidInstanceError()
	}
	if j.cancel == nil {
		return scerr.InvalidInstanceContentError("j.cancel", "cannot be nil")
	}
	j.cancel()
	j.cancel = nil
	return nil
}

// Aborted tells if the job has been aborted
func (j *job) Aborted() bool {
	status, _ := j.task.GetStatus()

	return status == concurrency.ABORTED
}

// Close tells the job to wait for end of operation; this ensure everything is cleaned up correctly
func (j *job) Close() {
	_ = deregister(j)
	if j.cancel != nil {
		j.cancel()
	}
}

// String returns a string representation of job information
func (j *job) String() string {
	return fmt.Sprintf("Job: %s (started at %s)", j.description, j.startTime.String())
}

var (
	jobMap          = map[string]Job{}
	mutexJobManager sync.Mutex
)

// register ...
func register(job Job) error {
	mutexJobManager.Lock()
	defer mutexJobManager.Unlock()

	jobMap[job.SafeGetID()] = job
	return nil
}

// deregister ...
func deregister(job Job) error {
	if job == nil {
		return scerr.InvalidParameterError("job", "cannot be nil")
	}
	return deregisterUUID(job.SafeGetID())
}

func deregisterUUID(uuid string) error {
	if uuid == "" {
		return scerr.InvalidParameterError("uuid", "cannot be empty string")
	}
	mutexJobManager.Lock()
	defer mutexJobManager.Unlock()

	if _, ok := jobMap[uuid]; !ok {
		return scerr.NotFoundError(fmt.Sprintf("no job identified by '%s' found", uuid))
	}
	delete(jobMap, uuid)
	return nil
}

// AbortJobByID asks the job identified by 'id' to abort
func AbortJobByID(id string) error {
	if id == "" {
		return scerr.InvalidParameterError("id", "cannot be empty string")
	}
	if job, ok := jobMap[id]; ok {
		err := job.Abort()
		if err != nil {
			return scerr.Wrap(err, fmt.Sprintf("failed to stop job '%s'", id))
		}
		return nil
	}
	return scerr.NotFoundError(fmt.Sprintf("no job identified by '%s' found", id))
}

// ListJobs ...
func ListJobs() map[string]string {
	listMap := map[string]string{}
	for uuid, job := range jobMap {
		listMap[uuid] = job.SafeGetName()
	}
	return listMap
}