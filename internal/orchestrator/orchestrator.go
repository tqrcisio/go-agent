package orchestrator

import (
	"context"
	"fmt"
	"go-agent/internal/state"
	"log"
	"time"
)

// Agent defines the interface for a pipeline step.
type Agent interface {
	Name() string
	Run(ctx context.Context, state *state.State) error
}

// Orchestrator manages the execution of agents.
type Orchestrator struct {
	Agents []Agent
}

// New creates a new Orchestrator with the given agents.
func New(agents ...Agent) *Orchestrator {
	return &Orchestrator{
		Agents: agents,
	}
}

// Run executes all agents sequentially.
func (o *Orchestrator) Run(ctx context.Context, state *state.State) error {
	for _, agent := range o.Agents {
		start := time.Now()
		log.Printf("[%s] Starting...", agent.Name())
		
		err := agent.Run(ctx, state)
		duration := time.Since(start)

		if err != nil {
			log.Printf("[%s] Failed after %s: %v", agent.Name(), duration, err)
			return fmt.Errorf("agent %s failed: %w", agent.Name(), err)
		}

		log.Printf("[%s] Finished in %s", agent.Name(), duration)
	}
	return nil
}
