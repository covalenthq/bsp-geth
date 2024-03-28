package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

type ExportBlockReplica struct {
	Type           string
	NetworkId      uint64
	Hash           common.Hash
	TotalDiff      *big.Int
	Header         *Header
	Transactions   []*TransactionExportRLP
	Uncles         []*Header
	Receipts       []*ReceiptExportRLP
	Senders        []common.Address
	State          *StateSpecimen
	Withdrawals    []*WithdrawalExportRLP
	BlobTxSidecars []*BlobTxSidecar
}

type LogsExportRLP struct {
	Address     common.Address `json:"address"`
	Topics      []common.Hash  `json:"topics"`
	Data        []byte         `json:"data"`
	BlockNumber uint64         `json:"blockNumber"`
	TxHash      common.Hash    `json:"transactionHash"`
	TxIndex     uint           `json:"transactionIndex"`
	BlockHash   common.Hash    `json:"blockHash"`
	Index       uint           `json:"logIndex"`
	Removed     bool           `json:"removed"`
}

type ReceiptForExport Receipt

type ReceiptExportRLP struct {
	PostStateOrStatus []byte
	CumulativeGasUsed uint64
	TxHash            common.Hash
	ContractAddress   common.Address
	Logs              []*LogsExportRLP
	GasUsed           uint64
}

type WithdrawalForExport Withdrawal

type WithdrawalExportRLP struct {
	Index     uint64         `json:"index"`          // monotonically increasing identifier issued by consensus layer
	Validator uint64         `json:"validatorIndex"` // index of validator associated with withdrawal
	Address   common.Address `json:"address"`        // target address for withdrawn ether
	Amount    uint64         `json:"amount"`         // value of withdrawal in Gwei
}

type TransactionForExport Transaction

type TransactionExportRLP struct {
	Type         byte            `json:"type"`
	AccessList   AccessList      `json:"accessList"`
	ChainId      *big.Int        `json:"chainId"`
	AccountNonce uint64          `json:"nonce"`
	Price        *big.Int        `json:"gasPrice"`
	GasLimit     uint64          `json:"gas"`
	GasTipCap    *big.Int        `json:"gasTipCap"`
	GasFeeCap    *big.Int        `json:"gasFeeCap"`
	Sender       *common.Address `json:"from" rlp:"nil"`
	Recipient    *common.Address `json:"to" rlp:"nil"` // nil means contract creation
	Amount       *big.Int        `json:"value"`
	Payload      []byte          `json:"input"`
	V            *big.Int        `json:"v" rlp:"nil"`
	R            *big.Int        `json:"r" rlp:"nil"`
	S            *big.Int        `json:"s" rlp:"nil"`
	BlobFeeCap   *big.Int        `json:"blobFeeCap" rlp:"optional"`
	BlobHashes   []common.Hash   `json:"blobHashes" rlp:"optional"`
	BlobGas      uint64          `json:"blobGas" rlp:"optional"`
}

type BlobTxSidecarData struct {
	Blobs       *BlobTxSidecar
	BlockNumber *big.Int
}

var BlobTxSidecarChan = make(chan *BlobTxSidecarData, 1000)

func (r *ReceiptForExport) ExportReceipt() *ReceiptExportRLP {
	enc := &ReceiptExportRLP{
		PostStateOrStatus: (*Receipt)(r).statusEncoding(),
		GasUsed:           r.GasUsed,
		CumulativeGasUsed: r.CumulativeGasUsed,
		TxHash:            r.TxHash,
		ContractAddress:   r.ContractAddress,
		Logs:              make([]*LogsExportRLP, len(r.Logs)),
	}
	for i, log := range r.Logs {
		enc.Logs[i] = (*LogsExportRLP)(log)
	}
	return enc
}

func (r *WithdrawalForExport) ExportWithdrawal() *WithdrawalExportRLP {
	return &WithdrawalExportRLP{
		Index:     r.Index,
		Validator: r.Validator,
		Address:   r.Address,
		Amount:    r.Amount,
	}
}

func (tx *TransactionForExport) ExportTx(chainConfig *params.ChainConfig, blockNumber *big.Int, baseFee *big.Int, blockTime uint64) *TransactionExportRLP {
	var inner_tx *Transaction = (*Transaction)(tx)
	v, r, s := tx.inner.rawSignatureValues()
	var signer Signer = MakeSigner(chainConfig, blockNumber, blockTime)
	from, _ := Sender(signer, inner_tx)

	txData := tx.inner

	if inner_tx.Type() == BlobTxType {
		return &TransactionExportRLP{
			AccountNonce: txData.nonce(),
			Price:        txData.effectiveGasPrice(&big.Int{}, baseFee),
			GasLimit:     txData.gas(),
			Sender:       &from,
			Recipient:    txData.to(),
			Amount:       txData.value(),
			Payload:      txData.data(),
			Type:         txData.txType(),
			ChainId:      txData.chainID(),
			AccessList:   txData.accessList(),
			GasTipCap:    txData.gasTipCap(),
			GasFeeCap:    txData.gasFeeCap(),
			V:            v,
			R:            r,
			S:            s,
			BlobFeeCap:   inner_tx.BlobGasFeeCap(),
			BlobHashes:   inner_tx.BlobHashes(),
			BlobGas:      inner_tx.BlobGas(),
		}
	} else {
		return &TransactionExportRLP{
			AccountNonce: txData.nonce(),
			Price:        txData.effectiveGasPrice(&big.Int{}, baseFee),
			GasLimit:     txData.gas(),
			Sender:       &from,
			Recipient:    txData.to(),
			Amount:       txData.value(),
			Payload:      txData.data(),
			Type:         txData.txType(),
			ChainId:      txData.chainID(),
			AccessList:   txData.accessList(),
			GasTipCap:    txData.gasTipCap(),
			GasFeeCap:    txData.gasFeeCap(),
			V:            v,
			R:            r,
			S:            s,
			BlobFeeCap:   &big.Int{},
			BlobHashes:   make([]common.Hash, 0),
			BlobGas:      0,
		}
	}
}
