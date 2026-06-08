package service

import (
	"sort"
	"strings"
)

const (
	apimartGPTImage2Model                     = "gpt-image-2"
	apimartGPTImage2OfficialModel             = "gpt-image-2-official"
	apimartOpenAIBaseURLHost                  = "api.apimart.ai"
	apimartGPTImage2OfficialBalanceMultiplier = 7 * 1.2
	apimartGPTImage2OfficialDefaultPrice      = 0.2109
)

type apimartImagePriceRow struct {
	Size     string
	Quality  string
	Official float64
}

var apimartGPTImage2OfficialPriceRows = []apimartImagePriceRow{
	{Size: "1536x512", Quality: "low", Official: 0.0018},
	{Size: "512x1536", Quality: "auto", Official: 0.0018},
	{Size: "512x1536", Quality: "low", Official: 0.0018},
	{Size: "1024x3072", Quality: "auto", Official: 0.0032},
	{Size: "1024x3072", Quality: "low", Official: 0.0032},
	{Size: "3072x1024", Quality: "auto", Official: 0.0032},
	{Size: "3072x1024", Quality: "low", Official: 0.0032},
	{Size: "2016x864", Quality: "auto", Official: 0.0033},
	{Size: "2016x864", Quality: "low", Official: 0.0033},
	{Size: "864x2016", Quality: "auto", Official: 0.0033},
	{Size: "864x2016", Quality: "low", Official: 0.0033},
	{Size: "1536x864", Quality: "auto", Official: 0.0038},
	{Size: "1536x864", Quality: "low", Official: 0.0038},
	{Size: "864x1536", Quality: "auto", Official: 0.0038},
	{Size: "864x1536", Quality: "low", Official: 0.0038},
	{Size: "1774x887", Quality: "auto", Official: 0.0041},
	{Size: "1774x887", Quality: "low", Official: 0.0041},
	{Size: "887x1774", Quality: "auto", Official: 0.0041},
	{Size: "887x1774", Quality: "low", Official: 0.0041},
	{Size: "1915x821", Quality: "auto", Official: 0.0042},
	{Size: "1915x821", Quality: "low", Official: 0.0042},
	{Size: "821x1915", Quality: "auto", Official: 0.0042},
	{Size: "821x1915", Quality: "low", Official: 0.0042},
	{Size: "1024x1536", Quality: "auto", Official: 0.0043},
	{Size: "1024x1536", Quality: "low", Official: 0.0043},
	{Size: "1536x1024", Quality: "auto", Official: 0.0043},
	{Size: "1536x1024", Quality: "low", Official: 0.0043},
	{Size: "1881x836", Quality: "auto", Official: 0.0043},
	{Size: "1881x836", Quality: "low", Official: 0.0043},
	{Size: "836x1881", Quality: "auto", Official: 0.0043},
	{Size: "836x1881", Quality: "low", Official: 0.0043},
	{Size: "1024x2048", Quality: "auto", Official: 0.0045},
	{Size: "1024x2048", Quality: "low", Official: 0.0045},
	{Size: "2048x1024", Quality: "auto", Official: 0.0045},
	{Size: "2048x1024", Quality: "low", Official: 0.0045},
	{Size: "1024x768", Quality: "auto", Official: 0.0049},
	{Size: "1024x768", Quality: "low", Official: 0.0049},
	{Size: "768x1024", Quality: "auto", Official: 0.0049},
	{Size: "768x1024", Quality: "low", Official: 0.0049},
	{Size: "1672x941", Quality: "auto", Official: 0.0054},
	{Size: "1672x941", Quality: "low", Official: 0.0054},
	{Size: "941x1672", Quality: "auto", Official: 0.0054},
	{Size: "941x1672", Quality: "low", Official: 0.0054},
	{Size: "1448x1086", Quality: "auto", Official: 0.0056},
	{Size: "1448x1086", Quality: "low", Official: 0.0056},
	{Size: "1086x1448", Quality: "auto", Official: 0.0056},
	{Size: "1086x1448", Quality: "low", Official: 0.0056},
	{Size: "1024x1024", Quality: "auto", Official: 0.0061},
	{Size: "1024x1024", Quality: "low", Official: 0.0061},
	{Size: "1152x2688", Quality: "auto", Official: 0.0065},
	{Size: "1152x2688", Quality: "low", Official: 0.0065},
	{Size: "2688x1152", Quality: "auto", Official: 0.0065},
	{Size: "2688x1152", Quality: "low", Official: 0.0065},
	{Size: "1024x1280", Quality: "auto", Official: 0.0072},
	{Size: "1024x1280", Quality: "low", Official: 0.0072},
	{Size: "1280x1024", Quality: "auto", Official: 0.0072},
	{Size: "1280x1024", Quality: "low", Official: 0.0072},
	{Size: "1152x2048", Quality: "auto", Official: 0.0076},
	{Size: "1152x2048", Quality: "low", Official: 0.0076},
	{Size: "2048x1152", Quality: "auto", Official: 0.0076},
	{Size: "2048x1152", Quality: "low", Official: 0.0076},
	{Size: "1122x1402", Quality: "auto", Official: 0.0092},
	{Size: "1122x1402", Quality: "low", Official: 0.0092},
	{Size: "1402x1122", Quality: "auto", Official: 0.0092},
	{Size: "1402x1122", Quality: "low", Official: 0.0092},
	{Size: "1360x2048", Quality: "auto", Official: 0.0113},
	{Size: "1360x2048", Quality: "low", Official: 0.0113},
	{Size: "2048x1360", Quality: "auto", Official: 0.0113},
	{Size: "2048x1360", Quality: "low", Official: 0.0113},
	{Size: "2160x3840", Quality: "auto", Official: 0.0113},
	{Size: "2160x3840", Quality: "low", Official: 0.0113},
	{Size: "3840x2160", Quality: "auto", Official: 0.0113},
	{Size: "3840x2160", Quality: "low", Official: 0.0113},
	{Size: "2048x2048", Quality: "auto", Official: 0.0121},
	{Size: "2048x2048", Quality: "low", Official: 0.0121},
	{Size: "1648x3840", Quality: "auto", Official: 0.0136},
	{Size: "1648x3840", Quality: "low", Official: 0.0136},
	{Size: "3840x1648", Quality: "auto", Official: 0.0136},
	{Size: "3840x1648", Quality: "low", Official: 0.0136},
	{Size: "1280x3840", Quality: "auto", Official: 0.0149},
	{Size: "1280x3840", Quality: "low", Official: 0.0149},
	{Size: "3840x1280", Quality: "auto", Official: 0.0149},
	{Size: "3840x1280", Quality: "low", Official: 0.0149},
	{Size: "2576x3216", Quality: "auto", Official: 0.0162},
	{Size: "2576x3216", Quality: "low", Official: 0.0162},
	{Size: "3216x2576", Quality: "auto", Official: 0.0162},
	{Size: "3216x2576", Quality: "low", Official: 0.0162},
	{Size: "2880x2880", Quality: "auto", Official: 0.0199},
	{Size: "2880x2880", Quality: "low", Official: 0.0199},
	{Size: "1536x512", Quality: "auto", Official: 0.0162},
	{Size: "1536x512", Quality: "medium", Official: 0.0162},
	{Size: "512x1536", Quality: "medium", Official: 0.0162},
	{Size: "1024x3072", Quality: "medium", Official: 0.0272},
	{Size: "3072x1024", Quality: "medium", Official: 0.0272},
	{Size: "2016x864", Quality: "medium", Official: 0.0285},
	{Size: "864x2016", Quality: "medium", Official: 0.0285},
	{Size: "1536x864", Quality: "medium", Official: 0.0298},
	{Size: "864x1536", Quality: "medium", Official: 0.0298},
	{Size: "1774x887", Quality: "medium", Official: 0.0325},
	{Size: "887x1774", Quality: "medium", Official: 0.0325},
	{Size: "1915x821", Quality: "medium", Official: 0.0336},
	{Size: "821x1915", Quality: "medium", Official: 0.0336},
	{Size: "1024x1536", Quality: "medium", Official: 0.0356},
	{Size: "1536x1024", Quality: "medium", Official: 0.0356},
	{Size: "1881x836", Quality: "medium", Official: 0.0363},
	{Size: "836x1881", Quality: "medium", Official: 0.0363},
	{Size: "1024x2048", Quality: "medium", Official: 0.0387},
	{Size: "2048x1024", Quality: "medium", Official: 0.0387},
	{Size: "1024x768", Quality: "medium", Official: 0.04},
	{Size: "768x1024", Quality: "medium", Official: 0.04},
	{Size: "1672x941", Quality: "medium", Official: 0.0413},
	{Size: "941x1672", Quality: "medium", Official: 0.0413},
	{Size: "1448x1086", Quality: "medium", Official: 0.0426},
	{Size: "1086x1448", Quality: "medium", Official: 0.0426},
	{Size: "1024x1024", Quality: "medium", Official: 0.0529},
	{Size: "1152x2688", Quality: "medium", Official: 0.0504},
	{Size: "2688x1152", Quality: "medium", Official: 0.0504},
	{Size: "1024x1280", Quality: "medium", Official: 0.0553},
	{Size: "1280x1024", Quality: "medium", Official: 0.0553},
	{Size: "1152x2048", Quality: "medium", Official: 0.0631},
	{Size: "2048x1152", Quality: "medium", Official: 0.0631},
	{Size: "1122x1402", Quality: "medium", Official: 0.0812},
	{Size: "1402x1122", Quality: "medium", Official: 0.0812},
	{Size: "1360x2048", Quality: "medium", Official: 0.0993},
	{Size: "2048x1360", Quality: "medium", Official: 0.0993},
	{Size: "2160x3840", Quality: "medium", Official: 0.1003},
	{Size: "3840x2160", Quality: "medium", Official: 0.1003},
	{Size: "2048x2048", Quality: "medium", Official: 0.1072},
	{Size: "1648x3840", Quality: "medium", Official: 0.1179},
	{Size: "3840x1648", Quality: "medium", Official: 0.1179},
	{Size: "1280x3840", Quality: "medium", Official: 0.1325},
	{Size: "3840x1280", Quality: "medium", Official: 0.1325},
	{Size: "2576x3216", Quality: "medium", Official: 0.1408},
	{Size: "3216x2576", Quality: "medium", Official: 0.1408},
	{Size: "2880x2880", Quality: "medium", Official: 0.178},
	{Size: "1536x512", Quality: "high", Official: 0.0643},
	{Size: "512x1536", Quality: "high", Official: 0.0643},
	{Size: "1024x3072", Quality: "high", Official: 0.1064},
	{Size: "3072x1024", Quality: "high", Official: 0.1064},
	{Size: "2016x864", Quality: "high", Official: 0.1106},
	{Size: "864x2016", Quality: "high", Official: 0.1106},
	{Size: "1536x864", Quality: "high", Official: 0.1187},
	{Size: "864x1536", Quality: "high", Official: 0.1187},
	{Size: "1774x887", Quality: "high", Official: 0.1295},
	{Size: "887x1774", Quality: "high", Official: 0.1295},
	{Size: "1915x821", Quality: "high", Official: 0.1336},
	{Size: "821x1915", Quality: "high", Official: 0.1336},
	{Size: "1024x1536", Quality: "high", Official: 0.1418},
	{Size: "1536x1024", Quality: "high", Official: 0.1418},
	{Size: "1881x836", Quality: "high", Official: 0.1446},
	{Size: "836x1881", Quality: "high", Official: 0.1446},
	{Size: "1024x2048", Quality: "high", Official: 0.1507},
	{Size: "2048x1024", Quality: "high", Official: 0.1507},
	{Size: "1024x768", Quality: "high", Official: 0.1595},
	{Size: "768x1024", Quality: "high", Official: 0.1595},
	{Size: "1672x941", Quality: "high", Official: 0.1648},
	{Size: "941x1672", Quality: "high", Official: 0.1648},
	{Size: "1448x1086", Quality: "high", Official: 0.1697},
	{Size: "1086x1448", Quality: "high", Official: 0.1697},
	{Size: "1024x1024", Quality: "high", Official: 0.2109},
	{Size: "1152x2688", Quality: "high", Official: 0.2},
	{Size: "2688x1152", Quality: "high", Official: 0.2},
	{Size: "1024x1280", Quality: "high", Official: 0.2207},
	{Size: "1280x1024", Quality: "high", Official: 0.2207},
	{Size: "1152x2048", Quality: "high", Official: 0.2461},
	{Size: "2048x1152", Quality: "high", Official: 0.2461},
	{Size: "1122x1402", Quality: "high", Official: 0.3241},
	{Size: "1402x1122", Quality: "high", Official: 0.3241},
	{Size: "1360x2048", Quality: "high", Official: 0.4004},
	{Size: "2048x1360", Quality: "high", Official: 0.4004},
	{Size: "2160x3840", Quality: "high", Official: 0.4004},
	{Size: "3840x2160", Quality: "high", Official: 0.4004},
	{Size: "2048x2048", Quality: "high", Official: 0.4283},
	{Size: "1648x3840", Quality: "high", Official: 0.4712},
	{Size: "3840x1648", Quality: "high", Official: 0.4712},
	{Size: "1280x3840", Quality: "high", Official: 0.5296},
	{Size: "3840x1280", Quality: "high", Official: 0.5296},
	{Size: "2576x3216", Quality: "high", Official: 0.5703},
	{Size: "3216x2576", Quality: "high", Official: 0.5703},
	{Size: "2880x2880", Quality: "high", Official: 0.7117},
}

