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
