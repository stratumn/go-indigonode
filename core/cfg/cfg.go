// Copyright © 2017 Stratumn SAS
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

/*
Package cfg provides a simple mechanism for creating and loading configuration
files.

The configuration files are saved using the TOML file format.

Modules need not worry about dealing with configuration files. They simply need
to implement the Configurable interface.
*/
package cfg

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/pelletier/go-toml"
	"github.com/pkg/errors"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
)

// log is the logger for the configuration package.
var log = logging.Logger("cfg")

// Configurable represents something that can be configured.
type Configurable interface {
	// ID returns the unique identifier of the configurable.
	ID() string

	// Config should return the current configuration or a default
	// configuration if it wasn't set.
	Config() interface{}

	// SetConfig should configure the configurable.
	SetConfig(interface{}) error
}

// Set represents a set of configurables.
type Set map[string]Configurable

// ConfigSet represents a set of configurations.
type ConfigSet map[string]interface{}

// Load loads a TOML file and sets the configurations of a set of
// configurables.
func Load(set Set, filename string) error {
	ctx := logging.ContextWithLoggable(context.Background(), logging.Metadata{
		"filename": filename,
		"set":      set,
	})
	event := log.EventBegin(ctx, "Load")
	defer event.Done()

	tree, err := toml.LoadFile(filename)
	if err != nil {
		event.SetError(err)
		return errors.WithStack(err)
	}

	if err := setValuesFromTree(ctx, set, tree); err != nil {
		event.SetError(err)
		return err
	}

	return nil
}

// Save saves the configurations of a set of configurables to a TOML file.
func Save(set Set, filename string, perms os.FileMode, overwrite bool) error {
	ctx := logging.ContextWithLoggable(context.Background(), logging.Metadata{
		"filename":  filename,
		"perms":     perms,
		"overwrite": overwrite,
	})
	event := log.EventBegin(ctx, "Save")
	defer event.Done()

	err := set.Configs().Save(filename, perms, overwrite)
	if err != nil {
		event.SetError(err)
	}

	return err
}

// Configs returns the current configurations of a set of configurables.
func (s Set) Configs() ConfigSet {
	cs := ConfigSet{}

	for id, configurable := range s {
		cs[id] = configurable.Config()
	}

	return cs
}

// Save saves a set of configurations to a file. It will return an error if the
// file already exists unless overwrite is true.
func (cs ConfigSet) Save(filename string, perms os.FileMode, overwrite bool) error {
	ctx := logging.ContextWithLoggable(context.Background(), logging.Metadata{
		"filename":  filename,
		"perms":     perms,
		"overwrite": overwrite,
	})
	event := log.EventBegin(ctx, "configSave")
	defer event.Done()

	// Backup file.
	if overwrite {
		if _, err := os.Stat(filename); !os.IsNotExist(err) {
			if err := backup(ctx, filename); err != nil {
				event.SetError(err)
				return errors.WithStack(err)
			}
		}
	}

	mode := os.O_WRONLY | os.O_CREATE

	if !overwrite {
		mode |= os.O_EXCL
	}

	f, err := os.OpenFile(filename, mode, perms)
	if err != nil {
		event.SetError(err)
		return errors.WithStack(err)
	}

	defer func() {
		if err := f.Close(); err != nil {
			log.Event(ctx, "closeError", logging.Metadata{
				"error": err.Error(),
			})
		}
	}()

	_, err = fmt.Fprintln(f, "# Alice configuration file. Keep private!!!")
	if err != nil {
		err := errors.WithStack(err)
		event.SetError(err)
		return err
	}

	enc := toml.NewEncoder(f)
	enc.QuoteMapKeys(true)

	s := reflect.Indirect(reflect.ValueOf(structuralize(ctx, cs))).Interface()

	if err := enc.Encode(s); err != nil {
		event.SetError(err)
		return errors.WithStack(err)
	}

	return errors.WithStack(f.Sync())
}

// setValuesFromTree sets the values of a set from a TOML tree.
func setValuesFromTree(ctx context.Context, set Set, tree *toml.Tree) error {
	s := structuralize(ctx, set.Configs())

	if err := tree.Unmarshal(s); err != nil {
		return errors.WithStack(err)
	}

	// Use reflect to get the values from the struct.
	v := reflect.Indirect(reflect.ValueOf(s))

	for id, configurable := range set {
		f := v.FieldByName(ucFirst(id)).Interface()

		if f == nil {
			continue
		}

		if err := configurable.SetConfig(f); err != nil {
			return err
		}
	}

	return nil
}

// Complicated, but go-toml wants a struct.
func structuralize(ctx context.Context, cs ConfigSet) interface{} {
	defer log.EventBegin(ctx, "structuralize").Done()

	var fields []reflect.StructField

	// Create the fields of a struct type dynamically.
	for id, config := range cs {
		if config == nil {
			continue
		}

		t := reflect.TypeOf(config)

		// Uppercase the name of the field so it's exported.
		name := ucFirst(id)

		tag := fmt.Sprintf(`toml:"%s" comment:"Settings for the %s module."`, id, id)

		// Append the struct field.
		fields = append(fields, reflect.StructField{
			Name: name,
			Tag:  reflect.StructTag(tag),
			Type: t,
		})
	}

	// Create the struct type.
	s := reflect.StructOf(fields)

	// Allocate an instance of the struct type.
	p := reflect.New(s)

	// Set the fields of the struct.
	for id, config := range cs {
		if config == nil {
			continue
		}

		f := p.Elem().FieldByName(ucFirst(id))
		v := reflect.ValueOf(config)
		f.Set(v)
	}

	// Return the struct pointer as an interface{}.
	return p.Interface()
}

// ucFirst makes the first letter of a string uppercase.
func ucFirst(str string) string {
	return strings.ToUpper(str[:1]) + str[1:]
}

// backup renames the configuration file by adding a timestamp.
func backup(ctx context.Context, filename string) error {
	event := log.EventBegin(ctx, "backup", logging.Metadata{
		"filename": filename,
	})
	defer event.Done()

	stamp := time.Now().UTC().Format(time.RFC3339)
	dir := filepath.Dir(filename)
	base := filepath.Base(filename)
	ext := filepath.Ext(filename)
	file := base[:len(base)-len(ext)] + "." + stamp + ext
	backup := filepath.Join(dir, file)

	if err := os.Rename(filename, backup); err != nil {
		event.SetError(err)
		return errors.WithStack(err)
	}

	return nil
}
