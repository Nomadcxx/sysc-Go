package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// AgentType defines the specialization of a sub-agent
type AgentType string

const (
	AgentTypePlanner  AgentType = "planner"
	AgentTypeCoder   AgentType = "coder"
	AgentTypeTester  AgentType = "tester"
	AgentTypeSearch  AgentType = "search"
	AgentTypeGeneral AgentType = "general"
)

// SubAgent represents a spawned specialist agent
type SubAgent struct {
	ID        string            `json:"id"`
	Type      AgentType         `json:"type"`
	Task      string            `json:"task"`
	Context   map[string]any    `json:"context"`
	Status    string            `json:"status"` // "running", "completed", "failed"
	CreatedAt time.Time         `json:"created_at"`
	CompletedAt time.Time        `json:"completed_at,omitempty"`
	Output    string            `json:"output,omitempty"`
	Error     error             `json:"error,omitempty"`
	Metadata  map[string]any    `json:"metadata"`
}

// Orchestrator manages multiple sub-agents
type Orchestrator struct {
	client       GLMClient
	agents       map[string]*SubAgent
	mu           sync.RWMutex
	ctx          context.Context
	cancelFunc   context.CancelFunc
	nextID       int64
}

// NewOrchestrator creates a new agent orchestrator
func NewOrchestrator(client GLMClient) *Orchestrator {
	ctx, cancel := context.WithCancel(context.Background())
	return &Orchestrator{
		client:     client,
		agents:     make(map[string]*SubAgent),
		ctx:        ctx,
		cancelFunc: cancel,
		nextID:     time.Now().Unix(),
	}
}

// Spawn creates a new specialized sub-agent
func (o *Orchestrator) Spawn(agentType AgentType, task string, parentContext map[string]any) (*SubAgent, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Generate unique agent ID
	agentID := o.generateAgentID(agentType, task)

	// Build agent context by cloning parent
	agentContext := o.cloneContext(parentContext)
	agentContext["parent_agent_id"] = "main_floyd"
	agentContext["agent_type"] = string(agentType)
	agentContext["task"] = task
	agentContext["spawned_at"] = time.Now().Format(time.RFC3339)

	// Create the sub-agent
	agent := &SubAgent{
		ID:        agentID,
		Type:      agentType,
		Task:      task,
		Context:   agentContext,
		Status:    "running",
		CreatedAt: time.Now(),
		Metadata:  make(map[string]any),
	}

	o.agents[agentID] = agent

	// Start the agent in background
	go o.runAgent(agent)

	return agent, nil
}

// runAgent executes the sub-agent's task
func (o *Orchestrator) runAgent(agent *SubAgent) {
	// Build specialized prompt based on agent type
	systemPrompt := o.buildSystemPrompt(agent.Type)

	// Add task to context
	messages := []Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: o.buildTaskPrompt(agent),
		},
	}

	// Build request
	req := ChatRequest{
		Messages:    messages,
		MaxTokens:   4096,
		Temperature: 0.7,
		Stream:      false, // Sub-agents use non-streaming for complete responses
	}

	// Execute the request
	response, err := o.client.Chat(o.ctx, req)

	o.mu.Lock()
	defer o.mu.Unlock()

	if err != nil {
		agent.Status = "failed"
		agent.Error = err
		agent.CompletedAt = time.Now()
		return
	}

	agent.Output = response
	agent.Status = "completed"
	agent.CompletedAt = time.Now()
}

// buildSystemPrompt creates the system prompt for each agent type
func (o *Orchestrator) buildSystemPrompt(agentType AgentType) string {
	switch agentType {
	case AgentTypePlanner:
		return `You are a PLANNER specialist. Your role is to:
1. Analyze the task and break it down into clear steps
2. Identify dependencies and potential risks
3. Suggest an optimal execution order
4. Provide estimates for complexity

Be concise and structured. Use bullet points and numbered lists.
Focus on WHAT needs to be done, not HOW (that's for the Coder agent).`

	case AgentTypeCoder:
		return `You are a CODER specialist. Your role is to:
1. Write clean, working, idiomatic code
2. Follow existing project patterns and conventions
3. Include necessary imports and error handling
4. Provide brief explanations of complex logic

Focus on practical, working solutions. Avoid "word salad" - get straight to the code.
If you're unsure, ask clarifying questions rather than making assumptions.`

	case AgentTypeTester:
		return `You are a TESTER specialist. Your role is to:
1. Identify edge cases and potential failure modes
2. Suggest test scenarios (unit, integration, e2e)
3. Review code for correctness and safety
4. Identify potential bugs before they reach production

Be thorough but practical. Prioritize high-impact test cases.`

	case AgentTypeSearch:
		return `You are a SEARCH specialist. Your role is to:
1. Analyze codebases to find specific patterns, functions, or bugs
2. Trace dependencies and data flows
3. Locate where features are implemented
4. Identify files relevant to a given task

Be methodical. Report what you find with file paths and line numbers.`

	default:
		return `You are a helpful assistant working as part of the FLOYD agent system.
Be concise, accurate, and practical. Focus on delivering working solutions.`
	}
}

