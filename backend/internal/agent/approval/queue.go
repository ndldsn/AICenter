package approval

import (
	"context"
	"sync"

	"github.com/aicenter/aicenter/internal/models"
)

// Queue manages pending approvals with in-process waiters (used by the agent loop).
type Queue struct {
	mu      sync.Mutex
	pending []*pendingEntry
}

type pendingEntry struct {
	req    *models.ApprovalRequest
	decide chan models.ApprovalStatus
}

func NewQueue() *Queue {
	return &Queue{}
}

// WaitForApproval blocks until an approval decision arrives for the given request id.
// Returns the resolved status.
func (q *Queue) WaitForApproval(ctx context.Context, reqID string) (models.ApprovalStatus, error) {
	ch := make(chan models.ApprovalStatus, 1)
	q.mu.Lock()
	q.pending = append(q.pending, &pendingEntry{req: &models.ApprovalRequest{ID: reqID}, decide: ch})
	q.mu.Unlock()
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case s := <-ch:
		return s, nil
	}
}

// Resolve looks up a pending waiter by request id and delivers the decision.
func (q *Queue) Resolve(reqID string, status models.ApprovalStatus) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	for i, e := range q.pending {
		if e.req.ID == reqID {
			e.decide <- status
			q.pending = append(q.pending[:i], q.pending[i+1:]...)
			return true
		}
	}
	return false
}
