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
	"github.com/stratumn/alice/core/protocol/indigo/store/audit"
	"github.com/stratumn/alice/core/protocol/indigo/store/audit/mockaudit"
	"github.com/stratumn/alice/core/protocol/indigo/store/constants"
	"github.com/stratumn/alice/core/protocol/indigo/store/mocknetwork"
	"github.com/stratumn/alice/core/protocol/indigo/store/mockstore"
	"github.com/stratumn/alice/core/protocol/indigo/store/mockvalidator"
	"github.com/stratumn/alice/core/protocol/indigo/store/sync/mocksync"
	"github.com/stratumn/alice/test"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stratumn/go-indigocore/dummystore"
	indigostore "github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/store/storetestcases"
	"github.com/stratumn/go-indigocore/types"
	"github.com/stratumn/go-indigocore/validator"
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
	govMgr      validator.GovernanceManager
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

func (b *TestStoreBuilder) WithGovernanceManager(govMgr validator.GovernanceManager) *TestStoreBuilder {
	b.govMgr = govMgr
	return b
}

func (b *TestStoreBuilder) Build() *store.Store {
	if b.networkMgr == nil {
		b.networkMgr = mocknetworkmanager.NewMockNetworkManager(b.ctrl)
		b.networkMgr.EXPECT().AddListener().Times(1)
		b.networkMgr.EXPECT().Publish(gomock.Any(), gomock.Any()).AnyTimes()
		b.networkMgr.EXPECT().NodeID().AnyTimes()
	}

	if b.syncEngine == nil {
		b.syncEngine = mocksync.NewMockEngine(b.ctrl)
		// No missing links to sync.
		b.syncEngine.EXPECT().
			GetMissingLinks(gomock.Any(), gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(nil, nil)
	}

	if b.auditStore == nil {
		b.auditStore = mockaudit.NewMockStore(b.ctrl)
		b.auditStore.EXPECT().AddSegment(gomock.Any(), gomock.Any()).AnyTimes()
	}

	if b.indigoStore == nil {
		b.indigoStore = dummystore.New(&dummystore.Config{})
	}

	if b.govMgr == nil {
		govMgr := mockvalidator.NewMockGovernanceManager(b.ctrl)
		govMgr.EXPECT().ListenAndUpdate(gomock.Any()).Times(1)
		b.govMgr = govMgr
	}

	return store.New(
		context.Background(),
		b.networkMgr,
		b.syncEngine,
		b.indigoStore,
		b.auditStore,
		b.govMgr,
	)
}

func TestNewStore(t *testing.T) {
	t.Run("add-network-and-governance-listener", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		networkMgr := mocknetworkmanager.NewMockNetworkManager(ctrl)
		networkMgr.EXPECT().AddListener().Times(1)

		governanceManager := mockvalidator.NewMockGovernanceManager(ctrl)
		governanceManager.EXPECT().ListenAndUpdate(gomock.Any()).Times(1)

		NewTestStoreBuilder(ctrl).WithNetworkManager(networkMgr).WithGovernanceManager(governanceManager).Build()
		<-time.After(10 * time.Millisecond)
	})
}

func TestClose(t *testing.T) {
	t.Run("remove-listener", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		listenChan := make(<-chan *cs.Segment)

		networkMgr := mocknetworkmanager.NewMockNetworkManager(ctrl)
		networkMgr.EXPECT().AddListener().Times(1).Return(listenChan)

		s := NewTestStoreBuilder(ctrl).WithNetworkManager(networkMgr).Build()
		<-time.After(10 * time.Millisecond)

		networkMgr.EXPECT().RemoveListener(listenChan).Times(1)
		s.Close(context.Background())
	})
}

