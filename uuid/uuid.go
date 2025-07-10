package uuid

import (
	"errors"
	"math/rand"
	"sync"
	"time"
)

const (
	PROJECT_ID_BITS = 8
	TIMESTAMP_BITS  = 32
	RAND_BITS       = 16
	COUNT_BITS      = 8
)

var (
	pool             = make(chan uint64, 5)
	lastSec   uint64 = 0
	lastCount uint8  = 1
	rng       sync.Mutex
	localRand = rand.New(rand.NewSource(time.Now().UnixNano())) // 独立的随机数生成器
)

func init() {
	go gen()
}

func gen() {
	for {
		currentSec := uint64(time.Now().Unix() & 0xFFFFFFFF)
		rng.Lock()
		if currentSec != lastSec {
			lastCount = 1
			lastSec = currentSec
		}

		c := uint64(lastSec << (RAND_BITS + COUNT_BITS))
		randVal := localRand.Uint64() & 0x0000000000FFFF00 // 无需每次重置 Seed
		c += randVal
		c += uint64(lastCount)
		lastCount++
		rng.Unlock()

		pool <- c
	}
}

func Get(projectID uint8) (uint64, error) {
	select {
	case c := <-pool:
		return uint64(projectID)<<(TIMESTAMP_BITS+RAND_BITS+COUNT_BITS) + c, nil
	default:
		return 0, errors.New("gen uuid fail")
	}
}

func MustGet(projectID uint8) uint64 {
	select {
	case c := <-pool:
		return uint64(projectID)<<(TIMESTAMP_BITS+RAND_BITS+COUNT_BITS) + c
	default:
		return uint64(time.Now().Unix())
	}
}
