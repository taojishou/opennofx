package market

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Data 市场数据结构
type Data struct {
	Symbol            string
	CurrentPrice      float64
	PriceChange1h     float64 // 1小时价格变化百分比
	PriceChange4h     float64 // 4小时价格变化百分比
	CurrentEMA20      float64
	CurrentMACD       float64
	CurrentRSI7       float64
	OpenInterest      *OIData
	FundingRate       float64
	IntradaySeries    *IntradayData
	LongerTermContext *LongerTermData
	AllTimeframes     []*TimeframeData // 所有配置的时间框架数据
}

// OIData Open Interest数据
type OIData struct {
	Latest  float64
	Average float64
}

// KlinePoint 完整K线数据点
type KlinePoint struct {
	Timestamp int64   // 时间戳（秒）
	Open      float64 // 开盘价
	High      float64 // 最高价
	Low       float64 // 最低价
	Close     float64 // 收盘价
	Volume    float64 // 成交量
	Change    float64 // 涨跌幅%
}

// IntradayData 日内数据(3分钟间隔)
type IntradayData struct {
	MidPrices   []float64     // 收盘价序列（保留兼容）
	EMA20Values []float64     // EMA20序列
	MACDValues  []float64     // MACD序列
	RSI7Values  []float64     // RSI7序列
	RSI14Values []float64     // RSI14序列
	Klines      []KlinePoint  // 完整K线数据（新增）
	HighestPrice float64      // 最高价
	LowestPrice  float64      // 最低价
	PriceRange   float64      // 价格区间
	Patterns     []string     // K线形态
}

// LongerTermData 长期数据(4小时时间框架)
type LongerTermData struct {
	Klines        []KlinePoint // K线数据
	EMA20         float64
	EMA50         float64
	ATR3          float64
	ATR14         float64
	CurrentVolume float64
	AverageVolume float64
	MACDValues    []float64
	RSI14Values   []float64
}

// TimeframeData 单个时间框架的完整数据
type TimeframeData struct {
	Interval      string       // 时间周期 (3m, 15m, 4h等)
	Limit         int          // 配置的K线数量
	ShowTable     bool         // 是否显示K线数据
	Klines        []KlinePoint // K线数据
	EMA20         float64
	EMA50         float64
	MACD          float64
	RSI7          float64
	RSI14         float64
	ATR3          float64
	ATR14         float64
	CurrentVolume float64
	AverageVolume float64
	Patterns      []string // K线形态
}

// Kline K线数据
type Kline struct {
	OpenTime  int64
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
	CloseTime int64
}

// KlineSettings K线配置（避免循环依赖，不直接使用config包）
type KlineSettings struct {
	Interval  string // "3m", "5m", "15m", "1h", "4h", "1d"
	Limit     int    // 显示多少根K线
	ShowTable bool   // 是否显示K线表格
}

var (
	// 默认K线配置（可被外部覆盖）
	DefaultKlineSettings = []KlineSettings{
		{Interval: "3m", Limit: 20, ShowTable: true},
		{Interval: "4h", Limit: 60, ShowTable: false},
	}
)

// SetKlineSettings 设置K线配置（由main函数在启动时调用）
func SetKlineSettings(settings []KlineSettings) {
	if len(settings) > 0 {
		DefaultKlineSettings = settings
		log.Printf("[Market] DefaultKlineSettings 已更新为 %d 个配置", len(DefaultKlineSettings))
		for i, s := range DefaultKlineSettings {
			log.Printf("[Market] [%d] %s × %d根 (显示表格: %v)", i, s.Interval, s.Limit, s.ShowTable)
		}
	}
}

// getIntervalMinutes 获取时间周期对应的分钟数
func getIntervalMinutes(interval string) int {
	minutes := map[string]int{
		"1m":  1,
		"3m":  3,
		"5m":  5,
		"15m": 15,
		"30m": 30,
		"1h":  60,
		"2h":  120,
		"4h":  240,
		"6h":  360,
		"8h":  480,
		"12h": 720,
		"1d":  1440,
	}
	if m, ok := minutes[interval]; ok {
		return m
	}
	return 1 // 默认1分钟
}

// getIntervalName 获取时间周期的可读名称
func getIntervalName(interval string) string {
	names := map[string]string{
		"1m":  "1分钟",
		"3m":  "3分钟",
		"5m":  "5分钟",
		"15m": "15分钟",
		"30m": "30分钟",
		"1h":  "1小时",
		"2h":  "2小时",
		"4h":  "4小时",
		"6h":  "6小时",
		"12h": "12小时",
		"1d":  "1天",
	}
	if name, ok := names[interval]; ok {
		return name
	}
	return interval
}

