package market

import (
	"math"
	"sort"
)

// EnhancedIndicators 增强技术指标
type EnhancedIndicators struct {
	// 趋势指标
	BollingerBands *BollingerBands
	VWAP           float64
	Ichimoku       *IchimokuCloud
	
	// 动量指标
	Stochastic     *StochasticOscillator
	Williams       float64 // Williams %R
	CCI            float64 // Commodity Channel Index
	
	// 成交量指标
	OBV            float64 // On Balance Volume
	VolumeProfile  *VolumeProfile
	VWMA           float64 // Volume Weighted Moving Average
	
	// 波动率指标
	TrueRange      float64
	HistoricalVol  float64
	
	// 市场结构指标
	SupportLevels  []float64
	ResistanceLevels []float64
	PivotPoints    *PivotPoints
}

// BollingerBands 布林带
type BollingerBands struct {
	Upper      float64
	Middle     float64 // SMA
	Lower      float64
	Width      float64 // 带宽
	Position   float64 // 价格在带内的位置 (0-1)
}

// IchimokuCloud 一目均衡表
type IchimokuCloud struct {
	TenkanSen   float64 // 转换线
	KijunSen    float64 // 基准线
	SenkouSpanA float64 // 先行带A
	SenkouSpanB float64 // 先行带B
	ChikouSpan  float64 // 迟行线
	CloudColor  string  // "bullish" or "bearish"
}

// StochasticOscillator 随机振荡器
type StochasticOscillator struct {
	K      float64 // %K值
	D      float64 // %D值
	Signal string  // "oversold", "overbought", "neutral"
}

// VolumeProfile 成交量分布
type VolumeProfile struct {
	VPOC       float64   // Volume Point of Control
	VAH        float64   // Value Area High
	VAL        float64   // Value Area Low
	VolumeNodes []VolumeNode
}

// VolumeNode 成交量节点
type VolumeNode struct {
	Price  float64
	Volume float64
}

// PivotPoints 枢轴点
type PivotPoints struct {
	Pivot float64
	R1    float64
	R2    float64
	R3    float64
	S1    float64
	S2    float64
	S3    float64
}

// MarketSentiment 市场情绪分析
type MarketSentiment struct {
	FearGreedIndex    int     // 0-100, 恐慌贪婪指数
	BullBearRatio     float64 // 多空比例
	VolumeStrength    string  // "strong", "weak", "normal"
	MomentumSignal    string  // "bullish", "bearish", "neutral"
	OverallSentiment  string  // "extreme_fear", "fear", "neutral", "greed", "extreme_greed"
}

// CalculateEnhancedIndicators 计算增强技术指标
func CalculateEnhancedIndicators(klines []Kline) *EnhancedIndicators {
	if len(klines) < 50 {
		return nil
	}

	indicators := &EnhancedIndicators{}
	
	// 计算布林带
	indicators.BollingerBands = calculateBollingerBands(klines, 20, 2.0)
	
	// 计算VWAP
	indicators.VWAP = calculateVWAP(klines)
	
	// 计算一目均衡表
	indicators.Ichimoku = calculateIchimoku(klines)
	
	// 计算随机振荡器
	indicators.Stochastic = calculateStochastic(klines, 14, 3)
	
	// 计算Williams %R
	indicators.Williams = calculateWilliamsR(klines, 14)
	
	// 计算CCI
	indicators.CCI = calculateCCI(klines, 20)
	
	// 计算OBV
	indicators.OBV = calculateOBV(klines)
	
	// 计算成交量加权移动平均
	indicators.VWMA = calculateVWMA(klines, 20)
	
	// 计算历史波动率
	indicators.HistoricalVol = calculateHistoricalVolatility(klines, 20)
	
	// 计算支撑阻力位
	indicators.SupportLevels, indicators.ResistanceLevels = calculateSupportResistance(klines)
	
	// 计算枢轴点
	indicators.PivotPoints = calculatePivotPoints(klines)
	
	return indicators
}

// calculateBollingerBands 计算布林带
func calculateBollingerBands(klines []Kline, period int, stdDev float64) *BollingerBands {
	if len(klines) < period {
		return nil
	}
	
	// 计算SMA
	sum := 0.0
	for i := len(klines) - period; i < len(klines); i++ {
		sum += klines[i].Close
	}
	sma := sum / float64(period)
	
	// 计算标准差
	variance := 0.0
	for i := len(klines) - period; i < len(klines); i++ {
		diff := klines[i].Close - sma
		variance += diff * diff
	}
	stdDeviation := math.Sqrt(variance / float64(period))
	
	upper := sma + (stdDev * stdDeviation)
	lower := sma - (stdDev * stdDeviation)
	
	// 计算价格在带内的位置
	currentPrice := klines[len(klines)-1].Close
	position := (currentPrice - lower) / (upper - lower)
	
	return &BollingerBands{
		Upper:    upper,
		Middle:   sma,
		Lower:    lower,
		Width:    (upper - lower) / sma * 100, // 带宽百分比
		Position: position,
	}
}

