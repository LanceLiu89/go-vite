package pmchain

import (
	"errors"
	"fmt"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/ledger"
	"github.com/vitelabs/go-vite/pmchain/block"
	"github.com/vitelabs/go-vite/vm_db"
)

/*
 *	1. prepare
 *	2.
 */
func (c *chain) InsertAccountBlock(vmAccountBlock *vm_db.VmAccountBlock) error {

	vmAbList := []*vm_db.VmAccountBlock{vmAccountBlock}
	c.em.Trigger(prepareInsertAbsEvent, vmAbList, nil, nil)

	accountBlock := vmAccountBlock.AccountBlock
	// write unconfirmed pool
	c.cache.InsertAccountBlock(accountBlock)

	// write index database
	if err := c.indexDB.InsertAccountBlock(vmAccountBlock); err != nil {
		cErr := errors.New(fmt.Sprintf("c.indexDB.InsertAccountBlock failed, error is %s, blockHash is %s", err.Error(), accountBlock.Hash))
		c.log.Error(cErr.Error(), "method", "InsertAccountBlock")
		return cErr
	}

	// write state db
	if err := c.stateDB.Write(vmAccountBlock); err != nil {
		cErr := errors.New(fmt.Sprintf("c.stateDB.Write failed, error is %s, blockHash is %s", err.Error(), accountBlock.Hash))
		c.log.Error(cErr.Error(), "method", "InsertAccountBlock")
		return cErr
	}

	c.em.Trigger(insertAbsEvent, vmAbList, nil, nil)
	return nil
}

// no lock
func (c *chain) InsertSnapshotBlock(snapshotBlock *ledger.SnapshotBlock) (map[types.Address][]*ledger.AccountBlock, error) {
	sbList := []*ledger.SnapshotBlock{snapshotBlock}
	c.em.Trigger(prepareInsertSbsEvent, nil, nil, sbList)

	unconfirmedBlocks := c.cache.GetUnconfirmedBlocks()
	canBeSnappedBlocks, invalidSubLedger, needDeletedAccountBlocks := c.filterCanBeSnapped(unconfirmedBlocks)

	canBeSnappedSubLedger := blocksToMap(canBeSnappedBlocks)

	// write block db
	_, snapshotBlockLocation, err := c.blockDB.Write(&chain_block.SnapshotSegment{
		SnapshotBlock: snapshotBlock,
		AccountBlocks: canBeSnappedBlocks,
	})

	if err != nil {
		cErr := errors.New(fmt.Sprintf("c.blockDB.Write failed, error is %s, snapshotBlock is %+v", err.Error(), snapshotBlock))
		c.log.Error(cErr.Error(), "method", "InsertSnapshotBlock")
		return nil, cErr
	}

	// insert index TODO
	if err := c.indexDB.InsertSnapshotBlock(snapshotBlock, canBeSnappedSubLedger, snapshotBlockLocation, nil); err != nil {
		cErr := errors.New(fmt.Sprintf("c.indexDB.InsertSnapshotBlock failed, error is %s, snapshotBlock is %+v", err.Error(), snapshotBlock))
		c.log.Error(cErr.Error(), "method", "InsertSnapshotBlock")
		return nil, cErr
	}

	// flush state db
	if err := c.stateDB.Flush(&snapshotBlock.Hash, canBeSnappedBlocks); err != nil {
		cErr := errors.New(fmt.Sprintf("c.stateDB.NewNext failed, error is %s, snapshotBlock is %+v", err.Error(), snapshotBlock))
		c.log.Error(cErr.Error(), "method", "InsertSnapshotBlock")
		return nil, cErr
	}

	// remove state db
	c.stateDB.DeleteInvalidAccountBlocks(invalidSubLedger)

	// remove invalid subLedger index
	c.indexDB.DeleteInvalidAccountBlocks(invalidSubLedger)

	// remove invalid subLedger cache
	c.cache.DeleteUnconfirmedAccountBlocks(append(canBeSnappedBlocks, needDeletedAccountBlocks...))

	// update latest snapshot block cache
	c.cache.InsertSnapshotBlock(snapshotBlock)

	c.em.Trigger(InsertSbsEvent, nil, nil, sbList)

	return invalidSubLedger, nil
}
