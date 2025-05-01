package core

import (
	"bytes"
	"fmt"
	"math/big"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
)

type BlockReplicationEvent struct {
	Hash string
	Data []byte
}

func (bc *BlockChain) createBlockReplica(block *types.Block, replicaConfig *ReplicaConfig, chainConfig *params.ChainConfig, stateSpecimen *types.StateSpecimen) error {

	// blobs
	var blobTxSidecars []*types.BlobTxSidecar
	// if replicaConfig.EnableBlob {
	// 	for sidecarData := range types.BlobTxSidecarChan {
	// 		if sidecarData.BlockNumber.Uint64() == block.NumberU64() {
	// 			log.Info("Consuming BlobTxSidecar Match From Chain Sync Channel", "Block Number:", sidecarData.BlockNumber.Uint64())
	// 			blobTxSidecars = append(blobTxSidecars, sidecarData.Blobs)
	// 		} else {
	// 			log.Info("Failing BlobTxSidecar Match from Chain Sync Channel", "Block Number:", sidecarData.BlockNumber.Uint64())
	// 		}
	// 		log.Info("BlobTxSidecar Header", "Block Number:", sidecarData.BlockNumber.Uint64())
	// 		log.Info("Chain Sync Sidecar Channel", "Length:", len(types.BlobTxSidecarChan))
	// 	}
	// }
	//block replica with blobs
	exportBlockReplica, err := bc.createReplica(block, replicaConfig, chainConfig, stateSpecimen, blobTxSidecars)
	if err != nil {
		return err
	}
	//encode to rlp
	blockReplicaRLP, err := rlp.EncodeToBytes(exportBlockReplica)
	if err != nil {
		log.Error("error encoding block replica rlp", "error", err)
		return err
	}

	sHash := block.Hash().String()

	if atomic.LoadUint32(replicaConfig.HistoricalBlocksSynced) == 0 {
		log.Info("BSP running in Live mode", "Unexported block ", block.NumberU64(), "hash", sHash)
		return nil
	} else if atomic.LoadUint32(replicaConfig.HistoricalBlocksSynced) == 1 {
		log.Info("Creating Block Specimen", "Exported block", block.NumberU64(), "hash", sHash)
		bc.blockReplicationFeed.Send(BlockReplicationEvent{
			sHash,
			blockReplicaRLP,
		})
		return nil
	} else {
		return fmt.Errorf("error in setting atomic config historical block sync: %v", replicaConfig.HistoricalBlocksSynced)
	}
}

