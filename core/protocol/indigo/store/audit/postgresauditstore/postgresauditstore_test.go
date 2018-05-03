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

package postgresauditstore_test

import (
	"testing"

	"fmt"
	"os"

	"github.com/stratumn/alice/core/protocol/indigo/store/audit"
	"github.com/stratumn/alice/core/protocol/indigo/store/audit/postgresauditstore"
	"github.com/stratumn/alice/core/protocol/indigo/store/audit/storetestcases"
	"github.com/stratumn/alice/test/containers"
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
