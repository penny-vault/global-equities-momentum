// Copyright 2021-2026
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	_ "embed"
	"fmt"
	"math"
	"time"

	"github.com/penny-vault/pvbt/asset"
	"github.com/penny-vault/pvbt/data"
	"github.com/penny-vault/pvbt/engine"
	"github.com/penny-vault/pvbt/portfolio"
	"github.com/penny-vault/pvbt/tradecron"
)

//go:embed README.md
var description string

// GlobalEquitiesMomentum implements Gary Antonacci's GEM strategy, which uses
// absolute momentum (vs T-bills) to decide whether to be in equities, and
// relative momentum to choose between US and international equities.
type GlobalEquitiesMomentum struct {
	USTicker            string `pvbt:"us-ticker" desc:"US equities ticker" default:"SPY" suggest:"GEM=SPY"`
	InternationalTicker string `pvbt:"intl-ticker" desc:"International equities ticker" default:"VEU" suggest:"GEM=VEU"`
	BondTicker          string `pvbt:"bond-ticker" desc:"Bond ticker for risk-off allocation" default:"AGG" suggest:"GEM=AGG"`
	TBillTicker         string `pvbt:"tbill-ticker" desc:"T-Bill ticker for absolute momentum comparison" default:"BIL" suggest:"GEM=BIL"`
}

func (s *GlobalEquitiesMomentum) Name() string {
	return "Global Equities Momentum"
}

func (s *GlobalEquitiesMomentum) Setup(eng *engine.Engine) {
	tc, err := tradecron.New("@monthend", tradecron.MarketHours{Open: 930, Close: 1600})
	if err != nil {
		panic(err)
	}

	eng.Schedule(tc)
	eng.SetBenchmark(eng.Asset("VFINX"))
	eng.RiskFreeAsset(eng.Asset("DGS3MO"))
}

func (s *GlobalEquitiesMomentum) Describe() engine.StrategyDescription {
	return engine.StrategyDescription{
		ShortCode:   "gem",
		Description: description,
		Source:      "https://papers.ssrn.com/sol3/papers.cfm?abstract_id=2042750",
		Version:     "1.0.0",
		VersionDate: time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC),
	}
}

func (s *GlobalEquitiesMomentum) Compute(ctx context.Context, eng *engine.Engine, strategyPortfolio portfolio.Portfolio) error {
	// 1. Build a universe of US equities and T-bills for the absolute momentum check.
	usAsset := eng.Asset(s.USTicker)
	intlAsset := eng.Asset(s.InternationalTicker)
	bondAsset := eng.Asset(s.BondTicker)
	tbillAsset := eng.Asset(s.TBillTicker)

	allUniverse := eng.Universe(usAsset, intlAsset, tbillAsset)

	priceDF, err := allUniverse.Window(ctx, portfolio.Months(12), data.MetricClose)
	if err != nil {
		return fmt.Errorf("failed to fetch prices: %w", err)
	}

	// 2. Downsample to monthly frequency.
	monthly := priceDF.Downsample(data.Monthly).Last()

	// Need at least 13 rows for Pct(12) to produce a valid value.
	if monthly.Len() < 13 {
		return nil
	}

	// 3. Compute 12-month returns.
	returns := monthly.Pct(12)
	returns = returns.Drop(math.NaN()).Last()

	if returns.Len() == 0 {
		return nil
	}

	usReturn := returns.Value(usAsset, data.MetricClose)
	intlReturn := returns.Value(intlAsset, data.MetricClose)
	tbillReturn := returns.Value(tbillAsset, data.MetricClose)

	ts := eng.CurrentDate().Unix()

	strategyPortfolio.Annotate(ts, "us-return-12m", fmt.Sprintf("%.4f", usReturn))
	strategyPortfolio.Annotate(ts, "intl-return-12m", fmt.Sprintf("%.4f", intlReturn))
	strategyPortfolio.Annotate(ts, "tbill-return-12m", fmt.Sprintf("%.4f", tbillReturn))

	// 4. Decision logic:
	//    - If US equities 12m return < T-bill 12m return -> bonds (absolute momentum)
	//    - Else pick US vs International by relative momentum (higher 12m return)
	var selectedAsset asset.Asset

	var justification string

	if usReturn < tbillReturn {
		// Absolute momentum is negative: equities underperforming T-bills.
		selectedAsset = bondAsset
		justification = fmt.Sprintf("absolute momentum negative: US %.2f%% < T-bill %.2f%%, 100%% %s",
			usReturn*100, tbillReturn*100, s.BondTicker)
	} else if usReturn >= intlReturn {
		// Relative momentum favors US.
		selectedAsset = usAsset
		justification = fmt.Sprintf("relative momentum: US %.2f%% >= Intl %.2f%%, 100%% %s",
			usReturn*100, intlReturn*100, s.USTicker)
	} else {
		// Relative momentum favors international.
		selectedAsset = intlAsset
		justification = fmt.Sprintf("relative momentum: Intl %.2f%% > US %.2f%%, 100%% %s",
			intlReturn*100, usReturn*100, s.InternationalTicker)
	}

	strategyPortfolio.Annotate(ts, "justification", justification)

	allocation := portfolio.Allocation{
		Date:          eng.CurrentDate(),
		Members:       map[asset.Asset]float64{selectedAsset: 1.0},
		Justification: justification,
	}

	if err := strategyPortfolio.RebalanceTo(ctx, allocation); err != nil {
		return fmt.Errorf("rebalance failed: %w", err)
	}

	return nil
}
