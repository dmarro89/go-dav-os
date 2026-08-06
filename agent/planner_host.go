//go:build !gccgo

package agent

type Planner interface {
	Plan(input string, context *Context) PlanningResult
	Available() bool
}

type DeterministicPlanner struct{}

var _ Planner = DeterministicPlanner{}

func (DeterministicPlanner) Plan(input string, context *Context) PlanningResult {
	return deterministicPlan(input, context)
}

func (DeterministicPlanner) Available() bool {
	return true
}
