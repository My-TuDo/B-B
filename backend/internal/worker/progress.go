package worker

import (
	"sync"

	"github.com/My-TuDo/B-B/backend/pkg/logger"
	"go.uber.org/zap"
)

// ProgressUpdate is pushed to SSE subscribers during transcode.
type ProgressUpdate struct {
	VideoID  uint   `json:"video_id"`
	Status   int8   `json:"status"`
	Progress uint8  `json:"progress"`
	Quality  string `json:"quality,omitempty"`
	ErrorMsg string `json:"error_msg,omitempty"`
}

// ProgressBroker manages SSE subscribers for real-time transcode progress.
type ProgressBroker struct {
	mu   sync.RWMutex
	subs map[uint]map[chan ProgressUpdate]struct{}
}

var brokerInstance = &ProgressBroker{
	subs: make(map[uint]map[chan ProgressUpdate]struct{}),
}

// GetBroker returns the global progress broker.
func GetBroker() *ProgressBroker {
	return brokerInstance
}

// Subscribe returns a buffered channel for videoID. Call Unsubscribe to clean up.
func (b *ProgressBroker) Subscribe(videoID uint) chan ProgressUpdate {
	ch := make(chan ProgressUpdate, 32)
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.subs[videoID] == nil {
		b.subs[videoID] = make(map[chan ProgressUpdate]struct{})
	}
	b.subs[videoID][ch] = struct{}{}
	return ch
}

// Unsubscribe removes a subscriber and closes the channel.
func (b *ProgressBroker) Unsubscribe(videoID uint, ch chan ProgressUpdate) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if m := b.subs[videoID]; m != nil {
		delete(m, ch)
		if len(m) == 0 {
			delete(b.subs, videoID)
		}
	}
	// Drain remaining messages before close to avoid goroutine leak
	for {
		select {
		case <-ch:
		default:
			close(ch)
			return
		}
	}
}

// Publish broadcasts an update to all subscribers of the given videoID.
// Non-blocking: drops message if a subscriber's buffer is full.
func (b *ProgressBroker) Publish(update ProgressUpdate) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.subs[update.VideoID] {
		select {
		case ch <- update:
		default:
			logger.Warn("progress broker: subscriber full, dropping update",
				zap.Uint("video_id", update.VideoID))
		}
	}
}
