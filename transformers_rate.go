package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"net/http"
	"time"
)

// RequestRateLimiterTransformer is a transformer that rate limits the requests based on configurable patterns.
// This transformer will need Redis to work.
type RequestRateLimiterTransformer struct {
	ActivateOnTags   []string
	RedisUri         string
	redisClient      *redis.Client
	Vary             string
	Limit            int64
	Range            string
	_range           time.Duration
	PrometheusPrefix string
}

func NewRateLimiterTransformer(activateOnTags []string, params map[string]interface{}) (*RequestRateLimiterTransformer, error) {
	transformer := RequestRateLimiterTransformer{ActivateOnTags: activateOnTags}
	err := DecodeAndTempl(params, &transformer, nil, []string{"Vary"})
	if err != nil {
		return nil, err
	}
	transformer._range, err = time.ParseDuration(transformer.Range)
	if err != nil {
		return nil, err
	}
	redisOptions, err := redis.ParseURL(transformer.RedisUri)
	if err != nil {
		return nil, err
	}
	transformer.redisClient = redis.NewClient(redisOptions)
	_, err = transformer.redisClient.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}
	return &transformer, nil
}

func (s *RequestRateLimiterTransformer) getPrometheusPrefix() string {
	if s.PrometheusPrefix == "" {
		return "rate_rejections"
	}
	return "rate_rejections_" + s.PrometheusPrefix
}

func (t *RequestRateLimiterTransformer) Transform(wrapper *APIWrapper) (*APIWrapper, error) {
	vary, _ := Templ(t.Vary, wrapper)
	cmd := t.redisClient.LLen(context.Background(), vary)
	wrapper.ApplyHeaders.Set("RateLimit-Limit", fmt.Sprintf("%d %d;window=%d", t.Limit, t.Limit, int(t._range.Seconds())))
	if cmd.Err() != nil {
		log.Error("error while reading length from redis", cmd.Err(), nil)
		return wrapper, nil
	}
	current := cmd.Val()
	if current > t.Limit {
		if prom != nil {
			prom.CustomCounter(t.getPrometheusPrefix()).Inc()
		}
		return nil, errors.New("rate_limit")
	} else {
		if t.redisClient.Exists(context.Background(), vary).Val() == 0 {
			pipeline := t.redisClient.TxPipeline()
			if err := pipeline.RPush(context.Background(), vary, vary).Err(); err != nil {
				log.Error("error while pushing to Redis in rate limiter", err, nil)
			}
			if err := pipeline.Expire(context.Background(), vary, t._range).Err(); err != nil {
				log.Error("error while setting Redis TTL in rate limiter", err, nil)
			}
			if _, err := pipeline.Exec(context.Background()); err != nil {
				log.Error("error while setting running Redis pipeline in rate limiter", err, nil)
			}
		} else {
			t.redisClient.RPushX(context.Background(), vary, vary)
		}
	}
	return wrapper, nil
}

func (t *RequestRateLimiterTransformer) ErrorMatches(err error) bool {
	return err.Error() == "rate_limit"
}

func (t *RequestRateLimiterTransformer) HandleError(writer *http.ResponseWriter) {
	(*writer).WriteHeader(429)
}

func (t *RequestRateLimiterTransformer) ShouldExpandRequest() bool {
	return false
}

func (t *RequestRateLimiterTransformer) ShouldExpandResponse() bool {
	return false
}

func (t *RequestRateLimiterTransformer) IsActive(wrapper *APIWrapper) bool {
	return wrapper.HasTag(t.ActivateOnTags)
}
