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

package manager

// ServiceGroup represents a group of services.
//
// It is implemented as a service.
type ServiceGroup struct {
	// GroupID is a unique identifier that will be used as the service ID.
	GroupID string

	// GroupName is a human friendly name for the group.
	GroupName string

	// GroupDesc is a description of what the group represents.
	GroupDesc string

	// Services are the services that this group starts.
	Services map[string]struct{}
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
