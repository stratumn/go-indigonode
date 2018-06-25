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

//go:generate mockgen -package mockchain -destination mockchain/mockchainreader.go github.com/stratumn/go-indigonode/app/coin/protocol/chain Reader
//go:generate mockgen -package mockchain -destination mockchain/mockchainwriter.go github.com/stratumn/go-indigonode/app/coin/protocol/chain Writer
//go:generate mockgen -package mockchain -destination mockchain/mockchain.go github.com/stratumn/go-indigonode/app/coin/protocol/chain Chain

package chain