func (bc *BlockChain) createReplica(block *types.Block, replicaConfig *ReplicaConfig, chainConfig *params.ChainConfig, stateSpecimen *types.StateSpecimen, blobSpecimen []*types.BlobTxSidecar) (*types.ExportBlockReplica, error) {
	log.Info("Creating Block Replica", "Block Number:", block.NumberU64(), "Block Hash:", block.Hash().String())
	bHash := block.Hash()
	bNum := block.NumberU64()

	//totalDifficulty
	//tdRLP := rawdb.ReadTdRLP(bc.db, bHash, bNum)
	td := new(big.Int)
	td.SetUint64(0)
	//if err := rlp.Decode(bytes.NewReader(tdRLP), td); err != nil {
	//	log.Error("Invalid block total difficulty RLP ", "hash ", bHash, "err", err)
	//	return nil, err
	//}

	//header
	headerRLP := rawdb.ReadHeaderRLP(bc.db, bHash, bNum)
	header := new(types.Header)
	if err := rlp.Decode(bytes.NewReader(headerRLP), header); err != nil {
		log.Error("Invalid block header RLP ", "hash ", bHash, "err ", err)
		return nil, err
	}

	//transactions
	txsExp := make([]*types.TransactionForExport, len(block.Transactions()))
	txsRlp := make([]*types.TransactionExportRLP, len(block.Transactions()))
	for i, tx := range block.Transactions() {
		txsExp[i] = (*types.TransactionForExport)(tx)
		txsRlp[i] = txsExp[i].ExportTx(chainConfig, block.Number(), header.BaseFee, header.Time)
		if !replicaConfig.EnableSpecimen {
			txsRlp[i].V, txsRlp[i].R, txsRlp[i].S = nil, nil, nil
		}
	}

	// withdrawals
	var withdrawalsRlp []*types.WithdrawalExportRLP = nil
	if chainConfig.IsShanghai(block.Number(), block.Time()) {
		withdrawalsExp := make([]*types.WithdrawalForExport, len(block.Withdrawals()))
		withdrawalsRlp = make([]*types.WithdrawalExportRLP, len(block.Withdrawals()))
		for i, withdrawal := range block.Withdrawals() {
			withdrawalsExp[i] = (*types.WithdrawalForExport)(withdrawal)
			withdrawalsRlp[i] = withdrawalsExp[i].ExportWithdrawal()
		}
	}

	//receipts
	receipts := rawdb.ReadRawReceipts(bc.db, bHash, bNum)
	receiptsExp := make([]*types.ReceiptForExport, len(receipts))
	receiptsRlp := make([]*types.ReceiptExportRLP, len(receipts))
	for i, receipt := range receipts {
		receiptsExp[i] = (*types.ReceiptForExport)(receipt)
		receiptsRlp[i] = receiptsExp[i].ExportReceipt()
	}

	//senders
	signer := types.MakeSigner(bc.chainConfig, block.Number(), block.Time())
	senders := make([]common.Address, 0, len(block.Transactions()))
	for _, tx := range block.Transactions() {
		sender, err := types.Sender(signer, tx)
		if err != nil {
			return nil, err
		} else {
			senders = append(senders, sender)
		}
	}

	//uncles
	uncles := block.Uncles()

	//block replica export
	if replicaConfig.EnableSpecimen && replicaConfig.EnableResult && replicaConfig.EnableBlob {
		exportBlockReplica := &types.ExportBlockReplica{
			Type:           "block-replica",
			NetworkId:      chainConfig.ChainID.Uint64(),
			Hash:           bHash,
			TotalDiff:      td,
			Header:         header,
			Transactions:   txsRlp,
			Uncles:         uncles,
			Receipts:       receiptsRlp,
			Senders:        senders,
			State:          stateSpecimen,
			Withdrawals:    withdrawalsRlp,
			BlobTxSidecars: []*types.BlobTxSidecar{},
		}
		log.Debug("Exporting full block-replica with blob-specimen")
		return exportBlockReplica, nil
	} else if replicaConfig.EnableSpecimen && !replicaConfig.EnableResult {
		exportBlockReplica := &types.ExportBlockReplica{
			Type:           "block-specimen",
			NetworkId:      chainConfig.ChainID.Uint64(),
			Hash:           bHash,
			TotalDiff:      td,
			Header:         header,
			Transactions:   txsRlp,
			Uncles:         uncles,
			Receipts:       []*types.ReceiptExportRLP{},
			Senders:        senders,
			State:          stateSpecimen,
			Withdrawals:    withdrawalsRlp,
			BlobTxSidecars: []*types.BlobTxSidecar{},
		}
		log.Debug("Exporting block-specimen only (no blob specimens)")
		return exportBlockReplica, nil
	} else if !replicaConfig.EnableSpecimen && replicaConfig.EnableResult {
		exportBlockReplica := &types.ExportBlockReplica{
			Type:           "block-result",
			NetworkId:      chainConfig.ChainID.Uint64(),
			Hash:           bHash,
			TotalDiff:      td,
			Header:         header,
			Transactions:   txsRlp,
			Uncles:         uncles,
			Receipts:       receiptsRlp,
			Senders:        senders,
			State:          &types.StateSpecimen{},
			BlobTxSidecars: []*types.BlobTxSidecar{},
		}
		log.Debug("Exporting block-result only (no blob specimens)")
		return exportBlockReplica, nil
	} else {
		return nil, fmt.Errorf("--replication.targets flag is invalid without --replica.specimen and/or --replica.result, ADD --replica.blob with both replica.specimen AND replica.result flags for complete unified state capture aka block-replica)")
	}
}

// SubscribeChainReplicationEvent registers a subscription of ChainReplicationEvent.
func (bc *BlockChain) SubscribeBlockReplicationEvent(ch chan<- BlockReplicationEvent) event.Subscription {
	return bc.scope.Track(bc.blockReplicationFeed.Subscribe(ch))
}

func (bc *BlockChain) SetBlockReplicaExports(replicaConfig *ReplicaConfig) bool {
	if replicaConfig.EnableResult {
		bc.ReplicaConfig.EnableResult = true
	}
	if replicaConfig.EnableSpecimen {
		bc.ReplicaConfig.EnableSpecimen = true
	}
	if replicaConfig.EnableBlob {
		bc.ReplicaConfig.EnableBlob = true
	}
	return true
}
