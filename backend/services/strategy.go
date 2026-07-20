package services

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"stock-app/models"
)

type StrategyService struct {
	mu         sync.RWMutex
	strategies []models.Strategy
	stockSvc   *StockService
	storePath  string
}

func NewStrategyService(stockSvc *StockService) *StrategyService {
	s := &StrategyService{
		stockSvc:  stockSvc,
		storePath: filepath.Join(os.TempDir(), "strategies.json"),
	}
	s.loadDefaults()
	s.loadFromDisk()
	return s
}

// ── CRUD ──

func (s *StrategyService) List() []models.Strategy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]models.Strategy, len(s.strategies))
	copy(result, s.strategies)
	return result
}

func (s *StrategyService) Get(id string) *models.Strategy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := range s.strategies {
		if s.strategies[i].ID == id {
			return &s.strategies[i]
		}
	}
	return nil
}

func (s *StrategyService) Create(strategy models.Strategy) (*models.Strategy, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isValidWeight(strategy.Indicators) {
		return nil, fmt.Errorf("指标权重总和必须为100%")
	}

	strategy.ID = fmt.Sprintf("str_%d", time.Now().UnixNano())
	strategy.IsSystem = false
	strategy.CreatedAt = time.Now().Format("2006-01-02 15:04:05")
	strategy.UpdatedAt = strategy.CreatedAt
	s.strategies = append(s.strategies, strategy)
	s.saveToDisk()
	return &s.strategies[len(s.strategies)-1], nil
}

func (s *StrategyService) Update(id string, updated models.Strategy) (*models.Strategy, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isValidWeight(updated.Indicators) {
		return nil, fmt.Errorf("指标权重总和必须为100%")
	}

	for i := range s.strategies {
		if s.strategies[i].ID == id {
			if s.strategies[i].IsSystem {
				return nil, fmt.Errorf("系统预置策略不可编辑")
			}
			updated.ID = id
			updated.IsSystem = s.strategies[i].IsSystem
			updated.CreatedAt = s.strategies[i].CreatedAt
			updated.UpdatedAt = time.Now().Format("2006-01-02 15:04:05")
			// Preserve default flag or set it
			if updated.IsDefault {
				s.clearOtherDefaults()
			}
			s.strategies[i] = updated
			s.saveToDisk()
			return &s.strategies[i], nil
		}
	}
	return nil, fmt.Errorf("策略不存在")
}

func (s *StrategyService) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.strategies {
		if s.strategies[i].ID == id {
			if s.strategies[i].IsSystem {
				return fmt.Errorf("系统预置策略不可删除")
			}
			s.strategies = append(s.strategies[:i], s.strategies[i+1:]...)
			s.saveToDisk()
			return nil
		}
	}
	return fmt.Errorf("策略不存在")
}

func (s *StrategyService) SetDefault(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clearOtherDefaults()
	for i := range s.strategies {
		if s.strategies[i].ID == id {
			s.strategies[i].IsDefault = true
			s.saveToDisk()
			return nil
		}
	}
	return fmt.Errorf("策略不存在")
}

func (s *StrategyService) GetDefault() *models.Strategy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := range s.strategies {
		if s.strategies[i].IsDefault {
			return &s.strategies[i]
		}
	}
	if len(s.strategies) > 0 {
		return &s.strategies[0]
	}
	return nil
}

func (s *StrategyService) isValidWeight(indicators []models.StrategyIndicator) bool {
	total := 0.0
	for _, ind := range indicators {
		total += ind.Weight
	}
	return math.Abs(total-100) < 0.01
}

func (s *StrategyService) clearOtherDefaults() {
	for i := range s.strategies {
		s.strategies[i].IsDefault = false
	}
}

func (s *StrategyService) Copy(id string) (*models.Strategy, error) {
	original := s.Get(id)
	if original == nil {
		return nil, fmt.Errorf("策略不存在")
	}
	copy := *original
	copy.ID = fmt.Sprintf("str_%d", time.Now().UnixNano())
	copy.Name = original.Name + " (副本)"
	copy.IsSystem = false
	copy.IsDefault = false
	copy.CreatedAt = time.Now().Format("2006-01-02 15:04:05")
	copy.UpdatedAt = copy.CreatedAt

	s.mu.Lock()
	s.strategies = append(s.strategies, copy)
	s.saveToDisk()
	s.mu.Unlock()
	return &copy, nil
}

// ── Evaluation ──