// calculateVWAP 计算成交量加权平均价格
func calculateVWAP(klines []Kline) float64 {
	if len(klines) == 0 {
		return 0
	}
	
	totalVolume := 0.0
	totalVolumePrice := 0.0
	
	for _, kline := range klines {
		typicalPrice := (kline.High + kline.Low + kline.Close) / 3
		totalVolumePrice += typicalPrice * kline.Volume
		totalVolume += kline.Volume
	}
	
	if totalVolume == 0 {
		return 0
	}
	
	return totalVolumePrice / totalVolume
}

// calculateIchimoku 计算一目均衡表
func calculateIchimoku(klines []Kline) *IchimokuCloud {
	if len(klines) < 52 {
		return nil
	}
	
	// 转换线 (9期最高价+最低价)/2
	tenkanSen := (getHighest(klines, 9) + getLowest(klines, 9)) / 2
	
	// 基准线 (26期最高价+最低价)/2
	kijunSen := (getHighest(klines, 26) + getLowest(klines, 26)) / 2
	
	// 先行带A (转换线+基准线)/2
	senkouSpanA := (tenkanSen + kijunSen) / 2
	
	// 先行带B (52期最高价+最低价)/2
	senkouSpanB := (getHighest(klines, 52) + getLowest(klines, 52)) / 2
	
	// 迟行线 (当前收盘价)
	chikouSpan := klines[len(klines)-1].Close
	
	// 云的颜色
	cloudColor := "bullish"
	if senkouSpanA < senkouSpanB {
		cloudColor = "bearish"
	}
	
	return &IchimokuCloud{
		TenkanSen:   tenkanSen,
		KijunSen:    kijunSen,
		SenkouSpanA: senkouSpanA,
		SenkouSpanB: senkouSpanB,
		ChikouSpan:  chikouSpan,
		CloudColor:  cloudColor,
	}
}

// calculateStochastic 计算随机振荡器
func calculateStochastic(klines []Kline, kPeriod, dPeriod int) *StochasticOscillator {
	if len(klines) < kPeriod {
		return nil
	}
	
	// 计算%K
	currentClose := klines[len(klines)-1].Close
	lowestLow := getLowest(klines, kPeriod)
	highestHigh := getHighest(klines, kPeriod)
	
	k := 100 * (currentClose - lowestLow) / (highestHigh - lowestLow)
	
	// 计算%D (简化为%K的移动平均)
	d := k // 简化实现
	
	// 判断信号
	signal := "neutral"
	if k < 20 {
		signal = "oversold"
	} else if k > 80 {
		signal = "overbought"
	}
	
	return &StochasticOscillator{
		K:      k,
		D:      d,
		Signal: signal,
	}
}

// calculateWilliamsR 计算Williams %R
func calculateWilliamsR(klines []Kline, period int) float64 {
	if len(klines) < period {
		return 0
	}
	
	currentClose := klines[len(klines)-1].Close
	highestHigh := getHighest(klines, period)
	lowestLow := getLowest(klines, period)
	
	return -100 * (highestHigh - currentClose) / (highestHigh - lowestLow)
}

// calculateCCI 计算商品通道指数
func calculateCCI(klines []Kline, period int) float64 {
	if len(klines) < period {
		return 0
	}
	
	// 计算典型价格
	typicalPrices := make([]float64, len(klines))
	for i, kline := range klines {
		typicalPrices[i] = (kline.High + kline.Low + kline.Close) / 3
	}
	
	// 计算移动平均
	sum := 0.0
	for i := len(typicalPrices) - period; i < len(typicalPrices); i++ {
		sum += typicalPrices[i]
	}
	sma := sum / float64(period)
	
	// 计算平均偏差
	deviation := 0.0
	for i := len(typicalPrices) - period; i < len(typicalPrices); i++ {
		deviation += math.Abs(typicalPrices[i] - sma)
	}
	meanDeviation := deviation / float64(period)
	
	currentTP := typicalPrices[len(typicalPrices)-1]
	
	if meanDeviation == 0 {
		return 0
	}
	
	return (currentTP - sma) / (0.015 * meanDeviation)
}

