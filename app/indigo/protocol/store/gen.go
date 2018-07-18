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

//go:generate mockgen -package mocknetworkmanager -destination mocknetwork/mocknetwork.go github.com/stratumn/go-indigonode/app/indigo/protocol/store NetworkManager
//go:generate mockgen -package mockstore -destination mockstore/mockstore.go github.com/stratumn/go-indigocore/store Adapter
//go:generate mockgen -package mockvalidator -destination mockvalidator/mockvalidator.go github.com/stratumn/go-indigocore/validator Validator
//go:generate mockgen -package mockvalidator -destination mockvalidator/mockgovernance.go github.com/stratumn/go-indigocore/validator GovernanceManager

package store