// Evaluate 对全市场股票执行策略评估
func (s *StrategyService) Evaluate(strategy models.Strategy, page, pageSize int) (*models.StrategyEvalResponse, error) {
	start := time.Now()

	// 获取全市场K线数据和行情
	klinesMap, quotes := s.stockSvc.GetAllKLineData()
	if len(quotes) == 0 {
		// 缓存为空时直接从新浪API拉取实时行情做预览
		log.Printf("[INFO] Evaluate: cache empty, fetching live quotes from Sina...")
		// 用常见A股代码段（缩减范围，仅主要板块前几千支）
		allCodes := generateAllAShareCodes()
		if len(allCodes) > 3000 {
			allCodes = allCodes[:3000]
		}
		liveQuotes := s.stockSvc.fetchBatchQuotesFromPool(allCodes)
		if len(liveQuotes) > 300 {
			liveQuotes = liveQuotes[:300]
		}
		klinesMap = make(map[string][]models.KLineData)
		// 并发获取 K 线
		var mu sync.Mutex
		var wg sync.WaitGroup
		sem := make(chan struct{}, 10)
		for _, q := range liveQuotes {
			code := q.Code
			code = strings.TrimPrefix(code, "sh")
			code = strings.TrimPrefix(code, "sz")
			code = strings.TrimPrefix(code, "SH")
			code = strings.TrimPrefix(code, "SZ")
			code = strings.TrimPrefix(code, "s_sh")
			code = strings.TrimPrefix(code, "s_sz")
			if code == "" || q.Price <= 0 || strings.Contains(q.Name, "ST") || strings.Contains(q.Name, "st") || strings.Contains(q.Name, "退") {
				continue
			}
			wg.Add(1)
			go func(c string, quote models.SinaQuote) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()
				klines := s.stockSvc.fetchKLineDataWithCache(c, 60)
				if len(klines) >= 20 {
					mu.Lock()
					klinesMap[c] = klines
					quotes = append(quotes, quote)
					mu.Unlock()
				}
			}(code, q)
		}
		wg.Wait()
		log.Printf("[INFO] Evaluate: fetched %d live stocks with K-lines in %v", len(quotes), time.Since(start))
	}

	if len(quotes) == 0 {
		return nil, fmt.Errorf("没有可用的行情数据")
	}

	type evalTask struct {
		code   string
		quote  models.SinaQuote
		klines []models.KLineData
	}

	var tasks []evalTask
	for _, q := range quotes {
		if q.Price <= 0 {
			continue
		}
		code := q.Code
		// Normalize code (remove prefix)
		code = strings.TrimPrefix(code, "sh")
		code = strings.TrimPrefix(code, "sz")
		code = strings.TrimPrefix(code, "SH")
		code = strings.TrimPrefix(code, "SZ")
		code = strings.TrimPrefix(code, "s_sh")
		code = strings.TrimPrefix(code, "s_sz")

		klines, ok := klinesMap[code]
		if !ok || len(klines) < 20 {
			continue
		}
		tasks = append(tasks, evalTask{code: code, quote: q, klines: klines})
	}

	// Concurrent indicator computation
	type stockResult struct {
		code    string
		results map[string]float64
		boolRes map[string]bool
		name    string
		price   float64
		changePct float64
	}

	resultCh := make(chan stockResult, len(tasks))
	sem := make(chan struct{}, 20) // 20 concurrent

	for _, task := range tasks {
		go func(t evalTask) {
			sem <- struct{}{}
			defer func() { <-sem }()
			
			name := t.quote.Name
			changePct := t.quote.ChangePercent
			price := t.quote.Price
			results := make(map[string]float64)
			boolRes := make(map[string]bool)

			// Compute all indicators
			computeIndicators(t.klines, results, boolRes)

			resultCh <- stockResult{code: t.code, name: name, price: price, changePct: changePct, results: results, boolRes: boolRes}
		}(task)
	}

	// Collect results
	rawResults := make(map[string]stockResult)
	for i := 0; i < len(tasks); i++ {
		r := <-resultCh
		rawResults[r.code] = r
	}

	// Build eval items per stock
	type evalItem struct {
		Code   string
		Name   string
		Price  float64
		ChangePct float64
		Indicators []models.IndicatorEvalItem
		TotalScore float64
	}

	var evalItems []evalItem

	for code, r := range rawResults {
		var items []models.IndicatorEvalItem
		totalWeight := 0.0
		weightedScore := 0.0

		for _, cfg := range strategy.Indicators {
			def := models.GetIndicatorByID(cfg.IndicatorID)
			if def == nil {
				continue
			}

			item := models.IndicatorEvalItem{
				IndicatorID: cfg.IndicatorID,
				Name:        def.Name,
				Weight:      cfg.Weight,
				IsBool:      def.IsBool,
			}

			// Get raw value
			if def.IsBool {
				val, ok := r.boolRes[cfg.IndicatorID]
				if !ok {
					item.InRange = false
					item.Score = 0
					items = append(items, item)
					continue
				}
				item.BoolValue = val
				item.RawValue = 0
				if val {
					item.RawValue = 1
				}

				// Filter check for bool
				item.InRange = true
				if cfg.MinVal != "" {
					minV := parseFloat(cfg.MinVal)
					if val != (minV > 0) {
						item.InRange = false
					}
				}
				if cfg.MaxVal != "" {
					maxV := parseFloat(cfg.MaxVal)
					if val != (maxV > 0) {
						item.InRange = false
					}
				}

				if item.InRange {
					if val {
						item.Score = 100
					} else {
						item.Score = 0
					}
				}
			} else {
				val, ok := r.results[cfg.IndicatorID]
				if !ok {
					item.InRange = false
					item.Score = 0
					items = append(items, item)
					continue
				}
				item.RawValue = val

				// Filter check
				item.InRange = true
				if cfg.MinVal != "" {
					minV := parseFloat(cfg.MinVal)
					if val < minV {
						item.InRange = false
					}
				}
				if cfg.MaxVal != "" {
					maxV := parseFloat(cfg.MaxVal)
					if val > maxV {
						item.InRange = false
					}
				}
			}

			if item.InRange && !def.IsBool {
				// Will compute percentile later
				items = append(items, item)
			} else {
				// Score is already set for bool or filtered-out items
				if item.InRange {
					// Bool with score set
				} else {
					item.Score = 0
				}
				items = append(items, item)
			}

			if item.InRange && def.IsBool {
				weightedScore += item.Score * cfg.Weight / 100
				totalWeight += cfg.Weight
			}
		}

		// Compute percentile for non-bool indicators that passed filter
		for i := range items {
			item := &items[i]
			if !item.InRange || item.IsBool {
				continue
			}
			
			// Gather all valid values for this indicator
			var vals []float64
			for _, other := range rawResults {
				if v, ok := other.results[item.IndicatorID]; ok {
					// Check if this stock's value passes filter
					vals = append(vals, v)
				}
			}

			if len(vals) == 0 {
				item.Score = 0
				continue
			}

			// Find the percentile rank
			sort.Float64s(vals)
			pos := sort.SearchFloat64s(vals, item.RawValue)
			percentile := float64(pos) / float64(len(vals)) * 100

			// Direction
			cfg := findIndicatorCfg(strategy.Indicators, item.IndicatorID)
			if cfg != nil && cfg.Direction == "down" {
				percentile = 100 - percentile
			}

			// Clamp
			if percentile < 0 {
				percentile = 0
			}
			if percentile > 100 {
				percentile = 100
			}

			item.Percentile = percentile
			item.Score = percentile
			weightedScore += item.Score * item.Weight / 100
			totalWeight += item.Weight
		}

		// Normalize total score
		totalScore := 0.0
		if totalWeight > 0 {
			totalScore = weightedScore / totalWeight * 100
		}
		if totalScore > 100 {
			totalScore = 100
		}

		evalItems = append(evalItems, evalItem{
			Code:       code,
			Name:       r.name,
			Price:      r.price,
			ChangePct:  r.changePct,
			Indicators: items,
			TotalScore: math.Round(totalScore*10) / 10,
		})
	}

	// Sort by total score desc
	sort.Slice(evalItems, func(i, j int) bool {
		return evalItems[i].TotalScore > evalItems[j].TotalScore
	})

	// Paginate
	total := len(evalItems)
	startIdx := (page - 1) * pageSize
	if startIdx < 0 {
		startIdx = 0
	}
	endIdx := startIdx + pageSize
	if endIdx > total {
		endIdx = total
	}
	if startIdx > total {
		startIdx = total
	}

	paged := evalItems[startIdx:endIdx]
	results := make([]models.StockEvalResult, len(paged))
	for i, item := range paged {
		results[i] = models.StockEvalResult{
			Code:        item.Code,
			Name:        item.Name,
			Price:       item.Price,
			ChangePct:   item.ChangePct,
			TotalScore:  item.TotalScore,
			Indicators:  item.Indicators,
			Rank:        startIdx + i + 1,
		}
	}

	elapsed := time.Since(start).Milliseconds()
	return &models.StrategyEvalResponse{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Items:    results,
		CostMs:   elapsed,
	}, nil
}

