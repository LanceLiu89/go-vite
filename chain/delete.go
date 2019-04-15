package chain

import (
	"errors"
	"fmt"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/ledger"
)

func (c *chain) DeleteSnapshotBlocks(toHash types.Hash) ([]*ledger.SnapshotChunk, error) {
	height, err := c.indexDB.GetSnapshotBlockHeight(&toHash)

	if err != nil {
		cErr := errors.New(fmt.Sprintf("c.indexDB.GetSnapshotBlockHeight failed, snapshotHash is %s. Error: %s", toHash, err.Error()))
		c.log.Error(cErr.Error(), "method", "deleteSnapshotBlocks")
		return nil, cErr
	}
	if height <= 1 {
		cErr := errors.New(fmt.Sprintf("height <= 1,  snapshotHash is %s. Error: %s", toHash, err.Error()))
		c.log.Error(cErr.Error(), "method", "deleteSnapshotBlocks")
		return nil, cErr
	}

	return c.DeleteSnapshotBlocksToHeight(height)
}

// delete and recover unconfirmed cache
func (c *chain) DeleteSnapshotBlocksToHeight(toHeight uint64) ([]*ledger.SnapshotChunk, error) {
	latestHeight := c.GetLatestSnapshotBlock().Height
	if toHeight > latestHeight || toHeight <= 1 {
		cErr := errors.New(fmt.Sprintf("toHeight is %d, GetLatestHeight is %d", toHeight, latestHeight))
		c.log.Error(cErr.Error(), "method", "DeleteSnapshotBlocksToHeight")
		return nil, cErr
	}

	tmpLocation, err := c.indexDB.GetSnapshotBlockLocation(toHeight - 1)
	if err != nil {
		cErr := errors.New(fmt.Sprintf("c.indexDB.GetSnapshotBlockLocation failed, height is %d. Error: %s", toHeight-1, err.Error()))
		c.log.Error(cErr.Error(), "method", "DeleteSnapshotBlocksToHeight")
		return nil, cErr
	}

	location, err := c.blockDB.GetNextLocation(tmpLocation)
	if err != nil {
		cErr := errors.New(fmt.Sprintf("c.blockDB.GetNextLocation failed. Error: %s", err.Error()))
		c.log.Error(cErr.Error(), "method", "DeleteSnapshotBlocksToHeight")
		return nil, cErr
	}

	if location == nil {
		cErr := errors.New(fmt.Sprintf("location is nil, toHeight is %d",
			toHeight))
		c.log.Error(cErr.Error(), "method", "DeleteSnapshotBlocksToHeight")

		return nil, cErr
	}

	// block db rollback
	snapshotChunks, err := c.blockDB.Rollback(location)

	if err != nil {
		cErr := errors.New(fmt.Sprintf("c.blockDB.RollbackAccountBlocks failed, location is %d. Error: %s,", location, err.Error()))
		c.log.Crit(cErr.Error(), "method", "deleteSnapshotBlocksToLocation")
	}
	if len(snapshotChunks) <= 0 {
		return nil, nil
	}

	// rollback blocks db
	hasStorageRedoLog, err := c.stateDB.StorageRedo().HasRedo(toHeight)

	if err != nil {
		cErr := errors.New(fmt.Sprintf("c.stateDB.StorageRedo().HasRedo() failed, toHeight is %d. Error: %s", toHeight, err.Error()))
		c.log.Error(cErr.Error(), "method", "DeleteSnapshotBlocksToHeight")
		return nil, cErr
	}

	var newUnconfirmedBlocks []*ledger.AccountBlock
	if hasStorageRedoLog {
		newUnconfirmedBlocks = snapshotChunks[0].AccountBlocks
	}
	// append old unconfirmed blocks
	snapshotChunks = append(snapshotChunks, &ledger.SnapshotChunk{
		AccountBlocks: c.cache.GetUnconfirmedBlocks(),
	})

	//FOR DEBUG
	//for _, chunk := range snapshotChunks {
	//	if chunk.SnapshotBlock != nil {
	//		fmt.Printf("Delete snapshot block %d\n", chunk.SnapshotBlock.Height)
	//		for addr, sc := range chunk.SnapshotBlock.SnapshotContent {
	//			fmt.Printf("%d SC: %s %d %s\n", chunk.SnapshotBlock.Height, addr, sc.Height, sc.Hash)
	//		}
	//	}
	//	for _, ab := range chunk.AccountBlocks {
	//		fmt.Printf("delete by sb %s %d %s\n", ab.AccountAddress, ab.Height, ab.Hash)
	//	}
	//}

	c.em.Trigger(prepareDeleteSbsEvent, nil, nil, nil, snapshotChunks)

	// rollback index db
	if err := c.indexDB.RollbackSnapshotBlocks(snapshotChunks, newUnconfirmedBlocks); err != nil {
		cErr := errors.New(fmt.Sprintf("c.indexDB.RollbackSnapshotBlocks failed, error is %s", err.Error()))
		c.log.Crit(cErr.Error(), "method", "deleteSnapshotBlocksToLocation")
	}

	// rollback cache
	if err := c.cache.RollbackSnapshotBlocks(snapshotChunks, newUnconfirmedBlocks); err != nil {
		cErr := errors.New(fmt.Sprintf("c.cache.RollbackSnapshotBlocks failed, error is %s", err.Error()))
		c.log.Crit(cErr.Error(), "method", "deleteSnapshotBlocksToLocation")
	}

	// rollback state db
	if err := c.stateDB.RollbackSnapshotBlocks(snapshotChunks, newUnconfirmedBlocks); err != nil {
		cErr := errors.New(fmt.Sprintf("c.stateDB.RollbackSnapshotBlocks failed, error is %s", err.Error()))
		c.log.Crit(cErr.Error(), "method", "deleteSnapshotBlocksToLocation")
	}

	c.flusher.Flush(true)

	c.em.Trigger(DeleteSbsEvent, nil, nil, nil, snapshotChunks)

	return snapshotChunks, nil
}

