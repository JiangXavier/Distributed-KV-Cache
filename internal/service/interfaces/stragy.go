package interfaces

import "time"

type CacheStrategy interface {
	Get(string) (Value, *time.Time, bool)
	Put(string, Value)
	CleanUp(ttl time.Duration)
	Len() int
}

type Value interface {
	Len() int
}

type Entry struct {
	Key      string
	Value    Value
	UpdateAt *time.Time // last time update the cache
}

// Expired ttl expired function
func (ele *Entry) Expired(duration time.Duration) (ok bool) {
	if ele.UpdateAt == nil {
		ok = false
	} else {
		ok = ele.UpdateAt.Add(duration).Before(time.Now())
	}
	return
}

// ttl touch function
func (ele *Entry) Touch() {
	nowTime := time.Now()
	ele.UpdateAt = &nowTime
}