// ── Indicator Computation ──

func computeIndicators(klines []models.KLineData, results map[string]float64, boolRes map[string]bool) {
	if len(klines) < 20 {
		return
	}

	n := len(klines)
	closes := make([]float64, n)
	highs := make([]float64, n)
	lows := make([]float64, n)
	volumes := make([]float64, n)
	opens := make([]float64, n)
	for i, k := range klines {
		closes[i] = k.Close
		highs[i] = k.High
		lows[i] = k.Low
		volumes[i] = k.Volume
		opens[i] = k.Open
	}
	last := closes[n-1]

	// ── MACD DIF-DEA ──
	if n >= 26 {
		ema12 := ema(closes, 12)
		ema26 := ema(closes, 26)
		dif := ema12[n-1] - ema26[n-1]
		dea := ema(append([]float64{}, difSeries(closes, 12, 26)...), 9)
		if len(dea) > 0 {
			results["macd_diff"] = dif - dea[len(dea)-1]
		}
	}

	// ── MA Alignment (5,10,20) ──
	ma5 := sma(closes, 5)
	ma10 := sma(closes, 10)
	ma20 := sma(closes, 20)
	if len(ma5) > 0 && len(ma10) > 0 && len(ma20) > 0 {
		boolRes["ma_alignment"] = ma5[len(ma5)-1] > ma10[len(ma10)-1] && ma10[len(ma10)-1] > ma20[len(ma20)-1]
	}

	// ── EMA Deviation ──
	if len(ma20) > 0 && ma20[len(ma20)-1] > 0 {
		results["ema_deviation"] = (last - ma20[len(ma20)-1]) / ma20[len(ma20)-1] * 100
	}

	// ── RSI(14) ──
	if n >= 15 {
		results["rsi"] = rsi(closes, 14)
	}

	// ── KDJ J ──
	if n >= 9 {
		k, d := kdj(highs, lows, closes, 9, 3, 3)
		results["kdj_j"] = 3*k - 2*d
	}

	// ── CCI(14) ──
	if n >= 15 {
		results["cci"] = cci(highs, lows, closes, 14)
	}

	// ── WR(14) ──
	if n >= 14 {
		results["wr"] = williamsR(highs, lows, closes, 14)
	}

	// ── MOM(10) ──
	if n >= 11 {
		results["mom"] = last - closes[n-11]
	}

	// ── Volume Ratio ──
	vol5 := sma(volumes, 5)
	if len(vol5) > 0 && vol5[len(vol5)-1] > 0 {
		results["volume_ratio"] = volumes[n-1] / vol5[len(vol5)-1]
	}

	// ── BOLL Position ──
	upper, lower := bollinger(closes, 20, 2)
	if upper-lower > 0 {
		results["boll_position"] = (last - lower) / (upper - lower)
	}

	// ── ATR(14) ──
	if n >= 14 {
		results["atr"] = atr(highs, lows, closes, 14)
	}

	// ── SAR ──
	boolRes["sar"] = sar(highs, lows, 0.02, 0.2)

	// ── 5d Change ──
	if n >= 6 {
		results["pct_5d"] = (last - closes[n-6]) / closes[n-6] * 100
	}

	// ── New 20d High ──
	if n >= 20 {
		max20 := closes[n-20]
		for i := n - 20; i < n-1; i++ {
			if closes[i] > max20 {
				max20 = closes[i]
			}
		}
		boolRes["new_20d_high"] = last > max20
	}

	// ── BIAS(6) ──
	if n >= 6 {
		ma6 := sma(closes, 6)
		if len(ma6) > 0 && ma6[len(ma6)-1] > 0 {
			results["bias"] = (last - ma6[len(ma6)-1]) / ma6[len(ma6)-1] * 100
		}
	}

	// ── PSY(12) ──
	if n >= 12 {
		upCount := 0
		for i := n - 12; i < n; i++ {
			if closes[i] > opens[i] {
				upCount++
			}
		}
		results["psy"] = float64(upCount) / 12.0 * 100
	}

	// ── OBV ──
	obv := 0.0
	for i := 1; i < n; i++ {
		if closes[i] > closes[i-1] {
			obv += volumes[i]
		} else if closes[i] < closes[i-1] {
			obv -= volumes[i]
		}
	}
	results["obv"] = obv

	// ── TRIX(12) ──
	if n >= 25 {
		tr := trix(closes, 12)
		results["trix"] = tr
	}

	// ── MFI(14) ──
	if n >= 15 {
		results["mfi"] = mfi(highs, lows, closes, volumes, 14)
	}

	// ═══ 新增指标 ═══

	// ── MA Golden Cross (5上穿10) ──
	if n >= 11 {
		boolRes["ma_golden_cross"] = ma5[len(ma5)-1] > ma10[len(ma10)-1] && ma5[len(ma5)-2] <= ma10[len(ma10)-2]
	}

	// ── ADX(14) ──
	if n >= 28 {
		results["adx"] = adx(highs, lows, closes, 14)
	}

	// ── BBI (3,6,12,24) ──
	ma3 := sma(closes, 3)
	ma6 := sma(closes, 6)
	ma12 := sma(closes, 12)
	ma24 := sma(closes, 24)
	if len(ma3) > 0 && len(ma6) > 0 && len(ma12) > 0 && len(ma24) > 0 {
		bbi := (ma3[len(ma3)-1] + ma6[len(ma6)-1] + ma12[len(ma12)-1] + ma24[len(ma24)-1]) / 4
		results["bbi"] = (last - bbi) / bbi * 100 // 偏离率
	}

	// ── EXPMA(12) ──
	ema12 := ema(closes, 12)
	if len(ema12) > 0 && ema12[len(ema12)-1] > 0 {
		results["expma"] = (last - ema12[len(ema12)-1]) / ema12[len(ema12)-1] * 100
	}

	// ── GMMA (顾比复合平均线) ──
	if n >= 30 {
		shortEMAs := []int{3, 5, 8, 10, 12, 15}
		longEMAs := []int{30, 35, 40, 45, 50, 60}
		shortAllAbove := true
		for i := 1; i < len(shortEMAs); i++ {
			e1 := ema(closes, shortEMAs[i-1])
			e2 := ema(closes, shortEMAs[i])
			if len(e1) == 0 || len(e2) == 0 || e1[len(e1)-1] <= e2[len(e2)-1] {
				shortAllAbove = false
				break
			}
		}
		longAllAbove := true
		for i := 1; i < len(longEMAs); i++ {
			e1 := ema(closes, longEMAs[i-1])
			e2 := ema(closes, longEMAs[i])
			if len(e1) == 0 || len(e2) == 0 || e1[len(e1)-1] <= e2[len(e2)-1] {
				longAllAbove = false
				break
			}
		}
		if shortAllAbove && longAllAbove {
			// Check short group completely above long group
			shortLast := ema(closes, shortEMAs[len(shortEMAs)-1])
			longFirst := ema(closes, longEMAs[0])
			if len(shortLast) > 0 && len(longFirst) > 0 {
				boolRes["gmma"] = shortLast[len(shortLast)-1] > longFirst[len(longFirst)-1]
			}
		} else {
			boolRes["gmma"] = false
		}
	}

	// ── ROC(12) ──
	if n >= 13 && closes[n-13] > 0 {
		results["roc"] = (last/closes[n-13] - 1) * 100
	}

	// ── BRAR ──
	if n >= 26 {
		sumH := 0.0
		sumL := 0.0
		for i := n - 26; i < n; i++ {
			sumH += highs[i] - opens[i]
			sumL += opens[i] - lows[i]
		}
		if sumL > 0 {
			results["brar"] = sumH / sumL * 100
		}
	}

	// ── WVAD ──
	if n >= 24 {
		wvadSum := 0.0
		for i := n - 24; i < n; i++ {
			rangeHL := highs[i] - lows[i]
			if rangeHL > 0 {
				wvadSum += volumes[i] * (closes[i] - opens[i]) / rangeHL
			}
		}
		results["wvad"] = wvadSum
	}

	// ── VR(24) ──
	if n >= 25 {
		upVol, downVol, flatVol := 0.0, 0.0, 0.0
		for i := n - 24; i < n; i++ {
			if closes[i] > opens[i] {
				upVol += volumes[i]
			} else if closes[i] < opens[i] {
				downVol += volumes[i]
			} else {
				flatVol += volumes[i]
			}
		}
		denom := downVol + flatVol/2
		if denom > 0 {
			results["vr"] = (upVol + flatVol/2) / denom * 100
		}
	}

	// ── Main Force Net Inflow (估算) ──
	if n >= 1 {
		rangeHL := highs[n-1] - lows[n-1]
		if rangeHL > 0 {
			// 基于当日成交额估算主力净流入
			results["main_force_net"] = volumes[n-1] * (closes[n-1] - (highs[n-1]+lows[n-1])/2) / rangeHL
		}
	}

	// ── Big Order Ratio (估算) ──
	if n >= 5 {
		avgVol := 0.0
		for i := n - 5; i < n; i++ {
			avgVol += volumes[i]
		}
		avgVol /= 5
		if avgVol > 0 {
			ratio := volumes[n-1] / avgVol
			if ratio > 1.5 {
				// 放量视为大单活跃
				results["big_order_ratio"] = (ratio - 1.5) * 100
			} else {
				results["big_order_ratio"] = (ratio - 1) * 100
			}
		}
	}

	// ── Chip Concentration (估算) ──
	if n >= 20 {
		avgPrice := 0.0
		for i := n - 20; i < n; i++ {
			avgPrice += (highs[i] + lows[i] + closes[i]) / 3
		}
		avgPrice /= 20
		if avgPrice > 0 {
			results["chip_concentration"] = (last - avgPrice) / avgPrice * 100
		}
	}

	// ── Volume Trend ──
	vol10 := sma(volumes, 10)
	if len(vol10) > 0 && len(vol5) > 0 && vol10[len(vol10)-1] > 0 {
		results["volume_trend"] = vol5[len(vol5)-1] / vol10[len(vol10)-1]
	}

	// ── STD Dev(20) ──
	if n >= 20 {
		mean := 0.0
		for i := n - 20; i < n; i++ {
			mean += closes[i]
		}
		mean /= 20
		variance := 0.0
		for i := n - 20; i < n; i++ {
			diff := closes[i] - mean
			variance += diff * diff
		}
		variance /= 20
		results["std_dev"] = math.Sqrt(variance)
	}

	// ── ENE Upper Break ──
	eneUpper, _ := ene(closes, 20, 11)
	if eneUpper > 0 {
		boolRes["ene_upper"] = last > eneUpper
	}

	// ── Xue Channel (薛斯通道) ──
	if n >= 20 {
		lowest := lows[n-1]
		for i := n - 20; i < n; i++ {
			if lows[i] < lowest {
				lowest = lows[i]
			}
		}
		highest := highs[n-1]
		for i := n - 20; i < n; i++ {
			if highs[i] > highest {
				highest = highs[i]
			}
		}
		xueLower := lowest + (highest-lowest)*0.236
		boolRes["xue_channel"] = last <= xueLower
	}

	// ── Historical Volatility(20) ──
	if n >= 21 {
		logReturns := make([]float64, 20)
		for i := 0; i < 20; i++ {
			idx := n - 20 + i
			if closes[idx-1] > 0 {
				logReturns[i] = math.Log(closes[idx] / closes[idx-1])
			}
		}
		mean := 0.0
		for _, r := range logReturns {
			mean += r
		}
		mean /= 20
		variance := 0.0
		for _, r := range logReturns {
			diff := r - mean
			variance += diff * diff
		}
		variance /= 19 // sample std
		hv := math.Sqrt(variance) * math.Sqrt(252) * 100 // annualized
		results["hist_volatility"] = hv
	}

	// ── New 20d Low ──
	if n >= 20 {
		min20 := closes[n-20]
		for i := n - 20; i < n-1; i++ {
			if closes[i] < min20 {
				min20 = closes[i]
			}
		}
		boolRes["new_20d_low"] = last < min20
	}

	// ── 20d Change ──
	if n >= 21 && closes[n-21] > 0 {
		results["pct_20d"] = (last - closes[n-21]) / closes[n-21] * 100
	}

	// ── Golden Ratio Break ──
	if n >= 20 {
		periodLow := lows[n-1]
		periodHigh := highs[n-1]
		for i := n - 20; i < n; i++ {
			if lows[i] < periodLow {
				periodLow = lows[i]
			}
			if highs[i] > periodHigh {
				periodHigh = highs[i]
			}
		}
		if periodHigh > periodLow {
			golden618 := periodLow + (periodHigh-periodLow)*0.618
			results["golden_ratio"] = (last - golden618) / golden618 * 100
		}
	}

	// ── Fibonacci Deep Retrace ──
	if n >= 40 {
		periodLow := lows[n-1]
		periodHigh := highs[n-1]
		for i := n - 40; i < n; i++ {
			if lows[i] < periodLow {
				periodLow = lows[i]
			}
			if highs[i] > periodHigh {
				periodHigh = highs[i]
			}
		}
		if periodHigh > periodLow {
			retrace := (last - periodLow) / (periodHigh - periodLow)
			boolRes["fib_deep_retrace"] = retrace < 0.382
		}
	}

	// ═══ 箱体形态检测 ═══
	if n >= 40 {
		boxTop, boxBot, touchesTop, touchesBot := detectBox(highs, lows, 40)
		if boxTop > boxBot && boxBot > 0 {
			boxHeight := (boxTop - boxBot) / boxBot * 100
			results["box_height_pct"] = boxHeight

			// 箱体位置：当前价在箱体中的相对位置 (0=箱底, 100=箱顶)
			if boxTop > boxBot {
				pos := (last - boxBot) / (boxTop - boxBot) * 100
				if pos < 0 {
					pos = 0
				} else if pos > 100 {
					pos = 100
				}
				results["box_position"] = pos
			}

			// 箱体缩量：最近5日均量 / 全周期均量
			vol5Avg := 0.0
			volAllAvg := 0.0
			for i := n - 5; i < n; i++ {
				vol5Avg += volumes[i]
			}
			vol5Avg /= 5
			for i := 0; i < n; i++ {
				volAllAvg += volumes[i]
			}
			volAllAvg /= float64(n)
			if volAllAvg > 0 {
				results["box_volume_shrink"] = vol5Avg / volAllAvg
			}

			// 箱体有效性条件：高度 > 5%，顶部≥2次触碰，底部≥2次触碰
			isBoxValid := boxHeight > 5 && touchesTop >= 2 && touchesBot >= 2
			boolRes["box_consolidating"] = isBoxValid

			// 箱底附近（买入区）：价格处于箱体下半区15%以内
			if isBoxValid {
				pos := results["box_position"]
				boolRes["box_at_bottom"] = pos <= 15

				// 放量突破箱顶：收盘价在箱顶上方且当日量>5日均量*1.5
				if last > boxTop && len(vol5) > 0 && vol5[len(vol5)-1] > 0 {
					volRatio := volumes[n-1] / vol5[len(vol5)-1]
					boolRes["box_breakout_up"] = volRatio > 1.5
				} else {
					boolRes["box_breakout_up"] = false
				}
			} else {
				boolRes["box_at_bottom"] = false
				boolRes["box_breakout_up"] = false
			}
		}
	}
}

