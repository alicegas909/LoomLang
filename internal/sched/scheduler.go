package sched

import "context"

// Scheduler owns attached patterns and the tempo anchor. Every ~10 ms it
// queries each pattern from its cursor up to now + ~100 ms and emits
// timestamped events. Hot-swaps land at the first cycle boundary at or after
// the cursor (or the next multiple of Quant cycles).
type Scheduler struct {
	clock Clock
}

// New returns a scheduler driven by clock.
func New(clock Clock) *Scheduler {
	return &Scheduler{clock: clock}
}

// Run blocks until ctx is cancelled. TODO(M4): the tick loop.
func (s *Scheduler) Run(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}
