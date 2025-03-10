package bibliophile

import (
	"math/big"

	"github.com/ava-labs/subnet-evm/precompile/contract"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	IOC_ORDERBOOK_ADDRESS         = "0x635c5F96989a4226953FE6361f12B96c5d50289b"
	IOC_ORDER_INFO_SLOT     int64 = 53
	IOC_EXPIRATION_CAP_SLOT int64 = 54
)

// State Reader
func iocGetBlockPlaced(stateDB contract.StateDB, orderHash [32]byte) *big.Int {
	orderInfo := iocOrderInfoMappingStorageSlot(orderHash)
	return new(big.Int).SetBytes(stateDB.GetState(common.HexToAddress(IOC_ORDERBOOK_ADDRESS), common.BigToHash(orderInfo)).Bytes())
}

func iocGetOrderFilledAmount(stateDB contract.StateDB, orderHash [32]byte) *big.Int {
	orderInfo := iocOrderInfoMappingStorageSlot(orderHash)
	return new(big.Int).SetBytes(stateDB.GetState(common.HexToAddress(IOC_ORDERBOOK_ADDRESS), common.BigToHash(new(big.Int).Add(orderInfo, big.NewInt(1)))).Bytes())
}

func iocGetOrderStatus(stateDB contract.StateDB, orderHash [32]byte) int64 {
	orderInfo := iocOrderInfoMappingStorageSlot(orderHash)
	return new(big.Int).SetBytes(stateDB.GetState(common.HexToAddress(IOC_ORDERBOOK_ADDRESS), common.BigToHash(new(big.Int).Add(orderInfo, big.NewInt(2)))).Bytes()).Int64()
}

func iocOrderInfoMappingStorageSlot(orderHash [32]byte) *big.Int {
	return new(big.Int).SetBytes(crypto.Keccak256(append(orderHash[:], common.LeftPadBytes(big.NewInt(IOC_ORDER_INFO_SLOT).Bytes(), 32)...)))
}

func iocGetExpirationCap(stateDB contract.StateDB) *big.Int {
	return new(big.Int).SetBytes(stateDB.GetState(common.HexToAddress(IOC_ORDERBOOK_ADDRESS), common.BigToHash(big.NewInt(IOC_EXPIRATION_CAP_SLOT))).Bytes())
}