// detectBox 检测箱体形态，返回箱顶、箱底、顶部触碰次数、底部触碰次数
func detectBox(highs, lows []float64, lookback int) (boxTop, boxBot float64, touchesTop, touchesBot int) {
	n := len(highs)
	if n < lookback {
		return 0, 0, 0, 0
	}
	start := n - lookback

	// 1. 找摆动高点和低点（窗口=3）
	var swingHighs, swingLows []float64
	for i := start + 1; i < n-1; i++ {
		if highs[i] > highs[i-1] && highs[i] > highs[i+1] {
			swingHighs = append(swingHighs, highs[i])
		}
		if lows[i] < lows[i-1] && lows[i] < lows[i+1] {
			swingLows = append(swingLows, lows[i])
		}
	}

	if len(swingHighs) < 2 || len(swingLows) < 2 {
		return 0, 0, 0, 0
	}

	// 2. 聚类：将相近的摆动点归为一组（容忍度3%）
	cluster := func(pts []float64, tolerance float64) (bestLevel float64, bestCount int) {
		maxCount := 0
		for _, p := range pts {
			count := 0
			for _, q := range pts {
				if math.Abs(p-q)/math.Max(p, 0.01) <= tolerance {
					count++
				}
			}
			if count > maxCount {
				maxCount = count
				bestLevel = p
			}
		}
		return bestLevel, maxCount
	}

	// 顶部聚类（从高处找，降低容忍度找到明确的阻力位）
	boxTop, touchesTop = cluster(swingHighs, 0.03)
	// 底部聚类（从低处找）
	boxBot, touchesBot = cluster(swingLows, 0.03)

	// 3. 验证：箱顶必须在箱底上方
	if boxTop <= boxBot {
		return 0, 0, 0, 0
	}

	return boxTop, boxBot, touchesTop, touchesBot
}

