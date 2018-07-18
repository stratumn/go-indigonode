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

//go:generate mockgen -package mockp2p -destination mockp2p/mockp2p.go github.com/stratumn/go-indigonode/app/coin/protocol/p2p P2P
//go:generate mockgen -package mockencoder -destination mockencoder/mockencoder.go github.com/stratumn/go-indigonode/app/coin/protocol/p2p Encoder

package p2p
