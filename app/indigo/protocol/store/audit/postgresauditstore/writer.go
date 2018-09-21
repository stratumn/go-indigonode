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

package postgresauditstore

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/types"

	peer "gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
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
