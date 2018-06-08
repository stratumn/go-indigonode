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

package proposal_test

import (
	"context"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stratumn/alice/core/protocol/bootstrap/proposal"
	"github.com/stratumn/alice/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaver(t *testing.T) {
	ctx := context.Background()

	dir, _ := ioutil.TempDir("", "alice")
	peer1 := test.GeneratePeerID(t)

	t.Run("save-on-add", func(t *testing.T) {
		filePath := path.Join(dir, "save.json")
		s, err := proposal.WrapWithSaver(ctx, proposal.NewInMemoryStore(), filePath)
		require.NoError(t, err)

		_, err = os.Stat(filePath)
		require.True(t, os.IsNotExist(err))

		err = s.Add(ctx, &proposal.Request{
			Type:   proposal.RemoveNode,
			PeerID: peer1,
		})
		require.NoError(t, err)

		_, err = os.Stat(filePath)
		assert.NoError(t, err)
	})

	t.Run("save-on-remove", func(t *testing.T) {
		filePath := path.Join(dir, "remove.json")
		s, err := proposal.WrapWithSaver(ctx, proposal.NewInMemoryStore(), filePath)
		require.NoError(t, err)

		_, err = os.Stat(filePath)
		require.True(t, os.IsNotExist(err))

		err = s.Remove(ctx, peer1)
		require.NoError(t, err)

		_, err = os.Stat(filePath)
		assert.NoError(t, err)
	})

	t.Run("save-on-list", func(t *testing.T) {
		filePath := path.Join(dir, "list.json")
		s, err := proposal.WrapWithSaver(ctx, proposal.NewInMemoryStore(), filePath)
		require.NoError(t, err)

		_, err = os.Stat(filePath)
		require.True(t, os.IsNotExist(err))

		_, err = s.List(ctx)
		require.NoError(t, err)

		_, err = os.Stat(filePath)
		assert.NoError(t, err)
	})

	t.Run("reject-invalid-data", func(t *testing.T) {
		filePath := path.Join(dir, "invalid.json")
		err := ioutil.WriteFile(filePath, []byte("YOLO"), 0744)
		require.NoError(t, err)

		_, err = proposal.WrapWithSaver(ctx, proposal.NewInMemoryStore(), filePath)
		assert.Error(t, err)
	})

	t.Run("load-previous-data", func(t *testing.T) {
		filePath := path.Join(dir, "valid.json")
		s, err := proposal.WrapWithSaver(ctx, proposal.NewInMemoryStore(), filePath)
		require.NoError(t, err)

		err = s.Add(ctx, &proposal.Request{
			Type:   proposal.RemoveNode,
			PeerID: peer1,
		})
		require.NoError(t, err)

		s2, err := proposal.WrapWithSaver(ctx, proposal.NewInMemoryStore(), filePath)
		require.NoError(t, err)

		r, err := s2.Get(ctx, peer1)
		require.NoError(t, err)
		require.NotNil(t, r)

		assert.Equal(t, proposal.RemoveNode, r.Type)
		assert.Equal(t, peer1, r.PeerID)
	})
}
