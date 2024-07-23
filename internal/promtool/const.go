// Lenstra - This file contains methods used by the promtool package. - https://github.com/prometheus/prometheus/blob/5a6c8f9c152dfab5f96f5b6f14703b801b014255/cmd/promtool/main.go

// Copyright 2015 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package promtool

const (
	successExitCode = 0
	failureExitCode = 1

	lintOptionAll            = "all"
	lintOptionDuplicateRules = "duplicate-rules"
	lintOptionNone           = "none"
)
