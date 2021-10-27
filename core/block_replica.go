package core

import (
	"bytes"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
)

type BlockReplicationEvent struct {
	Type     string
	Hash     string
	Data     []byte
	Datetime time.Time
}

func (bc *BlockChain) createBlockReplica(block *types.Block, stateSpecimen *types.StateSpecimen) error {

	//block result
	exportBlockResult, err := bc.createBlockResult(block)
	if err != nil {
		return err
	}
	//block specimen
	exportBlockSpecimen, err := bc.createBlockSpecimen(block, stateSpecimen)
	if err != nil {
		return err
	}
	//encode to rlp
	blockResultRLP, err := rlp.EncodeToBytes(exportBlockResult)
	if err != nil {
		return err
	}
	//specimen encode to rlp
	blockSpecimenRLP, err := rlp.EncodeToBytes(exportBlockSpecimen)
	if err != nil {
		return err
	}
	sHash := block.Hash().String()

	log.Info("Creating block-result replication event", "block number", block.NumberU64(), "hash", sHash)
	bc.blockReplicationFeed.Send(BlockReplicationEvent{
		"block-result",
		sHash,
		blockResultRLP,
		time.Now(),
	})

	log.Info("Creating block-specimen replication event", "block number", block.NumberU64(), "hash", sHash)
	bc.blockReplicationFeed.Send(BlockReplicationEvent{
		"block-specimen",
		sHash,
		blockSpecimenRLP,
		time.Now(),
	})

	return nil
}

func (bc *BlockChain) createBlockSpecimen(block *types.Block, stateSpecimen *types.StateSpecimen) (*types.BlockSpecimen, error) {

	bHash := block.Hash()
	bNum := block.NumberU64()

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
		txsRlp[i] = txsExp[i].ExportTx()
	}

	//uncles
	uncles := block.Uncles()

	//block specimen export
	exportBlockSpecimen := &types.BlockSpecimen{
		Hash:         bHash,
		Header:       header,
		Transactions: txsRlp,
		Uncles:       uncles,
		State:        stateSpecimen,
	}
	return exportBlockSpecimen, nil
}

func (bc *BlockChain) createBlockResult(block *types.Block) (*types.ExportBlockResult, error) {

	bHash := block.Hash()
	bNum := block.NumberU64()

	//totalDifficulty
	tdRLP := rawdb.ReadTdRLP(bc.db, bHash, bNum)
	td := new(big.Int)
	if err := rlp.Decode(bytes.NewReader(tdRLP), td); err != nil {
		log.Error("Invalid block total difficulty RLP ", "hash ", bHash, "err", err)
		return nil, err
	}

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
		txsRlp[i] = txsExp[i].ExportTx()
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
	signer := types.MakeSigner(bc.chainConfig, block.Number())
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

	//block result export
	exportBlockResult := &types.ExportBlockResult{
		Hash:         bHash,
		TotalDiff:    td,
		Header:       header,
		Transactions: txsRlp,
		Receipts:     receiptsRlp,
		Uncles:       uncles,
		Senders:      senders,
	}
	return exportBlockResult, nil
}

// SubscribeChainReplicationEvent registers a subscription of ChainReplicationEvent.
func (bc *BlockChain) SubscribeBlockReplicationEvent(ch chan<- BlockReplicationEvent) event.Subscription {
	return bc.scope.Track(bc.blockReplicationFeed.Subscribe(ch))
}
