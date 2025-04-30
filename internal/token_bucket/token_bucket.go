package token_bucket

import (
	"github.com/nogavadu/load_balancer/pkg/request"
	"net/http"
	"sync/atomic"
	"time"
)

type SessionStorage map[string]TokenBucket

type TokenBucket interface {
	increment()
	decrement()
	GetAmount() int
	Refill()
	HandleRequest() bool
}

type tokenBucket struct {
	cap          int
	curAmount    int32
	refillPeriod time.Duration
	lastRefill   time.Time
}

func New() func(next http.Handler) http.Handler {
	sessionStorage := make(SessionStorage)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := request.GetClientIP(r)
			_, exists := sessionStorage[clientIP]
			if !exists {
				sessionStorage[clientIP] = CreateBucket(6, time.Second*10)
			}

			if sessionStorage[clientIP].HandleRequest() {
				next.ServeHTTP(w, r)
			} else {
				http.Error(w, "too many requests", http.StatusTooManyRequests)
				return
			}
		})
	}
}

func CreateBucket(cap int, refillPeriod time.Duration) TokenBucket {
	return &tokenBucket{
		cap:          cap,
		curAmount:    int32(cap),
		refillPeriod: refillPeriod,
		lastRefill:   time.Now(),
	}
}

func (t *tokenBucket) increment() {
	if t.curAmount >= int32(t.cap) {
		return
	}
	atomic.AddInt32(&t.curAmount, 1)
}

func (t *tokenBucket) decrement() {
	if t.curAmount <= 0 {
		return
	}

	atomic.StoreInt32(&t.curAmount, t.curAmount-1)
}

func (t *tokenBucket) GetAmount() int {
	return int(t.curAmount)
}

func (t *tokenBucket) Refill() {
	if int(t.curAmount) < t.cap {
		t.increment()
		t.lastRefill = time.Now()
	}
}

func (t *tokenBucket) HandleRequest() bool {
	elapsed := time.Since(t.lastRefill)
	if elapsed >= t.refillPeriod {
		tokens := elapsed / t.refillPeriod
		for range tokens {
			t.Refill()
		}
	}
	if t.curAmount < 1 {
		return false
	}

	t.decrement()
	return true
}