// Get 获取指定代币的市场数据
func Get(symbol string) (*Data, error) {
	// 标准化symbol
	symbol = Normalize(symbol)

	// 根据配置获取K线数据（第一个配置作为短期，第二个作为长期）
	var klines3m, klines4h []Kline
	var err error

	if len(DefaultKlineSettings) > 0 {
		// 短期K线
		shortTerm := DefaultKlineSettings[0]
		klines3m, err = getKlines(symbol, shortTerm.Interval, shortTerm.Limit+20) // 多获取20根用于计算指标
		if err != nil {
			return nil, fmt.Errorf("获取%s K线失败: %v", shortTerm.Interval, err)
		}
	} else {
		// fallback 到默认值
		klines3m, err = getKlines(symbol, "3m", 40)
		if err != nil {
			return nil, fmt.Errorf("获取3分钟K线失败: %v", err)
		}
	}

	if len(DefaultKlineSettings) > 1 {
		// 长期K线
		longTerm := DefaultKlineSettings[1]
		klines4h, err = getKlines(symbol, longTerm.Interval, longTerm.Limit)
		if err != nil {
			return nil, fmt.Errorf("获取%s K线失败: %v", longTerm.Interval, err)
		}
	} else {
		// fallback 到默认值
		klines4h, err = getKlines(symbol, "4h", 60)
		if err != nil {
			return nil, fmt.Errorf("获取4小时K线失败: %v", err)
		}
	}

	// 计算当前指标 (基于3分钟最新数据)
	currentPrice := klines3m[len(klines3m)-1].Close
	currentEMA20 := calculateEMA(klines3m, 20)
	currentMACD := calculateMACD(klines3m)
	currentRSI7 := calculateRSI(klines3m, 7)

	// 计算价格变化百分比
	// 1小时价格变化 = 20个3分钟K线前的价格
	priceChange1h := 0.0
	if len(klines3m) >= 21 { // 至少需要21根K线 (当前 + 20根前)
		price1hAgo := klines3m[len(klines3m)-21].Close
		if price1hAgo > 0 {
			priceChange1h = ((currentPrice - price1hAgo) / price1hAgo) * 100
		}
	}

	// 4小时价格变化 = 1个4小时K线前的价格
	priceChange4h := 0.0
	if len(klines4h) >= 2 {
		price4hAgo := klines4h[len(klines4h)-2].Close
		if price4hAgo > 0 {
			priceChange4h = ((currentPrice - price4hAgo) / price4hAgo) * 100
		}
	}

	// 获取OI数据
	oiData, err := getOpenInterestData(symbol)
	if err != nil {
		// OI失败不影响整体,使用默认值
		oiData = &OIData{Latest: 0, Average: 0}
	}

	// 获取Funding Rate
	fundingRate, _ := getFundingRate(symbol)

	// 计算日内系列数据
	intradayData := calculateIntradaySeries(klines3m)

	// 计算长期数据
	longerTermData := calculateLongerTermData(klines4h)

	// 获取所有配置的时间框架数据
	allTimeframes := make([]*TimeframeData, 0, len(DefaultKlineSettings))
	for _, setting := range DefaultKlineSettings {
		tfData, err := fetchTimeframeData(symbol, setting)
		if err != nil {
			log.Printf("⚠️ 获取%s时间框架数据失败: %v", setting.Interval, err)
			continue
		}
		allTimeframes = append(allTimeframes, tfData)
	}

	return &Data{
		Symbol:            symbol,
		CurrentPrice:      currentPrice,
		PriceChange1h:     priceChange1h,
		PriceChange4h:     priceChange4h,
		CurrentEMA20:      currentEMA20,
		CurrentMACD:       currentMACD,
		CurrentRSI7:       currentRSI7,
		OpenInterest:      oiData,
		FundingRate:       fundingRate,
		IntradaySeries:    intradayData,
		LongerTermContext: longerTermData,
		AllTimeframes:     allTimeframes,
	}, nil
}

// fetchTimeframeData 获取单个时间框架的完整数据
func fetchTimeframeData(symbol string, setting KlineSettings) (*TimeframeData, error) {
	// 获取K线数据（多获取20根用于计算指标）
	klines, err := getKlines(symbol, setting.Interval, setting.Limit+20)
	if err != nil {
		return nil, err
	}
	
	if len(klines) == 0 {
		return nil, fmt.Errorf("没有获取到K线数据")
	}
	
	tfData := &TimeframeData{
		Interval:  setting.Interval,
		Limit:     setting.Limit,
		ShowTable: setting.ShowTable,
		Klines:    make([]KlinePoint, 0, len(klines)),
	}
	
	// 转换K线数据格式
	for _, k := range klines {
		change := 0.0
		if k.Open != 0 {
			change = (k.Close - k.Open) / k.Open * 100
		}
		tfData.Klines = append(tfData.Klines, KlinePoint{
			Timestamp: k.OpenTime / 1000,
			Open:      k.Open,
			High:      k.High,
			Low:       k.Low,
			Close:     k.Close,
			Volume:    k.Volume,
			Change:    change,
		})
	}
	
	// 计算技术指标
	tfData.EMA20 = calculateEMA(klines, 20)
	tfData.EMA50 = calculateEMA(klines, 50)
	tfData.MACD = calculateMACD(klines)
	tfData.RSI7 = calculateRSI(klines, 7)
	tfData.RSI14 = calculateRSI(klines, 14)
	tfData.ATR3 = calculateATR(klines, 3)
	tfData.ATR14 = calculateATR(klines, 14)
	
	// 计算成交量
	if len(klines) > 0 {
		tfData.CurrentVolume = klines[len(klines)-1].Volume
		sum := 0.0
		for _, k := range klines {
			sum += k.Volume
		}
		tfData.AverageVolume = sum / float64(len(klines))
	}
	
	// K线形态识别
	tfData.Patterns = identifyPatterns(klines)
	
	return tfData, nil
}

