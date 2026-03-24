# Global Equities Momentum

The **Global Equities Momentum** (GEM) strategy was developed by [Gary Antonacci](https://www.optimalmomentum.com/). It is based on his book *Dual Momentum Investing: An Innovative Strategy for Higher Returns with Less Risk* (McGraw-Hill, 2014) and his paper: [Risk Premia Harvesting Through Dual Momentum](https://papers.ssrn.com/sol3/papers.cfm?abstract_id=2042750). GEM applies both absolute and relative momentum to choose between US equities, international equities, and bonds. It allocates 100% of the portfolio to a single asset each month.

## Rules

The strategy uses three assets:

- **US Equities**: SPY (S&P 500)
- **International Equities**: VEU (FTSE All-World ex-US)
- **Bonds**: AGG (US Aggregate Bond)
- **T-Bills** (for absolute momentum comparison): BIL

1. On the last trading day of the month, compute the 12-month total return for SPY and BIL.
2. **Absolute momentum check**: If SPY's 12-month return is less than BIL's 12-month return, equities have negative excess momentum. Allocate 100% to AGG (bonds).
3. **Relative momentum check**: If SPY's 12-month return is greater than or equal to BIL's 12-month return, compare SPY vs VEU 12-month returns. Allocate 100% to whichever has the higher return.
4. Hold the position until the close of the following month.

The strategy generates approximately 1.4 trades per year on average.

## Assets Typically Held

| Ticker | Name                                                | Sector                              |
| ------ | --------------------------------------------------- | ----------------------------------- |
| SPY    | SPDR S&P 500 ETF                                    | Equity, U.S., Large Cap             |
| VEU    | Vanguard FTSE All-World ex-US ETF                   | Equity, International               |
| AGG    | iShares Core US Aggregate Bond ETF                  | Bond, U.S., Aggregate               |
| BIL    | SPDR Bloomberg 1-3 Month T-Bill ETF                 | Bond, U.S., Short-Term              |
