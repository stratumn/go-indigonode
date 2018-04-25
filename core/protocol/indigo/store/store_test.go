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
	"github.com/stratumn/alice/core/protocol/indigo/store/mockstore"
	"github.com/stratumn/alice/core/protocol/indigo/store/sync/mocksync"
	"github.com/stratumn/alice/pb/crypto"
	pb "github.com/stratumn/alice/pb/indigo/store"
	"github.com/stratumn/alice/test"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stratumn/go-indigocore/dummystore"
	indigostore "github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/store/storetestcases"
	"github.com/stratumn/go-indigocore/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	peer "gx/ipfs/QmZoWKhxUmZ2seW4BzX6fJkNR8hh9PsGModr7q171yq2SS/go-libp2p-peer"
	ic "gx/ipfs/QmaPbCnUMBohSGo3KnxEa2bHqyJVVeEEcwtqJAYxerieBo/go-libp2p-crypto"
)

type TestStoreBuilder struct {
	ctrl        *gomock.Controller
	networkMgr  *mocknetworkmanager.MockNetworkManager
	syncEngine  *mocksync.MockEngine
	indigoStore indigostore.Adapter
	auditStore  *mockaudit.MockStore
}

func NewTestStoreBuilder(ctrl *gomock.Controller) *TestStoreBuilder {
	return &TestStoreBuilder{ctrl: ctrl}
}

func (b *TestStoreBuilder) WithNetworkManager(networkMgr *mocknetworkmanager.MockNetworkManager) *TestStoreBuilder {
	b.networkMgr = networkMgr
	return b
}

func (b *TestStoreBuilder) WithSyncEngine(syncEngine *mocksync.MockEngine) *TestStoreBuilder {
	b.syncEngine = syncEngine
	return b
}

func (b *TestStoreBuilder) WithAuditStore(auditStore *mockaudit.MockStore) *TestStoreBuilder {
	b.auditStore = auditStore
	return b
}

func (b *TestStoreBuilder) WithIndigoStore(indigoStore indigostore.Adapter) *TestStoreBuilder {
	b.indigoStore = indigoStore
	return b
}

func (b *TestStoreBuilder) Build() *store.Store {
	if b.networkMgr == nil {
		b.networkMgr = mocknetworkmanager.NewMockNetworkManager(b.ctrl)
		b.networkMgr.EXPECT().AddListener().Times(1)
		b.networkMgr.EXPECT().Publish(gomock.Any(), gomock.Any()).AnyTimes()
	}

	if b.syncEngine == nil {
		b.syncEngine = mocksync.NewMockEngine(b.ctrl)
		// No missing links to sync.
		b.syncEngine.EXPECT().
			GetMissingLinks(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(nil, nil)
	}

	if b.auditStore == nil {
		b.auditStore = mockaudit.NewMockStore(b.ctrl)
		b.auditStore.EXPECT().AddLink(gomock.Any(), gomock.Any()).AnyTimes()
	}

	if b.indigoStore == nil {
		b.indigoStore = dummystore.New(&dummystore.Config{})
	}

	return store.New(
		b.networkMgr,
		b.syncEngine,
		b.indigoStore,
		b.auditStore,
	)
}

func TestNewStore(t *testing.T) {
	t.Run("add-network-listener", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		networkMgr := mocknetworkmanager.NewMockNetworkManager(ctrl)
		networkMgr.EXPECT().AddListener().Times(1)

		NewTestStoreBuilder(ctrl).WithNetworkManager(networkMgr).Build()
	})
}

func TestClose(t *testing.T) {
	t.Run("remove-listener", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		listenChan := make(<-chan *pb.SignedLink)

		networkMgr := mocknetworkmanager.NewMockNetworkManager(ctrl)
		networkMgr.EXPECT().AddListener().Times(1).Return(listenChan)

		s := NewTestStoreBuilder(ctrl).WithNetworkManager(networkMgr).Build()

		networkMgr.EXPECT().RemoveListener(listenChan).Times(1)
		s.Close(context.Background())
	})
}

