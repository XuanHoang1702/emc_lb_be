package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"emc_lb/src/pkg/entities"

	"github.com/redis/go-redis/v9"
)

const (
	couponListKey    = "cache:coupons:list"
	couponItemPrefix = "cache:coupon:"
	couponCacheTTL   = 5 * time.Minute
)

type CouponCacheStore interface {
	GetByCode(ctx context.Context, code string) (entities.Coupon, error)
	SetByCode(ctx context.Context, code string, coupon entities.Coupon) error
	GetAll(ctx context.Context) ([]entities.CouponResponse, error)
	SetAll(ctx context.Context, coupons []entities.CouponResponse) error
	InvalidateByCode(ctx context.Context, code string) error
	InvalidateAll(ctx context.Context) error
}

type RedisCouponCacheStore struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisCouponCacheStore(client *redis.Client) CouponCacheStore {
	return &RedisCouponCacheStore{client: client, ttl: couponCacheTTL}
}

func (s *RedisCouponCacheStore) GetByCode(ctx context.Context, code string) (entities.Coupon, error) {
	data, err := s.client.Get(ctx, buildCouponItemKey(code)).Bytes()
	if err != nil {
		return entities.Coupon{}, err
	}

	var coupon entities.Coupon
	if err := json.Unmarshal(data, &coupon); err != nil {
		return entities.Coupon{}, err
	}

	return coupon, nil
}

func (s *RedisCouponCacheStore) SetByCode(ctx context.Context, code string, coupon entities.Coupon) error {
	data, err := json.Marshal(coupon)
	if err != nil {
		return err
	}

	return s.client.Set(ctx, buildCouponItemKey(code), data, s.ttl).Err()
}

func (s *RedisCouponCacheStore) GetAll(ctx context.Context) ([]entities.CouponResponse, error) {
	data, err := s.client.Get(ctx, couponListKey).Bytes()
	if err != nil {
		return nil, err
	}

	var coupons []entities.CouponResponse
	if err := json.Unmarshal(data, &coupons); err != nil {
		return nil, err
	}

	return coupons, nil
}

func (s *RedisCouponCacheStore) SetAll(ctx context.Context, coupons []entities.CouponResponse) error {
	data, err := json.Marshal(coupons)
	if err != nil {
		return err
	}

	return s.client.Set(ctx, couponListKey, data, s.ttl).Err()
}

func (s *RedisCouponCacheStore) InvalidateByCode(ctx context.Context, code string) error {
	return s.client.Del(ctx, buildCouponItemKey(code), couponListKey).Err()
}

func (s *RedisCouponCacheStore) InvalidateAll(ctx context.Context) error {
	iter := s.client.Scan(ctx, 0, couponItemPrefix+"*", 100).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return err
	}

	keys = append(keys, couponListKey)

	if len(keys) > 0 {
		return s.client.Del(ctx, keys...).Err()
	}

	return nil
}

func buildCouponItemKey(code string) string {
	return fmt.Sprintf("%s%s", couponItemPrefix, strings.ToUpper(strings.TrimSpace(code)))
}
