//go:build !gccgo

package agent

type Planner interface {
	Plan(input string, context *Context) PlanningResult
}

type DeterministicPlanner struct{}

var _ Planner = DeterministicPlanner{}

func (DeterministicPlanner) Plan(input string, context *Context) PlanningResult {
	return deterministicPlan(input, context)
}