// calculateOBV 计算能量潮
func calculateOBV(klines []Kline) float64 {
	if len(klines) < 2 {
		return 0
	}
	
	obv := 0.0
	for i := 1; i < len(klines); i++ {
		if klines[i].Close > klines[i-1].Close {
			obv += klines[i].Volume
		} else if klines[i].Close < klines[i-1].Close {
			obv -= klines[i].Volume
		}
		// 如果价格相等，OBV不变
	}
	
	return obv
}

// calculateVWMA 计算成交量加权移动平均
func calculateVWMA(klines []Kline, period int) float64 {
	if len(klines) < period {
		return 0
	}
	
	totalVolumePrice := 0.0
	totalVolume := 0.0
	
	for i := len(klines) - period; i < len(klines); i++ {
		totalVolumePrice += klines[i].Close * klines[i].Volume
		totalVolume += klines[i].Volume
	}
	
	if totalVolume == 0 {
		return 0
	}
	
	return totalVolumePrice / totalVolume
}

// calculateHistoricalVolatility 计算历史波动率
func calculateHistoricalVolatility(klines []Kline, period int) float64 {
	if len(klines) < period+1 {
		return 0
	}
	
	// 计算对数收益率
	returns := make([]float64, 0, period)
	for i := len(klines) - period; i < len(klines); i++ {
		if i > 0 && klines[i-1].Close > 0 {
			ret := math.Log(klines[i].Close / klines[i-1].Close)
			returns = append(returns, ret)
		}
	}
	
	if len(returns) == 0 {
		return 0
	}
	
	// 计算平均收益率
	sum := 0.0
	for _, ret := range returns {
		sum += ret
	}
	mean := sum / float64(len(returns))
	
	// 计算方差
	variance := 0.0
	for _, ret := range returns {
		diff := ret - mean
		variance += diff * diff
	}
	variance /= float64(len(returns) - 1)
	
	// 年化波动率 (假设一年252个交易日)
	return math.Sqrt(variance * 252) * 100
}

// calculateSupportResistance 计算支撑阻力位
func calculateSupportResistance(klines []Kline) ([]float64, []float64) {
	if len(klines) < 20 {
		return nil, nil
	}
	
	// 简化实现：找到局部高点和低点
	supports := make([]float64, 0)
	resistances := make([]float64, 0)
	
	// 寻找局部极值点
	for i := 2; i < len(klines)-2; i++ {
		// 局部低点 (支撑位)
		if klines[i].Low < klines[i-1].Low && klines[i].Low < klines[i-2].Low &&
		   klines[i].Low < klines[i+1].Low && klines[i].Low < klines[i+2].Low {
			supports = append(supports, klines[i].Low)
		}
		
		// 局部高点 (阻力位)
		if klines[i].High > klines[i-1].High && klines[i].High > klines[i-2].High &&
		   klines[i].High > klines[i+1].High && klines[i].High > klines[i+2].High {
			resistances = append(resistances, klines[i].High)
		}
	}
	
	// 排序并去重
	sort.Float64s(supports)
	sort.Float64s(resistances)
	
	return supports, resistances
}

// calculatePivotPoints 计算枢轴点
func calculatePivotPoints(klines []Kline) *PivotPoints {
	if len(klines) == 0 {
		return nil
	}
	
	// 使用最后一根K线计算
	lastKline := klines[len(klines)-1]
	high := lastKline.High
	low := lastKline.Low
	close := lastKline.Close
	
	pivot := (high + low + close) / 3
	
	return &PivotPoints{
		Pivot: pivot,
		R1:    2*pivot - low,
		R2:    pivot + (high - low),
		R3:    high + 2*(pivot - low),
		S1:    2*pivot - high,
		S2:    pivot - (high - low),
		S3:    low - 2*(high - pivot),
	}
}

// AnalyzeMarketSentiment 分析市场情绪
func AnalyzeMarketSentiment(data *Data, indicators *EnhancedIndicators) *MarketSentiment {
	sentiment := &MarketSentiment{}
	
	// 计算恐慌贪婪指数 (简化版)
	sentiment.FearGreedIndex = calculateFearGreedIndex(data, indicators)
	
	// 计算多空比（使用1小时数据，更能反映当前市场情绪）
	sentiment.BullBearRatio = calculateBullBearRatio(data)
	
	// 分析成交量强度
	sentiment.VolumeStrength = analyzeVolumeStrength(data)
	
	// 分析动量信号
	sentiment.MomentumSignal = analyzeMomentumSignal(data, indicators)
	
	// 综合情绪评估
	sentiment.OverallSentiment = assessOverallSentiment(sentiment.FearGreedIndex)
	
	return sentiment
}