var apimartGPTImage2OfficialPriceIndex = buildAPIMartImagePriceIndex(apimartGPTImage2OfficialPriceRows)

func buildAPIMartImagePriceIndex(rows []apimartImagePriceRow) map[string]float64 {
	out := make(map[string]float64, len(rows))
	for _, row := range rows {
		out[apimartImagePriceKey(row.Size, row.Quality)] = row.Official
	}
	return out
}

func apimartImagePriceKey(size string, quality string) string {
	return normalizeAPIMartImageSize(size) + ":" + normalizeAPIMartImageQuality(quality)
}

func normalizeAPIMartImageSize(size string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(size), "×", "x"))
}

func normalizeAPIMartImageQuality(quality string) string {
	quality = strings.ToLower(strings.TrimSpace(quality))
	if quality == "" {
		return "auto"
	}
	return quality
}

func isAPIMartGPTImage2OfficialModel(model string) bool {
	return strings.EqualFold(strings.TrimSpace(model), apimartGPTImage2OfficialModel)
}

func isAPIMartGPTImage2Model(model string) bool {
	return strings.EqualFold(strings.TrimSpace(model), apimartGPTImage2Model)
}

func isAPIMartOpenAIAPIKeyAccount(account *Account) bool {
	if account == nil || !account.IsOpenAIApiKey() {
		return false
	}
	host := openAIBaseURLHost(account.GetOpenAIBaseURL())
	return strings.EqualFold(host, apimartOpenAIBaseURLHost)
}