func TestReceiveLinks(t *testing.T) {
	sk, _, _ := ic.GenerateEd25519Key(rand.Reader)
	peerID, _ := peer.IDFromPrivateKey(sk)

	createTestLink := func() (*pb.SignedLink, *cs.Link, *types.Bytes32) {
		link := cstesting.NewLinkBuilder().Build()
		linkHash, _ := link.Hash()
		signedLink, _ := pb.NewSignedLink(sk, link)
		return signedLink, link, linkHash
	}

	createNetworkMgrWithChan := func(ctrl *gomock.Controller) (chan *pb.SignedLink, *mocknetworkmanager.MockNetworkManager) {
		listenChan := make(chan *pb.SignedLink)
		networkMgr := mocknetworkmanager.NewMockNetworkManager(ctrl)
		networkMgr.EXPECT().AddListener().Times(1).Return(listenChan)
		return listenChan, networkMgr
	}

	tests := []struct {
		name string
		run  func(*testing.T)
	}{{
		"malformed-link",
		func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			malformed := &pb.SignedLink{Link: []byte("I'm not a valid JSON link")}
			malformed.From = []byte(peerID)
			malformed.Signature, _ = crypto.Sign(sk, malformed.Link)

			auditChan := make(chan struct{})
			auditStore := mockaudit.NewMockStore(ctrl)
			auditStore.EXPECT().AddLink(gomock.Any(), malformed).Times(1).
				Do(func(context.Context, *pb.SignedLink) { auditChan <- struct{}{} })

			listenChan, networkMgr := createNetworkMgrWithChan(ctrl)
			NewTestStoreBuilder(ctrl).
				WithAuditStore(auditStore).
				WithNetworkManager(networkMgr).
				Build()

			listenChan <- malformed
			<-auditChan
		},
	}, {
		"invalid-link-signature",
		func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			signedLink, _, linkHash := createTestLink()
			signedLink.Signature.PublicKey = nil

			listenChan, networkMgr := createNetworkMgrWithChan(ctrl)
			testStore := NewTestStoreBuilder(ctrl).
				WithNetworkManager(networkMgr).
				Build()

			listenChan <- signedLink

			verifySegmentNotStored(t, testStore, linkHash)
		},
	}, {
		"invalid-link",
		func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			link := cstesting.NewLinkBuilder().Invalid().Build()
			linkHash, _ := link.Hash()
			signedLink, _ := pb.NewSignedLink(sk, link)

			auditChan := make(chan struct{})
			auditStore := mockaudit.NewMockStore(ctrl)
			auditStore.EXPECT().AddLink(gomock.Any(), signedLink).Times(1).
				Do(func(context.Context, *pb.SignedLink) { auditChan <- struct{}{} })

			listenChan, networkMgr := createNetworkMgrWithChan(ctrl)
			testStore := NewTestStoreBuilder(ctrl).
				WithAuditStore(auditStore).
				WithNetworkManager(networkMgr).
				Build()

			listenChan <- signedLink
			<-auditChan

			verifySegmentNotStored(t, testStore, linkHash)
		},
	}, {
		"already-saved-link",
		func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			signedLink, link, linkHash := createTestLink()
			listenChan, networkMgr := createNetworkMgrWithChan(ctrl)
			testStore := NewTestStoreBuilder(ctrl).
				WithNetworkManager(networkMgr).
				Build()

			listenChan <- signedLink

			waitUntilSegmentStored(t, testStore, linkHash)

			listenChan <- signedLink

			<-time.After(20 * time.Millisecond)

			seg, err := testStore.GetSegment(context.Background(), linkHash)
			require.NoError(t, err, "testStore.GetSegment()")
			require.NotNil(t, seg, "testStore.GetSegment()")
			assert.Equal(t, link, &seg.Link, "testStore.GetSegment()")
		},
	}, {
		"valid-link",
		func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			signedLink, link, linkHash := createTestLink()
			listenChan, networkMgr := createNetworkMgrWithChan(ctrl)
			testStore := NewTestStoreBuilder(ctrl).
				WithNetworkManager(networkMgr).
				Build()

			listenChan <- signedLink

			// Network segments are handled in a separate goroutine,
			// so there is a race condition unless we wait.
			waitUntilSegmentStored(t, testStore, linkHash)

			seg, err := testStore.GetSegment(context.Background(), linkHash)
			require.NoError(t, err, "testStore.GetSegment()")
			require.NotNil(t, seg, "testStore.GetSegment()")
			assert.Equal(t, link, &seg.Link, "testStore.GetSegment()")
		},
	}, {
		"valid-link-sync-success",
		func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// link1 <----- link2 <----- link3
			link1 := cstesting.NewLinkBuilder().Build()
			linkHash1, _ := link1.Hash()
			link2 := cstesting.NewLinkBuilder().WithParent(link1).Build()
			linkHash2, _ := link2.Hash()
			link3 := cstesting.NewLinkBuilder().WithParent(link2).Build()
			linkHash3, _ := link3.Hash()
			signedLink3, _ := pb.NewSignedLink(sk, link3)

			syncEngine := mocksync.NewMockEngine(ctrl)
			syncEngine.EXPECT().
				GetMissingLinks(gomock.Any(), peerID, link3, gomock.Any()).
				Times(1).
				Return([]*cs.Link{link1, link2}, nil)

			listenChan, networkMgr := createNetworkMgrWithChan(ctrl)
			testStore := NewTestStoreBuilder(ctrl).
				WithNetworkManager(networkMgr).
				WithSyncEngine(syncEngine).
				Build()

			listenChan <- signedLink3

			waitUntilSegmentStored(t, testStore, linkHash1)
			waitUntilSegmentStored(t, testStore, linkHash2)
			waitUntilSegmentStored(t, testStore, linkHash3)
		},
	}, {
		"valid-link-sync-error",
		func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			signedLink, link, linkHash := createTestLink()

			syncEngine := mocksync.NewMockEngine(ctrl)
			syncEngine.EXPECT().
				GetMissingLinks(gomock.Any(), peerID, link, gomock.Any()).
				Times(1).
				Return(nil, errors.New("no more credits"))

			indigoStore := mockstore.NewMockAdapter(ctrl)
			indigoStore.EXPECT().GetSegment(gomock.Any(), linkHash).Times(1)

			listenChan, networkMgr := createNetworkMgrWithChan(ctrl)
			NewTestStoreBuilder(ctrl).
				WithNetworkManager(networkMgr).
				WithSyncEngine(syncEngine).
				WithIndigoStore(indigoStore).
				Build()

			listenChan <- signedLink
			<-time.After(20 * time.Millisecond)
			// The test will fail if a call to mockstore.CreateLink is made.
		},
	}, {
		"valid-link-synced-invalid-link",
		func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			invalidLink := cstesting.NewLinkBuilder().Invalid().Build()
			invalidLinkHash, _ := invalidLink.Hash()

			link := cstesting.NewLinkBuilder().WithParent(invalidLink).Build()
			linkHash, _ := link.Hash()
			signedLink, _ := pb.NewSignedLink(sk, link)

			auditChan := make(chan struct{})
			auditStore := mockaudit.NewMockStore(ctrl)
			auditStore.EXPECT().AddLink(gomock.Any(), signedLink).Times(1).
				Do(func(context.Context, *pb.SignedLink) { auditChan <- struct{}{} })

			syncEngine := mocksync.NewMockEngine(ctrl)
			syncEngine.EXPECT().
				GetMissingLinks(gomock.Any(), peerID, link, gomock.Any()).
				Times(1).
				Return([]*cs.Link{invalidLink}, nil)

			listenChan, networkMgr := createNetworkMgrWithChan(ctrl)
			testStore := NewTestStoreBuilder(ctrl).
				WithNetworkManager(networkMgr).
				WithSyncEngine(syncEngine).
				WithAuditStore(auditStore).
				Build()

			listenChan <- signedLink
			<-auditChan

			verifySegmentNotStored(t, testStore, linkHash)
			verifySegmentNotStored(t, testStore, invalidLinkHash)
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func waitUntilSegmentStored(t *testing.T, testStore *store.Store, linkHash *types.Bytes32) {
	test.WaitUntil(
		t,
		25*time.Millisecond,
		5*time.Millisecond,
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
}

func verifySegmentNotStored(t *testing.T, testStore *store.Store, linkHash *types.Bytes32) {
	seg, err := testStore.GetSegment(context.Background(), linkHash)
	require.NoError(t, err, "testStore.GetSegment()")
	require.Nil(t, seg, "testStore.GetSegment()")
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

			link := cstesting.NewLinkBuilder().Invalid().Build()
			_, err := s.CreateLink(context.Background(), link)
			assert.Error(t, err, "s.CreateLink()")
		},
	}, {
		"valid-link",
		func(t *testing.T, s *store.Store, mgr *mocknetworkmanager.MockNetworkManager) {
			link := cstesting.NewLinkBuilder().Build()

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

			networkMgr := mocknetworkmanager.NewMockNetworkManager(ctrl)
			networkMgr.EXPECT().AddListener().Times(1)

			s := NewTestStoreBuilder(ctrl).WithNetworkManager(networkMgr).Build()
			tt.run(t, s, networkMgr)
		})
	}
}

func TestIndigoStoreImpl(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storetestcases.Factory{
		New: func() (indigostore.Adapter, error) {
			store := NewTestStoreBuilder(ctrl).Build()
			return store, nil
		},
	}.RunStoreTests(t)
}