// getKlines 从Binance获取K线数据
func getKlines(symbol, interval string, limit int) ([]Kline, error) {
	url := fmt.Sprintf("https://fapi.binance.com/fapi/v1/klines?symbol=%s&interval=%s&limit=%d",
		symbol, interval, limit)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rawData [][]interface{}
	if err := json.Unmarshal(body, &rawData); err != nil {
		return nil, err
	}

	klines := make([]Kline, len(rawData))
	for i, item := range rawData {
		openTime := int64(item[0].(float64))
		open, _ := parseFloat(item[1])
		high, _ := parseFloat(item[2])
		low, _ := parseFloat(item[3])
		close, _ := parseFloat(item[4])
		volume, _ := parseFloat(item[5])
		closeTime := int64(item[6].(float64))

		klines[i] = Kline{
			OpenTime:  openTime,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    volume,
			CloseTime: closeTime,
		}
	}

	return klines, nil
}

// calculateEMA 计算EMA
func calculateEMA(klines []Kline, period int) float64 {
	if len(klines) < period {
		return 0
	}

	// 计算SMA作为初始EMA
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += klines[i].Close
	}
	ema := sum / float64(period)

	// 计算EMA
	multiplier := 2.0 / float64(period+1)
	for i := period; i < len(klines); i++ {
		ema = (klines[i].Close-ema)*multiplier + ema
	}

	return ema
}

// calculateMACD 计算MACD
func calculateMACD(klines []Kline) float64 {
	if len(klines) < 26 {
		return 0
	}

	// 计算12期和26期EMA
	ema12 := calculateEMA(klines, 12)
	ema26 := calculateEMA(klines, 26)

	// MACD = EMA12 - EMA26
	return ema12 - ema26
}

// calculateRSI 计算RSI
func calculateRSI(klines []Kline, period int) float64 {
	if len(klines) <= period {
		return 0
	}

	gains := 0.0
	losses := 0.0

	// 计算初始平均涨跌幅
	for i := 1; i <= period; i++ {
		change := klines[i].Close - klines[i-1].Close
		if change > 0 {
			gains += change
		} else {
			losses += -change
		}
	}

	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)

	// 使用Wilder平滑方法计算后续RSI
	for i := period + 1; i < len(klines); i++ {
		change := klines[i].Close - klines[i-1].Close
		if change > 0 {
			avgGain = (avgGain*float64(period-1) + change) / float64(period)
			avgLoss = (avgLoss * float64(period-1)) / float64(period)
		} else {
			avgGain = (avgGain * float64(period-1)) / float64(period)
			avgLoss = (avgLoss*float64(period-1) + (-change)) / float64(period)
		}
	}

	if avgLoss == 0 {
		return 100
	}

	rs := avgGain / avgLoss
	rsi := 100 - (100 / (1 + rs))

	return rsi
}

// calculateATR 计算ATR
func calculateATR(klines []Kline, period int) float64 {
	if len(klines) <= period {
		return 0
	}

	trs := make([]float64, len(klines))
	for i := 1; i < len(klines); i++ {
		high := klines[i].High
		low := klines[i].Low
		prevClose := klines[i-1].Close

		tr1 := high - low
		tr2 := math.Abs(high - prevClose)
		tr3 := math.Abs(low - prevClose)

		trs[i] = math.Max(tr1, math.Max(tr2, tr3))
	}

	// 计算初始ATR
	sum := 0.0
	for i := 1; i <= period; i++ {
		sum += trs[i]
	}
	atr := sum / float64(period)

	// Wilder平滑
	for i := period + 1; i < len(klines); i++ {
		atr = (atr*float64(period-1) + trs[i]) / float64(period)
	}

	return atr
}

