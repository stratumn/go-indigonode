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

package contacts

import (
	"context"
	"sort"

	"github.com/pkg/errors"
	pb "github.com/stratumn/alice/grpc/contacts"

	peer "gx/ipfs/Qma7H6RW8wRrfZpNSXwxYGcd1E149s42FpWNpDNieSVrnU/go-libp2p-peer"
)

var (
	// ErrUnavailable is returned from gRPC methods when the service is not
	// available.
	ErrUnavailable = errors.New("the service is not available")
)

// grpcServer is a gRPC server for the contacs service.
type grpcServer struct {
	GetManager func() *Manager
}

// Lists streams all the contacts.
func (s grpcServer) List(req *pb.ListReq, ss pb.Contacts_ListServer) error {
	mgr := s.GetManager()
	if mgr == nil {
		return errors.WithStack(ErrUnavailable)
	}

	list := mgr.List()

	for _, name := range sortedNames(list) {
		contact := list[name]
		msg := &pb.Contact{Name: name, PeerId: []byte(contact.PeerID)}

		if err := ss.Send(msg); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

// Get finds a contact by name.
func (s grpcServer) Get(ctx context.Context, req *pb.GetReq) (*pb.Contact, error) {
	mgr := s.GetManager()
	if mgr == nil {
		return nil, errors.WithStack(ErrUnavailable)
	}

	contact, err := mgr.Get(req.Name)
	if err != nil {
		return nil, err
	}

	return &pb.Contact{Name: req.Name, PeerId: []byte(contact.PeerID)}, nil
}

// Set sets or adds a contact.
func (s grpcServer) Set(ctx context.Context, req *pb.SetReq) (*pb.Contact, error) {
	mgr := s.GetManager()
	if mgr == nil {
		return nil, errors.WithStack(ErrUnavailable)
	}

	pid, err := peer.IDFromBytes(req.PeerId)
	if err != nil {
		return nil, err
	}

	err = mgr.Set(req.Name, &Contact{PeerID: pid})
	if err != nil {
		return nil, err
	}

	return &pb.Contact{Name: req.Name, PeerId: req.PeerId}, nil
}

// Delete deletes a contact.
func (s grpcServer) Delete(ctx context.Context, req *pb.DeleteReq) (*pb.Contact, error) {
	mgr := s.GetManager()
	if mgr == nil {
		return nil, errors.WithStack(ErrUnavailable)
	}

	contact, err := mgr.Get(req.Name)
	if err != nil {
		return nil, err
	}

	if err := mgr.Delete(req.Name); err != nil {
		return nil, err
	}

	return &pb.Contact{Name: req.Name, PeerId: []byte(contact.PeerID)}, nil
}

// sortedNames returns the keys of a map of contacts sorted alphabetically.
func sortedNames(set map[string]Contact) []string {
	if set == nil {
		return nil
	}

	var keys []string
	for k := range set {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}
