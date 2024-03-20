// Copyright © 2023 OpenIM. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/openimsdk/tools/log"
	"github.com/redis/go-redis/v9"

	"github.com/openimsdk/tools/errs"
	"github.com/openimsdk/tools/mw/specialerror"
)

var (
	// Singleton pattern.
	redisClient redis.UniversalClient
)

const (
	maxRetry = 10 // number of retries
)

type Redis struct {
	ClusterMode    bool
	Address        []string
	Username       string
	Password       string
	EnablePipeline bool
}

// NewRedis Initialize redis connection.
func NewRedis(ctx context.Context, redisConf *Redis) (redis.UniversalClient, error) {
	if redisClient != nil {
		return redisClient, nil
	}

	if len(redisConf.Address) == 0 {
		return nil, errs.Wrap(errors.New("redis address is empty"))
	}
	specialerror.AddReplace(redis.Nil, errs.ErrRecordNotFound)
	var rdb redis.UniversalClient
	if len(redisConf.Address) > 1 || redisConf.ClusterMode {
		rdb = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:      redisConf.Address,
			Username:   redisConf.Username,
			Password:   redisConf.Password, // no password set
			PoolSize:   50,
			MaxRetries: maxRetry,
		})
	} else {
		rdb = redis.NewClient(&redis.Options{
			Addr:       redisConf.Address[0],
			Username:   redisConf.Username,
			Password:   redisConf.Password,
			DB:         0,   // use default DB
			PoolSize:   100, // connection pool size
			MaxRetries: maxRetry,
		})
	}

	var err error
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	if err = rdb.Ping(ctx).Err(); err != nil {
		errMsg := fmt.Sprintf("address:%s, username:%s, password:%s, clusterMode:%t, enablePipeline:%t", redisConf.Address, redisConf.Username,
			redisConf.Password, redisConf.ClusterMode, redisConf.EnablePipeline)
		return nil, errs.WrapMsg(err, "redis connect failed", "errMsg", errMsg)
	}

	redisClient = rdb
	log.CInfo(ctx, "REDIS connected successfully", "address", redisConf.Address, "username", redisConf.Username, "password", redisConf.Password, "clusterMode", redisConf.ClusterMode, "enablePipeline", redisConf.EnablePipeline)
	return rdb, err
}