// calculateIntradaySeries 计算日内系列数据
func calculateIntradaySeries(klines []Kline) *IntradayData {
	data := &IntradayData{
		MidPrices:   make([]float64, 0, 20),
		EMA20Values: make([]float64, 0, 20),
		MACDValues:  make([]float64, 0, 20),
		RSI7Values:  make([]float64, 0, 20),
		RSI14Values: make([]float64, 0, 20),
		Klines:      make([]KlinePoint, 0, 20),
	}

	// 获取最近20个数据点（1小时数据）
	start := len(klines) - 20
	if start < 0 {
		start = 0
	}
	
	// 初始化最高最低价
	data.HighestPrice = 0
	data.LowestPrice = 999999999

	for i := start; i < len(klines); i++ {
		data.MidPrices = append(data.MidPrices, klines[i].Close)
		
		// 计算涨跌幅
		change := 0.0
		if i > 0 {
			change = (klines[i].Close - klines[i-1].Close) / klines[i-1].Close * 100
		}
		
		// 添加完整K线数据
		data.Klines = append(data.Klines, KlinePoint{
			Timestamp: klines[i].OpenTime / 1000, // 转为秒
			Open:      klines[i].Open,
			High:      klines[i].High,
			Low:       klines[i].Low,
			Close:     klines[i].Close,
			Volume:    klines[i].Volume,
			Change:    change,
		})
		
		// 更新最高最低价
		if klines[i].High > data.HighestPrice {
			data.HighestPrice = klines[i].High
		}
		if klines[i].Low < data.LowestPrice {
			data.LowestPrice = klines[i].Low
		}

		// 计算每个点的EMA20
		if i >= 19 {
			ema20 := calculateEMA(klines[:i+1], 20)
			data.EMA20Values = append(data.EMA20Values, ema20)
		}

		// 计算每个点的MACD
		if i >= 25 {
			macd := calculateMACD(klines[:i+1])
			data.MACDValues = append(data.MACDValues, macd)
		}

		// 计算每个点的RSI
		if i >= 7 {
			rsi7 := calculateRSI(klines[:i+1], 7)
			data.RSI7Values = append(data.RSI7Values, rsi7)
		}
		if i >= 14 {
			rsi14 := calculateRSI(klines[:i+1], 14)
			data.RSI14Values = append(data.RSI14Values, rsi14)
		}
	}
	
	// 计算价格区间
	data.PriceRange = data.HighestPrice - data.LowestPrice
	
	// 识别K线形态
	data.Patterns = identifyPatterns(klines[start:])

	return data
}

// calculateLongerTermData 计算长期数据
func calculateLongerTermData(klines []Kline) *LongerTermData {
	data := &LongerTermData{
		Klines:      make([]KlinePoint, 0, len(klines)),
		MACDValues:  make([]float64, 0, 10),
		RSI14Values: make([]float64, 0, 10),
	}
	
	// 转换K线数据格式
	for _, k := range klines {
		change := 0.0
		if k.Open != 0 {
			change = (k.Close - k.Open) / k.Open * 100
		}
		data.Klines = append(data.Klines, KlinePoint{
			Timestamp: k.OpenTime / 1000,
			Open:      k.Open,
			High:      k.High,
			Low:       k.Low,
			Close:     k.Close,
			Volume:    k.Volume,
			Change:    change,
		})
	}

	// 计算EMA
	data.EMA20 = calculateEMA(klines, 20)
	data.EMA50 = calculateEMA(klines, 50)

	// 计算ATR
	data.ATR3 = calculateATR(klines, 3)
	data.ATR14 = calculateATR(klines, 14)

	// 计算成交量
	if len(klines) > 0 {
		data.CurrentVolume = klines[len(klines)-1].Volume
		// 计算平均成交量
		sum := 0.0
		for _, k := range klines {
			sum += k.Volume
		}
		data.AverageVolume = sum / float64(len(klines))
	}

	// 计算MACD和RSI序列
	start := len(klines) - 10
	if start < 0 {
		start = 0
	}

	for i := start; i < len(klines); i++ {
		if i >= 25 {
			macd := calculateMACD(klines[:i+1])
			data.MACDValues = append(data.MACDValues, macd)
		}
		if i >= 14 {
			rsi14 := calculateRSI(klines[:i+1], 14)
			data.RSI14Values = append(data.RSI14Values, rsi14)
		}
	}

	return data
}

// getOpenInterestData 获取OI数据
func getOpenInterestData(symbol string) (*OIData, error) {
	url := fmt.Sprintf("https://fapi.binance.com/fapi/v1/openInterest?symbol=%s", symbol)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		OpenInterest string `json:"openInterest"`
		Symbol       string `json:"symbol"`
		Time         int64  `json:"time"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	oi, _ := strconv.ParseFloat(result.OpenInterest, 64)

	return &OIData{
		Latest:  oi,
		Average: oi * 0.999, // 近似平均值
	}, nil
}

// getFundingRate 获取资金费率
func getFundingRate(symbol string) (float64, error) {
	url := fmt.Sprintf("https://fapi.binance.com/fapi/v1/premiumIndex?symbol=%s", symbol)

	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var result struct {
		Symbol          string `json:"symbol"`
		MarkPrice       string `json:"markPrice"`
		IndexPrice      string `json:"indexPrice"`
		LastFundingRate string `json:"lastFundingRate"`
		NextFundingTime int64  `json:"nextFundingTime"`
		InterestRate    string `json:"interestRate"`
		Time            int64  `json:"time"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return 0, err
	}

	rate, _ := strconv.ParseFloat(result.LastFundingRate, 64)
	return rate, nil
}

// Format 格式化输出市场数据
func Format(data *Data) string {
	return FormatWithKlineTable(data, true)
}

// FormatSimple 格式化市场数据为字符串（不包含K线表格，用于候选币种）
func FormatSimple(data *Data) string {
	return FormatWithKlineTable(data, false)
}