// ── Indicator Helpers ──

func adx(high, low, close []float64, period int) float64 {
	n := len(close)
	if n < period*2 {
		return 0
	}
	// +DI and -DI
	plusDM := make([]float64, n-1)
	minusDM := make([]float64, n-1)
	tr := make([]float64, n-1)
	for i := 1; i < n; i++ {
		upMove := high[i] - high[i-1]
		downMove := low[i-1] - low[i]
		plusDM[i-1] = 0
		minusDM[i-1] = 0
		if upMove > downMove && upMove > 0 {
			plusDM[i-1] = upMove
		}
		if downMove > upMove && downMove > 0 {
			minusDM[i-1] = downMove
		}
		hL := high[i] - low[i]
		hC := math.Abs(high[i] - close[i-1])
		lC := math.Abs(low[i] - close[i-1])
		tr[i-1] = math.Max(hL, math.Max(hC, lC))
	}
	// Smooth
	avgTR := sma(tr, period)
	avgPlus := sma(plusDM, period)
	avgMinus := sma(minusDM, period)
	if len(avgTR) == 0 || len(avgPlus) == 0 || len(avgMinus) == 0 {
		return 0
	}
	lastIdx := len(avgTR) - 1
	plusDI := avgPlus[lastIdx] / avgTR[lastIdx] * 100
	minusDI := avgMinus[lastIdx] / avgTR[lastIdx] * 100
	dx := math.Abs(plusDI-minusDI) / (plusDI + minusDI) * 100
	if math.IsNaN(dx) {
		return 0
	}
	return dx
}

