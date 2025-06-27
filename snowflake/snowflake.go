package snowflake

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

const (
	workerBits   = 10 // 节点ID所占位数
	sequenceBits = 12 // 序列号所占位数

	maxSequence = -1 ^ (-1 << sequenceBits) // 最大序列号值（4095）

	workerMax = -1 ^ (-1 << workerBits) // 最大节点ID值（1023）
)

type Snowflake struct {
	mu        sync.Mutex
	sign      int8  //
	timestamp int64 // 时间戳 时间戳（毫秒），相对于某一时间起点
	workerID  int64 // 工作节点ID（最多支持1024个节点）
	sequence  int64 // 同一毫秒内的序列号（最多4095个）
}

// NewSnowflake 创建一个新的 Snowflake 实例
func NewSnowflake(workerID int64) (*Snowflake, error) {
	if workerID < 0 || workerID > workerMax {
		return nil, errors.New("workerID 不能超过 " + fmt.Sprintf("%d", workerMax))
	}
	return &Snowflake{
		sign:      0,
		timestamp: 0,
		workerID:  workerID,
		sequence:  0,
	}, nil
}

// NextID 生成下一个唯一ID
func (s *Snowflake) NextID() (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixNano() / 1e6 // 毫秒级时间戳

	if now < s.timestamp {
		return 0, errors.New("时间回拨")
	}

	if now == s.timestamp {
		// 当前毫秒内增加序列号
		s.sequence = (s.sequence + 1) & maxSequence
		if s.sequence == 0 {
			// 超过最大值则等待下一毫秒
			now = s.tilNextMillis(s.timestamp)
		}
	} else {
		// 新的毫秒，重置序列号
		s.sequence = 0
	}

	s.timestamp = now

	// 拼接 ID
	id := (now << (workerBits + sequenceBits)) |
		(s.workerID << sequenceBits) |
		s.sequence

	return id, nil
}

// tilNextMillis 获取下一毫秒
func (s *Snowflake) tilNextMillis(lastTimestamp int64) int64 {
	now := time.Now().UnixNano() / 1e6
	for now <= lastTimestamp {
		now = time.Now().UnixNano() / 1e6
	}
	return now
}
