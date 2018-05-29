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

//+build !lint

package containers

import (
	"database/sql"
	"flag"
	"fmt"
	"math/rand"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/stratumn/go-indigocore/utils"
)

const (
	// PostgresTestURL is the URL to connect to when running postgres inside a docker container.
	PostgresTestURL = "postgres://postgres@localhost?sslmode=disable"
)

// RunPostgres starts a docker container running postgres and initialize a database.
// The container needs to be killed at the end of a test.
func RunPostgres() (string, error) {
	const (
		domain = "0.0.0.0"
		port   = "5432"
	)

	seed := int64(time.Now().Nanosecond())
	fmt.Printf("using seed %d\n", seed)
	rand.Seed(seed)
	flag.Parse()

	// Postgres container configuration.
	imageName := "postgres:10.1"
	containerName := "indigo_postgresstore_test"
	p, _ := nat.NewPort("tcp", port)
	exposedPorts := map[nat.Port]struct{}{p: {}}
	portBindings := nat.PortMap{
		p: []nat.PortBinding{
			{
				HostIP:   domain,
				HostPort: port,
			},
		},
	}

	// Stop container if it is already running, swallow error.
	utils.KillContainer(containerName)

	// Start postgres container
	if err := utils.RunContainer(containerName, imageName, exposedPorts, portBindings); err != nil {
		return "", err
	}
	// Retry until container is ready.
	if err := utils.Retry(createDatabase, 10); err != nil {
		return "", err
	}

	return containerName, nil
}

func createDatabase(attempt int) (bool, error) {
	db, err := sql.Open("postgres", PostgresTestURL)
	if err != nil {
		time.Sleep(300 * time.Millisecond)
		return true, err
	}
	if _, err := db.Exec("CREATE DATABASE postgresstore_tests;"); err != nil {
		time.Sleep(300 * time.Millisecond)
		return true, err
	}
	return false, err
}