func ene(close []float64, period int, multiplier float64) (float64, float64) {
	ma := sma(close, period)
	if len(ma) == 0 {
		return 0, 0
	}
	lastMA := ma[len(ma)-1]
	// Calculate mean deviation
	n := len(close)
	sumDev := 0.0
	for i := n - period; i < n; i++ {
		dev := math.Abs(close[i] - lastMA)
		sumDev += dev
	}
	md := sumDev / float64(period)
	upper := lastMA + md*multiplier/10
	lower := lastMA - md*multiplier/10
	return upper, lower
}

func ema(data []float64, period int) []float64 {
	if len(data) < period {
		return nil
	}
	k := 2.0 / float64(period+1)
	result := make([]float64, len(data))
	result[0] = data[0]
	for i := 1; i < len(data); i++ {
		result[i] = data[i]*k + result[i-1]*(1-k)
	}
	return result
}

func sma(data []float64, period int) []float64 {
	if len(data) < period {
		return nil
	}
	result := make([]float64, len(data)-period+1)
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += data[i]
	}
	result[0] = sum / float64(period)
	for i := period; i < len(data); i++ {
		sum += data[i] - data[i-period]
		result[i-period+1] = sum / float64(period)
	}
	return result
}

func rsi(data []float64, period int) float64 {
	if len(data) < period+1 {
		return 50
	}
	gains, losses := 0.0, 0.0
	for i := len(data) - period; i < len(data); i++ {
		diff := data[i] - data[i-1]
		if diff > 0 {
			gains += diff
		} else {
			losses -= diff
		}
	}
	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)
	if avgLoss == 0 {
		return 100
	}
	rs := avgGain / avgLoss
	return 100 - 100/(1+rs)
}

