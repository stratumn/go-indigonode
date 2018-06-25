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

// Avoid cyclic dependencies.
package manager_test

import (
	"testing"

	"github.com/stratumn/go-indigonode/core/manager"
	"github.com/stratumn/go-indigonode/core/manager/testservice"
)

func TestServiceGroup_strings(t *testing.T) {
	testservice.CheckStrings(t, &manager.ServiceGroup{
		GroupID:   "mygroup",
		GroupName: "My Group",
		GroupDesc: "A test group.",
	})
}
