// Copyright © 2017-2018 Stratumn SAS
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

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
