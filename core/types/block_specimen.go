package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

type StateSpecimen struct {
	AccountRead      []*accountRead
	StorageRead      []*storageRead
	CodeRead         []*codeRead
	BlockhashRead    []*blockhashRead
	BlockhashReadMap map[uint64]common.Hash `json:"-" rlp:"-"` // ignore in encoding/decoding
}

type accountRead struct {
	Address  common.Address
	Nonce    uint64
	Balance  *big.Int
	CodeHash common.Hash
}

type storageRead struct {
	Account common.Address
	SlotKey common.Hash
	Value   common.Hash
}

type codeRead struct {
	Hash common.Hash
	Code []byte
}

type blockhashRead struct {
	BlockNumber uint64
	BlockHash   common.Hash
}

func NewStateSpecimen() *StateSpecimen {
	sp := &StateSpecimen{
		BlockhashReadMap: make(map[uint64]common.Hash),
	}
	return sp
}

func (sp *StateSpecimen) Copy() *StateSpecimen {

	cpy := StateSpecimen{
		AccountRead:      make([]*accountRead, 0),
		StorageRead:      make([]*storageRead, 0),
		CodeRead:         make([]*codeRead, 0),
		BlockhashRead:    make([]*blockhashRead, 0),
		BlockhashReadMap: make(map[uint64]common.Hash),
	}

	return &cpy
}

func (sp *StateSpecimen) LogAccountRead(addr common.Address, nonce uint64, balance *big.Int, codeHashB []byte) *StateSpecimen {
	codeHash := common.BytesToHash(codeHashB)
	log.Trace("Retrieved committed account", "addr", addr, "nonce", nonce, "balance", balance, "codeHash", codeHash)

	sp.AccountRead = append(sp.AccountRead, &accountRead{
		Address:  addr,
		Nonce:    nonce,
		Balance:  balance,
		CodeHash: codeHash,
	})

	return sp
}

func (sp *StateSpecimen) LogStorageRead(account common.Address, slotKey common.Hash, value common.Hash) *StateSpecimen {
	log.Trace("Retrieved committed storage", "account", account, "slotKey", slotKey, "value", value)

	sp.StorageRead = append(sp.StorageRead, &storageRead{
		Account: account,
		SlotKey: slotKey,
		Value:   value,
	})

	return sp
}

func (sp *StateSpecimen) LogCodeRead(hashB []byte, code []byte) *StateSpecimen {
	hash := common.BytesToHash(hashB)
	log.Trace("Retrieved code", "hash", hash, "len", len(code))

	sp.CodeRead = append(sp.CodeRead, &codeRead{
		Hash: hash,
		Code: code,
	})

	return sp
}

func (sp *StateSpecimen) LogBlockhashRead(blockN uint64, blockHash common.Hash) *StateSpecimen {
	log.Trace("Retrieved BlockHash", "block_number", blockN, "hash", blockHash)

	if _, ok := sp.BlockhashReadMap[blockN]; !ok {
		sp.BlockhashReadMap[blockN] = blockHash
		sp.BlockhashRead = append(sp.BlockhashRead, &blockhashRead{
			BlockNumber: blockN,
			BlockHash:   blockHash,
		})
	}

	return sp
}
