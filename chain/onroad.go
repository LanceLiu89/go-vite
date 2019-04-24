package chain

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/vitelabs/go-vite/chain/plugins"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/ledger"
)

func (c *chain) LoadOnRoad(gid types.Gid) (map[types.Address]map[types.Address][]ledger.HashHeight, error) {
	addrList, err := c.GetContractList(gid)
	if err != nil {
		return nil, err
	}

	onRoadData, err := c.indexDB.Load(addrList)
	if err != nil {
		cErr := errors.New(fmt.Sprintf("c.indexDB.Load failed, addrList is %+v。 Error: %s", addrList, err))
		c.log.Error(cErr.Error(), "method", "LoadOnRoad")
	}

	return onRoadData, nil

}

func (c *chain) DeleteOnRoad(toAddress types.Address, sendBlockHash types.Hash) {
	c.indexDB.DeleteOnRoad(toAddress, sendBlockHash)
}

func (c *chain) GetAccountOnRoadInfo(addr types.Address) (*ledger.AccountInfo, error) {
	if c.plugins == nil {
		return nil, errors.New("plugins-OnRoadInfo's service not provided.")
	}
	onRoadInfo, ok := c.plugins.GetPlugin("onRoadInfo").(*chain_plugins.OnRoadInfo)
	if !ok || onRoadInfo == nil {
		return nil, errors.New("plugins-OnRoadInfo's service not provided.")
	}
	info, err := onRoadInfo.GetAccountInfo(&addr)
	if err != nil {
		return nil, err
	}
	return info, nil
}
