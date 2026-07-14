// Package worker 提供视频转码的后台任务处理功能（本文件为进度广播模块）。
package worker

import (
	"sync"

	"github.com/My-TuDo/B-B/backend/pkg/logger"
	"go.uber.org/zap"
)

// ProgressUpdate 是转码过程中推送给 SSE 订阅者的进度快照。
// 包含视频 ID、当前状态、进度百分比、当前处理的清晰度和错误信息。
type ProgressUpdate struct {
	VideoID  uint   `json:"video_id"`
	Status   int8   `json:"status"`
	Progress uint8  `json:"progress"`
	Quality  string `json:"quality,omitempty"`
	ErrorMsg string `json:"error_msg,omitempty"`
}

// ProgressBroker 管理 SSE 订阅者，负责实时转码进度的发布-订阅。
// 每个 videoID 对应一组订阅者 channel，支持非阻塞广播。
type ProgressBroker struct {
	mu   sync.RWMutex
	subs map[uint]map[chan ProgressUpdate]struct{} // videoID → 订阅者 channel 集合
}

// brokerInstance 是全局唯一的 ProgressBroker 单例。
var brokerInstance = &ProgressBroker{
	subs: make(map[uint]map[chan ProgressUpdate]struct{}),
}

// GetBroker 返回全局的进度 broker 实例。
func GetBroker() *ProgressBroker {
	return brokerInstance
}

// Subscribe 为指定 videoID 注册一个新的订阅者，返回一个带缓冲的 channel。
// 调用方应在不再需要时调用 Unsubscribe 清理资源。
func (b *ProgressBroker) Subscribe(videoID uint) chan ProgressUpdate {
	ch := make(chan ProgressUpdate, 32) // 32 条缓冲，避免短时阻塞
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.subs[videoID] == nil {
		b.subs[videoID] = make(map[chan ProgressUpdate]struct{})
	}
	b.subs[videoID][ch] = struct{}{}
	return ch
}

// Unsubscribe 移除指定订阅者，排空 channel 中的残留消息后关闭 channel。
func (b *ProgressBroker) Unsubscribe(videoID uint, ch chan ProgressUpdate) {
	b.mu.Lock()
	defer b.mu.Unlock()
	// 从订阅映射中删除该 channel
	if m := b.subs[videoID]; m != nil {
		delete(m, ch)
		// 如果该 videoID 下没有订阅者了，清理整个映射条目
		if len(m) == 0 {
			delete(b.subs, videoID)
		}
	}
	// 关闭前排空残留消息，防止 goroutine 泄漏
	for {
		select {
		case <-ch:
		default:
			close(ch)
			return
		}
	}
}

// Publish 向指定 videoID 的所有订阅者广播一条进度更新。
// 非阻塞：如果某个订阅者的缓冲区已满，则丢弃该消息并记录警告。
func (b *ProgressBroker) Publish(update ProgressUpdate) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.subs[update.VideoID] {
		select {
		case ch <- update:
		default:
			// 订阅者消费过慢，丢弃消息
			logger.Warn("progress broker: subscriber full, dropping update",
				zap.Uint("video_id", update.VideoID))
		}
	}
}
