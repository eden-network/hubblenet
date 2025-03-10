package orderbook

import (
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/ava-labs/subnet-evm/metrics"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

var positionType = SHORT
var userAddress = "0x22Bb736b64A0b4D4081E103f83bccF864F0404aa"
var trader = common.HexToAddress(userAddress)
var price = big.NewInt(20)
var status Status = Placed
var blockNumber = big.NewInt(2)

var market = Market(0)

func TestgetDatabase(t *testing.T) {
	inMemoryDatabase := getDatabase()
	assert.NotNil(t, inMemoryDatabase)
}

func TestAdd(t *testing.T) {
	baseAssetQuantity := big.NewInt(-10)
	inMemoryDatabase := getDatabase()
	salt := big.NewInt(time.Now().Unix())
	limitOrder := createLimitOrder(positionType, userAddress, baseAssetQuantity, price, status, blockNumber, salt)
	inMemoryDatabase.Add(&limitOrder)
	returnedOrder := inMemoryDatabase.OrderMap[limitOrder.Id]
	assert.Equal(t, limitOrder.PositionType, returnedOrder.PositionType)
	assert.Equal(t, limitOrder.UserAddress, returnedOrder.UserAddress)
	assert.Equal(t, limitOrder.BaseAssetQuantity, returnedOrder.BaseAssetQuantity)
	assert.Equal(t, limitOrder.Price, returnedOrder.Price)
	assert.Equal(t, limitOrder.getOrderStatus().Status, returnedOrder.getOrderStatus().Status)
	assert.Equal(t, limitOrder.BlockNumber, returnedOrder.BlockNumber)
}

func TestGetAllOrders(t *testing.T) {
	baseAssetQuantity := big.NewInt(-10)
	inMemoryDatabase := getDatabase()
	totalOrders := uint64(5)
	for i := uint64(0); i < totalOrders; i++ {
		salt := big.NewInt(0).Add(big.NewInt(int64(i)), big.NewInt(time.Now().Unix()))
		limitOrder := createLimitOrder(positionType, userAddress, baseAssetQuantity, price, status, blockNumber, salt)
		inMemoryDatabase.Add(&limitOrder)
	}
	returnedOrders := inMemoryDatabase.GetAllOrders()
	assert.Equal(t, totalOrders, uint64(len(returnedOrders)))
	for _, returnedOrder := range returnedOrders {
		assert.Equal(t, positionType, returnedOrder.PositionType)
		assert.Equal(t, userAddress, returnedOrder.UserAddress)
		assert.Equal(t, baseAssetQuantity, returnedOrder.BaseAssetQuantity)
		assert.Equal(t, price, returnedOrder.Price)
		assert.Equal(t, status, returnedOrder.getOrderStatus().Status)
		assert.Equal(t, blockNumber, returnedOrder.BlockNumber)
	}
}

func TestGetShortOrders(t *testing.T) {
	baseAssetQuantity := big.NewInt(0).Mul(big.NewInt(-3), _1e18)
	inMemoryDatabase := getDatabase()
	totalLongOrders := uint64(2)
	longOrderPrice := big.NewInt(0).Add(price, big.NewInt(1))
	longOrderBaseAssetQuantity := big.NewInt(0).Mul(big.NewInt(10), _1e18)
	for i := uint64(0); i < totalLongOrders; i++ {
		salt := big.NewInt(0).Add(big.NewInt(int64(i)), big.NewInt(time.Now().Unix()))
		limitOrder := createLimitOrder(LONG, userAddress, longOrderBaseAssetQuantity, longOrderPrice, status, blockNumber, salt)
		inMemoryDatabase.Add(&limitOrder)
	}
	//Short order with price 10 and blockNumber 2
	price1 := big.NewInt(10)
	blockNumber1 := big.NewInt(2)
	salt1 := big.NewInt(time.Now().Unix())
	shortOrder1 := createLimitOrder(SHORT, userAddress, baseAssetQuantity, price1, status, blockNumber1, salt1)
	inMemoryDatabase.Add(&shortOrder1)

	//Short order with price 9 and blockNumber 2
	price2 := big.NewInt(9)
	blockNumber2 := big.NewInt(2)
	salt2 := big.NewInt(0).Add(salt1, big.NewInt(1))
	shortOrder2 := createLimitOrder(SHORT, userAddress, baseAssetQuantity, price2, status, blockNumber2, salt2)
	inMemoryDatabase.Add(&shortOrder2)

	//Short order with price 9.01 and blockNumber 3
	price3 := big.NewInt(9)
	blockNumber3 := big.NewInt(3)
	salt3 := big.NewInt(0).Add(salt2, big.NewInt(1))
	shortOrder3 := createLimitOrder(SHORT, userAddress, baseAssetQuantity, price3, status, blockNumber3, salt3)
	inMemoryDatabase.Add(&shortOrder3)

	//Reduce only short order with price 9 and blockNumber 4
	price4 := big.NewInt(9)
	blockNumber4 := big.NewInt(4)
	salt4 := big.NewInt(0).Add(salt3, big.NewInt(1))
	shortOrder4 := createLimitOrder(SHORT, userAddress, baseAssetQuantity, price4, status, blockNumber4, salt4)
	shortOrder4.ReduceOnly = true
	inMemoryDatabase.Add(&shortOrder4)

	returnedShortOrders := inMemoryDatabase.GetShortOrders(market, nil, nil)
	assert.Equal(t, 3, len(returnedShortOrders))

	for _, returnedOrder := range returnedShortOrders {
		assert.Equal(t, SHORT, returnedOrder.PositionType)
		assert.Equal(t, userAddress, returnedOrder.UserAddress)
		assert.Equal(t, baseAssetQuantity, returnedOrder.BaseAssetQuantity)
		assert.Equal(t, status, returnedOrder.getOrderStatus().Status)
	}

	//Test returnedShortOrders are sorted by price lowest to highest first and then block number from lowest to highest
	assert.Equal(t, price2, returnedShortOrders[0].Price)
	assert.Equal(t, blockNumber2, returnedShortOrders[0].BlockNumber)
	assert.Equal(t, price3, returnedShortOrders[1].Price)
	assert.Equal(t, blockNumber3, returnedShortOrders[1].BlockNumber)
	assert.Equal(t, price1, returnedShortOrders[2].Price)
	assert.Equal(t, blockNumber1, returnedShortOrders[2].BlockNumber)

	// now test with one reduceOnly order when there's a long position

	size := big.NewInt(0).Mul(big.NewInt(2), _1e18)
	inMemoryDatabase.UpdatePosition(trader, market, size, big.NewInt(0).Mul(big.NewInt(100), _1e6), false)

	returnedShortOrders = inMemoryDatabase.GetShortOrders(market, nil, nil)
	assert.Equal(t, 4, len(returnedShortOrders))

	// at least one of the orders should be reduce only
	reduceOnlyOrder := Order{}
	for _, order := range returnedShortOrders {
		if order.ReduceOnly {
			reduceOnlyOrder = order
		}
	}
	assert.Equal(t, reduceOnlyOrder.Salt, salt4)
	assert.Equal(t, reduceOnlyOrder.BaseAssetQuantity, baseAssetQuantity)
	assert.Equal(t, reduceOnlyOrder.FilledBaseAssetQuantity, big.NewInt(0).Neg(_1e18))
}

func TestGetLongOrders(t *testing.T) {
	baseAssetQuantity := big.NewInt(-10)
	inMemoryDatabase := getDatabase()
	for i := uint64(0); i < 3; i++ {
		salt := big.NewInt(0).Add(big.NewInt(time.Now().Unix()), big.NewInt(int64(i)))
		limitOrder := createLimitOrder(SHORT, userAddress, baseAssetQuantity, price, status, blockNumber, salt)
		inMemoryDatabase.Add(&limitOrder)
	}

	//Long order with price 9 and blockNumber 2
	longOrderBaseAssetQuantity := big.NewInt(10)
	price1 := big.NewInt(9)
	blockNumber1 := big.NewInt(2)
	salt1 := big.NewInt(time.Now().Unix())
	longOrder1 := createLimitOrder(LONG, userAddress, longOrderBaseAssetQuantity, price1, status, blockNumber1, salt1)
	inMemoryDatabase.Add(&longOrder1)

	//long order with price 9 and blockNumber 3
	price2 := big.NewInt(9)
	blockNumber2 := big.NewInt(3)
	salt2 := big.NewInt(0).Add(salt1, big.NewInt(1))
	longOrder2 := createLimitOrder(LONG, userAddress, longOrderBaseAssetQuantity, price2, status, blockNumber2, salt2)
	inMemoryDatabase.Add(&longOrder2)

	//long order with price 10 and blockNumber 3
	price3 := big.NewInt(10)
	blockNumber3 := big.NewInt(3)
	salt3 := big.NewInt(0).Add(salt2, big.NewInt(1))
	longOrder3 := createLimitOrder(LONG, userAddress, longOrderBaseAssetQuantity, price3, status, blockNumber3, salt3)
	inMemoryDatabase.Add(&longOrder3)

	returnedLongOrders := inMemoryDatabase.GetLongOrders(market, nil, nil)
	assert.Equal(t, 3, len(returnedLongOrders))

	//Test returnedLongOrders are sorted by price highest to lowest first and then block number from lowest to highest
	assert.Equal(t, price3, returnedLongOrders[0].Price)
	assert.Equal(t, blockNumber3, returnedLongOrders[0].BlockNumber)
	assert.Equal(t, price1, returnedLongOrders[1].Price)
	assert.Equal(t, blockNumber1, returnedLongOrders[1].BlockNumber)
	assert.Equal(t, price2, returnedLongOrders[2].Price)
	assert.Equal(t, blockNumber2, returnedLongOrders[2].BlockNumber)

	for _, returnedOrder := range returnedLongOrders {
		assert.Equal(t, LONG, returnedOrder.PositionType)
		assert.Equal(t, userAddress, returnedOrder.UserAddress)
		assert.Equal(t, longOrderBaseAssetQuantity, returnedOrder.BaseAssetQuantity)
		assert.Equal(t, status, returnedOrder.getOrderStatus().Status)
	}
}

func TestGetCancellableOrders(t *testing.T) {
	// also tests getTotalNotionalPositionAndUnrealizedPnl
	inMemoryDatabase := getDatabase()
	getReservedMargin := func(order Order) *big.Int {
		notional := big.NewInt(0).Abs(big.NewInt(0).Div(big.NewInt(0).Mul(order.BaseAssetQuantity, order.Price), _1e18))
		return divideByBasePrecision(big.NewInt(0).Mul(notional, inMemoryDatabase.configService.getMinAllowableMargin()))
	}

	blockNumber1 := big.NewInt(2)
	baseAssetQuantity := big.NewInt(0).Mul(big.NewInt(-3), _1e18)

	salt1 := big.NewInt(101)
	price1 := multiplyBasePrecision(big.NewInt(10))
	shortOrder1 := createLimitOrder(SHORT, userAddress, baseAssetQuantity, price1, status, blockNumber1, salt1)

	salt2 := big.NewInt(102)
	price2 := multiplyBasePrecision(big.NewInt(9))
	shortOrder2 := createLimitOrder(SHORT, userAddress, baseAssetQuantity, price2, status, blockNumber1, salt2)

	salt3 := big.NewInt(103)
	price3 := multiplyBasePrecision(big.NewInt(8))
	shortOrder3 := createLimitOrder(SHORT, userAddress, baseAssetQuantity, price3, status, blockNumber1, salt3)

	depositMargin := multiplyBasePrecision(big.NewInt(40))
	inMemoryDatabase.UpdateMargin(trader, HUSD, depositMargin)

	// 3 different short orders with price = 10, 9, 8
	inMemoryDatabase.Add(&shortOrder1)
	inMemoryDatabase.UpdateReservedMargin(trader, getReservedMargin(shortOrder1))
	inMemoryDatabase.Add(&shortOrder2)
	inMemoryDatabase.UpdateReservedMargin(trader, getReservedMargin(shortOrder2))
	inMemoryDatabase.Add(&shortOrder3)
	inMemoryDatabase.UpdateReservedMargin(trader, getReservedMargin(shortOrder3))

	// 1 fulfilled order at price = 10, size = 9
	size := big.NewInt(0).Mul(big.NewInt(-9), _1e18)
	fulfilPrice := multiplyBasePrecision(big.NewInt(10))
	inMemoryDatabase.UpdatePosition(trader, market, size, dividePrecisionSize(new(big.Int).Mul(new(big.Int).Abs(size), fulfilPrice)), false)
	inMemoryDatabase.UpdateLastPrice(market, fulfilPrice)

	// price has moved from 10 to 11 now
	priceMap := map[Market]*big.Int{
		market: multiplyBasePrecision(big.NewInt(11)),
	}
	// Setup completed, assertions start here
	_trader := inMemoryDatabase.TraderMap[trader]
	assert.Equal(t, big.NewInt(0), getTotalFunding(_trader, []Market{market}))
	assert.Equal(t, depositMargin, getNormalisedMargin(_trader))

	// last price based notional = 9 * 10 = 90, pnl = 0, mf = (40-0)/90 = 0.44
	// oracle price based notional = 9 * 11 = 99, pnl = -9, mf = (40-9)/99 = 0.31
	// for Min_Allowable_Margin we select the min of 2 hence, oracle based mf
	notionalPosition, unrealizePnL := getTotalNotionalPositionAndUnrealizedPnl(_trader, depositMargin, Min_Allowable_Margin, priceMap, inMemoryDatabase.GetLastPrices(), []Market{market})
	assert.Equal(t, multiplyBasePrecision(big.NewInt(99)), notionalPosition)
	assert.Equal(t, multiplyBasePrecision(big.NewInt(-9)), unrealizePnL)

	// for Maintenance_Margin we select the max of 2 hence, last price based mf
	notionalPosition, unrealizePnL = getTotalNotionalPositionAndUnrealizedPnl(_trader, depositMargin, Maintenance_Margin, priceMap, inMemoryDatabase.GetLastPrices(), []Market{market})
	assert.Equal(t, multiplyBasePrecision(big.NewInt(90)), notionalPosition)
	assert.Equal(t, big.NewInt(0), unrealizePnL)

	marginFraction := calcMarginFraction(_trader, big.NewInt(0), priceMap, inMemoryDatabase.GetLastPrices(), []Market{market})
	assert.Equal(t, new(big.Int).Div(multiplyBasePrecision(depositMargin /* uPnL = 0 */), notionalPosition), marginFraction)

	availableMargin := getAvailableMargin(_trader, big.NewInt(0), priceMap, inMemoryDatabase.GetLastPrices(), inMemoryDatabase.configService.getMinAllowableMargin(), []Market{market})
	// availableMargin = 40 - 9 - (99 + (10+9+8) * 3)/5 = -5
	assert.Equal(t, multiplyBasePrecision(big.NewInt(-5)), availableMargin)
	_, ordersToCancel := inMemoryDatabase.GetNaughtyTraders(priceMap, []Market{market})

	// t.Log("####", "ordersToCancel", ordersToCancel)
	assert.Equal(t, 1, len(ordersToCancel)) // only one trader
	// orders will be cancelled in the order of price, hence orderId3, 2, 1
	// orderId3 will free up 8*3/5 = 4.8
	// orderId2 will free up 9*3/5 = 5.4
	assert.Equal(t, 2, len(ordersToCancel[trader])) // 2 orders
	assert.Equal(t, ordersToCancel[trader][0].Id, shortOrder3.Id)
	assert.Equal(t, ordersToCancel[trader][1].Id, shortOrder2.Id)
}

func TestUpdateFulfilledBaseAssetQuantityLimitOrder(t *testing.T) {
	baseAssetQuantity := big.NewInt(-10)
	t.Run("when order id does not exist", func(t *testing.T) {
		inMemoryDatabase := getDatabase()
		filledQuantity := big.NewInt(1)
		randomOrderID := common.BigToHash(big.NewInt(1))
		counter := metrics.GetOrRegisterCounter("update_filled_base_asset_quantity_order_id_not_found", nil)
		assert.Equal(t, counter.Count(), int64(0))

		inMemoryDatabase.UpdateFilledBaseAssetQuantity(filledQuantity, randomOrderID, 69)
		counter = metrics.GetOrRegisterCounter("update_filled_base_asset_quantity_order_id_not_found", nil)
		assert.Equal(t, counter.Count(), int64(1))
	})
	t.Run("when filled quantity is not equal to baseAssetQuantity", func(t *testing.T) {
		t.Run("When order type is short order", func(t *testing.T) {
			inMemoryDatabase := getDatabase()
			salt := big.NewInt(time.Now().Unix())
			limitOrder := createLimitOrder(positionType, userAddress, baseAssetQuantity, price, status, blockNumber, salt)
			inMemoryDatabase.Add(&limitOrder)

			filledQuantity := big.NewInt(2)

			inMemoryDatabase.UpdateFilledBaseAssetQuantity(filledQuantity, limitOrder.Id, 69)
			updatedLimitOrder := inMemoryDatabase.OrderMap[limitOrder.Id]

			assert.Equal(t, updatedLimitOrder.FilledBaseAssetQuantity, big.NewInt(0).Neg(filledQuantity))
			assert.Equal(t, updatedLimitOrder.FilledBaseAssetQuantity, filledQuantity.Mul(filledQuantity, big.NewInt(-1)))
		})
		t.Run("When order type is long order", func(t *testing.T) {
			inMemoryDatabase := getDatabase()
			positionType = LONG
			baseAssetQuantity = big.NewInt(10)
			salt := big.NewInt(time.Now().Unix())
			limitOrder := createLimitOrder(positionType, userAddress, baseAssetQuantity, price, status, blockNumber, salt)
			inMemoryDatabase.Add(&limitOrder)

			filledQuantity := big.NewInt(2)
			inMemoryDatabase.UpdateFilledBaseAssetQuantity(filledQuantity, limitOrder.Id, 69)
			updatedLimitOrder := inMemoryDatabase.OrderMap[limitOrder.Id]

			assert.Equal(t, updatedLimitOrder.FilledBaseAssetQuantity, filledQuantity)
		})
	})
	t.Run("when filled quantity is equal to baseAssetQuantity", func(t *testing.T) {
		t.Run("When order type is short order", func(t *testing.T) {
			inMemoryDatabase := getDatabase()
			salt := big.NewInt(time.Now().Unix())
			limitOrder := createLimitOrder(positionType, userAddress, baseAssetQuantity, price, status, blockNumber, salt)
			inMemoryDatabase.Add(&limitOrder)

			filledQuantity := big.NewInt(0).Abs(limitOrder.BaseAssetQuantity)
			inMemoryDatabase.UpdateFilledBaseAssetQuantity(filledQuantity, limitOrder.Id, 69)
			assert.Equal(t, int64(0), limitOrder.GetUnFilledBaseAssetQuantity().Int64())

			allOrders := inMemoryDatabase.GetAllOrders()
			assert.Equal(t, 1, len(allOrders))
			inMemoryDatabase.Accept(70, 70)
			allOrders = inMemoryDatabase.GetAllOrders()
			assert.Equal(t, 0, len(allOrders))
		})
		t.Run("When order type is long order", func(t *testing.T) {
			inMemoryDatabase := getDatabase()
			positionType = LONG
			baseAssetQuantity = big.NewInt(10)
			salt := big.NewInt(time.Now().Unix())
			limitOrder := createLimitOrder(positionType, userAddress, baseAssetQuantity, price, status, blockNumber, salt)
			inMemoryDatabase.Add(&limitOrder)

			filledQuantity := big.NewInt(0).Abs(limitOrder.BaseAssetQuantity)
			inMemoryDatabase.UpdateFilledBaseAssetQuantity(filledQuantity, limitOrder.Id, 420)

			assert.Equal(t, int64(0), limitOrder.GetUnFilledBaseAssetQuantity().Int64())

			allOrders := inMemoryDatabase.GetAllOrders()
			assert.Equal(t, 1, len(allOrders))
			inMemoryDatabase.Accept(420, 420)
			allOrders = inMemoryDatabase.GetAllOrders()
			assert.Equal(t, 0, len(allOrders))
		})
	})
}

func TestUpdatePosition(t *testing.T) {
	t.Run("When no positions exists for trader, it updates trader map with new positions", func(t *testing.T) {
		inMemoryDatabase := getDatabase()
		address := common.HexToAddress("0x22Bb736b64A0b4D4081E103f83bccF864F0404aa")
		var market Market = 1
		size := big.NewInt(20.00)
		openNotional := big.NewInt(200.00)
		inMemoryDatabase.UpdatePosition(address, market, size, openNotional, false)
		position := inMemoryDatabase.TraderMap[address].Positions[market]
		assert.Equal(t, size, position.Size)
		assert.Equal(t, openNotional, position.OpenNotional)
	})
	t.Run("When positions exists for trader, it overwrites old positions with new data", func(t *testing.T) {
		inMemoryDatabase := getDatabase()
		address := common.HexToAddress("0x22Bb736b64A0b4D4081E103f83bccF864F0404aa")
		var market Market = 1
		size := big.NewInt(20.00)
		openNotional := big.NewInt(200.00)
		inMemoryDatabase.UpdatePosition(address, market, size, openNotional, false)

		newSize := big.NewInt(25.00)
		newOpenNotional := big.NewInt(250.00)
		inMemoryDatabase.UpdatePosition(address, market, newSize, newOpenNotional, false)
		position := inMemoryDatabase.TraderMap[address].Positions[market]
		assert.Equal(t, newSize, position.Size)
		assert.Equal(t, newOpenNotional, position.OpenNotional)
	})
}

func TestUpdateMargin(t *testing.T) {
	t.Run("when adding margin for first time it updates margin in tradermap", func(t *testing.T) {
		inMemoryDatabase := getDatabase()
		address := common.HexToAddress("0x22Bb736b64A0b4D4081E103f83bccF864F0404aa")
		var collateral Collateral = 1
		amount := big.NewInt(20.00)
		inMemoryDatabase.UpdateMargin(address, collateral, amount)
		margin := inMemoryDatabase.TraderMap[address].Margin.Deposited[collateral]
		assert.Equal(t, amount, margin)
	})
	t.Run("When more margin is added, it updates margin in tradermap", func(t *testing.T) {
		inMemoryDatabase := getDatabase()
		address := common.HexToAddress("0x22Bb736b64A0b4D4081E103f83bccF864F0404aa")
		var collateral Collateral = 1
		amount := big.NewInt(20.00)
		inMemoryDatabase.UpdateMargin(address, collateral, amount)

		removedMargin := big.NewInt(15.00)
		inMemoryDatabase.UpdateMargin(address, collateral, removedMargin)
		margin := inMemoryDatabase.TraderMap[address].Margin.Deposited[collateral]
		assert.Equal(t, big.NewInt(0).Add(amount, removedMargin), margin)
	})
	t.Run("When margin is removed, it updates margin in tradermap", func(t *testing.T) {
		inMemoryDatabase := getDatabase()
		address := common.HexToAddress("0x22Bb736b64A0b4D4081E103f83bccF864F0404aa")
		var collateral Collateral = 1
		amount := big.NewInt(20.00)
		inMemoryDatabase.UpdateMargin(address, collateral, amount)

		removedMargin := big.NewInt(-15.00)
		inMemoryDatabase.UpdateMargin(address, collateral, removedMargin)
		margin := inMemoryDatabase.TraderMap[address].Margin.Deposited[collateral]
		assert.Equal(t, big.NewInt(0).Add(amount, removedMargin), margin)
	})
}

func TestAccept(t *testing.T) {
	t.Run("Order is fulfilled, should be deleted when block is accepted", func(t *testing.T) {
		inMemoryDatabase := getDatabase()
		orderId1 := addLimitOrder(inMemoryDatabase)
		orderId2 := addLimitOrder(inMemoryDatabase)

		err := inMemoryDatabase.SetOrderStatus(orderId1, FulFilled, "", 51)
		assert.Nil(t, err)
		assert.Equal(t, inMemoryDatabase.OrderMap[orderId1].getOrderStatus().Status, FulFilled)

		inMemoryDatabase.Accept(51, 51)

		// fulfilled order is deleted
		_, ok := inMemoryDatabase.OrderMap[orderId1]
		assert.False(t, ok)
		// unfulfilled order still exists
		_, ok = inMemoryDatabase.OrderMap[orderId2]
		assert.True(t, ok)
	})

	t.Run("Order is fulfilled, should be deleted when a future block is accepted", func(t *testing.T) {
		inMemoryDatabase := getDatabase()
		orderId := addLimitOrder(inMemoryDatabase)
		err := inMemoryDatabase.SetOrderStatus(orderId, FulFilled, "", 51)
		assert.Nil(t, err)
		assert.Equal(t, inMemoryDatabase.OrderMap[orderId].getOrderStatus().Status, FulFilled)

		inMemoryDatabase.Accept(52, 52)

		_, ok := inMemoryDatabase.OrderMap[orderId]
		assert.False(t, ok)
	})

	t.Run("Order is fulfilled, should not be deleted when a past block is accepted", func(t *testing.T) {
		inMemoryDatabase := getDatabase()
		orderId := addLimitOrder(inMemoryDatabase)
		err := inMemoryDatabase.SetOrderStatus(orderId, FulFilled, "", 51)
		assert.Nil(t, err)
		assert.Equal(t, inMemoryDatabase.OrderMap[orderId].getOrderStatus().Status, FulFilled)

		inMemoryDatabase.Accept(50, 50)

		_, ok := inMemoryDatabase.OrderMap[orderId]
		assert.True(t, ok)
	})

	t.Run("Order is placed, should not be deleted when a block is accepted", func(t *testing.T) {
		inMemoryDatabase := getDatabase()
		orderId := addLimitOrder(inMemoryDatabase)
		inMemoryDatabase.Accept(50, 50)

		_, ok := inMemoryDatabase.OrderMap[orderId]
		assert.True(t, ok)
	})
}

func TestRevertLastStatus(t *testing.T) {
	t.Run("revert status for order that doesn't exist - expect error", func(t *testing.T) {
		inMemoryDatabase := getDatabase()
		orderId := common.BytesToHash([]byte("order id"))
		err := inMemoryDatabase.RevertLastStatus(orderId)

		assert.Error(t, err)
	})

	t.Run("revert status for placed order", func(t *testing.T) {
		inMemoryDatabase := getDatabase()
		orderId := addLimitOrder(inMemoryDatabase)

		err := inMemoryDatabase.RevertLastStatus(orderId)
		assert.Nil(t, err)

		assert.Equal(t, len(inMemoryDatabase.OrderMap[orderId].LifecycleList), 0)
	})

	t.Run("revert status for fulfilled order", func(t *testing.T) {
		inMemoryDatabase := getDatabase()
		orderId := addLimitOrder(inMemoryDatabase)
		err := inMemoryDatabase.SetOrderStatus(orderId, FulFilled, "", 3)
		assert.Nil(t, err)

		err = inMemoryDatabase.RevertLastStatus(orderId)
		assert.Nil(t, err)

		assert.Equal(t, len(inMemoryDatabase.OrderMap[orderId].LifecycleList), 1)
		assert.Equal(t, inMemoryDatabase.OrderMap[orderId].LifecycleList[0].BlockNumber, uint64(2))
	})

	t.Run("revert status for accepted + fulfilled order - expect error", func(t *testing.T) {
		inMemoryDatabase := getDatabase()
		orderId := addLimitOrder(inMemoryDatabase)
		err := inMemoryDatabase.SetOrderStatus(orderId, FulFilled, "", 3)
		assert.Nil(t, err)

		inMemoryDatabase.Accept(3, 3)
		err = inMemoryDatabase.RevertLastStatus(orderId)
		assert.Error(t, err)
	})
}

func TestUpdateUnrealizedFunding(t *testing.T) {
	t.Run("When trader has no positions, it does not update anything", func(t *testing.T) {
		inMemoryDatabase := getDatabase()
		address := common.HexToAddress("0x22Bb736b64A0b4D4081E103f83bccF864F0404aa")
		var market Market = 1
		cumulativePremiumFraction := big.NewInt(2)
		trader := inMemoryDatabase.TraderMap[address]
		inMemoryDatabase.UpdateUnrealisedFunding(market, cumulativePremiumFraction)
		updatedTrader := inMemoryDatabase.TraderMap[address]
		assert.Equal(t, trader, updatedTrader)
	})
	t.Run("When trader has positions", func(t *testing.T) {
		t.Run("when unrealized funding is zero, it updates unrealized funding in trader's positions", func(t *testing.T) {
			inMemoryDatabase := getDatabase()
			addresses := [2]common.Address{common.HexToAddress("0x22Bb736b64A0b4D4081E103f83bccF864F0404aa"), common.HexToAddress("0x710bf5F942331874dcBC7783319123679033b63b")}
			var market Market = 1
			openNotional := big.NewInt(200.00)
			cumulativePremiumFraction := big.NewInt(0)
			for i, address := range addresses {
				iterator := i + 1
				size := big.NewInt(int64(20 * iterator))
				inMemoryDatabase.UpdatePosition(address, market, size, openNotional, false)
				inMemoryDatabase.ResetUnrealisedFunding(market, address, cumulativePremiumFraction)
			}
			newCumulativePremiumFraction := big.NewInt(5)
			inMemoryDatabase.UpdateUnrealisedFunding(market, newCumulativePremiumFraction)
			for _, address := range addresses {
				unrealizedFunding := inMemoryDatabase.TraderMap[address].Positions[market].UnrealisedFunding
				size := inMemoryDatabase.TraderMap[address].Positions[market].Size
				expectedUnrealizedFunding := big.NewInt(0).Div(big.NewInt(0).Mul(big.NewInt(0).Sub(newCumulativePremiumFraction, cumulativePremiumFraction), size), SIZE_BASE_PRECISION)
				assert.Equal(t, expectedUnrealizedFunding, unrealizedFunding)
			}
		})
		t.Run("when unrealized funding is not zero, it adds new funding to old unrealized funding in trader's positions", func(t *testing.T) {
			inMemoryDatabase := getDatabase()
			address := common.HexToAddress("0x22Bb736b64A0b4D4081E103f83bccF864F0404aa")
			var market Market = 1
			openNotional := big.NewInt(200.00)
			size := big.NewInt(20.00)
			inMemoryDatabase.UpdatePosition(address, market, size, openNotional, false)
			cumulativePremiumFraction := big.NewInt(2)
			inMemoryDatabase.ResetUnrealisedFunding(market, address, cumulativePremiumFraction)

			newCumulativePremiumFraction := big.NewInt(-1)
			inMemoryDatabase.UpdateUnrealisedFunding(market, newCumulativePremiumFraction)
			newUnrealizedFunding := inMemoryDatabase.TraderMap[address].Positions[market].UnrealisedFunding
			expectedUnrealizedFunding := big.NewInt(0).Div(big.NewInt(0).Mul(big.NewInt(0).Sub(newCumulativePremiumFraction, cumulativePremiumFraction), size), SIZE_BASE_PRECISION)
			assert.Equal(t, expectedUnrealizedFunding, newUnrealizedFunding)
		})
	})
}

func TestResetUnrealisedFunding(t *testing.T) {
	t.Run("When trader has no positions, it does not update anything", func(t *testing.T) {
		inMemoryDatabase := getDatabase()
		address := common.HexToAddress("0x22Bb736b64A0b4D4081E103f83bccF864F0404aa")
		var market Market = 1
		trader := inMemoryDatabase.TraderMap[address]
		cumulativePremiumFraction := big.NewInt(5)
		inMemoryDatabase.ResetUnrealisedFunding(market, address, cumulativePremiumFraction)
		updatedTrader := inMemoryDatabase.TraderMap[address]
		assert.Equal(t, trader, updatedTrader)
	})
	t.Run("When trader has positions, it resets unrealized funding to zero", func(t *testing.T) {
		inMemoryDatabase := getDatabase()
		address := common.HexToAddress("0x22Bb736b64A0b4D4081E103f83bccF864F0404aa")
		var market Market = 1
		openNotional := big.NewInt(200)
		size := big.NewInt(20)
		inMemoryDatabase.UpdatePosition(address, market, size, openNotional, false)
		cumulativePremiumFraction := big.NewInt(1)
		inMemoryDatabase.ResetUnrealisedFunding(market, address, cumulativePremiumFraction)
		unrealizedFundingFee := inMemoryDatabase.TraderMap[address].Positions[market].UnrealisedFunding
		assert.Equal(t, big.NewInt(0), unrealizedFundingFee)
	})
}

func TestUpdateNextFundingTime(t *testing.T) {
	inMemoryDatabase := getDatabase()
	nextFundingTime := uint64(time.Now().Unix())
	inMemoryDatabase.UpdateNextFundingTime(nextFundingTime)
	assert.Equal(t, nextFundingTime, inMemoryDatabase.NextFundingTime)
}

func TestGetNextFundingTime(t *testing.T) {
	t.Run("when funding time is not set", func(t *testing.T) {
		inMemoryDatabase := getDatabase()
		assert.Equal(t, uint64(0), inMemoryDatabase.GetNextFundingTime())
	})
	t.Run("when funding time is set", func(t *testing.T) {
		inMemoryDatabase := getDatabase()
		nextFundingTime := uint64(time.Now().Unix())
		inMemoryDatabase.UpdateNextFundingTime(nextFundingTime)
		assert.Equal(t, nextFundingTime, inMemoryDatabase.GetNextFundingTime())
	})
}

func TestUpdateLastPrice(t *testing.T) {
	inMemoryDatabase := getDatabase()
	var market Market = 1
	lastPrice := big.NewInt(20)
	inMemoryDatabase.UpdateLastPrice(market, lastPrice)
	assert.Equal(t, lastPrice, inMemoryDatabase.LastPrice[market])
}
func TestGetLastPrice(t *testing.T) {
	inMemoryDatabase := getDatabase()
	var market Market = 1
	lastPrice := big.NewInt(20)
	inMemoryDatabase.UpdateLastPrice(market, lastPrice)
	assert.Equal(t, lastPrice, inMemoryDatabase.GetLastPrice(market))
}

func TestUpdateReservedMargin(t *testing.T) {
	address := common.HexToAddress("0x22Bb736b64A0b4D4081E103f83bccF864F0404aa")
	amount := big.NewInt(20 * 1e6)
	inMemoryDatabase := getDatabase()
	inMemoryDatabase.UpdateReservedMargin(address, amount)
	assert.Equal(t, amount, inMemoryDatabase.TraderMap[address].Margin.Reserved)

	// subtract some amount
	amount = big.NewInt(-5 * 1e6)
	inMemoryDatabase.UpdateReservedMargin(address, amount)
	assert.Equal(t, big.NewInt(15*1e6), inMemoryDatabase.TraderMap[address].Margin.Reserved)
}

func createLimitOrder(positionType PositionType, userAddress string, baseAssetQuantity *big.Int, price *big.Int, status Status, blockNumber *big.Int, salt *big.Int) Order {
	lo := Order{
		Market:                  market,
		PositionType:            positionType,
		UserAddress:             userAddress,
		FilledBaseAssetQuantity: big.NewInt(0),
		BaseAssetQuantity:       baseAssetQuantity,
		Price:                   price,
		Salt:                    salt,
		BlockNumber:             blockNumber,
		ReduceOnly:              false,
	}
	lo.Id = getIdFromLimitOrder(lo)
	return lo
}

func TestGetUnfilledBaseAssetQuantity(t *testing.T) {
	t.Run("When limit FilledBaseAssetQuantity is zero, it returns BaseAssetQuantity", func(t *testing.T) {
		baseAssetQuantityLongOrder := big.NewInt(10)
		salt1 := big.NewInt(time.Now().Unix())
		longOrder := createLimitOrder(LONG, "0x22Bb736b64A0b4D4081E103f83bccF864F0404aa", baseAssetQuantityLongOrder, big.NewInt(21), Placed, big.NewInt(2), salt1)
		longOrder.FilledBaseAssetQuantity = big.NewInt(0)
		//baseAssetQuantityLongOrder - filledBaseAssetQuantity
		expectedUnFilledForLongOrder := big.NewInt(10)
		assert.Equal(t, expectedUnFilledForLongOrder, longOrder.GetUnFilledBaseAssetQuantity())

		baseAssetQuantityShortOrder := big.NewInt(-10)
		salt2 := big.NewInt(0).Add(salt1, big.NewInt(1))
		shortOrder := createLimitOrder(SHORT, "0x22Bb736b64A0b4D4081E103f83bccF864F0404aa", baseAssetQuantityShortOrder, big.NewInt(21), Placed, big.NewInt(2), salt2)
		shortOrder.FilledBaseAssetQuantity = big.NewInt(0)
		//baseAssetQuantityLongOrder - filledBaseAssetQuantity
		expectedUnFilledForShortOrder := big.NewInt(-10)
		assert.Equal(t, expectedUnFilledForShortOrder, shortOrder.GetUnFilledBaseAssetQuantity())
	})
	t.Run("When limit FilledBaseAssetQuantity is not zero, it returns BaseAssetQuantity - FilledBaseAssetQuantity", func(t *testing.T) {
		baseAssetQuantityLongOrder := big.NewInt(10)
		salt1 := big.NewInt(time.Now().Unix())
		longOrder := createLimitOrder(LONG, "0x22Bb736b64A0b4D4081E103f83bccF864F0404aa", baseAssetQuantityLongOrder, big.NewInt(21), Placed, big.NewInt(2), salt1)
		longOrder.FilledBaseAssetQuantity = big.NewInt(5)
		//baseAssetQuantityLongOrder - filledBaseAssetQuantity
		expectedUnFilledForLongOrder := big.NewInt(5)
		assert.Equal(t, expectedUnFilledForLongOrder, longOrder.GetUnFilledBaseAssetQuantity())

		baseAssetQuantityShortOrder := big.NewInt(-10)
		salt2 := big.NewInt(0).Add(salt1, big.NewInt(1))
		shortOrder := createLimitOrder(SHORT, "0x22Bb736b64A0b4D4081E103f83bccF864F0404aa", baseAssetQuantityShortOrder, big.NewInt(21), Placed, big.NewInt(2), salt2)
		shortOrder.FilledBaseAssetQuantity = big.NewInt(-5)
		//baseAssetQuantityLongOrder - filledBaseAssetQuantity
		expectedUnFilledForShortOrder := big.NewInt(-5)
		assert.Equal(t, expectedUnFilledForShortOrder, shortOrder.GetUnFilledBaseAssetQuantity())
	})
}

func addLimitOrder(db *InMemoryDatabase) common.Hash {
	salt := big.NewInt(time.Now().Unix() + int64(rand.Intn(200)))
	limitOrder := createLimitOrder(positionType, userAddress, big.NewInt(50), price, status, blockNumber, salt)
	db.Add(&limitOrder)
	return limitOrder.Id
}