func apimartGPTImage2OfficialBalancePrice(upstreamPrice float64) float64 {
	return upstreamPrice * apimartGPTImage2OfficialBalanceMultiplier
}

func apimartGPTImage2UsageMultiplierForModels(account *Account, models []string, baseMultiplier float64) float64 {
	apimartAccount := isAPIMartOpenAIAPIKeyAccount(account)
	for _, model := range models {
		if isAPIMartGPTImage2OfficialModel(model) {
			return baseMultiplier * apimartGPTImage2OfficialBalanceMultiplier
		}
		if apimartAccount && isAPIMartGPTImage2Model(model) {
			return baseMultiplier * apimartGPTImage2OfficialBalanceMultiplier
		}
	}
	return baseMultiplier
}

func lookupAPIMartGPTImage2OfficialPrice(size string, quality string) (float64, bool) {
	size = normalizeAPIMartImageSize(size)
	if size == "" {
		return 0, false
	}
	quality = normalizeAPIMartImageQuality(quality)
	if price, ok := apimartGPTImage2OfficialPriceIndex[size+":"+quality]; ok {
		return price, true
	}
	if quality != "auto" {
		if price, ok := apimartGPTImage2OfficialPriceIndex[size+":auto"]; ok {
			return price, true
		}
	}
	if quality != "low" {
		if price, ok := apimartGPTImage2OfficialPriceIndex[size+":low"]; ok {
			return price, true
		}
	}
	return 0, false
}

