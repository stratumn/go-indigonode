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

package monitoring

import (
	"context"

	log "github.com/sirupsen/logrus"

	"go.opencensus.io/tag"
)

// Common tags that can be used by all apps.
var (
	ErrorTag      = NewTag("indigo-node/keys/error")
	PeerIDTag     = NewTag("indigo-node/keys/peerid")
	ProtocolIDTag = NewTag("indigo-node/keys/protocolid")
)

// Tag can be associated to metrics.
type Tag struct {
	OCTag tag.Key
}

// NewTag creates a new tag with the given name.
func NewTag(name string) Tag {
	key, err := tag.NewKey(name)
	if err != nil {
		panic(err)
	}

	return Tag{
		OCTag: key,
	}
}

// TaggedContext is a builder to add tags to a context.
type TaggedContext struct {
	ctx context.Context
}

// NewTaggedContext adds the given tags to the context.
func NewTaggedContext(ctx context.Context) *TaggedContext {
	return &TaggedContext{ctx: ctx}
}

// Tag tags the context with the given value.
func (c *TaggedContext) Tag(t Tag, val string) *TaggedContext {
	var err error
	c.ctx, err = tag.New(c.ctx, tag.Upsert(t.OCTag, val))
	if err != nil {
		// The only failure reason is that the value is too long or contains
		// non-ASCII characters, so it's safe to simply skip and log.
		log.Warnf("could not add tag %s: %s", t.OCTag.Name(), err.Error())
	}

	return c
}

// Build returns the tagged context or an error.
func (c *TaggedContext) Build() context.Context {
	return c.ctx
}