func TestReceiveLinks(t *testing.T) {
	sk, _, _ := ic.GenerateEd25519Key(rand.Reader)
	peerID, _ := peer.IDFromPrivateKey(sk)

	createSignedSegment := func() *cs.Segment {
		link := cstesting.NewLinkBuilder().Build()
		constants.SetLinkNodeID(link, peerID)
		segment, err := audit.SignLink(context.Background(), sk, link)
		require.NoError(t, err, "audit.SignLink()")
		return segment
	}

	createNetworkMgrWithChan := func(ctrl *gomock.Controller) (chan *cs.Segment, *mocknetworkmanager.MockNetworkManager) {
		listenChan := make(chan *cs.Segment)
		networkMgr := mocknetworkmanager.NewMockNetworkManager(ctrl)
		networkMgr.EXPECT().AddListener().Times(1).Return(listenChan)
		return listenChan, networkMgr
	}

	createGovernanceMgr := func(ctrl *gomock.Controller) *mockvalidator.MockGovernanceManager {
		governanceManager := mockvalidator.NewMockGovernanceManager(ctrl)
		governanceManager.EXPECT().ListenAndUpdate(gomock.Any()).Times(1)
		return governanceManager
	}

	tests := []struct {
		name string
		run  func(*testing.T)
	}{{
		"empty-segment",
		func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			listenChan, networkMgr := createNetworkMgrWithChan(ctrl)
			NewTestStoreBuilder(ctrl).
				WithNetworkManager(networkMgr).
				Build()

			listenChan <- nil
		},
	}, {
		"invalid-link-signature",
		func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			signedSegment := createSignedSegment()
			signedSegment.Meta.Evidences[0].Proof.(*audit.PeerSignature).Signature.PublicKey = nil

			listenChan, networkMgr := createNetworkMgrWithChan(ctrl)
			testStore := NewTestStoreBuilder(ctrl).
				WithNetworkManager(networkMgr).
				Build()

			listenChan <- signedSegment

			verifySegmentNotStored(t, testStore, signedSegment.GetLinkHash())
		},
	}, {
		"invalid-link",
		func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			link := cstesting.NewLinkBuilder().
				Invalid().
				WithMetadata(constants.NodeIDKey, peerID.Pretty()).
				Build()
			signedSegment, _ := audit.SignLink(context.Background(), sk, link)

			auditChan := make(chan struct{})
			auditStore := mockaudit.NewMockStore(ctrl)
			auditStore.EXPECT().AddSegment(gomock.Any(), signedSegment).Times(1).
				Do(func(context.Context, *cs.Segment) { auditChan <- struct{}{} })

			listenChan, networkMgr := createNetworkMgrWithChan(ctrl)
			testStore := NewTestStoreBuilder(ctrl).
				WithAuditStore(auditStore).
				WithNetworkManager(networkMgr).
				Build()

			listenChan <- signedSegment
			<-auditChan

			verifySegmentNotStored(t, testStore, signedSegment.GetLinkHash())
		},
	}, {
		"invalid-evidence-provider",
		func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			link := cstesting.NewLinkBuilder().
				WithMetadata(constants.NodeIDKey, peerID.Pretty()).
				Build()
			signedSegment, err := audit.SignLink(context.Background(), sk, link)
			require.NoError(t, err, "audit.SignLink()")

			signedSegment.Meta.Evidences[0].Provider = "spongebob"

			auditChan := make(chan struct{})
			auditStore := mockaudit.NewMockStore(ctrl)
			auditStore.EXPECT().AddSegment(gomock.Any(), signedSegment).Times(1).
				Do(func(context.Context, *cs.Segment) { auditChan <- struct{}{} })

			listenChan, networkMgr := createNetworkMgrWithChan(ctrl)
			testStore := NewTestStoreBuilder(ctrl).
				WithAuditStore(auditStore).
				WithNetworkManager(networkMgr).
				Build()

			listenChan <- signedSegment
			<-auditChan

			verifySegmentNotStored(t, testStore, signedSegment.GetLinkHash())
		},
	}, {
		"invalid-meta-node-id",
		func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			signedSegment := cstesting.NewLinkBuilder().
				WithMetadata(constants.NodeIDKey, "segment").
				Build().
				Segmentify()

			proof, _ := audit.NewPeerSignature(sk, signedSegment)
			signedSegment.Meta.AddEvidence(cs.Evidence{
				Backend:  audit.PeerSignatureBackend,
				Provider: peerID.Pretty(),
				Proof:    proof,
			})

			auditChan := make(chan struct{})
			auditStore := mockaudit.NewMockStore(ctrl)
			auditStore.EXPECT().AddSegment(gomock.Any(), signedSegment).Times(1).
				Do(func(context.Context, *cs.Segment) { auditChan <- struct{}{} })

			listenChan, networkMgr := createNetworkMgrWithChan(ctrl)
			testStore := NewTestStoreBuilder(ctrl).
				WithAuditStore(auditStore).
				WithNetworkManager(networkMgr).
				Build()

			listenChan <- signedSegment
			<-auditChan

			verifySegmentNotStored(t, testStore, signedSegment.GetLinkHash())
		},
	}, {
		"already-saved-link",
		func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			govMgr := createGovernanceMgr(ctrl)
			govMgr.EXPECT().Current().Times(1).Return(nil)

			signedSegment := createSignedSegment()
			listenChan, networkMgr := createNetworkMgrWithChan(ctrl)
			testStore := NewTestStoreBuilder(ctrl).
				WithNetworkManager(networkMgr).
				WithGovernanceManager(govMgr).
				Build()

			listenChan <- signedSegment

			waitUntilSegmentStored(t, testStore, signedSegment.GetLinkHash())

			listenChan <- signedSegment

			<-time.After(20 * time.Millisecond)

			seg, err := testStore.GetSegment(context.Background(), signedSegment.GetLinkHash())
			require.NoError(t, err, "testStore.GetSegment()")
			require.NotNil(t, seg, "testStore.GetSegment()")
			assert.Equal(t, signedSegment.Link, seg.Link, "testStore.GetSegment()")
		},
	}, {
		"valid-link",
		func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			govMgr := createGovernanceMgr(ctrl)
			govMgr.EXPECT().Current().Times(1).Return(nil)

			signedSegment := createSignedSegment()
			listenChan, networkMgr := createNetworkMgrWithChan(ctrl)
			testStore := NewTestStoreBuilder(ctrl).
				WithNetworkManager(networkMgr).
				WithGovernanceManager(govMgr).
				Build()

			listenChan <- signedSegment

			// Network segments are handled in a separate goroutine,
			// so there is a race condition unless we wait.
			waitUntilSegmentStored(t, testStore, signedSegment.GetLinkHash())

			seg, err := testStore.GetSegment(context.Background(), signedSegment.GetLinkHash())
			require.NoError(t, err, "testStore.GetSegment()")
			require.NotNil(t, seg, "testStore.GetSegment()")
			assert.Equal(t, signedSegment.Link, seg.Link, "testStore.GetSegment()")
		},
	}, {
		"valid-link-sync-success",
		func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// link1 <----- link2 <----- link3
			link1 := cstesting.NewLinkBuilder().
				WithMetadata(constants.NodeIDKey, peerID.Pretty()).
				Build()
			linkHash1, _ := link1.Hash()
			link2 := cstesting.NewLinkBuilder().
				WithParent(link1).
				WithMetadata(constants.NodeIDKey, peerID.Pretty()).
				Build()
			linkHash2, _ := link2.Hash()
			link3 := cstesting.NewLinkBuilder().
				WithParent(link2).
				WithMetadata(constants.NodeIDKey, peerID.Pretty()).
				Build()
			linkHash3, _ := link3.Hash()
			signedSegment3, _ := audit.SignLink(context.Background(), sk, link3)

			syncEngine := mocksync.NewMockEngine(ctrl)
			syncEngine.EXPECT().
				GetMissingLinks(gomock.Any(), link3, gomock.Any()).
				Times(1).
				Return([]*cs.Link{link1, link2}, nil)

			govMgr := createGovernanceMgr(ctrl)
			govMgr.EXPECT().Current().Times(1).Return(nil)

			listenChan, networkMgr := createNetworkMgrWithChan(ctrl)
			testStore := NewTestStoreBuilder(ctrl).
				WithNetworkManager(networkMgr).
				WithGovernanceManager(govMgr).
				WithSyncEngine(syncEngine).
				Build()

			listenChan <- signedSegment3

			waitUntilSegmentStored(t, testStore, linkHash1)
			waitUntilSegmentStored(t, testStore, linkHash2)
			waitUntilSegmentStored(t, testStore, linkHash3)
		},
	}, {
		"valid-link-sync-error",
		func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			signedSegment := createSignedSegment()

			syncEngine := mocksync.NewMockEngine(ctrl)
			syncEngine.EXPECT().
				GetMissingLinks(gomock.Any(), &signedSegment.Link, gomock.Any()).
				Times(1).
				Return(nil, errors.New("no more credits"))

			indigoStore := mockstore.NewMockAdapter(ctrl)
			indigoStore.EXPECT().GetSegment(gomock.Any(), signedSegment.GetLinkHash()).Times(1)

			listenChan, networkMgr := createNetworkMgrWithChan(ctrl)
			NewTestStoreBuilder(ctrl).
				WithNetworkManager(networkMgr).
				WithSyncEngine(syncEngine).
				WithIndigoStore(indigoStore).
				Build()

			listenChan <- signedSegment
			<-time.After(20 * time.Millisecond)
			// The test will fail if a call to mockstore.CreateLink is made.
		},
	}, {
		"valid-link-synced-invalid-link",
		func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			invalidLink := cstesting.NewLinkBuilder().
				Invalid().
				WithMetadata(constants.NodeIDKey, peerID.Pretty()).
				Build()
			invalidLinkHash, _ := invalidLink.Hash()

			link := cstesting.NewLinkBuilder().
				WithParent(invalidLink).
				WithMetadata(constants.NodeIDKey, peerID.Pretty()).
				Build()
			linkHash, _ := link.Hash()
			signedSegment, _ := audit.SignLink(context.Background(), sk, link)

			auditChan := make(chan struct{})
			auditStore := mockaudit.NewMockStore(ctrl)
			auditStore.EXPECT().AddSegment(gomock.Any(), signedSegment).Times(1).
				Do(func(context.Context, *cs.Segment) { auditChan <- struct{}{} })

			syncEngine := mocksync.NewMockEngine(ctrl)
			syncEngine.EXPECT().
				GetMissingLinks(gomock.Any(), link, gomock.Any()).
				Times(1).
				Return([]*cs.Link{invalidLink}, nil)

			listenChan, networkMgr := createNetworkMgrWithChan(ctrl)
			testStore := NewTestStoreBuilder(ctrl).
				WithNetworkManager(networkMgr).
				WithSyncEngine(syncEngine).
				WithAuditStore(auditStore).
				Build()

			listenChan <- signedSegment
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
		run  func(*testing.T, *store.Store, *mocknetworkmanager.MockNetworkManager, *mockvalidator.MockGovernanceManager)
	}{{
		"invalid-link",
		func(t *testing.T, s *store.Store, networkMgr *mocknetworkmanager.MockNetworkManager, govMgr *mockvalidator.MockGovernanceManager) {
			// We don't add any assertions on the network manager.
			// In case of invalid link, it shouldn't be shared with the network.

			link := cstesting.NewLinkBuilder().Invalid().Build()
			_, err := s.CreateLink(context.Background(), link)
			assert.Error(t, err, "s.CreateLink()")
		},
	}, {
		"publish-valid-link",
		func(t *testing.T, s *store.Store, networkMgr *mocknetworkmanager.MockNetworkManager, govMgr *mockvalidator.MockGovernanceManager) {
			link := cstesting.NewLinkBuilder().Build()

			networkMgr.EXPECT().Publish(gomock.Any(), link).Times(1)
			networkMgr.EXPECT().NodeID().Times(1)

			govMgr.EXPECT().Current().Times(1).Return(nil)

			lh, err := s.CreateLink(context.Background(), link)
			assert.NoError(t, err, "s.CreateLink()")
			assert.NotNil(t, lh, "s.CreateLink()")
		},
	}, {
		"add-node-id-meta",
		func(t *testing.T, s *store.Store, networkMgr *mocknetworkmanager.MockNetworkManager, govMgr *mockvalidator.MockGovernanceManager) {
			link := cstesting.NewLinkBuilder().Build()
			link.Meta.Data = nil

			peerID, _ := peer.IDB58Decode("QmeZjNhdKPNNEtCbmL6THvMfTRPZMgC1wfYe9s3DdoQZcM")

			networkMgr.EXPECT().Publish(gomock.Any(), link).Times(1)
			networkMgr.EXPECT().NodeID().Times(1).Return(peerID)

			govMgr.EXPECT().Current().Times(1).Return(nil)

			lh, _ := s.CreateLink(context.Background(), link)
			segment, err := s.GetSegment(context.Background(), lh)
			assert.NoError(t, err, "s.GetSegment()")
			assert.Equal(t, "QmeZjNhdKPNNEtCbmL6THvMfTRPZMgC1wfYe9s3DdoQZcM", segment.Link.Meta.Data[constants.NodeIDKey])
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			networkMgr := mocknetworkmanager.NewMockNetworkManager(ctrl)
			networkMgr.EXPECT().AddListener().Times(1)

			governanceManager := mockvalidator.NewMockGovernanceManager(ctrl)
			governanceManager.EXPECT().ListenAndUpdate(gomock.Any()).Times(1)

			s := NewTestStoreBuilder(ctrl).WithNetworkManager(networkMgr).WithGovernanceManager(governanceManager).Build()
			<-time.After(10 * time.Millisecond)
			tt.run(t, s, networkMgr, governanceManager)
		})
	}
}

func TestIndigoStoreImpl(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	governanceManager := mockvalidator.NewMockGovernanceManager(ctrl)
	governanceManager.EXPECT().ListenAndUpdate(gomock.Any()).AnyTimes()
	governanceManager.EXPECT().Current().AnyTimes().Return(nil)

	storetestcases.Factory{
		New: func() (indigostore.Adapter, error) {
			store := NewTestStoreBuilder(ctrl).WithGovernanceManager(governanceManager).Build()
			return store, nil
		},
	}.RunStoreTests(t)
}
