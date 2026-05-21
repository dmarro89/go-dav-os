//go:build !gccgo

package agent

type DeterministicPlanner struct{}

func (DeterministicPlanner) Plan(input string, context *Context) PlanningResult {
	return deterministicPlan(input, context)
}
