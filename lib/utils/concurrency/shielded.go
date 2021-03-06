/*
 * Copyright 2018-2021, CS Systemes d'Information, http://csgroup.eu
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

package concurrency

import (
	"encoding/json"

	"github.com/sirupsen/logrus"

	"github.com/CS-SI/SafeScale/lib/utils/data"
	"github.com/CS-SI/SafeScale/lib/utils/fail"
)

// Shielded allows to store data with controlled access to it
type Shielded struct {
	witness data.Clonable
	lock    TaskedLock
}

// NewShielded creates a new protected data
func NewShielded(witness data.Clonable) *Shielded {
	return &Shielded{
		witness: witness,
		lock:    NewTaskedLock(),
	}
}

// Clone ...
func (d *Shielded) Clone() *Shielded {
	return NewShielded(d.witness.Clone())
}

// Inspect is used to lock a clonable for read
func (d *Shielded) Inspect(task Task, inspector func(clonable data.Clonable) fail.Error) (xerr fail.Error) {
	if d == nil {
		return fail.InvalidInstanceError()
	}
	if task.IsNull() {
		return fail.InvalidParameterError("task", "cannot be null value of 'concurrency.Task'")
	}
	if inspector == nil {
		return fail.InvalidParameterError("inspector", "cannot be nil")
	}
	if d.witness == nil {
		return fail.InvalidParameterError("d.witness", "cannot be nil; use concurrency.NewShielded() to instantiate")
	}

	if xerr = d.lock.RLock(task); xerr != nil {
		return xerr
	}

	defer func() {
		unlockErr := d.lock.RUnlock(task)
		if unlockErr != nil {
			logrus.Warn(unlockErr)
		}
		if xerr == nil && unlockErr != nil {
			xerr = unlockErr
		}
	}()

	return inspector(d.witness.Clone())
}

// Alter allows to update a cloneable using a write lock
// 'alterer' can use a special error to tell the outside there was no change : fail.ErrAlteredNothing, which can be
// generated with fail.AlteredNothingError().
// The caller of the Alter() method will then be able to known, when an error occurs, if it's because there was no change.
func (d *Shielded) Alter(task Task, alterer func(data.Clonable) fail.Error) (xerr fail.Error) {
	if d == nil {
		return fail.InvalidInstanceError()
	}
	if task.IsNull() {
		return fail.InvalidParameterError("task", "cannot be null value of 'concurrency.Task'")
	}
	if alterer == nil {
		return fail.InvalidParameterError("alterer", "cannot be nil")
	}
	if d.witness == nil {
		return fail.InvalidParameterError("d.witness", "cannot be nil; use concurrency.NewData() to instantiate")
	}

	xerr = d.lock.Lock(task)
	if xerr != nil {
		return xerr
	}
	defer func() {
		unlockErr := d.lock.Unlock(task)
		if unlockErr != nil {
			logrus.Warn(unlockErr)
		}
		if xerr == nil && unlockErr != nil {
			xerr = unlockErr
		}
	}()

	clone := d.witness.Clone()
	xerr = alterer(clone)
	if xerr != nil {
		return xerr
	}
	_ = d.witness.Replace(clone)
	return nil
}

// Serialize transforms content of Shielded instance to data suitable for serialization
// Note: doesn't follow interface data.Serializable (task parameter not used in it)
func (d *Shielded) Serialize(task Task) ([]byte, fail.Error) {
	if d == nil {
		return nil, fail.InvalidInstanceError()
	}
	if task.IsNull() {
		return nil, fail.InvalidParameterError("task", "cannot be null value of 'concurrency.Task'")
	}

	var jsoned []byte
	xerr := d.Inspect(task, func(clonable data.Clonable) fail.Error {
		var innerErr error
		jsoned, innerErr = json.Marshal(clonable)
		if innerErr != nil {
			return fail.SyntaxError("failed to marshal: %s", innerErr.Error())
		}
		return nil
	})
	if xerr != nil {
		return nil, xerr
	}

	return jsoned, nil
}

// Deserialize transforms serialization data to valid content of Shielded instance
// Note: doesn't follow interface data.Serializable (task parameter not used in it)
func (d *Shielded) Deserialize(task Task, buf []byte) fail.Error {
	if d == nil {
		return fail.InvalidInstanceError()
	}
	if task.IsNull() {
		return fail.InvalidParameterError("task", "cannot be null value of 'concurrency.Task'")
	}
	if len(buf) == 0 {
		return fail.InvalidParameterError("buf", "cannot be empty []byte")
	}

	return d.Alter(task, func(clonable data.Clonable) fail.Error {
		innerErr := json.Unmarshal(buf, clonable)
		if innerErr != nil {
			return fail.SyntaxError("failed to unmarshal: %s", innerErr.Error())
		}
		return nil
	})
}
