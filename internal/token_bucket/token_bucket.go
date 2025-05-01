package token_bucket

import (
	"github.com/nogavadu/load_balancer/internal/lib/request"
	"github.com/nogavadu/load_balancer/internal/lib/response"
	"net/http"
	"sync/atomic"
	"time"
)

// TODO: add methods to patch token bucket defaults settings
type SessionStorage map[string]TokenBucket

type TokenBucket interface {
	increment()
	decrement()
	Refill()
	HandleRequest() bool
}

type tokenBucket struct {
	cap          int
	curAmount    int32
	refillPeriod time.Duration
}

// New returns the middleware that implements the token bucket algorithm.
// If the bucket is empty, return an error to the client
func New() func(next http.Handler) http.Handler {
	newSession := make(chan string)
	// TODO: Replace by normal data base (sqlite, postgres)
	sessionStorage := make(SessionStorage)
	go func() {
		for {
			select {
			case clientIP := <-newSession:
				go sessionStorage[clientIP].Refill()
			}
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := request.GetClientIP(r)
			_, exists := sessionStorage[clientIP]
			if !exists {
				sessionStorage[clientIP] = CreateBucket(6, time.Second*10)
				newSession <- clientIP
			}

			if sessionStorage[clientIP].HandleRequest() {
				next.ServeHTTP(w, r)
			} else {
				response.Err(w, "too many requests", http.StatusTooManyRequests)
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

func (t *tokenBucket) Refill() {
	ticker := time.NewTicker(t.refillPeriod)
	defer ticker.Stop()

	for range ticker.C {
		if int(t.curAmount) < t.cap {
			t.increment()
		}
	}
}

func (t *tokenBucket) HandleRequest() bool {
	if t.curAmount < 1 {
		return false
	}

	t.decrement()
	return true
}