// FormatWithKlineTable 格式化市场数据，可选是否包含K线表格
func FormatWithKlineTable(data *Data, showKlineTable bool) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("current_price = %.2f, current_ema20 = %.3f, current_macd = %.3f, current_rsi (7 period) = %.3f\n\n",
		data.CurrentPrice, data.CurrentEMA20, data.CurrentMACD, data.CurrentRSI7))

	sb.WriteString(fmt.Sprintf("In addition, here is the latest %s open interest and funding rate for perps:\n\n",
		data.Symbol))

	if data.OpenInterest != nil {
		sb.WriteString(fmt.Sprintf("Open Interest: Latest: %.2f Average: %.2f\n\n",
			data.OpenInterest.Latest, data.OpenInterest.Average))
	}

	sb.WriteString(fmt.Sprintf("Funding Rate: %.2e\n\n", data.FundingRate))

	if data.IntradaySeries != nil {
		// 获取短期K线配置
		shortTerm := DefaultKlineSettings[0]
		intervalName := getIntervalName(shortTerm.Interval)
		
		sb.WriteString(fmt.Sprintf("Intraday series (%s intervals, oldest → latest):\n\n", intervalName))
		
		// 输出K线数据（根据配置决定，且调用方允许显示）
		if len(data.IntradaySeries.Klines) > 0 && shortTerm.ShowTable && showKlineTable {
			// 只显示配置数量的K线（数据里有更多用于计算指标）
			displayCount := shortTerm.Limit
			if displayCount > len(data.IntradaySeries.Klines) {
				displayCount = len(data.IntradaySeries.Klines)
			}
			startIdx := len(data.IntradaySeries.Klines) - displayCount
			
			// 提取时间范围
			startTime := formatTimestamp(data.IntradaySeries.Klines[startIdx].Timestamp)
			endTime := formatTimestamp(data.IntradaySeries.Klines[len(data.IntradaySeries.Klines)-1].Timestamp)
			
			sb.WriteString(fmt.Sprintf("**%sK线数据**（最近%d根，%s~%s）:\n\n", intervalName, displayCount, startTime, endTime))
			
			// 提取各项数据为数组
			opens := make([]string, displayCount)
			highs := make([]string, displayCount)
			lows := make([]string, displayCount)
			closes := make([]string, displayCount)
			changes := make([]string, displayCount)
			volumes := make([]string, displayCount)
			
			for i, idx := 0, startIdx; idx < len(data.IntradaySeries.Klines); i, idx = i+1, idx+1 {
				kline := data.IntradaySeries.Klines[idx]
				opens[i] = fmt.Sprintf("%.2f", kline.Open)
				highs[i] = fmt.Sprintf("%.2f", kline.High)
				lows[i] = fmt.Sprintf("%.2f", kline.Low)
				closes[i] = fmt.Sprintf("%.2f", kline.Close)
				changes[i] = fmt.Sprintf("%+.2f%%", kline.Change)
				volumes[i] = fmt.Sprintf("%.0f", kline.Volume)
			}
			
			sb.WriteString(fmt.Sprintf("Open: [%s]\n", strings.Join(opens, ", ")))
			sb.WriteString(fmt.Sprintf("High: [%s]\n", strings.Join(highs, ", ")))
			sb.WriteString(fmt.Sprintf("Low: [%s]\n", strings.Join(lows, ", ")))
			sb.WriteString(fmt.Sprintf("Close: [%s]\n", strings.Join(closes, ", ")))
			sb.WriteString(fmt.Sprintf("Change: [%s]\n", strings.Join(changes, ", ")))
			sb.WriteString(fmt.Sprintf("Volume: [%s]\n\n", strings.Join(volumes, ", ")))
		}
		
		// K线形态识别
		if len(data.IntradaySeries.Patterns) > 0 {
			sb.WriteString(fmt.Sprintf("**K线形态识别**: 检测到 %d 个信号\n", len(data.IntradaySeries.Patterns)))
			for i, pattern := range data.IntradaySeries.Patterns {
				sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, pattern))
			}
			sb.WriteString("\n")
		}

		// 技术指标序列（保持原有格式，便于AI分析）
		sb.WriteString("**技术指标序列**:\n\n")
		
		if len(data.IntradaySeries.MidPrices) > 0 {
			sb.WriteString(fmt.Sprintf("Mid prices: %s\n\n", formatFloatSlice(data.IntradaySeries.MidPrices)))
		}

		if len(data.IntradaySeries.EMA20Values) > 0 {
			sb.WriteString(fmt.Sprintf("EMA indicators (20‑period): %s\n\n", formatFloatSlice(data.IntradaySeries.EMA20Values)))
		}

		if len(data.IntradaySeries.MACDValues) > 0 {
			sb.WriteString(fmt.Sprintf("MACD indicators: %s\n\n", formatFloatSlice(data.IntradaySeries.MACDValues)))
		}

		if len(data.IntradaySeries.RSI7Values) > 0 {
			sb.WriteString(fmt.Sprintf("RSI indicators (7‑Period): %s\n\n", formatFloatSlice(data.IntradaySeries.RSI7Values)))
		}

		if len(data.IntradaySeries.RSI14Values) > 0 {
			sb.WriteString(fmt.Sprintf("RSI indicators (14‑Period): %s\n\n", formatFloatSlice(data.IntradaySeries.RSI14Values)))
		}
	}

	if data.LongerTermContext != nil && len(DefaultKlineSettings) > 1 {
		longTerm := DefaultKlineSettings[1]
		intervalName := getIntervalName(longTerm.Interval)
		
		sb.WriteString(fmt.Sprintf("Longer‑term context (%s timeframe):\n\n", intervalName))

		// 如果配置了显示K线数据，则显示
		if len(data.LongerTermContext.Klines) > 0 && longTerm.ShowTable && showKlineTable {
			displayCount := longTerm.Limit
			if displayCount > len(data.LongerTermContext.Klines) {
				displayCount = len(data.LongerTermContext.Klines)
			}
			startIdx := len(data.LongerTermContext.Klines) - displayCount
			
			startTime := formatTimestamp(data.LongerTermContext.Klines[startIdx].Timestamp)
			endTime := formatTimestamp(data.LongerTermContext.Klines[len(data.LongerTermContext.Klines)-1].Timestamp)
			
			sb.WriteString(fmt.Sprintf("**%sK线数据**（最近%d根，%s~%s）:\n\n", intervalName, displayCount, startTime, endTime))
			
			opens := make([]string, displayCount)
			highs := make([]string, displayCount)
			lows := make([]string, displayCount)
			closes := make([]string, displayCount)
			changes := make([]string, displayCount)
			
			for i, idx := 0, startIdx; idx < len(data.LongerTermContext.Klines); i, idx = i+1, idx+1 {
				kline := data.LongerTermContext.Klines[idx]
				opens[i] = fmt.Sprintf("%.2f", kline.Open)
				highs[i] = fmt.Sprintf("%.2f", kline.High)
				lows[i] = fmt.Sprintf("%.2f", kline.Low)
				closes[i] = fmt.Sprintf("%.2f", kline.Close)
				changes[i] = fmt.Sprintf("%+.2f%%", kline.Change)
			}
			
			sb.WriteString(fmt.Sprintf("Open: [%s]\n", strings.Join(opens, ", ")))
			sb.WriteString(fmt.Sprintf("High: [%s]\n", strings.Join(highs, ", ")))
			sb.WriteString(fmt.Sprintf("Low: [%s]\n", strings.Join(lows, ", ")))
			sb.WriteString(fmt.Sprintf("Close: [%s]\n", strings.Join(closes, ", ")))
			sb.WriteString(fmt.Sprintf("Change: [%s]\n\n", strings.Join(changes, ", ")))
		}

		// 技术指标
		sb.WriteString(fmt.Sprintf("20‑Period EMA: %.3f vs. 50‑Period EMA: %.3f\n\n",
			data.LongerTermContext.EMA20, data.LongerTermContext.EMA50))

		sb.WriteString(fmt.Sprintf("3‑Period ATR: %.3f vs. 14‑Period ATR: %.3f\n\n",
			data.LongerTermContext.ATR3, data.LongerTermContext.ATR14))

		sb.WriteString(fmt.Sprintf("Current Volume: %.3f vs. Average Volume: %.3f\n\n",
			data.LongerTermContext.CurrentVolume, data.LongerTermContext.AverageVolume))

		if len(data.LongerTermContext.MACDValues) > 0 {
			sb.WriteString(fmt.Sprintf("MACD indicators: %s\n\n", formatFloatSlice(data.LongerTermContext.MACDValues)))
		}

		if len(data.LongerTermContext.RSI14Values) > 0 {
			sb.WriteString(fmt.Sprintf("RSI indicators (14‑Period): %s\n\n", formatFloatSlice(data.LongerTermContext.RSI14Values)))
		}
	}
	
	// 输出所有配置的时间框架K线数据（如果show_table为true）
	if len(data.AllTimeframes) > 0 && showKlineTable {
		sb.WriteString("## 📊 配置的时间框架K线数据\n\n")
		
		for _, tf := range data.AllTimeframes {
			if !tf.ShowTable || len(tf.Klines) == 0 {
				continue
			}
			
			intervalName := getIntervalName(tf.Interval)
			displayCount := tf.Limit
			if displayCount > len(tf.Klines) {
				displayCount = len(tf.Klines)
			}
			startIdx := len(tf.Klines) - displayCount
			
			startTime := formatTimestamp(tf.Klines[startIdx].Timestamp)
			endTime := formatTimestamp(tf.Klines[len(tf.Klines)-1].Timestamp)
			
			// 根据K线数量和周期计算时间跨度
			timeSpanMinutes := getIntervalMinutes(tf.Interval) * displayCount
			timeSpanStr := ""
			if timeSpanMinutes >= 60 {
				hours := timeSpanMinutes / 60
				mins := timeSpanMinutes % 60
				if mins > 0 {
					timeSpanStr = fmt.Sprintf("%d小时%d分钟", hours, mins)
				} else {
					timeSpanStr = fmt.Sprintf("%d小时", hours)
				}
			} else {
				timeSpanStr = fmt.Sprintf("%d分钟", timeSpanMinutes)
			}
			
			sb.WriteString(fmt.Sprintf("### %s K线 - %d根/%s数据（%s~%s）\n\n", intervalName, displayCount, timeSpanStr, startTime, endTime))
			
			// 提取数据为数组
			opens := make([]string, displayCount)
			highs := make([]string, displayCount)
			lows := make([]string, displayCount)
			closes := make([]string, displayCount)
			changes := make([]string, displayCount)
			volumes := make([]string, displayCount)
			
			for i, idx := 0, startIdx; idx < len(tf.Klines); i, idx = i+1, idx+1 {
				kline := tf.Klines[idx]
				opens[i] = fmt.Sprintf("%.2f", kline.Open)
				highs[i] = fmt.Sprintf("%.2f", kline.High)
				lows[i] = fmt.Sprintf("%.2f", kline.Low)
				closes[i] = fmt.Sprintf("%.2f", kline.Close)
				changes[i] = fmt.Sprintf("%+.2f%%", kline.Change)
				volumes[i] = fmt.Sprintf("%.0f", kline.Volume)
			}
			
			sb.WriteString(fmt.Sprintf("Open: [%s]\n", strings.Join(opens, ", ")))
			sb.WriteString(fmt.Sprintf("High: [%s]\n", strings.Join(highs, ", ")))
			sb.WriteString(fmt.Sprintf("Low: [%s]\n", strings.Join(lows, ", ")))
			sb.WriteString(fmt.Sprintf("Close: [%s]\n", strings.Join(closes, ", ")))
			sb.WriteString(fmt.Sprintf("Change: [%s]\n", strings.Join(changes, ", ")))
			sb.WriteString(fmt.Sprintf("Volume: [%s]\n", strings.Join(volumes, ", ")))
			
			// 技术指标
			sb.WriteString(fmt.Sprintf("\n**技术指标**: EMA20=%.2f, EMA50=%.2f, MACD=%.3f, RSI7=%.1f, RSI14=%.1f, ATR14=%.3f\n\n",
				tf.EMA20, tf.EMA50, tf.MACD, tf.RSI7, tf.RSI14, tf.ATR14))
			
			// K线形态
			if len(tf.Patterns) > 0 {
				sb.WriteString(fmt.Sprintf("**形态**: %s\n\n", strings.Join(tf.Patterns, ", ")))
			}
		}
	}

	return sb.String()
}

