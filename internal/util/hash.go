package util

import (
	"encoding/hex"
	"fmt"
	"hash"
	"sync"

	"github.com/spaolacci/murmur3"
)

var hasherPool = sync.Pool{
	New: func() interface{} {
		return murmur3.New64()
	},
}

func Hash(str string) (string, error) {
	hasher := hasherPool.Get().(hash.Hash64)
	_, err := hasher.Write([]byte(str))
	if err != nil {
		return "", fmt.Errorf("failed to write string to hasher: %w", err)
	}
	hashed := hex.EncodeToString(hasher.Sum(nil))
	hasher.Reset()
	hasherPool.Put(hasher)
	return hashed, nil
}
