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

package trie

import (
	"fmt"

	"github.com/pkg/errors"
)

// Dump returns a string containing a dump of all the values in the trie for
// debugging and testing. Not pretty but does the job.
func (t *Trie) Dump() (string, error) {
	root, err := t.cache.Get(nil)
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

		target, err := t.cache.Get(path)
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
