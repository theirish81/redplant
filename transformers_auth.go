package main

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"net/http"
)

type RequestCookieToTokenTransformer struct {
	ActivateOnTags []string
	RedisUri       string
	redisClient    *redis.Client
	CookieName     string
}

func NewCookieToTokenTransformer(activateOnTags []string, params map[string]interface{}) (*RequestCookieToTokenTransformer, error) {
	transformer := RequestCookieToTokenTransformer{ActivateOnTags: activateOnTags}
	err := DecodeAndTempl(params, &transformer, nil, []string{})
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

func (t *RequestCookieToTokenTransformer) Transform(wrapper *APIWrapper) (*APIWrapper, error) {
	cookie, err := wrapper.Request.Cookie(t.CookieName)
	if err != nil {
		return nil, errors.New("no_auth")
	}
	cmd := t.redisClient.Get(context.Background(), cookie.Value)
	if cmd.Err() != nil {
		if cmd.Err() == redis.Nil {
			return nil, errors.New("no_auth")
		}
		return nil, cmd.Err()
	}
	wrapper.Request.Header.Set("Authorization", "Bearer "+cmd.Val())
	return wrapper, nil
}

func (t *RequestCookieToTokenTransformer) ShouldExpandRequest() bool {
	return false
}

func (t *RequestCookieToTokenTransformer) ShouldExpandResponse() bool {
	return false
}

func (t *RequestCookieToTokenTransformer) ErrorMatches(err error) bool {
	return err.Error() == "no_auth"
}

func (t *RequestCookieToTokenTransformer) HandleError(writer *http.ResponseWriter) {
	(*writer).WriteHeader(401)
}

func (t *RequestCookieToTokenTransformer) IsActive(wrapper *APIWrapper) bool {
	return wrapper.HasTag(t.ActivateOnTags)
}
