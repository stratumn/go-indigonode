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

package file

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	pb "github.com/stratumn/alice/pb/storage"
	"github.com/stretchr/testify/assert"
)

func TestFileHandler_Write(t *testing.T) {

	tmpPath := "/tmp"

	fileHandler := NewHandler(tmpPath)

	t.Run("successfully-write", func(t *testing.T) {
		ctx := context.Background()
		chunkCh := make(chan *pb.FileChunk)
		doneCh := make(chan struct{})

		fileName := fmt.Sprintf("TheFile-%d", time.Now().UnixNano())
		path := "/tmp/" + fileName
		defer os.Remove(path)

		go func() {
			f, err := fileHandler.WriteFile(ctx, chunkCh)
			assert.NoError(t, err, "WriteFile")

			// Check that right file pointer is returned.
			assert.Equal(t, path, f.Name(), "f.Name()")

			// Check that file has the right content.
			assert.NoError(t, err, "os.Open()")
			expected := []byte("Hello there! How are you today?")
			content := make([]byte, 100)
			zeros := make([]byte, 100-len(expected))
			f.Seek(0, 0)
			_, err = f.Read(content)
			assert.NoError(t, err, "ReadFile()")
			assert.Equal(t, expected, content[:len(expected)], "ReadFile()")
			assert.Equal(t, zeros, content[len(expected):], "ReadFile()")

			doneCh <- struct{}{}
		}()

		chunkCh <- &pb.FileChunk{
			FileName: fileName,
			Data:     []byte("Hello there! "),
		}
		chunkCh <- &pb.FileChunk{
			Data: []byte("How are you today?"),
		}

		close(chunkCh)

		<-doneCh
	})

	t.Run("missing-name", func(t *testing.T) {
		ctx := context.Background()
		chunkCh := make(chan *pb.FileChunk)
		doneCh := make(chan struct{})

		fileName := fmt.Sprintf("TheFile-%d", time.Now().UnixNano())
		path := "/tmp/" + fileName
		defer os.Remove(path)

		go func() {
			_, err := fileHandler.WriteFile(ctx, chunkCh)
			assert.EqualError(t, err, ErrFileNameMissing.Error(), "WriteFile()")

			// Check that file has been deleted.
			_, err = os.Stat(path)
			assert.True(t, os.IsNotExist(err), "os.Stat()")

			doneCh <- struct{}{}
		}()

		chunkCh <- &pb.FileChunk{
			Data: []byte("Hello there! "),
		}

		close(chunkCh)

		<-doneCh
	})

	t.Run("bad-path", func(t *testing.T) {
		ctx := context.Background()
		chunkCh := make(chan *pb.FileChunk)
		doneCh := make(chan struct{})

		fileName := fmt.Sprintf("thisdoesnotexist/TheFile-%d", time.Now().UnixNano())
		path := "/tmp/" + fileName
		defer os.Remove(path)

		go func() {
			_, err := fileHandler.WriteFile(ctx, chunkCh)
			assert.True(t, os.IsNotExist(err), "WriteFile()")

			// Check that file has not been written anyway.
			_, err = os.Stat(path)
			assert.True(t, os.IsNotExist(err), "os.Stat()")

			doneCh <- struct{}{}
		}()

		chunkCh <- &pb.FileChunk{
			FileName: fileName,
			Data:     []byte("Hello there! "),
		}

		close(chunkCh)

		<-doneCh
	})

	t.Run("context-done", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		chunkCh := make(chan *pb.FileChunk)
		doneCh := make(chan struct{})

		fileName := fmt.Sprintf("TheFile-%d", time.Now().UnixNano())
		path := "/tmp/" + fileName
		defer os.Remove(path)

		go func() {
			_, err := fileHandler.WriteFile(ctx, chunkCh)
			assert.Error(t, err, "WriteFile")

			// Check that file has not been written anyway.
			_, err = os.Stat(path)
			assert.True(t, os.IsNotExist(err), "os.Stat()")

			doneCh <- struct{}{}
		}()

		chunkCh <- &pb.FileChunk{
			FileName: fileName,
			Data:     []byte("Hello there! "),
		}

		cancel()

		<-doneCh
	})

}
