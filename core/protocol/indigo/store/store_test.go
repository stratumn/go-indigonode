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

package store_test

import (
	"context"
	"crypto/rand"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/protocol/indigo/store"
	"github.com/stratumn/alice/core/protocol/indigo/store/mocknetwork"
	pb "github.com/stratumn/alice/pb/indigo/store"
	"github.com/stratumn/alice/test"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stratumn/go-indigocore/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ic "gx/ipfs/QmaPbCnUMBohSGo3KnxEa2bHqyJVVeEEcwtqJAYxerieBo/go-libp2p-crypto"
)

func createTestStore(ctrl *gomock.Controller) (*store.Store, *mocknetworkmanager.MockNetworkManager) {
	networkMgr := mocknetworkmanager.NewMockNetworkManager(ctrl)
	networkMgr.EXPECT().AddListener().Times(1)

	return store.New(networkMgr), networkMgr
}

func createTestStoreWithChan(ctrl *gomock.Controller, listenChan chan *pb.SignedLink) *store.Store {
	networkMgr := mocknetworkmanager.NewMockNetworkManager(ctrl)
	networkMgr.EXPECT().AddListener().Times(1).Return(listenChan)

	return store.New(networkMgr)
}

func TestNewStore(t *testing.T) {
	t.Run("add-network-listener", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		createTestStore(ctrl)
	})
}

func TestClose(t *testing.T) {
	t.Run("remove-listener", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		listenChan := make(<-chan *pb.SignedLink)

		networkMgr := mocknetworkmanager.NewMockNetworkManager(ctrl)
		networkMgr.EXPECT().AddListener().Times(1).Return(listenChan)

		s := store.New(networkMgr)

		networkMgr.EXPECT().RemoveListener(listenChan).Times(1)
		s.Close(context.Background())
	})
}

func TestReceiveLinks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sk, _, _ := ic.GenerateEd25519Key(rand.Reader)

	listenChan := make(chan *pb.SignedLink)
	testStore := createTestStoreWithChan(ctrl, listenChan)

	tests := []struct {
		name string
		run  func(*testing.T)
	}{{
		"malformed-link",
		func(t *testing.T) {
			listenChan <- &pb.SignedLink{Link: []byte("I'm not a valid JSON link")}

			// TODO: verify that the link is added to the store of shame.
		},
	}, {
		"invalid-link-signature",
		func(t *testing.T) {
			link := cstesting.RandomLink()
			linkHash, _ := link.Hash()
			signedLink, _ := pb.NewSignedLink(sk, link)
			signedLink.Signature.PublicKey = nil

			listenChan <- signedLink

			// TODO: verify that the link is added to the store of shame.

			seg, err := testStore.GetSegment(context.Background(), linkHash)
			assert.NoError(t, err, "testStore.GetSegment()")
			assert.Nil(t, seg, "testStore.GetSegment()")
		},
	}, {
		"invalid-link",
		func(t *testing.T) {
			link := cstesting.RandomLink()
			link.Meta.MapID = ""
			linkHash, _ := link.Hash()
			signedLink, _ := pb.NewSignedLink(sk, link)

			listenChan <- signedLink

			// TODO: verify that the link is added to the store of shame.

			seg, err := testStore.GetSegment(context.Background(), linkHash)
			assert.NoError(t, err, "testStore.GetSegment()")
			assert.Nil(t, seg, "testStore.GetSegment()")
		},
	}, {
		"valid-link",
		func(t *testing.T) {
			link := cstesting.RandomLink()
			linkHash, _ := link.Hash()
			signedLink, _ := pb.NewSignedLink(sk, link)

			listenChan <- signedLink

			// Network segments are handled in a separate goroutine,
			// so there is a race condition unless we wait.
			test.WaitUntil(
				t,
				100*time.Millisecond,
				10*time.Millisecond,
				func() error {
					seg, err := testStore.GetSegment(context.Background(), linkHash)
					require.NoError(t, err, "testStore.GetSegment()")

					if seg == nil {
						return errors.New("segment missing")
					}

					return nil
				},
				"link wasn't added to store",
			)

			seg, err := testStore.GetSegment(context.Background(), linkHash)
			require.NoError(t, err, "testStore.GetSegment()")
			assert.Equal(t, link, &seg.Link, "testStore.GetSegment()")
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestCreateLink(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T, *store.Store, *mocknetworkmanager.MockNetworkManager)
	}{{
		"invalid-link",
		func(t *testing.T, s *store.Store, mgr *mocknetworkmanager.MockNetworkManager) {
			// We don't add any assertions on the network manager.
			// In case of invalid link, it shouldn't be shared with the network.

			link := cstesting.RandomLink()
			link.Meta.MapID = ""

			_, err := s.CreateLink(context.Background(), link)
			assert.Error(t, err, "s.CreateLink()")
		},
	}, {
		"valid-link",
		func(t *testing.T, s *store.Store, mgr *mocknetworkmanager.MockNetworkManager) {
			link := cstesting.RandomLink()

			mgr.EXPECT().Publish(gomock.Any(), link).Times(1)

			lh, err := s.CreateLink(context.Background(), link)
			assert.NoError(t, err, "s.CreateLink()")
			assert.NotNil(t, lh, "s.CreateLink()")
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			s, networkMgr := createTestStore(ctrl)
			tt.run(t, s, networkMgr)
		})
	}
}

func TestGetSegment(t *testing.T) {
	linkHashNotFound, _ := types.NewBytes32FromString("4242424242424242424242424242424242424242424242424242424242424242")
	testLink := cstesting.RandomLink()
	testLinkHash, _ := testLink.Hash()

	tests := []struct {
		name    string
		prepare func(*testing.T, *store.Store, *mocknetworkmanager.MockNetworkManager)
		run     func(*testing.T, *store.Store)
	}{{
		"missing-link",
		func(*testing.T, *store.Store, *mocknetworkmanager.MockNetworkManager) {},
		func(t *testing.T, s *store.Store) {
			seg, err := s.GetSegment(context.Background(), linkHashNotFound)
			assert.NoError(t, err, "s.GetSegment()")
			assert.Nil(t, seg, "s.GetSegment()")
		},
	}, {
		"valid-link",
		func(t *testing.T, s *store.Store, n *mocknetworkmanager.MockNetworkManager) {
			n.EXPECT().Publish(gomock.Any(), testLink).Times(1)

			_, err := s.CreateLink(context.Background(), testLink)
			assert.NoError(t, err, "s.CreateLink()")
		},
		func(t *testing.T, s *store.Store) {
			seg, err := s.GetSegment(context.Background(), testLinkHash)
			assert.NoError(t, err, "s.GetSegment()")
			assert.NotNil(t, seg, "s.GetSegment()")

			assert.Equal(t, testLink, &seg.Link, "seg.Link")
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			s, networkMgr := createTestStore(ctrl)
			tt.prepare(t, s, networkMgr)
			tt.run(t, s)
		})
	}
}