func kdj(high, low, close []float64, n, k, d int) (float64, float64) {
	if len(close) < n {
		return 50, 50
	}
	last := len(close) - 1
	hh := high[last-n+1]
	ll := low[last-n+1]
	for i := last - n + 1; i <= last; i++ {
		if high[i] > hh {
			hh = high[i]
		}
		if low[i] < ll {
			ll = low[i]
		}
	}
	rsv := 50.0
	if hh-ll > 0 {
		rsv = (close[last] - ll) / (hh - ll) * 100
	}
	kVal := (2.0/3.0)*50 + (1.0/3.0)*rsv
	dVal := (2.0/3.0)*50 + (1.0/3.0)*kVal
	return kVal, dVal
}

func difSeries(close []float64, short, long int) []float64 {
	emaS := ema(close, short)
	emaL := ema(close, long)
	if emaS == nil || emaL == nil {
		return nil
	}
	minLen := len(emaS)
	if len(emaL) < minLen {
		minLen = len(emaL)
	}
	result := make([]float64, minLen)
	for i := 0; i < minLen; i++ {
		result[i] = emaS[i] - emaL[i]
	}
	return result
}

func cci(high, low, close []float64, period int) float64 {
	n := len(close)
	if n < period {
		return 0
	}
	tp := (high[n-1] + low[n-1] + close[n-1]) / 3
	tpSum := 0.0
	for i := n - period; i < n; i++ {
		tpSum += (high[i] + low[i] + close[i]) / 3
	}
	avgTP := tpSum / float64(period)

	md := 0.0
	for i := n - period; i < n; i++ {
		t := (high[i] + low[i] + close[i]) / 3
		diff := t - avgTP
		if diff < 0 {
			diff = -diff
		}
		md += diff
	}
	md /= float64(period)
	if md == 0 {
		return 0
	}
	return (tp - avgTP) / (0.015 * md)
}

func williamsR(high, low, close []float64, period int) float64 {
	n := len(close)
	if n < period {
		return 50
	}
	hh := high[n-1]
	ll := low[n-1]
	for i := n - period; i < n; i++ {
		if high[i] > hh {
			hh = high[i]
		}
		if low[i] < ll {
			ll = low[i]
		}
	}
	if hh-ll == 0 {
		return 50
	}
	return (hh - close[n-1]) / (hh - ll) * 100
}

func bollinger(close []float64, period int, multiplier float64) (float64, float64) {
	n := len(close)
	if n < period {
		return close[n-1], close[n-1]
	}
	ma := 0.0
	for i := n - period; i < n; i++ {
		ma += close[i]
	}
	ma /= float64(period)

	variance := 0.0
	for i := n - period; i < n; i++ {
		diff := close[i] - ma
		variance += diff * diff
	}
	variance /= float64(period)
	std := math.Sqrt(variance)
	return ma + multiplier*std, ma - multiplier*std
}

func atr(high, low, close []float64, period int) float64 {
	n := len(close)
	if n < period+1 {
		return 0
	}
	var trs []float64
	for i := 1; i < n; i++ {
		hL := high[i] - low[i]
		hC := math.Abs(high[i] - close[i-1])
		lC := math.Abs(low[i] - close[i-1])
		tr := math.Max(hL, math.Max(hC, lC))
		trs = append(trs, tr)
	}
	if len(trs) < period {
		return 0
	}
	// SMA of TR
	sum := 0.0
	for _, v := range trs[len(trs)-period:] {
		sum += v
	}
	return sum / float64(period)
}

func sar(high, low []float64, acceleration, maxA float64) bool {
	n := len(high)
	if n < 5 {
		return true
	}
	// Simple: latest close > SAR > previous low means uptrend
	// Just return if recent trend is up
	upDays := 0
	for i := n - 5; i < n; i++ {
		if i > 0 && high[i] > high[i-1] {
			upDays++
		}
	}
	return upDays >= 3
}

func trix(close []float64, period int) float64 {
	ema1 := ema(close, period)
	if ema1 == nil || len(ema1) < period*3 {
		return 0
	}
	ema2 := ema(ema1, period)
	if ema2 == nil || len(ema2) < period*2 {
		return 0
	}
	ema3 := ema(ema2, period)
	if ema3 == nil || len(ema3) < 2 {
		return 0
	}
	last := ema3[len(ema3)-1]
	prev := ema3[len(ema3)-2]
	if prev == 0 {
		return 0
	}
	return (last - prev) / prev * 100
}

func mfi(high, low, close, volume []float64, period int) float64 {
	n := len(close)
	if n < period+1 {
		return 50
	}
	var posFlow, negFlow float64
	for i := n - period; i < n; i++ {
		tp := (high[i] + low[i] + close[i]) / 3
		prevTP := (high[i-1] + low[i-1] + close[i-1]) / 3
		mfv := tp * volume[i]
		if tp > prevTP {
			posFlow += mfv
		} else {
			negFlow += mfv
		}
	}
	if negFlow == 0 {
		return 100
	}
	mfr := posFlow / negFlow
	return 100 - 100/(1+mfr)
}

