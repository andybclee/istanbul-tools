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

package setup

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/p2p/discover"
	istcommon "github.com/getamis/istanbul-tools/common"
	"github.com/getamis/istanbul-tools/docker/compose"
	"github.com/getamis/istanbul-tools/genesis"
	"github.com/urfave/cli"
)

type validatorInfo struct {
	Address  common.Address
	Nodekey  string
	NodeInfo string
}

var (
	SetupCommand = cli.Command{
		Name:  "setup",
		Usage: "Setup your Istanbul network in seconds",
		Description: `This tool helps generate:

		* Genesis block
		* Static nodes for all validators
		* Validator details

	    for Istanbul consensus.
`,
		Action: gen,
		Flags: []cli.Flag{
			numOfValidatorsFlag,
			staticNodesFlag,
			verboseFlag,
			quorumFlag,
			dockerComposeFlag,
			saveFlag,
			keysFolderFlag,
			scriptFolderFlag,
			fundingAddrFlag,
		},
	}
)

func gen(ctx *cli.Context) error {
	num := ctx.Int(numOfValidatorsFlag.Name)

	keys, nodekeys, addrs := istcommon.GenerateKeys(num)
	var nodes []string

	if ctx.Bool(verboseFlag.Name) {
		fmt.Println("validators")
	}

	for i := 0; i < num; i++ {
		v := &validatorInfo{
			Address: addrs[i],
			Nodekey: nodekeys[i],
			NodeInfo: discover.NewNode(
				discover.PubkeyID(&keys[i].PublicKey),
				net.ParseIP("0.0.0.0"),
				0,
				uint16(21000)).String(),
		}

		nodes = append(nodes, string(v.NodeInfo))

		if ctx.Bool(verboseFlag.Name) {
			str, _ := json.MarshalIndent(v, "", "\t")
			fmt.Println(string(str))

			if ctx.Bool(saveFlag.Name) {
				folderName := strconv.Itoa(i)
				if keyFolder != "" {
					folderName = keyFolder
				}
				os.MkdirAll(folderName, os.ModePerm)
				ioutil.WriteFile(path.Join(folderName, "nodekey"+strconv.Itoa(i+1)), []byte(nodekeys[i]), os.ModePerm)

				//validators 정보를 validators.txt에 저장
				ioutil.WriteFile(path.Join(folderName, "validator"+strconv.Itoa(i+1)), []byte(str), os.ModePerm)
			}
		}
	}

	if ctx.Bool(verboseFlag.Name) {
		fmt.Print("\n\n\n")
	}

	staticNodes, _ := json.MarshalIndent(nodes, "", "\t")
	if ctx.Bool(staticNodesFlag.Name) {
		name := "static-nodes.json"
		fmt.Println(name)

		// 스크립트 저장 위치가 지정되면 해당 위치에 저장.
		if scriptFolder != "" {
			os.MkdirAll(scriptFolder, os.ModePerm)
			name = path.Join(scriptFolder, name)
		}

		fmt.Println(string(staticNodes))
		fmt.Print("\n\n\n")

		if ctx.Bool(saveFlag.Name) {
			ioutil.WriteFile(name, staticNodes, os.ModePerm)
		}
	}

	var jsonBytes []byte
	isQuorum := ctx.Bool(quorumFlag.Name)

	g := genesis.New(
		genesis.Validators(addrs...),
		genesis.Alloc(addInitialFund(addrs), new(big.Int).Exp(big.NewInt(10), big.NewInt(50), nil)),
	)

	if isQuorum {
		jsonBytes, _ = json.MarshalIndent(genesis.ToQuorum(g, true), "", "    ")
	} else {
		jsonBytes, _ = json.MarshalIndent(g, "", "    ")
	}

	genesisname := "genesis.json"
	fmt.Println(genesisname)

	// 스크립트 저장 위치가 지정되면 해당 위치에 저장.
	if scriptFolder != "" {
		os.MkdirAll(scriptFolder, os.ModePerm)
		genesisname = path.Join(scriptFolder, genesisname)
	}

	fmt.Println(string(jsonBytes))

	if ctx.Bool(saveFlag.Name) {
		ioutil.WriteFile(genesisname, jsonBytes, os.ModePerm)
	}

	if ctx.Bool(dockerComposeFlag.Name) {
		fmt.Print("\n\n\n")
		compose := compose.New(
			"172.16.239",
			num,
			"bb98a0b6442386d0cdf8a31b267892c1",
			nodekeys,
			removeSpacesAndLines(jsonBytes),
			removeSpacesAndLines(staticNodes),
			isQuorum)
		fmt.Println("docker-compose.yml")
		fmt.Println(compose.String())

		if ctx.Bool(saveFlag.Name) {
			ioutil.WriteFile("docker-compose.yml", []byte(compose.String()), os.ModePerm)
		}
	}

	return nil
}

func addInitialFund(addr []common.Address) []common.Address {
	// 어드레스 하나에도 많은 펀드를 추가한다. 이 계정은 key1, passwords.txt에 있음.
	fmt.Println("addInitialFund: " + fundingAddr)
	newAddr := append(addr, common.HexToAddress(fundingAddr))
	return newAddr
}

func removeSpacesAndLines(b []byte) string {
	out := string(b)
	out = strings.Replace(out, " ", "", -1)
	out = strings.Replace(out, "\t", "", -1)
	out = strings.Replace(out, "\n", "", -1)
	return out
}
