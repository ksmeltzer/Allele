package strategy

import (
	"crypto/ecdsa"
	"crypto/rand"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/shopspring/decimal"

	"arbitrage/internal/dashboard"
	"arbitrage/internal/execution"
	"arbitrage/internal/market"
	"arbitrage/internal/storage"
)

// CompletenessArbitrage is a phase 0 smoke test strategy.
// It checks if buying all mutually exclusive outcomes (e.g., Yes + No)
// costs less than the guaranteed $1.00 payout.
type CompletenessArbitrage struct {
	clobManager     *market.CLOBManager
	execClient      *execution.Client
	privateKey      *ecdsa.PrivateKey
	makerAddress    common.Address
	marketOutcomes  map[string][]string
	takerFee        decimal.Decimal
	minProfitMargin decimal.Decimal
	broadcaster     *dashboard.Broadcaster
}

func NewCompletenessArbitrage(clobManager *market.CLOBManager, execClient *execution.Client, privateKey *ecdsa.PrivateKey, makerAddress common.Address, takerFee decimal.Decimal, minProfitMargin decimal.Decimal, broadcaster *dashboard.Broadcaster) *CompletenessArbitrage {
	return &CompletenessArbitrage{
		clobManager:     clobManager,
		execClient:      execClient,
		privateKey:      privateKey,
		makerAddress:    makerAddress,
		marketOutcomes:  make(map[string][]string),
		takerFee:        takerFee,
		minProfitMargin: minProfitMargin,
		broadcaster:     broadcaster,
	}
}

// RegisterMarket lets the strategy know which assets are mutually exclusive.
func (ca *CompletenessArbitrage) RegisterMarket(conditionID string, assetIDs []string) {
	ca.marketOutcomes[conditionID] = assetIDs
}

// Evaluate runs the arbitrage check across all registered markets.
func (ca *CompletenessArbitrage) Evaluate() {
	one := decimal.NewFromInt(1)
	for conditionID, assetIDs := range ca.marketOutcomes {
		totalCost := decimal.Zero
		minSize := decimal.NewFromInt(999999) // Arbitrary large number
		canTrade := true

		for _, assetID := range assetIDs {
			book := ca.clobManager.GetBook(assetID)
			_, _, bestAsk, askSize, err := book.TopOfBook()
			if err != nil {
				canTrade = false
				break // Cannot execute completeness arb if an outcome has no asks
			}

			totalCost = totalCost.Add(bestAsk)
			if askSize.LessThan(minSize) {
				minSize = askSize
			}
		}

		// Apply taker fee: totalCost * (1 + takerFee)
		totalCostWithFee := totalCost.Mul(one.Add(ca.takerFee))

		if canTrade && one.Sub(totalCostWithFee).GreaterThanOrEqual(ca.minProfitMargin) {
			// Found an opportunity!
			profitPerShare := one.Sub(totalCostWithFee)
			log.Printf("[ARBITRAGE] Condition: %s | Cost to Buy All (inc fees): %s | Guaranteed Profit/Share: %s | Max Size: %s",
				conditionID, totalCostWithFee.String(), profitPerShare.String(), minSize.String())

			profitFloat, _ := profitPerShare.Float64()
			if err := storage.LogTrade("completeness", conditionID, profitFloat); err != nil {
				log.Printf("Failed to log trade: %v", err)
			}
			if ca.broadcaster != nil {
				ca.broadcaster.Broadcast(map[string]interface{}{
					"type":      "ARBITRAGE",
					"condition": conditionID,
					"profit":    profitFloat,
					"timestamp": time.Now(),
				})
			}

			if ca.execClient == nil || ca.privateKey == nil {
				log.Println("[ARBITRAGE] Execution client or private key not set, skipping execution.")
				continue
			}

			for _, assetID := range assetIDs {
				tokenId := new(big.Int)
				tokenId.SetString(assetID, 10)

				salt, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 256))

				order := execution.Order{
					Salt:          salt,
					Maker:         ca.makerAddress,
					Signer:        ca.makerAddress,
					Taker:         common.HexToAddress("0x0000000000000000000000000000000000000000"),
					TokenId:       tokenId,
					MakerAmount:   big.NewInt(1000000), // 1 USDC = 1,000,000
					TakerAmount:   big.NewInt(1000000), // 1 Share = 1,000,000
					Expiration:    big.NewInt(time.Now().Add(2 * time.Second).Unix()),
					Nonce:         big.NewInt(0),
					FeeRateBps:    big.NewInt(0),
					Side:          0, // BUY
					SignatureType: 1, // EOA
				}

				sig, err := execution.SignOrder(order, ca.privateKey)
				if err != nil {
					log.Printf("[ARBITRAGE] Failed to sign order for asset %s: %v", assetID, err)
					continue
				}

				// standard Ethereum signature modification for V
				if len(sig) == 65 {
					sig[64] += 27
				}
				sigHex := hexutil.Encode(sig)

				err = ca.execClient.PlaceOrder(&order, sigHex)
				if err != nil {
					log.Printf("[ARBITRAGE] Failed to place order for asset %s: %v", assetID, err)
				} else {
					log.Printf("[ARBITRAGE] Successfully placed order for asset %s", assetID)
				}
			}
		}
	}
}
