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
	log              *STLogHelper
}

// NewRequestRateLimiterTransformer is the constructor for RequestRateLimiterTransformer
func NewRequestRateLimiterTransformer(activateOnTags []string, logCfg *STLogConfig, params map[string]any) (*RequestRateLimiterTransformer, error) {
	transformer := RequestRateLimiterTransformer{ActivateOnTags: activateOnTags, log: NewSTLogHelper(logCfg)}
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
	transformer.log.PrometheusRegisterCounter("rate_limited")
	return &transformer, nil
}

func (t *RequestRateLimiterTransformer) Transform(wrapper *APIWrapper) (*APIWrapper, error) {
	t.log.Log("triggering rate limiter", wrapper, t.log.Debug)
	// compiling the `vary` template in real time
	vary, _ := Templ(t.Vary, wrapper)
	// getting the length of the item retrieved with the value of `vary` as key
	cmd := t.redisClient.LLen(context.Background(), vary)
	// setting response header displaying the rate limit
	wrapper.ApplyHeaders.Set("RateLimit-Limit", fmt.Sprintf("%d %d;window=%d", t.Limit, t.Limit, int(t._range.Seconds())))
	if cmd.Err() != nil {
		t.log.LogErr("error while reading length from redis", cmd.Err(), wrapper, t.log.Error)
		return wrapper, nil
	}
	current := cmd.Val()
	// if the retrieved count is greater than the configured limit
	if current > t.Limit {
		t.log.PrometheusCounterInc("rate_limited")
		t.log.LogErr("rate limited. Request dropped", nil, wrapper, t.log.Warn)
		// we cut the request and return an error
		return nil, errors.New("rate_limit")
	} else {
		// if the retrieved count is less than the configured limit and
		// the entry does not exist in Redis
		if t.redisClient.Exists(context.Background(), vary).Val() == 0 {
			pipeline := t.redisClient.TxPipeline()
			// we push the item
			if err := pipeline.RPush(context.Background(), vary, vary).Err(); err != nil {
				t.log.LogErr("error while pushing to Redis in rate limiter", err, wrapper, t.log.Error)
			}
			// set the expiry time
			if err := pipeline.Expire(context.Background(), vary, t._range).Err(); err != nil {
				t.log.LogErr("error while setting Redis TTL in rate limiter", err, wrapper, t.log.Error)
			}
			// and execute the pipeline
			if _, err := pipeline.Exec(context.Background()); err != nil {
				log.Error("error while setting running Redis pipeline in rate limiter", err, nil)
			}
		} else {
			// if the entry already existed, then we just push a new item
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
