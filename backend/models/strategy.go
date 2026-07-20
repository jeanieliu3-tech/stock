package models

// IndicatorDefinition 指标定义
type IndicatorDefinition struct {
	ID            string `json:"id"`
	Name          string `json:"name"`          // 显示名称
	Category      string `json:"category"`      // trend, oscillator, volume, volatility, pattern
	DefaultMin    string `json:"defaultMin"`    // 默认最小值（空字符串=不限制）
	DefaultMax    string `json:"defaultMax"`    // 默认最大值
	IsBool        bool   `json:"isBool"`        // 是否为布尔指标
	DefaultParams string `json:"defaultParams"` // 默认参数描述
}

// StrategyIndicator 策略中的指标配置
type StrategyIndicator struct {
	IndicatorID string  `json:"indicatorId"` // 指标ID
	Direction   string  `json:"direction"`   // "up"=越大越好, "down"=越小越好
	Weight      float64 `json:"weight"`      // 权重 0-100
	MinVal      string  `json:"minVal"`      // 筛选最小值（空=不限制）
	MaxVal      string  `json:"maxVal"`      // 筛选最大值（空=不限制）
}

// Strategy 策略
type Strategy struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	IsDefault   bool                `json:"isDefault"`
	IsSystem    bool                `json:"isSystem"` // 系统预置不可删除
	Indicators  []StrategyIndicator `json:"indicators"`
	CreatedAt   string              `json:"createdAt"`
	UpdatedAt   string              `json:"updatedAt"`
}

// StrategyListResponse 策略列表响应
type StrategyListResponse struct {
	Strategies []Strategy `json:"strategies"`
}

// IndicatorRawValue 指标原始值
type IndicatorRawValue struct {
	IndicatorID string  `json:"indicatorId"`
	Name        string  `json:"name"`
	Value       float64 `json:"value"`
	IsBool      bool    `json:"isBool"`
	BoolValue   bool    `json:"boolValue"`
}

// StockEvalResult 单只股票评估结果
type StockEvalResult struct {
	Code        string              `json:"code"`
	Name        string              `json:"name"`
	Price       float64             `json:"price"`
	ChangePct   float64             `json:"changePercent"`
	TotalScore  float64             `json:"totalScore"`
	Indicators  []IndicatorEvalItem `json:"indicators"`
	Rank        int                 `json:"rank"`
}

// IndicatorEvalItem 指标评估项
type IndicatorEvalItem struct {
	IndicatorID   string  `json:"indicatorId"`
	Name          string  `json:"name"`
	RawValue      float64 `json:"rawValue"`
	BoolValue     bool    `json:"boolValue"`
	IsBool        bool    `json:"isBool"`
	InRange       bool    `json:"inRange"`
	Score         float64 `json:"score"`
	Weight        float64 `json:"weight"`
	Percentile    float64 `json:"percentile"`
}

// StrategyEvalRequest 策略评估请求
type StrategyEvalRequest struct {
	StrategyID string     `json:"strategyId"`
	Strategy   *Strategy  `json:"strategy"` // 不保存直接评估
	Page       int        `json:"page"`
	PageSize   int        `json:"pageSize"`
}

// StrategyEvalResponse 策略评估响应
type StrategyEvalResponse struct {
	Total    int               `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"pageSize"`
	Items    []StockEvalResult `json:"items"`
	CostMs   int64             `json:"costMs"`
}

