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
	"strconv"
	"strings"
	"time"

	"github.com/pelletier/go-toml"
	"github.com/pkg/errors"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
)

var (
	// log is the logger for the configuration package.
	log = logging.Logger("cfg")

	// ErrUnexistingKey is the error returned when trying to get/set an unknown key.
	ErrUnexistingKey = errors.New("setting not found")

	// ErrEditGroupConfig is the error returned when trying to set
	// group of configuration attributes.
	ErrEditGroupConfig = errors.New("cannot edit a group of attribute")
)

//sliceSeparator is the separator used for delimiting the elements of a slice.
const sliceSeparator = ","

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

// NewSet returns a new set of configurables.
func NewSet(configurables []Configurable) Set {
	set := Set{}
	for _, cfg := range configurables {
		set[cfg.ID()] = cfg
	}

	return set
}

// Load loads a TOML file and sets the configurations of a set of
// configurables.
func (s Set) Load(filename string) error {
	ctx := logging.ContextWithLoggable(context.Background(), logging.Metadata{
		"filename": filename,
		"set":      s,
	})
	event := log.EventBegin(ctx, "Load")
	defer event.Done()

	tree, err := toml.LoadFile(filename)
	if err != nil {
		event.SetError(err)
		return errors.WithStack(err)
	}

	if err := s.fromTree(ctx, tree); err != nil {
		event.SetError(err)
		return err
	}

	return nil
}

// Save saves the configurations of a set of configurables to a TOML file.
func (s Set) Save(filename string, perms os.FileMode, overwrite bool) error {
	ctx := logging.ContextWithLoggable(context.Background(), logging.Metadata{
		"filename":  filename,
		"perms":     perms,
		"overwrite": overwrite,
	})
	event := log.EventBegin(ctx, "Save")
	defer event.Done()

	err := s.Configs().Save(filename, perms, overwrite)
	if err != nil {
		event.SetError(err)
	}

	return err
}

// Tree returns a toml tree filled with the configuration set's data.
func (s Set) Tree() (*Tree, error) {
	structuredSet := structuralize(context.Background(), s.Configs())
	abstractVal := reflect.ValueOf(structuredSet)
	concreteVal := reflect.Indirect(abstractVal).Interface()
	tree, err := treeFromStruct(concreteVal)
	if err != nil {
		return nil, err
	}
	return &Tree{tree}, nil
}

// Get returns the value of the tree indexed by the provided key.
// The key is a dot-separated path (e.g. a.b.c) without single/double quoted strings.
func (s Set) Get(key string) (interface{}, error) {
	tree, err := s.Tree()
	if err != nil {
		return nil, err
	}
	value := tree.GetDefault(key, nil)
	if value == nil {
		return nil, errors.Wrapf(ErrUnexistingKey, "could not get %q", key)
	}
	return value, nil
}

// Set edits the value of the tree indexed by the provided key.
// The value must be convertible to the right type for this setting (int, bool, str or slice).
// It fails if the key does not exist.
func (s Set) Set(key string, value string) error {
	tree, err := s.Tree()
	if err != nil {
		return err
	}

	currentVal := tree.Get(key)
	if currentVal == nil {
		return errors.Wrapf(ErrUnexistingKey, "could not set %q", key)
	}

	newValue, err := getArgValue(currentVal, value)
	if err != nil {
		return errors.Wrapf(err, "could not set %q", key)
	}

	if err := tree.Set(key, newValue); err != nil {
		return err
	}
	return s.fromTree(context.Background(), tree.tree)
}

// setValuesFromTree sets the values of a set from a TOML tree.
func (s Set) fromTree(ctx context.Context, tree *toml.Tree) error {
	structuredSet := structuralize(ctx, s.Configs())

	if err := tree.Unmarshal(structuredSet); err != nil {
		return errors.WithStack(err)
	}

	// Use reflect to get the values from the struct.
	v := reflect.Indirect(reflect.ValueOf(structuredSet))

	for id, configurable := range s {
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

// Configs returns the current configurations of a set of configurables.
func (s Set) Configs() ConfigSet {
	cs := ConfigSet{}

	for id, configurable := range s {
		cs[id] = configurable.Config()
	}

	return cs
}

// ConfigSet represents a set of configurations.
type ConfigSet map[string]interface{}

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

// getArgValue converts value to its right type based on the current value's type.
// the value can be a coma-separated list of items if the value is a slice.
func getArgValue(current interface{}, value string) (interface{}, error) {

	// We may not know the slice underlying value's type, in which case
	// a slice of strings is returned.
	caseSlice := func(currentSlice []interface{}) (ret []interface{}, err error) {
		sliceValues := strings.Split(value, sliceSeparator)
		for _, elem := range sliceValues {
			var val interface{}
			if len(currentSlice) > 0 {
				val, err = getArgValue(currentSlice[0], elem)
			} else {
				val, err = getArgValue(elem, elem)
			}
			if err != nil {
				return nil, err
			}
			ret = append(ret, val)
		}
		return ret, nil
	}

	switch t := current.(type) {
	case string:
		return value, nil
	case int64:
		val, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, errors.Errorf("wrong type for value %q (expected int)", value)
		}
		return val, nil
	case bool:
		val, err := strconv.ParseBool(value)
		if err != nil {
			return nil, errors.Errorf("wrong type for value %q (expected bool)", value)
		}
		return val, nil
	case []interface{}:
		return caseSlice(t)
	case *Tree:
		return nil, ErrEditGroupConfig
	default:
		return nil, errors.Errorf("unsupported type: %T", t)
	}
}
