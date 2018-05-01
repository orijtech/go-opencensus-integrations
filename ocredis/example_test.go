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

package ocredis_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/go-redis/redis"
	ocredis "github.com/orijtech/go-opencensus-integrations/ocredis"

	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"
)

type printExporter int

func (pe *printExporter) ExportSpan(sd *trace.SpanData) {
	fmt.Printf("\nSpanData:\nName: %s\nTraceID: %x\nSpanID: %x\nParentSpanID: %x\n\n",
		sd.Name, sd.TraceID, sd.SpanID, sd.ParentSpanID)
}

var pe = new(printExporter)

func Example_PerCommandTracer() {
	// Just some pre-requisites to demo the interaction
	// between OpenCensus and the integration.
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	trace.RegisterExporter(pe)
	defer trace.UnregisterExporter(pe)

	client := redis.NewClient(&redis.Options{Addr: ":6379"})
	client.WrapProcess(ocredis.PerCommandTracer(context.Background()))
	_, err := client.HMSet("names", map[string]interface{}{
		"space": 1961, "tracing": "OpenCensus",
		"monitoring": "OpenCensus",
	}).Result()
	if err != nil {
		log.Fatalf("Failed to HMSet: %v", err)
	}
}

func Example_NewClient() {
	// Just some pre-requisites to demo the interaction
	// between OpenCensus and the integration.
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	trace.RegisterExporter(pe)
	defer trace.UnregisterExporter(pe)

	// Create and use the client
	client := ocredis.NewClient(context.Background(), ":6379")
	if _, err := client.HSet("programs", "space", 1961).Result(); err != nil {
		log.Fatalf("HSet error: %v", err)
	}
}

func Example_NewClient_comprehensive() {
	// Just some pre-requisites to demo the interaction
	// between OpenCensus and the integration.
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	trace.RegisterExporter(pe)
	defer trace.UnregisterExporter(pe)

	ctx := context.Background()
	// Now. Download the website
	req, _ := http.NewRequest("POST", "https://example.org/", nil)
	req = req.WithContext(ctx)
	res, err := (&http.Client{Transport: &ochttp.Transport{}}).Do(req)
	if err != nil {
		log.Fatalf("Failed to retrieve response: %v", err)
	}
	blob, err := ioutil.ReadAll(res.Body)
	_ = res.Body.Close()

	// Then cache if for later use
	client := ocredis.NewClient(ctx, ":6379")
	if _, err := client.HSet("cached", "example.org", blob).Result(); err != nil {
		log.Fatalf("Failed to cache website: %v", err)
	}

	// Later use, a client fetching the cached version of the website.
	client = ocredis.NewClient(ctx, ":6379")
	cached, err := client.HGet("cached", "example.org").Result()
	if err != nil {
		log.Fatalf("Failed to lookup from cache: %v", err)
	}
	log.Printf("%s", cached)
}