func (c *chain) DeleteAccountBlocks(addr types.Address, toHash types.Hash) ([]*ledger.AccountBlock, error) {
	return c.deleteAccountBlockByHeightOrHash(addr, 0, &toHash)
}

func (c *chain) DeleteAccountBlocksToHeight(addr types.Address, toHeight uint64) ([]*ledger.AccountBlock, error) {
	return c.deleteAccountBlockByHeightOrHash(addr, toHeight, nil)
}

func (c *chain) deleteAccountBlockByHeightOrHash(addr types.Address, toHeight uint64, toHash *types.Hash) ([]*ledger.AccountBlock, error) {
	unconfirmedBlocks := c.cache.GetUnconfirmedBlocks()
	if len(unconfirmedBlocks) <= 0 {
		cErr := errors.New(fmt.Sprintf("blocks is not unconfirmed, Addr is %s, toHeight is %d", addr, toHeight))
		c.log.Error(cErr.Error(), "method", "deleteAccountBlockByHeightOrHash")
		return nil, cErr
	}
	var planDeleteBlocks []*ledger.AccountBlock
	for i, unconfirmedBlock := range unconfirmedBlocks {
		if (toHash != nil && unconfirmedBlock.Hash == *toHash) ||
			(toHeight > 0 && unconfirmedBlock.Height == toHeight) {
			planDeleteBlocks = unconfirmedBlocks[i:]
			break
		}
	}
	if len(planDeleteBlocks) <= 0 {
		cErr := errors.New(fmt.Sprintf("len(planDeleteBlocks) <= 0"))
		c.log.Error(cErr.Error(), "method", "deleteAccountBlockByHeightOrHash")
		return nil, cErr
	}

	needDeleteBlocks := c.computeDependencies(planDeleteBlocks)

	c.deleteAccountBlocks(needDeleteBlocks)

	return needDeleteBlocks, nil
}

func (c *chain) deleteAccountBlocks(blocks []*ledger.AccountBlock) {
	//FOR DEBUG
	//for _, ab := range blocks {
	//	fmt.Printf("delete by ab %s %d %s\n", ab.AccountAddress, ab.Height, ab.Hash)
	//}

	c.em.Trigger(prepareDeleteAbsEvent, nil, blocks, nil, nil)

	// rollback index db
	if err := c.indexDB.RollbackAccountBlocks(blocks); err != nil {
		cErr := errors.New(fmt.Sprintf("c.indexDB.RollbackAccountBlocks failed. Error: %s", err.Error()))
		c.log.Crit(cErr.Error(), "method", "deleteAccountBlocks")
	}

	// rollback cache
	if err := c.cache.RollbackAccountBlocks(blocks); err != nil {
		cErr := errors.New(fmt.Sprintf("c.cache.RollbackAccountBlocks failed. Error: %s", err.Error()))
		c.log.Crit(cErr.Error(), "method", "deleteAccountBlocks")
	}

	// rollback state db
	if err := c.stateDB.RollbackAccountBlocks(blocks); err != nil {
		cErr := errors.New(fmt.Sprintf("c.stateDB.RollbackAccountBlocks failed. Error: %s", err.Error()))
		c.log.Crit(cErr.Error(), "method", "deleteAccountBlocks")
	}

	c.em.Trigger(DeleteAbsEvent, nil, blocks, nil, nil)
}
