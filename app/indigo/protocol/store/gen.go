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

//go:generate mockgen -package mocknetworkmanager -destination mocknetwork/mocknetwork.go github.com/stratumn/alice/app/indigo/protocol/store NetworkManager
//go:generate mockgen -package mockstore -destination mockstore/mockstore.go github.com/stratumn/go-indigocore/store Adapter
//go:generate mockgen -package mockvalidator -destination mockvalidator/mockvalidator.go github.com/stratumn/go-indigocore/validator Validator
//go:generate mockgen -package mockvalidator -destination mockvalidator/mockgovernance.go github.com/stratumn/go-indigocore/validator GovernanceManager

package store
