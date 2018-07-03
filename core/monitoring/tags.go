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

package monitoring

import (
	"context"

	"go.opencensus.io/tag"
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
	err error
}

// NewTaggedContext adds the given tags to the context.
func NewTaggedContext(ctx context.Context) *TaggedContext {
	return &TaggedContext{ctx: ctx}
}

// Tag tags the context with the given value.
func (c *TaggedContext) Tag(t Tag, val string) *TaggedContext {
	newCtx, err := tag.New(c.ctx, tag.Upsert(t.OCTag, val))
	if err != nil {
		c.err = err
	} else {
		c.ctx = newCtx
	}

	return c
}

// Build returns the tagged context or an error.
func (c *TaggedContext) Build() (context.Context, error) {
	return c.ctx, c.err
}
