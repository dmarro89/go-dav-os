// Package agent defines the v0.5.0 Minimum Working Agent runtime boundary.
//
// The package intentionally models agent execution as typed plans and actions,
// not shell strings. A planner may be deterministic or backed by an external
// LLM bridge, but the shared runtime always validates the returned plan, routes
// risky work through a safety gate, and executes only known action kinds through
// the constrained Executor interface.
package agent
