// Copyright 2018, OpenCensus Authors
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

package ocredis

import (
	"context"
	"fmt"

	"go.opencensus.io/trace"

	"github.com/go-redis/redis"
)

// NewClient creates a trace instrumented redis.Client.
// It takes context.Context as a first argument because
// clients can cheaply be created an discarded, and for a
// full trace you'll want to create a client that uses the provided
// context. For example:
//    req, _ := http.NewRequest()
//    req = req.WithContext(ctx)
//    ...
//    client := NewClient(req.Context(), redisAddr)
//    client.HMSet("results", keyValueDict)
func NewClient(ctx context.Context, addr string) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	}).WithContext(ctx)

	client.WrapProcess(PerCommandTracer(client.Context()))
	return client
}

// PerCommandTracer provides the instrumented WrapProcess function that you can attach to any
// client. It specifically takes in a context.Context as the first argument because
// you could be using the same client but wrapping it in a different context.
func PerCommandTracer(ctx context.Context) func(oldProcess func(cmd redis.Cmder) error) func(redis.Cmder) error {
	return func(fn func(cmd redis.Cmder) error) func(redis.Cmder) error {
		return func(cmd redis.Cmder) error {
			_, span := trace.StartSpan(ctx, fmt.Sprintf("redis-go/%s", cmd.Name()))
			err := fn(cmd)
			if err != nil {
				span.SetStatus(trace.Status{Code: int32(trace.StatusCodeUnknown), Message: err.Error()})
			}
			span.End()
			return err
		}
	}
}