// buildTaskPrompt creates the user prompt for the agent
func (o *Orchestrator) buildTaskPrompt(agent *SubAgent) string {
	var prompt string

	// Add context about the task
	prompt += fmt.Sprintf("TASK: %s\n\n", agent.Task)

	// Add relevant context if available
	if ctx, err := json.MarshalIndent(agent.Context, "", "  "); err == nil {
		prompt += fmt.Sprintf("CONTEXT:\n%s\n\n", string(ctx))
	}

	prompt += "Please complete this task with your specialized expertise."

	return prompt
}

// generateAgentID creates a unique ID for a sub-agent
func (o *Orchestrator) generateAgentID(agentType AgentType, task string) string {
	o.nextID++
	// Create short hash of task
	taskHash := fmt.Sprintf("%x", len(task)+int(o.nextID))
	return fmt.Sprintf("%s_%d_%s", agentType, o.nextID, taskHash[:min(8, len(taskHash))])
}

// cloneContext creates a copy of the parent context for the sub-agent
func (o *Orchestrator) cloneContext(parent map[string]any) map[string]any {
	cloned := make(map[string]any, len(parent))
	for k, v := range parent {
		cloned[k] = v
	}
	return cloned
}

// CollectResults retrieves the output from a completed sub-agent
func (o *Orchestrator) CollectResults(agentID string) (*SubAgent, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	agent, exists := o.agents[agentID]
	if !exists {
		return nil, fmt.Errorf("agent not found: %s", agentID)
	}

	if agent.Status != "completed" {
		return nil, fmt.Errorf("agent not completed: %s (status: %s)", agentID, agent.Status)
	}

	return agent, nil
}

// WaitForCompletion waits for an agent to complete
func (o *Orchestrator) WaitForCompletion(agentID string, timeout time.Duration) (*SubAgent, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		o.mu.RLock()
		agent, exists := o.agents[agentID]
		o.mu.RUnlock()

		if !exists {
			return nil, fmt.Errorf("agent not found: %s", agentID)
		}

		if agent.Status == "completed" {
			return agent, nil
		}

		if agent.Status == "failed" {
			return agent, agent.Error
		}

		time.Sleep(100 * time.Millisecond)
	}

	return nil, fmt.Errorf("timeout waiting for agent: %s", agentID)
}

// ListAgents returns all active agents
func (o *Orchestrator) ListAgents() []*SubAgent {
	o.mu.RLock()
	defer o.mu.RUnlock()

	agents := make([]*SubAgent, 0, len(o.agents))
	for _, agent := range o.agents {
		agents = append(agents, agent)
	}
	return agents
}

// GetAgent returns a specific agent by ID
func (o *Orchestrator) GetAgent(agentID string) (*SubAgent, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	agent, exists := o.agents[agentID]
	if !exists {
		return nil, fmt.Errorf("agent not found: %s", agentID)
	}

	return agent, nil
}

// Cancel stops all running agents
func (o *Orchestrator) Cancel() {
	o.cancelFunc()

	o.mu.Lock()
	defer o.mu.Unlock()

	for _, agent := range o.agents {
		if agent.Status == "running" {
			agent.Status = "cancelled"
			agent.CompletedAt = time.Now()
		}
	}
}

// Cleanup removes completed agents older than the specified duration
func (o *Orchestrator) Cleanup(olderThan time.Duration) int {
	o.mu.Lock()
	defer o.mu.Unlock()

	cutoff := time.Now().Add(-olderThan)
	removed := 0

	for id, agent := range o.agents {
		if agent.CompletedAt.Before(cutoff) {
			delete(o.agents, id)
			removed++
		}
	}

	return removed
}

// SpawnParallel spawns multiple agents in parallel
func (o *Orchestrator) SpawnParallel(tasks []Task) ([]*SubAgent, error) {
	agents := make([]*SubAgent, 0, len(tasks))
	var wg sync.WaitGroup
	errChan := make(chan error, len(tasks))

	for _, task := range tasks {
		wg.Add(1)
		go func(t Task) {
			defer wg.Done()
			agent, err := o.Spawn(t.Type, t.Task, t.Context)
			if err != nil {
				errChan <- err
				return
			}
			agents = append(agents, agent)
		}(task)
	}

	wg.Wait()
	close(errChan)

	// Check for errors
	if len(errChan) > 0 {
		return agents, <-errChan
	}

	return agents, nil
}

// Task represents a unit of work for a sub-agent
type Task struct {
	Type    AgentType      `json:"type"`
	Task    string         `json:"task"`
	Context map[string]any `json:"context"`
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
