// Package danmaku 提供弹幕相关的业务逻辑服务，
// 包括弹幕获取、发送弹幕、Redis缓存池管理以及WebSocket实时广播。
package danmaku

import (
	"context"
	"encoding/json"
	"fmt"

	danmakumodel "github.com/My-TuDo/B-B/backend/internal/model/danmaku"
	usermodel "github.com/My-TuDo/B-B/backend/internal/model/user"
	danmakurepo "github.com/My-TuDo/B-B/backend/internal/repository/danmaku"
	"github.com/My-TuDo/B-B/backend/internal/ws"
	"github.com/My-TuDo/B-B/backend/pkg/storage"
	"github.com/redis/go-redis/v9"
)

// Redis中弹幕池的 key 前缀和最大容量。
const poolKeyPrefix = "danmaku:pool:"
const maxPoolSize = 500

// Service 弹幕服务，封装弹幕相关的业务逻辑。
type Service struct {
	repo *danmakurepo.Repository
	rdb  *redis.Client
}

// NewService 创建弹幕服务实例。
func NewService(repo *danmakurepo.Repository, rdb *redis.Client) *Service {
	return &Service{repo: repo, rdb: rdb}
}

// GetDanmaku 获取指定视频的全部弹幕。优先从Redis缓存读取，
// 缓存未命中时回退到MySQL，同时将MySQL结果回填到Redis。
func (s *Service) GetDanmaku(ctx context.Context, videoID uint) ([]danmakumodel.DanmakuResp, error) {
	poolKey := fmt.Sprintf("%s%d", poolKeyPrefix, videoID)

	// Try Redis first — 优先从Redis ZSet读取弹幕池
	if s.rdb != nil {
		members, err := s.rdb.ZRange(ctx, poolKey, 0, -1).Result()
		if err == nil && len(members) > 0 {
			resps := make([]danmakumodel.DanmakuResp, 0, len(members))
			for _, m := range members {
				var d danmakumodel.DanmakuResp
				if err := json.Unmarshal([]byte(m), &d); err == nil {
					resps = append(resps, d)
				}
			}
			return resps, nil
		}
	}

	// Fallback to MySQL — Redis未命中，从MySQL查询
	list, err := s.repo.FindByVideoID(ctx, videoID)
	if err != nil {
		return nil, fmt.Errorf("danmaku.service.GetDanmaku: %w", err)
	}

	// 转换为响应结构
	resps := make([]danmakumodel.DanmakuResp, 0, len(list))
	for _, d := range list {
		resp := toDanmakuResp(ctx, &d)
		resps = append(resps, *resp)
	}

	// Populate Redis cache — 将查询结果回填到Redis ZSet缓存
	if s.rdb != nil && len(resps) > 0 {
		pipe := s.rdb.Pipeline()
		for i := range resps {
			data, _ := json.Marshal(resps[i])
			pipe.ZAdd(ctx, poolKey, redis.Z{Score: float64(resps[i].PlayTime), Member: string(data)})
		}
		pipe.Expire(ctx, poolKey, 0) // No expiry for danmaku pool — 弹幕池不设过期
		pipe.Exec(ctx)
	}

	return resps, nil
}

// SendDanmaku 发送一条弹幕。弹幕数据写入MySQL，同时更新Redis弹幕池，
// 并通过WebSocket向同视频房间的所有在线用户实时广播。
func (s *Service) SendDanmaku(ctx context.Context, videoID uint, userID uint, req *danmakumodel.DanmakuReq) (*danmakumodel.DanmakuResp, error) {
	// 处理可选参数默认值
	position := int8(0)
	if req.Position != nil {
		position = *req.Position
	}
	size := int8(1)
	if req.Size != nil {
		size = *req.Size
	}
	color := req.Color
	if color == "" {
		color = "#ffffff" // 默认白色
	}

	// 构建弹幕实体
	d := &danmakumodel.Danmaku{
		VideoID:  videoID,
		UserID:   userID,
		Content:  req.Content,
		Color:    color,
		Position: position,
		Size:     size,
		PlayTime: req.PlayTime,
	}

	// 写入MySQL
	if err := s.repo.Create(ctx, d); err != nil {
		return nil, fmt.Errorf("danmaku.service.SendDanmaku: %w", err)
	}

	// Re-fetch with User preloaded — 重新查询以获取关联的用户信息
	list, _ := s.repo.FindByVideoID(ctx, videoID)
	var created *danmakumodel.Danmaku
	for i := range list {
		if list[i].ID == d.ID {
			created = &list[i]
			break
		}
	}
	if created == nil {
		created = d
	}

	resp := toDanmakuResp(ctx, created)

	// Add to Redis ZSet — 将新弹幕加入Redis弹幕池
	if s.rdb != nil {
		poolKey := fmt.Sprintf("%s%d", poolKeyPrefix, videoID)
		data, _ := json.Marshal(resp)
		s.rdb.ZAdd(ctx, poolKey, redis.Z{Score: float64(resp.PlayTime), Member: string(data)})

		// Trim to maxPoolSize — 超出上限时裁剪最早的弹幕
		count, _ := s.rdb.ZCard(ctx, poolKey).Result()
		if count > maxPoolSize {
			s.rdb.ZRemRangeByRank(ctx, poolKey, 0, count-maxPoolSize-1)
		}
	}

	// Broadcast to WebSocket room so all clients see it in real-time
	// 通过WebSocket向同房间所有客户端实时广播
	if ws.DefaultHub != nil {
		data, _ := json.Marshal(resp)
		ws.DefaultHub.Broadcast <- &ws.BroadcastMessage{
			VideoID: videoID,
			Data:    data,
		}
	}

	return resp, nil
}

// toDanmakuResp 将弹幕模型转换为对外响应结构，包括用户头像预签名。
func toDanmakuResp(ctx context.Context, d *danmakumodel.Danmaku) *danmakumodel.DanmakuResp {
	resp := &danmakumodel.DanmakuResp{
		ID:       d.ID,
		Content:  d.Content,
		Color:    d.Color,
		Position: d.Position,
		Size:     d.Size,
		PlayTime: d.PlayTime,
	}
	// 填充发送者简要信息
	if d.User.ID != 0 {
		avatar := ""
		if d.User.Avatar != "" {
			avatar = storage.GetObjectURL(d.User.Avatar)
		}
		resp.User = &usermodel.UserBrief{
			ID:       d.User.ID,
			Username: d.User.Username,
			Nickname: d.User.Nickname,
			Avatar:   avatar,
		}
	}
	return resp
}
