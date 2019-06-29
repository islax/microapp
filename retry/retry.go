package retry

import "time"

// Do Performs repeated calls with a time delay for specific number of attempts
// or till the function returns no error
// Return Stop{err}, if you want to stop despite Error
func Do(attempts int, sleep time.Duration, fn func() error) error {
	if err := fn(); err != nil {
		if s, ok := err.(Stop); ok {
			// Return the original error for later checking
			return s.OriginalError
		}

		if attempts--; attempts > 0 {
			time.Sleep(sleep)
			return Do(attempts, 2*sleep, fn)
		}
		return err
	}
	return nil
}

// Stop is used to return error and stop retrying
// Return Stop{err}, if you want to stop despite Error
type Stop struct {
	OriginalError error
}

func (stop Stop) Error() string {
	return stop.OriginalError.Error()
}
