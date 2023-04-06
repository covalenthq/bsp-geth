//nolint:stylecheck,revive
package t8ntool

import (
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const (
	BloomByteLength = 256
	BloomBitLength  = 8 * BloomByteLength
)

// BlockReplica it's actually the "block-specimen" portion of the block replica. Fields
// like Receipts, senders are block-result-specific and won't actually be present in the input.
type BlockReplica struct {
	Type            string
	NetworkId       uint64
	Hash            common.Hash
	TotalDifficulty *BigInt
	Header          *Header
	Transactions    []*Transaction2
	Uncles          []*Header `json:"uncles"`
	Receipts        []*Receipt
	Senders         []common.Address
	State           *StateSpecimen `json:"State"`
}
type StateSpecimen struct {
	AccountRead   []*AccountRead
	StorageRead   []*StorageRead
	CodeRead      []*CodeRead
	BlockhashRead []*BlockhashRead
}

type BlockNonce [8]byte

// EncodeNonce converts the given integer to a block nonce.
func EncodeNonce(i uint64) BlockNonce {
	var n BlockNonce
	binary.BigEndian.PutUint64(n[:], i)
	return n
}

type Bloom [BloomByteLength]byte

// BytesToBloom converts a byte slice to a bloom filter.
// It panics if b is not of suitable size.
func BytesToBloom(b []byte) Bloom {
	var bloom Bloom
	bloom.SetBytes(b)
	return bloom
}

// SetBytes sets the content of b to the given bytes.
// It panics if d is not of suitable size.
func (b *Bloom) SetBytes(d []byte) {
	if len(b) < len(d) {
		panic(fmt.Sprintf("bloom bytes too big %d %d", len(b), len(d)))
	}
	copy(b[BloomByteLength-len(d):], d)
}

type Header struct {
	ParentHash  common.Hash    `json:"parentHash"`
	UncleHash   common.Hash    `json:"sha3Uncles"`
	Coinbase    common.Address `json:"miner"`
	Root        common.Hash    `json:"stateRoot"`
	TxHash      common.Hash    `json:"transactionsRoot"`
	ReceiptHash common.Hash    `json:"receiptsRoot"`
	Bloom       Bloom          `json:"logsBloom"`
	Difficulty  *BigInt        `json:"difficulty"`
	Number      *BigInt        `json:"number"`
	GasLimit    uint64         `json:"gasLimit"`
	GasUsed     uint64         `json:"gasUsed"`
	Time        uint64         `json:"timestamp"`
	Extra       []byte         `json:"extraData"`
	MixDigest   common.Hash    `json:"mixHash"`
	Nonce       BlockNonce     `json:"nonce"`
	BaseFee     *BigInt        `json:"baseFeePerGas"`
	Random      *BigInt        `json:"random" rlp:"nil,optional"`
}

type Transaction2 struct {
	Type         byte             `json:"type"`
	AccessList   types.AccessList `json:"accessList"`
	ChainId      *BigInt          `json:"chainId"`
	AccountNonce uint64           `json:"nonce"`
	Price        *BigInt          `json:"gasPrice"`
	GasLimit     uint64           `json:"gas"`
	GasTipCap    *BigInt          `json:"gasTipCap"`
	GasFeeCap    *BigInt          `json:"gasFeeCap"`
	Sender       *common.Address  `json:"from"`
	Recipient    *common.Address  `json:"to" rlp:"nil"` // nil means contract creation
	Amount       *BigInt          `json:"value"`
	Payload      []byte           `json:"input"`
	V            *BigInt          `json:"v"`
	R            *BigInt          `json:"r"`
	S            *BigInt          `json:"s"`
}

type Logs struct {
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

type Receipt struct {
	PostStateOrStatus []byte
	CumulativeGasUsed uint64
	TxHash            common.Hash
	ContractAddress   common.Address
	Logs              []*Logs
	GasUsed           uint64
}

type AccountRead struct {
	Address  common.Address
	Nonce    uint64
	Balance  *BigInt
	CodeHash common.Hash
}

type StorageRead struct {
	Account common.Address
	SlotKey common.Hash
	Value   common.Hash
}

type CodeRead struct {
	Hash common.Hash
	Code []byte
}

type BlockhashRead struct {
	BlockNumber uint64
	BlockHash   common.Hash
}

func adaptHeader(header *types.Header) (*Header, error) {
	return &Header{
		ParentHash:  header.ParentHash,
		UncleHash:   header.UncleHash,
		Coinbase:    header.Coinbase,
		Root:        header.Root,
		TxHash:      header.TxHash,
		ReceiptHash: header.ReceiptHash,
		Bloom:       BytesToBloom(header.Bloom.Bytes()),
		Difficulty:  &BigInt{header.Difficulty},
		Number:      &BigInt{header.Number},
		GasLimit:    header.GasLimit,
		GasUsed:     header.GasUsed,
		Time:        header.Time,
		Extra:       header.Extra,
		MixDigest:   header.MixDigest,
		Nonce:       EncodeNonce(header.Nonce.Uint64()),
		BaseFee:     &BigInt{header.BaseFee},
	}, nil
}

func (tx *Transaction2) adaptTransaction(signer types.Signer) (*types.Transaction, error) {
	gasPrice, value := tx.Price.Int, tx.Amount.Int
	chainId := tx.ChainId.Int

	var v, r, s *big.Int
	if tx.V != nil {
		v = tx.V.Int
	}

	if tx.R != nil {
		r = tx.R.Int
	}

	if tx.S != nil {
		s = tx.S.Int
	}
	switch tx.Type {
	case types.LegacyTxType, types.AccessListTxType:
		var legacyTx *types.Transaction
		if tx.Recipient == nil {
			legacyTx = types.NewContractCreation(uint64(tx.AccountNonce), value, uint64(tx.GasLimit), gasPrice, tx.Payload)
		} else {
			legacyTx = types.NewTransaction(uint64(tx.AccountNonce), *tx.Recipient, value, uint64(tx.GasLimit), gasPrice, tx.Payload)
		}
		legacyTx.ChainId().Set(chainId)
		legacyTx.SetData(tx.Payload)
		legacyTx.SetSignatureValues(chainId, v, r, s)
		if tx.Sender != nil {
			legacyTx.SetFrom(signer, *tx.Sender)
		}

		if tx.Type == types.AccessListTxType {
			accessListTx := types.AccessListTx{
				ChainID:    chainId,
				Nonce:      legacyTx.Nonce(),
				GasPrice:   legacyTx.GasPrice(),
				Gas:        legacyTx.Gas(),
				To:         legacyTx.To(),
				Value:      legacyTx.Value(),
				Data:       legacyTx.Data(),
				AccessList: tx.AccessList,
			}

			wTx := types.NewTx(&accessListTx)
			if tx.Sender != nil {
				fmt.Println("setting from")
				wTx.SetFrom(signer, *tx.Sender)
			}
			wTx.SetSignatureValues(chainId, v, r, s)

			return wTx, nil
		} else {
			return legacyTx, nil
		}

	case types.DynamicFeeTxType:

		dynamicFeeTx := types.DynamicFeeTx{
			ChainID:    chainId,
			Nonce:      tx.AccountNonce,
			Gas:        tx.GasLimit,
			To:         tx.Recipient,
			Value:      value,
			Data:       tx.Payload,
			AccessList: tx.AccessList,
			GasTipCap:  tx.GasTipCap.Int,
			GasFeeCap:  tx.GasFeeCap.Int,
			V:          v,
			R:          r,
			S:          s,
		}

		wTx := types.NewTx(&dynamicFeeTx)

		if tx.Sender != nil {
			wTx.SetFrom(signer, *tx.Sender)
		}
		return wTx, nil

	default:
		return nil, nil

	}
}

func convertReceipts(Receipts types.Receipts) []*Receipt {
	receipts := make([]*Receipt, 0)
	for _, rec := range Receipts {
		result_logs := make([]*Logs, 0)
		for _, logs := range rec.Logs {
			log := &Logs{
				Address:     logs.Address,
				Topics:      logs.Topics,
				Data:        logs.Data,
				BlockNumber: logs.BlockNumber,
				TxHash:      logs.TxHash,
				TxIndex:     logs.TxIndex,
				BlockHash:   logs.BlockHash,
				Index:       logs.Index,
				Removed:     logs.Removed,
			}
			result_logs = append(result_logs, log)
		}
		receipt := &Receipt{
			PostStateOrStatus: rec.PostState,
			CumulativeGasUsed: rec.CumulativeGasUsed,
			TxHash:            rec.TxHash,
			ContractAddress:   rec.ContractAddress,
			Logs:              result_logs,
			GasUsed:           rec.GasUsed,
		}
		receipts = append(receipts, receipt)
	}
	return receipts
}

func convertTransactions(txs types.Transactions) ([]*Transaction2, error) {
	var Transactions []*Transaction2
	for i, tx := range txs {
		sender, ok := tx.GetSender()
		if !ok {
			return Transactions, fmt.Errorf("tx index %d failed to get sender", i)
		}
		new_tx := &Transaction2{
			Type:         tx.Type(),
			AccessList:   tx.AccessList(),
			ChainId:      &BigInt{tx.ChainId()},
			AccountNonce: tx.Nonce(),
			Price:        &BigInt{tx.GasPrice()},
			GasLimit:     tx.Gas(),
			GasTipCap:    &BigInt{tx.GasTipCap()},
			GasFeeCap:    &BigInt{tx.GasFeeCap()},
			Sender:       &sender,
			Recipient:    tx.To(),
			Amount:       &BigInt{tx.Value()},
			Payload:      tx.Data(),
		}
		Transactions = append(Transactions, new_tx)
	}
	return Transactions, nil
}

func converUncles(ommerHeaders []*types.Header) []*Header {
	var new_uncles []*Header
	for _, uncle := range ommerHeaders {
		adapted_uncle, _ := adaptHeader(uncle)
		new_uncles = append(new_uncles, adapted_uncle)
	}
	return new_uncles
}