func apimartGPTImage2OfficialReferencePricing() *ChannelModelPricing {
	intervals := make([]PricingInterval, 0, len(apimartGPTImage2OfficialPriceRows)+1)
	defaultPrice := apimartGPTImage2OfficialDefaultPrice
	intervals = append(intervals, PricingInterval{
		MinTokens:       0,
		MaxTokens:       nil,
		TierLabel:       "default",
		PerRequestPrice: &defaultPrice,
		SortOrder:       0,
	})
	for i, row := range apimartGPTImage2OfficialPriceRows {
		price := row.Official
		intervals = append(intervals, PricingInterval{
			MinTokens:       0,
			MaxTokens:       nil,
			TierLabel:       apimartImagePriceKey(row.Size, row.Quality),
			PerRequestPrice: &price,
			SortOrder:       i + 1,
		})
	}
	return &ChannelModelPricing{
		Platform:        PlatformOpenAI,
		Models:          []string{apimartGPTImage2OfficialModel},
		BillingMode:     BillingModeImage,
		PerRequestPrice: &defaultPrice,
		Intervals:       intervals,
	}
}

func appendAPIMartGPTImage2OfficialIntervals(intervals []PricingInterval) []PricingInterval {
	existing := make(map[string]struct{}, len(intervals))
	maxSort := 0
	for _, interval := range intervals {
		existing[strings.TrimSpace(interval.TierLabel)] = struct{}{}
		if interval.SortOrder > maxSort {
			maxSort = interval.SortOrder
		}
	}
	rows := append([]apimartImagePriceRow(nil), apimartGPTImage2OfficialPriceRows...)
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].Size == rows[j].Size {
			return rows[i].Quality < rows[j].Quality
		}
		return rows[i].Size < rows[j].Size
	})
	for _, row := range rows {
		label := apimartImagePriceKey(row.Size, row.Quality)
		if _, ok := existing[label]; ok {
			continue
		}
		price := row.Official
		maxSort++
		intervals = append(intervals, PricingInterval{
			MinTokens:       0,
			MaxTokens:       nil,
			TierLabel:       label,
			PerRequestPrice: &price,
			SortOrder:       maxSort,
		})
	}
	return intervals
}