func parseFloat(s string) float64 {
	var v float64
	fmt.Sscanf(s, "%f", &v)
	return v
}

func findIndicatorCfg(indicators []models.StrategyIndicator, id string) *models.StrategyIndicator {
	for i := range indicators {
		if indicators[i].IndicatorID == id {
			return &indicators[i]
		}
	}
	return nil
}

// ── Storage ──

func (s *StrategyService) loadDefaults() {
	s.strategies = []models.Strategy{
		{
			ID:          "sys_momentum",
			Name:        "动量突破",
			Description: "关注强势突破、量能放大的股票，适合短线交易",
			IsDefault:   true,
			IsSystem:    true,
			Indicators: []models.StrategyIndicator{
				{IndicatorID: "rsi", Direction: "up", Weight: 20, MinVal: "30", MaxVal: "80"},
				{IndicatorID: "volume_ratio", Direction: "up", Weight: 20, MinVal: "1", MaxVal: ""},
				{IndicatorID: "macd_diff", Direction: "up", Weight: 20, MinVal: "0", MaxVal: ""},
				{IndicatorID: "pct_5d", Direction: "up", Weight: 15, MinVal: "0", MaxVal: ""},
				{IndicatorID: "boll_position", Direction: "up", Weight: 15, MinVal: "0.3", MaxVal: "0.9"},
				{IndicatorID: "ma_alignment", Direction: "up", Weight: 10, MinVal: "", MaxVal: ""},
			},
		},
		{
			ID:          "sys_value",
			Name:        "超跌反弹",
			Description: "捕捉超卖后反弹机会，适合左侧交易",
			IsDefault:   false,
			IsSystem:    true,
			Indicators: []models.StrategyIndicator{
				{IndicatorID: "rsi", Direction: "up", Weight: 25, MinVal: "0", MaxVal: "30"},
				{IndicatorID: "boll_position", Direction: "up", Weight: 20, MinVal: "0", MaxVal: "0.2"},
				{IndicatorID: "wr", Direction: "up", Weight: 15, MinVal: "80", MaxVal: "100"},
				{IndicatorID: "macd_diff", Direction: "up", Weight: 20, MinVal: "", MaxVal: ""},
				{IndicatorID: "volume_ratio", Direction: "up", Weight: 20, MinVal: "0.5", MaxVal: ""},
			},
		},
		{
			ID:          "sys_trend",
			Name:        "趋势跟踪",
			Description: "跟随上升趋势、均线多头的股票，适合中线持有",
			IsDefault:   false,
			IsSystem:    true,
			Indicators: []models.StrategyIndicator{
				{IndicatorID: "ma_alignment", Direction: "up", Weight: 25, MinVal: "", MaxVal: ""},
				{IndicatorID: "ema_deviation", Direction: "up", Weight: 15, MinVal: "0", MaxVal: ""},
				{IndicatorID: "macd_diff", Direction: "up", Weight: 20, MinVal: "0", MaxVal: ""},
				{IndicatorID: "atr", Direction: "up", Weight: 10, MinVal: "", MaxVal: ""},
				{IndicatorID: "psY", Direction: "up", Weight: 15, MinVal: "50", MaxVal: "100"},
				{IndicatorID: "boll_position", Direction: "up", Weight: 15, MinVal: "0.4", MaxVal: "0.9"},
			},
		},
		{
			ID:          "sys_box_trading",
			Name:        "箱体策略",
			Description: "识别箱体震荡形态，箱底买入箱顶卖出，结合放量突破确认",
			IsDefault:   false,
			IsSystem:    true,
			Indicators: []models.StrategyIndicator{
				{IndicatorID: "box_consolidating", Direction: "up", Weight: 25, MinVal: "", MaxVal: ""},
				{IndicatorID: "box_position", Direction: "down", Weight: 25, MinVal: "0", MaxVal: "30"},
				{IndicatorID: "box_height_pct", Direction: "up", Weight: 15, MinVal: "8", MaxVal: "40"},
				{IndicatorID: "box_volume_shrink", Direction: "up", Weight: 15, MinVal: "0", MaxVal: "1.0"},
				{IndicatorID: "box_at_bottom", Direction: "up", Weight: 10, MinVal: "", MaxVal: ""},
				{IndicatorID: "box_breakout_up", Direction: "up", Weight: 10, MinVal: "", MaxVal: ""},
			},
		},
	}
}

func (s *StrategyService) loadFromDisk() {
	data, err := os.ReadFile(s.storePath)
	if err != nil {
		return
	}
	var custom []models.Strategy
	if err := json.Unmarshal(data, &custom); err != nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	// Append custom strategies after system ones
	for _, c := range custom {
		c.IsSystem = false
		// Check for duplicates
		found := false
		for _, existing := range s.strategies {
			if existing.ID == c.ID {
				found = true
				break
			}
		}
		if !found {
			s.strategies = append(s.strategies, c)
		} else {
			// Update existing
			for i, existing := range s.strategies {
				if existing.ID == c.ID {
					s.strategies[i] = c
					break
				}
			}
		}
	}
}

func (s *StrategyService) saveToDisk() {
	var custom []models.Strategy
	for _, st := range s.strategies {
		if !st.IsSystem {
			custom = append(custom, st)
		}
	}
	data, err := json.MarshalIndent(custom, "", "  ")
	if err != nil {
		log.Printf("[Strategy] save error: %v", err)
		return
	}
	if err := os.WriteFile(s.storePath, data, 0644); err != nil {
		log.Printf("[Strategy] write error: %v", err)
	}
}