// formatFloatSlice 格式化float64切片为字符串
func formatFloatSlice(values []float64) string {
	strValues := make([]string, len(values))
	for i, v := range values {
		strValues[i] = fmt.Sprintf("%.3f", v)
	}
	return "[" + strings.Join(strValues, ", ") + "]"
}

// formatTimestamp 格式化时间戳为可读时间
func formatTimestamp(timestamp int64) string {
	t := time.Unix(timestamp, 0)
	// 只显示时:分，更简洁
	return t.Format("15:04")
}

// identifyPatterns 识别K线形态
func identifyPatterns(klines []Kline) []string {
	patterns := []string{}
	
	if len(klines) < 3 {
		return patterns
	}
	
	last := klines[len(klines)-1]
	prev := klines[len(klines)-2]
	
	// 锤子线（看涨）
	if isHammer(last) {
		patterns = append(patterns, "🔨 锤子线（看涨信号）")
	}
	
	// 倒锤子（潜在反转）
	if isInvertedHammer(last) {
		patterns = append(patterns, "🔨 倒锤子（潜在反转）")
	}
	
	// 看涨吞没
	if isBullishEngulfing(prev, last) {
		patterns = append(patterns, "📈 看涨吞没（强烈看涨）")
	}
	
	// 看跌吞没
	if isBearishEngulfing(prev, last) {
		patterns = append(patterns, "📉 看跌吞没（强烈看跌）")
	}
	
	// 十字星（犹豫）
	if isDoji(last) {
		patterns = append(patterns, "✨ 十字星（方向不明）")
	}
	
	// 射击之星（看跌）
	if isShootingStar(last) {
		patterns = append(patterns, "💫 射击之星（看跌信号）")
	}
	
	// 三连阳
	if len(klines) >= 3 {
		prev2 := klines[len(klines)-3]
		if isThreeWhiteSoldiers(prev2, prev, last) {
			patterns = append(patterns, "🚀 三连阳（强势上涨）")
		}
		
		// 三连阴
		if isThreeBlackCrows(prev2, prev, last) {
			patterns = append(patterns, "💀 三连阴（强势下跌）")
		}
	}
	
	return patterns
}

