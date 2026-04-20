package rum

import (
	"context"
)

// Service the implementation guideline for the kits
// careful: only use that much func as much require
// becuase the formula will run all the dsipatch methods first than move on to the next call
// Rank current Rank of the serivce
type Service[in, out any] struct {
	context  context.Context
	Format   *TimeFormat
	Budget   *Budget
	dispatch *Dispatcher[in, out]
	Rank     int
	Name     string
}

func NewService[in, out any](ctx context.Context, settings Settings, name string) *Service[in, out] {
	b := NewBudget(0, 0)
	return &Service[in, out]{
		context:  ctx,
		Format:   NewTimeFormat(),
		Budget:   b,
		dispatch: NewDispatcher[in, out](settings),
		Rank:     1,
		Name:     name,
	}
}

// get funcs

func (d *Service[in, out]) GetContext() context.Context {
	return d.context
}

func (d *Service[in, out]) GetDispatch() *Dispatcher[in, out] {
	if d.dispatch == nil {
		return nil
	}
	return d.dispatch
}

func (d *Service[in, out]) GetName() string {
	return d.Name
}
func (d *Service[in, out]) GetBudget() *Budget {
	return d.Budget
}
func (d *Service[in, out]) GetFormat() *TimeFormat {
	return d.Format
}
func (d *Service[in, out]) GetRank() int {
	return d.Rank
}

// end

// set funcs

func (d *Service[in, out]) SetFormat(t *TimeFormat) {
	d.Format = t
}
func (d *Service[in, out]) SetBudget(b *Budget) {
	d.Budget = b
}
func (d *Service[in, out]) SetRank(Rank int) {
	d.Rank = Rank
}
func (d *Service[in, out]) SetDispatch(dp *Dispatcher[in, out]) {
	d.dispatch = dp
}

func (d *Service[in, out]) SetName(name string) {
	d.Name = name
}

// end
