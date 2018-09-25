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

package postgresauditstore_test

import (
	"testing"

	"fmt"
	"os"

	"github.com/stratumn/go-node/app/indigo/protocol/store/audit"
	"github.com/stratumn/go-node/app/indigo/protocol/store/audit/postgresauditstore"
	"github.com/stratumn/go-node/app/indigo/protocol/store/audit/storetestcases"
	"github.com/stratumn/go-node/test/containers"
	"github.com/stratumn/go-indigocore/postgresstore"
	"github.com/stretchr/testify/assert"

	"github.com/stratumn/go-indigocore/utils"
)

func TestMain(m *testing.M) {
	container, err := containers.RunPostgres()
	if err != nil {
		fmt.Println("Could not start docker container, aborting: ", err)
		os.Exit(1)
	}
	testResult := m.Run()

	if err := utils.KillContainer(container); err != nil {
		os.Exit(1)
	}

	os.Exit(testResult)
}
func TestNew(t *testing.T) {
	s, err := postgresauditstore.New(&postgresstore.Config{
		URL: containers.PostgresTestURL,
	})

	assert.NoError(t, err)
	assert.NotNil(t, s)
}

func TestCreate(t *testing.T) {
	s, err := postgresauditstore.New(&postgresstore.Config{
		URL: containers.PostgresTestURL,
	})
	assert.NoError(t, err)
	assert.NotNil(t, s)

	defer freeStore(s)

	err = s.Create()
	assert.NoError(t, err)
}

func TestPrepare(t *testing.T) {
	s, err := postgresauditstore.New(&postgresstore.Config{
		URL: containers.PostgresTestURL,
	})
	assert.NoError(t, err)
	assert.NotNil(t, s)

	defer freeStore(s)

	err = s.Create()
	assert.NoError(t, err)
	err = s.Prepare()
	assert.NoError(t, err)
}

func TestPostgresAuditStore(t *testing.T) {
	storetestcases.Factory{
		New:  createAdapter,
		Free: freeAdapter,
	}.RunTests(t)
}

func createStore() (*postgresauditstore.PostgresAuditStore, error) {
	a, err := postgresauditstore.New(&postgresstore.Config{URL: "postgres://postgres@localhost/postgresstore_tests?sslmode=disable"})
	if err := a.Create(); err != nil {
		return nil, err
	}
	if err := a.Prepare(); err != nil {
		return nil, err
	}
	return a, err
}
func createAdapter() (audit.Store, error) {
	return createStore()
}

func freeStore(s *postgresauditstore.PostgresAuditStore) {
	if err := s.Drop(); err != nil {
		panic(err)
	}
	if err := s.Close(); err != nil {
		panic(err)
	}
}

func freeAdapter(s audit.Store) {
	freeStore(s.(*postgresauditstore.PostgresAuditStore))
}
