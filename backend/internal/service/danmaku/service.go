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

const poolKeyPrefix = "danmaku:pool:"
const maxPoolSize = 500

type Service struct {
	repo *danmakurepo.Repository
	rdb  *redis.Client
}

func NewService(repo *danmakurepo.Repository, rdb *redis.Client) *Service {
	return &Service{repo: repo, rdb: rdb}
}

func (s *Service) GetDanmaku(ctx context.Context, videoID uint) ([]danmakumodel.DanmakuResp, error) {
	poolKey := fmt.Sprintf("%s%d", poolKeyPrefix, videoID)

	// Try Redis first
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

	// Fallback to MySQL
	list, err := s.repo.FindByVideoID(ctx, videoID)
	if err != nil {
		return nil, fmt.Errorf("danmaku.service.GetDanmaku: %w", err)
	}

	resps := make([]danmakumodel.DanmakuResp, 0, len(list))
	for _, d := range list {
		resp := toDanmakuResp(ctx, &d)
		resps = append(resps, *resp)
	}

	// Populate Redis cache
	if s.rdb != nil && len(resps) > 0 {
		pipe := s.rdb.Pipeline()
		for i := range resps {
			data, _ := json.Marshal(resps[i])
			pipe.ZAdd(ctx, poolKey, redis.Z{Score: float64(resps[i].PlayTime), Member: string(data)})
		}
		pipe.Expire(ctx, poolKey, 0) // No expiry for danmaku pool
		pipe.Exec(ctx)
	}

	return resps, nil
}

func (s *Service) SendDanmaku(ctx context.Context, videoID uint, userID uint, req *danmakumodel.DanmakuReq) (*danmakumodel.DanmakuResp, error) {
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
		color = "#ffffff"
	}

	d := &danmakumodel.Danmaku{
		VideoID:  videoID,
		UserID:   userID,
		Content:  req.Content,
		Color:    color,
		Position: position,
		Size:     size,
		PlayTime: req.PlayTime,
	}

	if err := s.repo.Create(ctx, d); err != nil {
		return nil, fmt.Errorf("danmaku.service.SendDanmaku: %w", err)
	}

	// Re-fetch with User preloaded
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

	// Add to Redis ZSet
	if s.rdb != nil {
		poolKey := fmt.Sprintf("%s%d", poolKeyPrefix, videoID)
		data, _ := json.Marshal(resp)
		s.rdb.ZAdd(ctx, poolKey, redis.Z{Score: float64(resp.PlayTime), Member: string(data)})

		// Trim to maxPoolSize
		count, _ := s.rdb.ZCard(ctx, poolKey).Result()
		if count > maxPoolSize {
			s.rdb.ZRemRangeByRank(ctx, poolKey, 0, count-maxPoolSize-1)
		}
	}

	// Broadcast to WebSocket room so all clients see it in real-time
	if ws.DefaultHub != nil {
		data, _ := json.Marshal(resp)
		ws.DefaultHub.Broadcast <- &ws.BroadcastMessage{
			VideoID: videoID,
			Data:    data,
		}
	}

	return resp, nil
}

func toDanmakuResp(ctx context.Context, d *danmakumodel.Danmaku) *danmakumodel.DanmakuResp {
	resp := &danmakumodel.DanmakuResp{
		ID:       d.ID,
		Content:  d.Content,
		Color:    d.Color,
		Position: d.Position,
		Size:     d.Size,
		PlayTime: d.PlayTime,
	}
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
