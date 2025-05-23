// Copyright 2017 Google Inc.
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

//go:build !prod
// +build !prod

// When included in the main build process, this file PERMANENTLY caches all
// HTTP requests. This is useful for quickly prototyping customisations to the
// bar without incurring HTTP costs or consuming quota on remote services.
// The cache is stored in ~/.cache/barista/http (using XDG_CACHE_HOME if set),
// and individual responses can be deleted if a fresher copy is needed.
// Once you are satisfied with the bar, simply omit this file (or build with
// the "prod" tag) to build a production version of the bar with no caching.

package main

import (
	"net/http"

	"github.com/leosunmo/barista/testing/httpcache"
)

func init() {
	http.DefaultTransport = httpcache.Wrap(http.DefaultTransport)
}
