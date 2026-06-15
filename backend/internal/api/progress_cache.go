package api

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"codecopybook/internal/storage"
)

// ProgressCache 是进度保存的 Redis 写回缓冲（T1）。
// 热路径 PATCH 只写 Redis 并标脏；后台 flusher 周期性把脏会话批量 UPSERT 落库。
// 这样把「每人每 1.2s 一次 SQL 写」合并成「每实例每几秒一批」，是冲 10k 并发的关键。
type ProgressCache struct {
	rdb   *redis.Client
	store *storage.Storage
}

const (
	progressKeyPrefix = "sess:"
	dirtySetKey       = "sess:dirty"
	progressTTL       = 7 * 24 * time.Hour
	flushInterval     = 2 * time.Second
	flushBatchSize    = 500
)

func NewProgressCache(rdb *redis.Client, store *storage.Storage) *ProgressCache {
	return &ProgressCache{rdb: rdb, store: store}
}

func progressKey(id string) string { return progressKeyPrefix + id }

func sessionToHash(s *storage.Session) map[string]interface{} {
	return map[string]interface{}{
		"userId":          s.UserID,
		"assetId":         s.AssetID,
		"relPath":         s.RelPath,
		"cursor":          s.Cursor,
		"errors":          s.Errors,
		"durationSeconds": s.DurationSeconds,
		"createdAt":       s.CreatedAt.Format(time.RFC3339Nano),
		"updatedAt":       s.UpdatedAt.Format(time.RFC3339Nano),
	}
}

func hashToSession(id string, h map[string]string) *storage.Session {
	atoi := func(k string) int { n, _ := strconv.Atoi(h[k]); return n }
	tparse := func(k string) time.Time { t, _ := time.Parse(time.RFC3339Nano, h[k]); return t }
	return &storage.Session{
		ID:              id,
		UserID:          h["userId"],
		AssetID:         h["assetId"],
		RelPath:         h["relPath"],
		Cursor:          atoi("cursor"),
		Errors:          atoi("errors"),
		DurationSeconds: atoi("durationSeconds"),
		CreatedAt:       tparse("createdAt"),
		UpdatedAt:       tparse("updatedAt"),
	}
}

// Seed 把一个会话写入 Redis（不标脏——它已经在 DB 里）。
func (c *ProgressCache) Seed(ctx context.Context, s *storage.Session) error {
	key := progressKey(s.ID)
	if err := c.rdb.HSet(ctx, key, sessionToHash(s)).Err(); err != nil {
		return err
	}
	return c.rdb.Expire(ctx, key, progressTTL).Err()
}

// loadOrSeed 确保会话在 Redis 中，并校验归属（user_id 必须匹配）。miss 时回源 DB。
func (c *ProgressCache) loadOrSeed(ctx context.Context, userID, id string) (map[string]string, error) {
	key := progressKey(id)
	h, err := c.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if len(h) == 0 {
		sess, derr := c.store.GetSession(ctx, userID, id)
		if derr != nil {
			return nil, derr // 含 storage.ErrNotFound
		}
		if serr := c.Seed(ctx, sess); serr != nil {
			return nil, serr
		}
		h = sessionToHashStr(sess)
	}
	if h["userId"] != userID {
		return nil, storage.ErrNotFound // 归属不符，按未找到处理（不泄漏他人会话）
	}
	return h, nil
}

func sessionToHashStr(s *storage.Session) map[string]string {
	return map[string]string{
		"userId":          s.UserID,
		"assetId":         s.AssetID,
		"relPath":         s.RelPath,
		"cursor":          strconv.Itoa(s.Cursor),
		"errors":          strconv.Itoa(s.Errors),
		"durationSeconds": strconv.Itoa(s.DurationSeconds),
		"createdAt":       s.CreatedAt.Format(time.RFC3339Nano),
		"updatedAt":       s.UpdatedAt.Format(time.RFC3339Nano),
	}
}

// Update 热路径：部分更新写入 Redis 并标脏，返回合并后的会话。不碰 DB。
func (c *ProgressCache) Update(ctx context.Context, userID, id string, cursor, errs, dur *int) (*storage.Session, error) {
	if _, err := c.loadOrSeed(ctx, userID, id); err != nil {
		return nil, err
	}
	key := progressKey(id)
	fields := map[string]interface{}{"updatedAt": time.Now().Format(time.RFC3339Nano)}
	if cursor != nil {
		fields["cursor"] = *cursor
	}
	if errs != nil {
		fields["errors"] = *errs
	}
	if dur != nil {
		fields["durationSeconds"] = *dur
	}
	pipe := c.rdb.TxPipeline()
	pipe.HSet(ctx, key, fields)
	pipe.Expire(ctx, key, progressTTL)
	pipe.SAdd(ctx, dirtySetKey, id)
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, err
	}
	h, err := c.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	return hashToSession(id, h), nil
}

// Get 单会话读取（Redis-first，回源 DB），校验归属。
func (c *ProgressCache) Get(ctx context.Context, userID, id string) (*storage.Session, error) {
	h, err := c.loadOrSeed(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	return hashToSession(id, h), nil
}

// GetByAssetPath 按 asset+path 查（id 未知）：DB 行取 id 与基线，再用 Redis 中更新的进度覆盖（跨设备同步看到最新）。
func (c *ProgressCache) GetByAssetPath(ctx context.Context, userID, assetID, relPath string) (*storage.Session, error) {
	sess, err := c.store.GetSessionByAssetAndPath(ctx, userID, assetID, relPath)
	if err != nil {
		return nil, err
	}
	if h, herr := c.rdb.HGetAll(ctx, progressKey(sess.ID)).Result(); herr == nil && len(h) > 0 && h["userId"] == userID {
		return hashToSession(sess.ID, h), nil
	}
	return sess, nil
}

// FlushAll 把脏会话批量落库（flusher 周期调用，也用于优雅关闭）。返回落库条数。
func (c *ProgressCache) FlushAll(ctx context.Context) (int, error) {
	total := 0
	for {
		ids, err := c.rdb.SPopN(ctx, dirtySetKey, flushBatchSize).Result()
		if err != nil && err != redis.Nil {
			return total, err
		}
		if len(ids) == 0 {
			return total, nil
		}
		sessions := make([]*storage.Session, 0, len(ids))
		for _, id := range ids {
			if h, herr := c.rdb.HGetAll(ctx, progressKey(id)).Result(); herr == nil && len(h) > 0 {
				sessions = append(sessions, hashToSession(id, h))
			}
		}
		if err := c.store.BatchUpsertSessions(ctx, sessions); err != nil {
			// 落库失败：把 id 放回脏集合，下个周期重试，避免丢更新
			back := make([]interface{}, len(ids))
			for i, id := range ids {
				back[i] = id
			}
			c.rdb.SAdd(ctx, dirtySetKey, back...)
			return total, err
		}
		total += len(sessions)
	}
}

// StartFlusher 启动后台批量落库循环（每 flushInterval 一次），ctx 取消时退出。
func (c *ProgressCache) StartFlusher(ctx context.Context, logger *log.Logger) {
	go func() {
		t := time.NewTicker(flushInterval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				if _, err := c.FlushAll(ctx); err != nil {
					logger.Printf("progress flush error: %v", err)
				}
			}
		}
	}()
}
