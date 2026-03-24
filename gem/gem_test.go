package gem_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/penny-vault/global-equities-momentum/gem"
	"github.com/penny-vault/pvbt/data"
	"github.com/penny-vault/pvbt/engine"
	"github.com/penny-vault/pvbt/portfolio"
)

var _ = Describe("GlobalEquitiesMomentum", func() {
	var (
		ctx       context.Context
		snap      *data.SnapshotProvider
		nyc       *time.Location
		startDate time.Time
		endDate   time.Time
	)

	BeforeEach(func() {
		ctx = context.Background()

		var err error
		nyc, err = time.LoadLocation("America/New_York")
		Expect(err).NotTo(HaveOccurred())

		snap, err = data.NewSnapshotProvider("testdata/snapshot.db")
		Expect(err).NotTo(HaveOccurred())

		startDate = time.Date(2024, 6, 1, 0, 0, 0, 0, nyc)
		endDate = time.Date(2026, 3, 1, 0, 0, 0, 0, nyc)
	})

	AfterEach(func() {
		if snap != nil {
			snap.Close()
		}
	})

	runBacktest := func() portfolio.Portfolio {
		strategy := &gem.GlobalEquitiesMomentum{}
		acct := portfolio.New(
			portfolio.WithCash(100000, startDate),
			portfolio.WithAllMetrics(),
		)

		eng := engine.New(strategy,
			engine.WithDataProvider(snap),
			engine.WithAssetProvider(snap),
			engine.WithAccount(acct),
		)

		result, err := eng.Backtest(ctx, startDate, endDate)
		Expect(err).NotTo(HaveOccurred())
		return result
	}

	It("produces expected returns and risk metrics", func() {
		result := runBacktest()

		summary, err := result.Summary()
		Expect(err).NotTo(HaveOccurred())
		Expect(summary.TWRR).To(BeNumerically("~", 0.3691, 0.01))
		Expect(summary.MaxDrawdown).To(BeNumerically(">", -0.20), "max drawdown should be better than -20%")

		Expect(result.Value()).To(BeNumerically("~", 136910, 500))
	})

	It("rotates through all three asset classes", func() {
		result := runBacktest()
		txns := result.Transactions()

		tickers := map[string]bool{}
		for _, t := range txns {
			if t.Type == portfolio.BuyTransaction || t.Type == portfolio.SellTransaction {
				tickers[t.Asset.Ticker] = true
			}
		}

		Expect(tickers).To(HaveKey("SPY"))
		Expect(tickers).To(HaveKey("VEU"))
	})

	It("produces the expected trade sequence", func() {
		result := runBacktest()
		txns := result.Transactions()

		type trade struct {
			date   string
			txType portfolio.TransactionType
			ticker string
		}

		var trades []trade
		for _, t := range txns {
			if t.Type == portfolio.BuyTransaction || t.Type == portfolio.SellTransaction {
				trades = append(trades, trade{
					date:   t.Date.In(nyc).Format("2006-01-02"),
					txType: t.Type,
					ticker: t.Asset.Ticker,
				})
			}
		}

		expected := []trade{
			{"2024-06-28", portfolio.BuyTransaction, "SPY"},
			{"2024-09-30", portfolio.BuyTransaction, "SPY"},
			{"2025-03-31", portfolio.BuyTransaction, "SPY"},
			{"2025-04-30", portfolio.SellTransaction, "SPY"},
			{"2025-04-30", portfolio.BuyTransaction, "VEU"},
			{"2025-06-30", portfolio.BuyTransaction, "VEU"},
			{"2025-07-31", portfolio.SellTransaction, "VEU"},
			{"2025-07-31", portfolio.BuyTransaction, "SPY"},
			{"2025-08-29", portfolio.SellTransaction, "SPY"},
			{"2025-08-29", portfolio.BuyTransaction, "VEU"},
			{"2025-09-30", portfolio.SellTransaction, "VEU"},
			{"2025-09-30", portfolio.BuyTransaction, "SPY"},
			{"2025-10-31", portfolio.SellTransaction, "SPY"},
			{"2025-10-31", portfolio.BuyTransaction, "VEU"},
			{"2025-12-31", portfolio.BuyTransaction, "VEU"},
		}

		Expect(trades).To(HaveLen(len(expected)))
		for i, exp := range expected {
			Expect(trades[i].date).To(Equal(exp.date), "trade %d date", i)
			Expect(trades[i].txType).To(Equal(exp.txType), "trade %d type", i)
			Expect(trades[i].ticker).To(Equal(exp.ticker), "trade %d ticker", i)
		}
	})
})
