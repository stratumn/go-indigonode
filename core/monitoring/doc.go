// Copyright © 2017-2018 Stratumn SAS
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

// Package monitoring contains thin wrappers around the monitoring libraries
// used.
//
// Its goal is not to create a completely new metrics/trace abstraction but
// rather to minimize places where the underlying monitoring libraries are
// directly used to make it easy to replace them if needed with the help
// of the compiler.
// Packages that expose metrics should define those metrics in a metrics.go
// file which would be the only file in the package to import OpenCensus.
// The rest of the package would use our monitoring wrappers to record metrics.
package monitoring
