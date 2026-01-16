package agent

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// StreamProcessor handles LLM streaming and converts to BubbleTea messages
type StreamProcessor struct {
	client   GLMClient
	mu       sync.Mutex
	active   map[string]context.CancelFunc
}

// NewStreamProcessor creates a new stream processor
func NewStreamProcessor(client GLMClient) *StreamProcessor {
	return &StreamProcessor{
		client: client,
		active: make(map[string]context.CancelFunc),
	}
}

// ProcessStream starts a streaming chat request and returns a receive-only channel
// The channel sends StreamChunk messages as they arrive
func (s *StreamProcessor) ProcessStream(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error) {
	// Create a sub-context with cancellation
	streamCtx, cancel := context.WithCancel(ctx)

	// Generate stream ID
	streamID := fmt.Sprintf("stream_%d", len(s.active))

	// Store cancel function
	s.mu.Lock()
	s.active[streamID] = cancel
	s.mu.Unlock()

	// Start streaming from client
	chunkChan, err := s.client.StreamChat(streamCtx, req)
	if err != nil {
		s.mu.Lock()
		delete(s.active, streamID)
		s.mu.Unlock()
		cancel()
		return nil, fmt.Errorf("start stream: %w", err)
	}

	// Create output channel with buffering
	outChan := make(chan StreamChunk, 100)

	// Start goroutine to bridge chunks
	go s.bridgeChunks(streamCtx, streamID, chunkChan, outChan)

	return outChan, nil
}

// bridgeChunks bridges chunks from the client to the output channel
func (s *StreamProcessor) bridgeChunks(ctx context.Context, streamID string, in <-chan StreamChunk, out chan<- StreamChunk) {
	defer close(out)
	defer func() {
		s.mu.Lock()
		delete(s.active, streamID)
		s.mu.Unlock()
	}()

	for {
		select {
		case <-ctx.Done():
			out <- StreamChunk{Done: true, Error: ctx.Err()}
			return

		case chunk, ok := <-in:
			if !ok {
				out <- StreamChunk{Done: true}
				return
			}
			out <- chunk

			if chunk.Done {
				return
			}
		}
	}
}

// CancelStream cancels an active stream
func (s *StreamProcessor) CancelStream(streamID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if cancel, ok := s.active[streamID]; ok {
		cancel()
		delete(s.active, streamID)
	}
}

// CancelAll cancels all active streams
func (s *StreamProcessor) CancelAll() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for streamID, cancel := range s.active {
		cancel()
		delete(s.active, streamID)
		_ = streamID
	}
}

// ActiveCount returns the number of active streams
func (s *StreamProcessor) ActiveCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.active)
}

// BubblteaStreamMsg is the message type for BubbleTea
// This would be imported by the TUI package
type BubblteaStreamMsg struct {
	StreamID string
	Chunk    StreamChunk
}

// ToBubbleTea converts stream chunks to BubbleTea commands
// This is a helper function for integration
func ToBubbleTea(streamID string, chunkChan <-chan StreamChunk) <-chan BubblteaStreamMsg {
	out := make(chan BubblteaStreamMsg, 100)

	go func() {
		defer close(out)

		for chunk := range chunkChan {
			out <- BubblteaStreamMsg{
				StreamID: streamID,
				Chunk:    chunk,
			}

			if chunk.Done {
				return
			}
		}
	}()

	return out
}

// ThrottledStream wraps a stream channel to limit message rate
// This prevents UI flooding when tokens arrive very quickly
func ThrottledStream(ctx context.Context, in <-chan StreamChunk, maxTokensPerSecond int) <-chan StreamChunk {
	out := make(chan StreamChunk, 100)

	go func() {
		defer close(out)

		ticker := time.NewTicker(time.Second / time.Duration(maxTokensPerSecond))
		defer ticker.Stop()

		buffer := make([]StreamChunk, 0, 10)

		for {
			select {
			case <-ctx.Done():
				// Flush remaining buffer
				for _, chunk := range buffer {
					out <- chunk
				}
				return

			case chunk, ok := <-in:
				if !ok {
					// Flush remaining buffer
					for _, chunk := range buffer {
						out <- chunk
					}
					return
				}

				buffer = append(buffer, chunk)

				// If this is the end, flush immediately
				if chunk.Done {
					for _, c := range buffer {
						out <- c
					}
					return
				}

			case <-ticker.C:
				// Send one chunk from buffer
				if len(buffer) > 0 {
					out <- buffer[0]
					buffer = buffer[1:]
				}
			}
		}
	}()

	return out
}

// BatchedStream batches tokens together to reduce UI updates
// This combines multiple tokens into single messages for smoother rendering
func BatchedStream(ctx context.Context, in <-chan StreamChunk, batchSize int, maxDelay time.Duration) <-chan StreamChunk {
	out := make(chan StreamChunk, 100)

	go func() {
		defer close(out)

		batch := make([]string, 0, batchSize)
		timer := time.NewTimer(maxDelay)
		timer.Stop()

		flush := func() {
			if len(batch) > 0 {
				out <- StreamChunk{Token: strings.Join(batch, "")}
				batch = batch[:0]
			}
			timer.Stop()
			select {
			case <-timer.C:
			default:
			}
		}

		for {
			select {
			case <-ctx.Done():
				flush()
				return

			case chunk, ok := <-in:
				if !ok {
					flush()
					return
				}

				if chunk.Done {
					flush()
					out <- StreamChunk{Done: true}
					return
				}

				if chunk.Error != nil {
					flush()
					out <- chunk
					return
				}

				batch = append(batch, chunk.Token)

				if len(batch) >= batchSize {
					flush()
				} else if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				timer.Reset(maxDelay)

			case <-timer.C:
				flush()
			}
		}
	}()

	return out
}