// calculateBullBearRatio 计算多空比
func calculateBullBearRatio(data *Data) float64 {
	if data.LongShortRatios == nil || len(data.LongShortRatios) == 0 {
		return 0.0
	}
	
	// 优先使用1小时数据，更能反映当前市场情绪
	if ratio, ok := data.LongShortRatios["1h"]; ok {
		return ratio.LongShortRatio
	}
	
	// 如果没有1小时数据，使用15分钟
	if ratio, ok := data.LongShortRatios["15m"]; ok {
		return ratio.LongShortRatio
	}
	
	// 否则使用任意可用的数据
	for _, ratio := range data.LongShortRatios {
		return ratio.LongShortRatio
	}
	
	return 0.0
}

// calculateFearGreedIndex 计算恐慌贪婪指数
func calculateFearGreedIndex(data *Data, indicators *EnhancedIndicators) int {
	score := 50 // 中性起点
	
	// RSI影响 (25%)
	if data.CurrentRSI7 > 70 {
		score += int((data.CurrentRSI7 - 70) / 30 * 25)
	} else if data.CurrentRSI7 < 30 {
		score -= int((30 - data.CurrentRSI7) / 30 * 25)
	}
	
	// 价格动量影响 (25%)
	if data.PriceChange4h > 5 {
		score += 25
	} else if data.PriceChange4h < -5 {
		score -= 25
	} else {
		score += int(data.PriceChange4h / 5 * 25)
	}
	
	// 波动率影响 (25%)
	if indicators != nil && indicators.HistoricalVol > 50 {
		score -= 15 // 高波动率增加恐慌
	} else if indicators != nil && indicators.HistoricalVol < 20 {
		score += 10 // 低波动率增加贪婪
	}
	
	// 成交量影响 (25%)
	if data.LongerTermContext != nil {
		volumeRatio := data.LongerTermContext.CurrentVolume / data.LongerTermContext.AverageVolume
		if volumeRatio > 1.5 {
			score += 15 // 高成交量
		} else if volumeRatio < 0.5 {
			score -= 10 // 低成交量
		}
	}
	
	// 确保在0-100范围内
	if score > 100 {
		score = 100
	} else if score < 0 {
		score = 0
	}
	
	return score
}

// analyzeVolumeStrength 分析成交量强度
func analyzeVolumeStrength(data *Data) string {
	if data.LongerTermContext == nil {
		return "normal"
	}
	
	ratio := data.LongerTermContext.CurrentVolume / data.LongerTermContext.AverageVolume
	
	if ratio > 1.5 {
		return "strong"
	} else if ratio < 0.7 {
		return "weak"
	} else {
		return "normal"
	}
}

// analyzeMomentumSignal 分析动量信号
func analyzeMomentumSignal(data *Data, indicators *EnhancedIndicators) string {
	bullishSignals := 0
	bearishSignals := 0
	
	// MACD信号
	if data.CurrentMACD > 0 {
		bullishSignals++
	} else {
		bearishSignals++
	}
	
	// RSI信号
	if data.CurrentRSI7 > 50 {
		bullishSignals++
	} else {
		bearishSignals++
	}
	
	// 价格趋势
	if data.PriceChange4h > 0 {
		bullishSignals++
	} else {
		bearishSignals++
	}
	
	// 布林带位置
	if indicators != nil && indicators.BollingerBands != nil {
		if indicators.BollingerBands.Position > 0.7 {
			bullishSignals++
		} else if indicators.BollingerBands.Position < 0.3 {
			bearishSignals++
		}
	}
	
	if bullishSignals > bearishSignals {
		return "bullish"
	} else if bearishSignals > bullishSignals {
		return "bearish"
	} else {
		return "neutral"
	}
}

// assessOverallSentiment 评估整体情绪
func assessOverallSentiment(fearGreedIndex int) string {
	if fearGreedIndex >= 80 {
		return "extreme_greed"
	} else if fearGreedIndex >= 60 {
		return "greed"
	} else if fearGreedIndex >= 40 {
		return "neutral"
	} else if fearGreedIndex >= 20 {
		return "fear"
	} else {
		return "extreme_fear"
	}
}

// 辅助函数
func getHighest(klines []Kline, period int) float64 {
	if len(klines) < period {
		period = len(klines)
	}
	
	highest := klines[len(klines)-period].High
	for i := len(klines) - period + 1; i < len(klines); i++ {
		if klines[i].High > highest {
			highest = klines[i].High
		}
	}
	return highest
}

func getLowest(klines []Kline, period int) float64 {
	if len(klines) < period {
		period = len(klines)
	}
	
	lowest := klines[len(klines)-period].Low
	for i := len(klines) - period + 1; i < len(klines); i++ {
		if klines[i].Low < lowest {
			lowest = klines[i].Low
		}
	}
	return lowest
}