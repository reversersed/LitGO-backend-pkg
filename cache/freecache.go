package freecache

import (
	"sync"

	"github.com/coocood/freecache"
)

type freecacherepo struct {
	sync.Mutex
	cache *freecache.Cache
}

func NewFreeCache(size int) *freecacherepo {
	return &freecacherepo{cache: freecache.NewCache(size)}
}
func (c *freecacherepo) EntryCount() int64 {
	c.Lock()
	defer c.Unlock()

	return c.cache.EntryCount()
}
func (r *freecacherepo) Get(uuid []byte) ([]byte, error) {
	r.Lock()
	defer r.Unlock()
	got, err := r.cache.Get(uuid)
	return got, err
}

func (r *freecacherepo) Set(key, val []byte, expireIn int) error {
	r.Lock()
	defer r.Unlock()

	err := r.cache.Set(key, val, expireIn)
	if err != nil {
		return err
	}
	return nil
}

func (r *freecacherepo) Delete(key []byte) (affected bool) {
	r.Lock()
	defer r.Unlock()

	return r.cache.Del(key)
}
