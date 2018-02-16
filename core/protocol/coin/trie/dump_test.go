// Copyright © 2017-2018  Stratumn SAS
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

package trie

import (
	"fmt"

	"github.com/pkg/errors"
)

// Dump returns a string containing a dump of all the values in the trie for
// debugging and testing. Not pretty but does the job.
func (t *Trie) Dump() (string, error) {
	root, err := t.rootNode()
	if err != nil {
		return "", err
	}

	return t.dumpRec(root, nil, "")
}

func (t *Trie) dumpRec(n node, prefix []uint8, ident string) (string, error) {
	s := ""

	switch v := n.(type) {
	case *edge:
		s = ident + fmt.Sprintf("edge %v", newNibsFromNibs(v.Path...))

		if len(v.Hash) > 0 {
			s += " " + v.Hash.B58String()
		}

		path := append(prefix, v.Path...)

		target, err := t.getNode(path)
		if err != nil {
			return "", err
		}

		targetStr, err := t.dumpRec(target, path, ident+"  ")
		if err != nil {
			return "", err
		}

		s += "\n" + targetStr

	case *branch:
		if len(v.Value) > 0 {
			if len(prefix) > 0 {
				s = ident + fmt.Sprintf(
					"branch %v [%x]",
					newNibsFromNibs(prefix...),
					v.Value,
				)
			} else {
				s = ident + fmt.Sprintf("branch [%x]", v.Value)
			}
		} else if len(prefix) > 0 {
			s = ident + fmt.Sprintf("branch %v", newNibsFromNibs(prefix...))
		} else {
			s = ident + "branch"
		}

		for _, n := range v.EmbeddedNodes {
			switch v := n.(type) {
			case *edge:
				edgeStr, err := t.dumpRec(v, prefix, ident+"  ")
				if err != nil {
					return "", err
				}

				s += "\n" + edgeStr

			case null:

			default:
				return "", errors.WithStack(ErrInvalidNodeType)
			}
		}

	case *leaf:
		if len(v.Value) > 0 {
			if len(prefix) > 0 {
				s = ident + fmt.Sprintf(
					"leaf %v [%x]",
					newNibsFromNibs(prefix...),
					v.Value,
				)
			} else {
				s = ident + fmt.Sprintf("leaf [%x]", v.Value)
			}
		} else if len(prefix) > 0 {
			s = ident + fmt.Sprintf("leaf %v", newNibsFromNibs(prefix...))
		} else {
			s = ident + "leaf"
		}

	case null:
		if len(prefix) > 0 {
			s = ident + fmt.Sprintf("null %v", newNibsFromNibs(prefix...))
		} else {
			s = ident + "null"
		}

	default:
		return "", errors.WithStack(ErrInvalidNodeType)
	}

	return s, nil
}
