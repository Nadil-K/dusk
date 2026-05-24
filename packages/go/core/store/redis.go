package store

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisHitStore struct {
	client *redis.Client
	prefix string
	ttl    time.Duration
}

func NewRedisHitStore(url, keyPrefix string, ttlDays int) (*RedisHitStore, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	return &RedisHitStore{
		client: redis.NewClient(opts),
		prefix: keyPrefix,
		ttl:    time.Duration(ttlDays) * 24 * time.Hour,
	}, nil
}

func (s *RedisHitStore) streamKey() string  { return s.prefix + ":hits" }
func (s *RedisHitStore) counterKey(ep string) string { return s.prefix + ":count:" + ep }
func (s *RedisHitStore) callersKey(ep string) string { return s.prefix + ":callers:" + ep }
func (s *RedisHitStore) endpointsKey() string         { return s.prefix + ":endpoints" }

func (s *RedisHitStore) Record(hit HitEvent) error {
	ctx := context.Background()
	pipe := s.client.Pipeline()

	daysLeft := ""
	if hit.DaysUntilSunset != nil {
		daysLeft = strconv.Itoa(*hit.DaysUntilSunset)
	}
	enforced := "0"
	if hit.Enforced {
		enforced = "1"
	}
	caller := ""
	if hit.CallerID != nil {
		caller = *hit.CallerID
	}
	ua := ""
	if hit.UserAgent != nil {
		ua = *hit.UserAgent
	}

	pipe.XAdd(ctx, &redis.XAddArgs{
		Stream: s.streamKey(),
		MaxLen: 100_000,
		Approx: true,
		Values: map[string]interface{}{
			"ts":       hit.Ts.UTC().Format(time.RFC3339),
			"path":     hit.Path,
			"method":   hit.Method,
			"caller":   caller,
			"ua":       ua,
			"ep":       hit.EndpointKey,
			"days_left": daysLeft,
			"enforced": enforced,
		},
	})

	counterKey := s.counterKey(hit.EndpointKey)
	pipe.Incr(ctx, counterKey)
	pipe.Expire(ctx, counterKey, s.ttl)

	callerName := caller
	if callerName == "" {
		callerName = "anonymous"
	}
	callersKey := s.callersKey(hit.EndpointKey)
	pipe.ZIncrBy(ctx, callersKey, 1, callerName)
	pipe.Expire(ctx, callersKey, s.ttl)

	endpointsKey := s.endpointsKey()
	pipe.SAdd(ctx, endpointsKey, hit.EndpointKey)
	pipe.Expire(ctx, endpointsKey, s.ttl)

	_, err := pipe.Exec(ctx)
	return err
}

func (s *RedisHitStore) RecentHits(query HitQuery) ([]HitEvent, error) {
	limit := query.Limit
	if limit == 0 {
		limit = 100
	}
	ctx := context.Background()

	entries, err := s.client.XRevRangeN(ctx, s.streamKey(), "+", "-", int64(limit)).Result()
	if err != nil {
		return nil, err
	}

	hits := make([]HitEvent, 0, len(entries))
	for _, entry := range entries {
		hit, err := deserializeHit(entry.Values)
		if err != nil {
			continue
		}
		if query.EndpointKey != nil && hit.EndpointKey != *query.EndpointKey {
			continue
		}
		if query.CallerID != nil && (hit.CallerID == nil || *hit.CallerID != *query.CallerID) {
			continue
		}
		hits = append(hits, hit)
	}
	return hits, nil
}

func deserializeHit(fields map[string]interface{}) (HitEvent, error) {
	get := func(k string) string {
		if v, ok := fields[k]; ok {
			return fmt.Sprint(v)
		}
		return ""
	}

	ts, err := time.Parse(time.RFC3339, get("ts"))
	if err != nil {
		return HitEvent{}, err
	}

	hit := HitEvent{
		Ts:          ts,
		Path:        get("path"),
		Method:      get("method"),
		EndpointKey: get("ep"),
		Enforced:    get("enforced") == "1",
	}
	if c := get("caller"); c != "" {
		hit.CallerID = &c
	}
	if ua := get("ua"); ua != "" {
		hit.UserAgent = &ua
	}
	if dl := get("days_left"); dl != "" {
		d, err := strconv.Atoi(dl)
		if err == nil {
			hit.DaysUntilSunset = &d
		}
	}
	return hit, nil
}

func (s *RedisHitStore) EndpointSummaries(_ int) ([]EndpointSummary, error) {
	ctx := context.Background()
	endpoints, err := s.client.SMembers(ctx, s.endpointsKey()).Result()
	if err != nil {
		return nil, err
	}

	summaries := make([]EndpointSummary, 0, len(endpoints))
	for _, ep := range endpoints {
		countStr, _ := s.client.Get(ctx, s.counterKey(ep)).Result()
		total, _ := strconv.Atoi(countStr)

		topRaw, _ := s.client.ZRevRangeWithScores(ctx, s.callersKey(ep), 0, 9).Result()
		topCallers := make([]TopCaller, 0, len(topRaw))
		for _, z := range topRaw {
			topCallers = append(topCallers, TopCaller{CallerID: z.Member.(string), Count: int(z.Score)})
		}

		unique, _ := s.client.ZCard(ctx, s.callersKey(ep)).Result()
		summaries = append(summaries, EndpointSummary{
			EndpointKey:   ep,
			TotalHits:     total,
			UniqueCallers: int(unique),
			TopCallers:    topCallers,
		})
	}

	// sort by total_hits desc
	for i := 1; i < len(summaries); i++ {
		for j := i; j > 0 && summaries[j].TotalHits > summaries[j-1].TotalHits; j-- {
			summaries[j], summaries[j-1] = summaries[j-1], summaries[j]
		}
	}
	return summaries, nil
}

func (s *RedisHitStore) TotalSummary(_ int) (TotalSummary, error) {
	ctx := context.Background()
	endpoints, err := s.client.SMembers(ctx, s.endpointsKey()).Result()
	if err != nil {
		return TotalSummary{}, err
	}

	var totalHits int
	for _, ep := range endpoints {
		countStr, _ := s.client.Get(ctx, s.counterKey(ep)).Result()
		n, _ := strconv.Atoi(countStr)
		totalHits += n
	}

	var uniqueCallers int
	if len(endpoints) > 0 {
		callerKeys := make([]string, len(endpoints))
		for i, ep := range endpoints {
			callerKeys[i] = s.callersKey(ep)
		}
		tmpKey := s.prefix + ":_tmp_callers"
		n, _ := s.client.ZUnionStore(ctx, tmpKey, &redis.ZStore{Keys: callerKeys}).Result()
		uniqueCallers = int(n)
		s.client.Del(ctx, tmpKey)
	}

	return TotalSummary{
		TotalHits:             totalHits,
		UniqueCallers:         uniqueCallers,
		EndpointsWithTraffic:  len(endpoints),
		PastSunsetWithTraffic: 0,
	}, nil
}

func (s *RedisHitStore) Close() error {
	return s.client.Close()
}