// 预定义指标库（六大类，40+个）
var IndicatorLibrary = []IndicatorDefinition{
	// ═══ 1. 趋势跟踪类（10个）═══
	{ID: "ma_alignment", Name: "均线多头排列(5,10,20,60)", Category: "trend", DefaultMin: "", DefaultMax: "", IsBool: true},
	{ID: "ma_golden_cross", Name: "MA金叉(5上穿10)", Category: "trend", DefaultMin: "", DefaultMax: "", IsBool: true},
	{ID: "ema_deviation", Name: "EMA乖离率(收盘价/EMA20-1)", Category: "trend", DefaultMin: "", DefaultMax: "", IsBool: false},
	{ID: "macd_diff", Name: "MACD金叉强度(DIF-DEA)", Category: "trend", DefaultMin: "0", DefaultMax: "", IsBool: false},
	{ID: "adx", Name: "ADX趋向指标(14)", Category: "trend", DefaultMin: "25", DefaultMax: "", IsBool: false},
	{ID: "sar", Name: "SAR抛物线转向", Category: "trend", DefaultMin: "", DefaultMax: "", IsBool: true},
	{ID: "trix", Name: "TRIX三重指数平滑(12)", Category: "trend", DefaultMin: "", DefaultMax: "", IsBool: false},
	{ID: "bbi", Name: "BBI多空指标(3,6,12,24)", Category: "trend", DefaultMin: "", DefaultMax: "", IsBool: false},
	{ID: "expma", Name: "EXPMA指数平均数(12)", Category: "trend", DefaultMin: "", DefaultMax: "", IsBool: false},
	{ID: "gmma", Name: "顾比复合平均线(GMMA)", Category: "trend", DefaultMin: "", DefaultMax: "", IsBool: true},

	// ═══ 2. 动量与摆动类（10个）═══
	{ID: "rsi", Name: "RSI相对强弱(14)", Category: "oscillator", DefaultMin: "0", DefaultMax: "100", IsBool: false},
	{ID: "kdj_j", Name: "KDJ随机指标J值(9,3,3)", Category: "oscillator", DefaultMin: "0", DefaultMax: "100", IsBool: false},
	{ID: "cci", Name: "CCI商品通道(14)", Category: "oscillator", DefaultMin: "", DefaultMax: "", IsBool: false},
	{ID: "wr", Name: "WR威廉指标(14)", Category: "oscillator", DefaultMin: "0", DefaultMax: "100", IsBool: false},
	{ID: "mom", Name: "MOM动量(10)", Category: "oscillator", DefaultMin: "", DefaultMax: "", IsBool: false},
	{ID: "roc", Name: "ROC变动速率(12)", Category: "oscillator", DefaultMin: "", DefaultMax: "", IsBool: false},
	{ID: "bias", Name: "BIAS乖离率(6)", Category: "oscillator", DefaultMin: "", DefaultMax: "", IsBool: false},
	{ID: "psy", Name: "PSY心理线(12)", Category: "oscillator", DefaultMin: "25", DefaultMax: "75", IsBool: false},
	{ID: "brar", Name: "BRAR多空意愿", Category: "oscillator", DefaultMin: "", DefaultMax: "", IsBool: false},
	{ID: "wvad", Name: "WVAD威廉变异离散量", Category: "oscillator", DefaultMin: "", DefaultMax: "", IsBool: false},

	// ═══ 3. 量能与资金流类（8个）═══
	{ID: "volume_ratio", Name: "量比(成交量/5日均量)", Category: "volume", DefaultMin: "0.5", DefaultMax: "", IsBool: false},
	{ID: "obv", Name: "OBV能量潮", Category: "volume", DefaultMin: "", DefaultMax: "", IsBool: false},
	{ID: "mfi", Name: "MFI资金流量(14)", Category: "volume", DefaultMin: "0", DefaultMax: "100", IsBool: false},
	{ID: "vr", Name: "VR成交量比率(24)", Category: "volume", DefaultMin: "", DefaultMax: "", IsBool: false},
	{ID: "main_force_net", Name: "主力净流入(估算)", Category: "volume", DefaultMin: "", DefaultMax: "", IsBool: false},
	{ID: "big_order_ratio", Name: "大单比率(估算)", Category: "volume", DefaultMin: "", DefaultMax: "", IsBool: false},
	{ID: "chip_concentration", Name: "筹码集中度(估算)", Category: "volume", DefaultMin: "", DefaultMax: "", IsBool: false},
	{ID: "volume_trend", Name: "量能趋势(vol/MAvol)", Category: "volume", DefaultMin: "", DefaultMax: "", IsBool: false},

	// ═══ 4. 波动与通道类（6个）═══
	{ID: "boll_position", Name: "BOLL位置((收盘价-下轨)/(上轨-下轨))", Category: "volatility", DefaultMin: "0", DefaultMax: "1", IsBool: false},
	{ID: "atr", Name: "ATR平均真实波幅(14)", Category: "volatility", DefaultMin: "", DefaultMax: "", IsBool: false},
	{ID: "std_dev", Name: "STD标准差(20)", Category: "volatility", DefaultMin: "", DefaultMax: "", IsBool: false},
	{ID: "ene_upper", Name: "ENE轨道线突破上轨", Category: "volatility", DefaultMin: "", DefaultMax: "", IsBool: true},
	{ID: "xue_channel", Name: "薛斯通道触及下轨", Category: "volatility", DefaultMin: "", DefaultMax: "", IsBool: true},
	{ID: "hist_volatility", Name: "历史波动率(20日)", Category: "volatility", DefaultMin: "", DefaultMax: "", IsBool: false},

	// ═══ 5. 支撑压力与形态类（6个）═══
	{ID: "new_20d_high", Name: "创20日新高", Category: "pattern", DefaultMin: "", DefaultMax: "", IsBool: true},
	{ID: "new_20d_low", Name: "创20日新低", Category: "pattern", DefaultMin: "", DefaultMax: "", IsBool: true},
	{ID: "pct_5d", Name: "5日涨跌幅(%)", Category: "pattern", DefaultMin: "", DefaultMax: "", IsBool: false},
	{ID: "pct_20d", Name: "20日涨跌幅(%)", Category: "pattern", DefaultMin: "", DefaultMax: "", IsBool: false},
	{ID: "golden_ratio", Name: "黄金分割突破(0.618)", Category: "pattern", DefaultMin: "", DefaultMax: "", IsBool: false},
	{ID: "fib_deep_retrace", Name: "斐波那契深度回调", Category: "pattern", DefaultMin: "", DefaultMax: "", IsBool: true},

	// ═══ 6. 箱体形态类（6个）═══
	{ID: "box_consolidating", Name: "箱体震荡(有清晰顶底)", Category: "pattern", DefaultMin: "", DefaultMax: "", IsBool: true},
	{ID: "box_position", Name: "箱体位置(0=箱底,100=箱顶)", Category: "pattern", DefaultMin: "0", DefaultMax: "100", IsBool: false},
	{ID: "box_height_pct", Name: "箱体高度(%)", Category: "pattern", DefaultMin: "5", DefaultMax: "50", IsBool: false},
	{ID: "box_volume_shrink", Name: "箱体缩量(vol/均量)", Category: "pattern", DefaultMin: "0", DefaultMax: "1.2", IsBool: false},
	{ID: "box_at_bottom", Name: "箱底附近(买入区)", Category: "pattern", DefaultMin: "", DefaultMax: "", IsBool: true},
	{ID: "box_breakout_up", Name: "放量突破箱顶", Category: "pattern", DefaultMin: "", DefaultMax: "", IsBool: true},
}

func GetIndicatorByID(id string) *IndicatorDefinition {
	for _, ind := range IndicatorLibrary {
		if ind.ID == id {
			return &ind
		}
	}
	return nil
}
