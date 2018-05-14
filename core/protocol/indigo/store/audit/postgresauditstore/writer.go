// Copyright 2017 Stratumn SAS. All rights reserved.
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

package postgresauditstore

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/types"

	peer "gx/ipfs/QmZoWKhxUmZ2seW4BzX6fJkNR8hh9PsGModr7q171yq2SS/go-libp2p-peer"
)

type writer struct {
	stmts writeStmts
}

func (a *writer) createLink(ctx context.Context, link *cs.Link, peerID peer.ID) (*types.Bytes32, error) {
	linkHash, err := link.Hash()
	if err != nil {
		return linkHash, err
	}

	data, err := json.Marshal(link)
	if err != nil {
		return linkHash, errors.WithStack(err)
	}

	_, err = a.stmts.CreateLink.Exec(linkHash[:], string(data), peerID.Pretty())

	return linkHash, errors.WithStack(err)
}

func (a *writer) addEvidence(ctx context.Context, linkHash *types.Bytes32, evidence *cs.Evidence) error {
	evidenceBytes, err := json.Marshal(evidence)
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = a.stmts.AddEvidence.Exec(linkHash[:], evidenceBytes, evidence.Provider)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
