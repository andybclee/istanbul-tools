// Copyright 2017 AMIS Technologies
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package genesis

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"

	"github.com/getamis/istanbul-tools/cmd/istanbul/extra"
)

type Option func(*core.Genesis)

func Validators(addrs ...common.Address) Option {
	return func(genesis *core.Genesis) {
		extraData, err := extra.Encode("0x00", addrs)
		if err != nil {
			log.Error("Failed to encode extra data", "err", err)
			return
		}
		genesis.ExtraData = hexutil.MustDecode(extraData)
	}
}

func GasLimit(limit uint64) Option {
	return func(genesis *core.Genesis) {
		genesis.GasLimit = limit
	}
}

func Alloc(addrs []common.Address, balance *big.Int) Option {
	return func(genesis *core.Genesis) {
		// 어드레스 하나에도 많은 펀드를 추가한다. 이 계정은 key1, passwords.txt에 있음.
		// addrs = append(addrs, common.HexToAddress("f1112d590851764745499c855bd4a4574ffe9079"))
		alloc := make(map[common.Address]core.GenesisAccount)
		for _, addr := range addrs {
			alloc[addr] = core.GenesisAccount{Balance: balance}
		}
		genesis.Alloc = alloc
	}
}