// isHammer 判断是否为锤子线
func isHammer(k Kline) bool {
	body := math.Abs(k.Close - k.Open)
	upperShadow := k.High - math.Max(k.Open, k.Close)
	lowerShadow := math.Min(k.Open, k.Close) - k.Low
	totalRange := k.High - k.Low
	
	if totalRange == 0 {
		return false
	}
	
	// 下影线至少是实体的2倍，上影线很短，实体在上部
	return lowerShadow > body*2 && upperShadow < body*0.5 && body/totalRange < 0.3
}

// isInvertedHammer 判断是否为倒锤子线
func isInvertedHammer(k Kline) bool {
	body := math.Abs(k.Close - k.Open)
	upperShadow := k.High - math.Max(k.Open, k.Close)
	lowerShadow := math.Min(k.Open, k.Close) - k.Low
	totalRange := k.High - k.Low
	
	if totalRange == 0 {
		return false
	}
	
	// 上影线至少是实体的2倍，下影线很短，实体在下部
	return upperShadow > body*2 && lowerShadow < body*0.5 && body/totalRange < 0.3
}

// isShootingStar 判断是否为射击之星
func isShootingStar(k Kline) bool {
	body := math.Abs(k.Close - k.Open)
	upperShadow := k.High - math.Max(k.Open, k.Close)
	lowerShadow := math.Min(k.Open, k.Close) - k.Low
	totalRange := k.High - k.Low
	
	if totalRange == 0 {
		return false
	}
	
	// 上影线很长，实体小，下影线很短，且收盘价接近最低价
	isRedCandle := k.Close < k.Open
	return upperShadow > body*2 && lowerShadow < body*0.3 && body/totalRange < 0.3 && isRedCandle
}

