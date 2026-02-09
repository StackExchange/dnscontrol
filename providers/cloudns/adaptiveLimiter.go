package cloudns

import (
	"context"
	"time"

	"golang.org/x/time/rate"
)

type AdaptiveLimiter struct {
	limiter *rate.Limiter
}

func NewAdaptiveLimiter(r rate.Limit, burst int) *AdaptiveLimiter {
	return &AdaptiveLimiter{
		limiter: rate.NewLimiter(r, burst),
	}
}

func (al *AdaptiveLimiter) Wait(ctx context.Context) error {
	return al.limiter.Wait(ctx)
}

// NotifyRateLimited reserves enough tokens to pause for a period of time.
func (al *AdaptiveLimiter) NotifyRateLimited() {
	tokensToReserve := max(int(float64(al.limiter.Limit())*0.5), 1)
	al.limiter.ReserveN(time.Now(), tokensToReserve)
}
