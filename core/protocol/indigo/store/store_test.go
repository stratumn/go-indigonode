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
	"github.com/stratumn/alice/core/protocol/indigo/store/audit/mockaudit"
	"github.com/stratumn/alice/core/protocol/indigo/store/mocknetwork"
	"github.com/stratumn/alice/pb/crypto"
	pb "github.com/stratumn/alice/pb/indigo/store"
	"github.com/stratumn/alice/test"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stratumn/go-indigocore/dummystore"
	indigostore "github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/store/storetestcases"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	peer "gx/ipfs/QmZoWKhxUmZ2seW4BzX6fJkNR8hh9PsGModr7q171yq2SS/go-libp2p-peer"
	ic "gx/ipfs/QmaPbCnUMBohSGo3KnxEa2bHqyJVVeEEcwtqJAYxerieBo/go-libp2p-crypto"
)

func createTestStore(ctrl *gomock.Controller) (*store.Store, *mocknetworkmanager.MockNetworkManager) {
	networkMgr := mocknetworkmanager.NewMockNetworkManager(ctrl)
	networkMgr.EXPECT().AddListener().Times(1)

	auditStore := mockaudit.NewMockStore(ctrl)
	auditStore.EXPECT().AddLink(gomock.Any(), gomock.Any()).AnyTimes()

	return store.New(networkMgr, dummystore.New(&dummystore.Config{}), auditStore), networkMgr
}

func createTestStoreWithChan(ctrl *gomock.Controller, listenChan chan *pb.SignedLink) (*store.Store, *mockaudit.MockStore) {
	networkMgr := mocknetworkmanager.NewMockNetworkManager(ctrl)
	networkMgr.EXPECT().AddListener().Times(1).Return(listenChan)

	auditStore := mockaudit.NewMockStore(ctrl)

	return store.New(networkMgr, dummystore.New(&dummystore.Config{}), auditStore), auditStore
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

		s := store.New(networkMgr, nil, nil)

		networkMgr.EXPECT().RemoveListener(listenChan).Times(1)
		s.Close(context.Background())
	})
}

func TestReceiveLinks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sk, _, _ := ic.GenerateEd25519Key(rand.Reader)
	peerID, _ := peer.IDFromPrivateKey(sk)

	listenChan := make(chan *pb.SignedLink)
	testStore, auditStore := createTestStoreWithChan(ctrl, listenChan)

	tests := []struct {
		name string
		run  func(*testing.T)
	}{{
		"malformed-link",
		func(t *testing.T) {
			malformed := &pb.SignedLink{Link: []byte("I'm not a valid JSON link")}
			malformed.From = []byte(peerID)
			malformed.Signature, _ = crypto.Sign(sk, malformed.Link)

			auditStore.EXPECT().AddLink(gomock.Any(), malformed)
			listenChan <- malformed
		},
	}, {
		"invalid-link-signature",
		func(t *testing.T) {
			link := cstesting.RandomLink()
			linkHash, _ := link.Hash()
			signedLink, _ := pb.NewSignedLink(sk, link)
			signedLink.Signature.PublicKey = nil

			listenChan <- signedLink

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

			auditStore.EXPECT().AddLink(gomock.Any(), signedLink)
			listenChan <- signedLink

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

func TestIndigoStoreImpl(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storetestcases.Factory{
		New: func() (indigostore.Adapter, error) {
			store, networkMgr := createTestStore(ctrl)
			networkMgr.EXPECT().Publish(gomock.Any(), gomock.Any()).AnyTimes()

			return store, nil
		},
	}.RunStoreTests(t)
}