// isDoji 判断是否为十字星
func isDoji(k Kline) bool {
	body := math.Abs(k.Close - k.Open)
	totalRange := k.High - k.Low
	
	if totalRange == 0 {
		return false
	}
	
	// 实体非常小（< 10%的总区间）
	return body/totalRange < 0.1
}

// isBullishEngulfing 判断是否为看涨吞没
func isBullishEngulfing(prev, curr Kline) bool {
	prevIsRed := prev.Close < prev.Open
	currIsGreen := curr.Close > curr.Open
	
	// 前一根是阴线，当前是阳线，且当前完全吞没前一根
	return prevIsRed && currIsGreen && 
		curr.Open < prev.Close && 
		curr.Close > prev.Open
}

// isBearishEngulfing 判断是否为看跌吞没
func isBearishEngulfing(prev, curr Kline) bool {
	prevIsGreen := prev.Close > prev.Open
	currIsRed := curr.Close < curr.Open
	
	// 前一根是阳线，当前是阴线，且当前完全吞没前一根
	return prevIsGreen && currIsRed && 
		curr.Open > prev.Close && 
		curr.Close < prev.Open
}

// isThreeWhiteSoldiers 判断是否为三连阳
func isThreeWhiteSoldiers(k1, k2, k3 Kline) bool {
	// 三根都是阳线
	all3Green := k1.Close > k1.Open && k2.Close > k2.Open && k3.Close > k3.Open
	
	// 收盘价逐步升高
	ascending := k2.Close > k1.Close && k3.Close > k2.Close
	
	// 每根K线的涨幅相似（避免单根暴涨）
	gain1 := (k1.Close - k1.Open) / k1.Open
	gain2 := (k2.Close - k2.Open) / k2.Open
	gain3 := (k3.Close - k3.Open) / k3.Open
	
	avgGain := (gain1 + gain2 + gain3) / 3
	consistent := math.Abs(gain1-avgGain) < avgGain*0.5 &&
		math.Abs(gain2-avgGain) < avgGain*0.5 &&
		math.Abs(gain3-avgGain) < avgGain*0.5
	
	return all3Green && ascending && consistent
}

// isThreeBlackCrows 判断是否为三连阴
func isThreeBlackCrows(k1, k2, k3 Kline) bool {
	// 三根都是阴线
	all3Red := k1.Close < k1.Open && k2.Close < k2.Open && k3.Close < k3.Open
	
	// 收盘价逐步降低
	descending := k2.Close < k1.Close && k3.Close < k2.Close
	
	// 每根K线的跌幅相似
	loss1 := (k1.Open - k1.Close) / k1.Open
	loss2 := (k2.Open - k2.Close) / k2.Open
	loss3 := (k3.Open - k3.Close) / k3.Open
	
	avgLoss := (loss1 + loss2 + loss3) / 3
	consistent := math.Abs(loss1-avgLoss) < avgLoss*0.5 &&
		math.Abs(loss2-avgLoss) < avgLoss*0.5 &&
		math.Abs(loss3-avgLoss) < avgLoss*0.5
	
	return all3Red && descending && consistent
}

// Normalize 标准化symbol,确保是USDT交易对
func Normalize(symbol string) string {
	symbol = strings.ToUpper(symbol)
	if strings.HasSuffix(symbol, "USDT") {
		return symbol
	}
	return symbol + "USDT"
}

// parseFloat 解析float值
func parseFloat(v interface{}) (float64, error) {
	switch val := v.(type) {
	case string:
		return strconv.ParseFloat(val, 64)
	case float64:
		return val, nil
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	default:
		return 0, fmt.Errorf("unsupported type: %T", v)
	}
}
