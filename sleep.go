package gosync

import (
	"time"
)

// Sleep sleeps for the given duration
// but is interruptable by calling Interrupt() at the current SyncGroup.
func (instance *Group) Sleep(duration time.Duration) error {
	mutex := instance.NewMutex()
	condition := instance.NewCondition(mutex)
	err := condition.wait(duration, false)
	if IsTimeout(err) {
		return nil
	}
	return err
}
