// Copyright © 2017 Stratumn SAS
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

package manager

// ServiceGroup represents a group of services. It is implemented as a service.
type ServiceGroup struct {
	GroupID   string
	GroupName string
	GroupDesc string
	Services  map[string]struct{}
}

// ID returns the unique identifier of the service.
func (s *ServiceGroup) ID() string {
	return s.GroupID
}

// Name returns the human friendly name of the service.
func (s *ServiceGroup) Name() string {
	return s.GroupName
}

// Desc returns a description of what the service does.
func (s *ServiceGroup) Desc() string {
	return s.GroupDesc
}

// Needs returns the set of services this service depends on.
func (s *ServiceGroup) Needs() map[string]struct{} {
	return s.Services
}
