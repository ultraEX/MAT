// RESTful
package server

import (
	//	"fmt"
	"strconv"
	"strings"
	"time"

	. "../comm"
	"../config"
	"./neko"
)

func GetTickerType(strType string) TickerType {
	switch strType {
	case "Ticker_1min":
		return TickerType_1min
	case "Ticker_5min":
		return TickerType_5min
	case "Ticker_15min":
		return TickerType_15min
	case "Ticker_30min":
		return TickerType_30min
	case "Ticker_1hour":
		return TickerType_1hour
	case "Ticker_1day":
		return TickerType_1day
	case "Ticker_1week":
		return TickerType_1week
	case "Ticker_1month":
		return TickerType_1month
	default:
		return TickerType_Invalid
	}
}

func GetTickerTypeFromResolution(strType string) TickerType {
	switch strType {
	case "1":
		return TickerType_1min
	case "5":
		return TickerType_5min
	case "15":
		return TickerType_15min
	case "30":
		return TickerType_30min
	case "60":
		return TickerType_1hour
	case "D":
		return TickerType_1day
	case "W":
		return TickerType_1week
	case "M":
		return TickerType_1month
	default:
		return TickerType_Invalid
	}
}

/// Request: "/k_line?symbol=ETH:BTC&resolution=<resolution>"
/// Example: http://localhost:7933/k_line?symbol=UBT:BTC&resolution=1
func Get_KLine(ctx *neko.Context) {
	// Parmas:
	sym := ctx.Params.ByGet("symbol")
	resolution := ctx.Params.ByGet("resolution")

	// Response:
	sym = strings.Replace(sym, ":", "/", -1)
	tickerType := GetTickerTypeFromResolution(resolution)
	matchEng, _ := marketsRef.GetMatchEngine(sym)

	if matchEng != nil && tickerType != TickerType_Invalid {
		tp := matchEng.GetTickersEngine()
		kLine, _ := tp.GetTicker(tickerType)

		ctx.SetHeader("Access-Control-Allow-Origin", "*")
		ctx.Json(
			neko.JSON{
				"s": "ok",
				"d": kLine,
			},
		)
	}
}

/// Request:GET /k_line_limit?symbol=<ticker_name>&resolution=<resolution>&limit=<limit>
/// eg.: GET /k_line_limit?symbol=UBT:BTC&resolution=1&limit=60
func Get_k_line_limit(ctx *neko.Context) {
	// Parmas:
	limit := ctx.Params.ByGet("limit")
	limitInt, err := strconv.ParseInt(limit, 10, 64)

	if err != nil {
		ctx.SetHeader("Access-Control-Allow-Origin", "*")
		ctx.Json(
			neko.JSON{
				"s":      "error",
				"errmsg": "params error",
			},
		)
		return
	}

	symbol := ctx.Params.ByGet("symbol")
	sym := strings.Replace(symbol, ":", "/", -1)
	resolution := ctx.Params.ByGet("resolution")
	tickerType := GetTickerTypeFromResolution(resolution)

	// Response:
	matchEng, _ := marketsRef.GetMatchEngine(sym)

	if matchEng != nil && tickerType != TickerType_Invalid {
		tp := matchEng.GetTickersEngine()
		kLine, _ := tp.GetTickerLimit(GetTickerTypeFromResolution(resolution), limitInt)

		ctx.SetHeader("Access-Control-Allow-Origin", "*")
		ctx.Json(
			neko.JSON{
				"s": "ok",
				"d": kLine,
			},
		)
	}
}

/// Request:GET /symbol_info?group=<group_name>
/// Example:GET /symbol_info?group=NYSE
func Get_symbol_info(ctx *neko.Context) {
	// Parmas:
	///group := ctx.Params.ByGet("group")

	syms := []string{"UBT:BTC"}
	groupInfo := neko.JSON{
		"symbol": syms,
	}
	ctx.SetHeader("Access-Control-Allow-Origin", "*")
	ctx.Json(groupInfo)
}

/// Request: GET /time
func Get_time(ctx *neko.Context) {
	// Response:
	loc, _ := time.LoadLocation("Local")
	ctx.SetHeader("Access-Control-Allow-Origin", "*")
	ctx.Text(strconv.FormatInt(time.Now().In(loc).Unix(), 10))
}

/// Request: GET /config
func Get_config(ctx *neko.Context) {
	// Response:
	resolution := [8]string{"1", "5", "15", "30", "60", "1D", "1W", "1M"}
	ctx.SetHeader("Access-Control-Allow-Origin", "*")
	ctx.Json(neko.JSON{
		"supports_search":        true,
		"supports_group_request": false,
		"supported_resolutions":  resolution,
		"supports_marks":         false,
		"supports_time":          true,
	})
}

/// Request:GET /symbols?symbol=<symbol>
func Get_symbols(ctx *neko.Context) {
	// Parmas:
	symbol := ctx.Params.ByGet("symbol")
	sym := strings.Replace(symbol, ":", "/", -1)
	// Response:
	ctx.SetHeader("Access-Control-Allow-Origin", "*")
	ctx.Json(neko.JSON{
		"name":         symbol,
		"ticker":       symbol,
		"description":  "trade pair: " + sym,
		"type":         "bitcoin",
		"session":      "24x7",
		"timezone":     "Asia/Shanghai",
		"has_intraday": true,
		"exchange":     "ultraex.io",
	})
}

/// Request:GET /search?query=<query>&type=<type>&exchange=<exchange>&limit=<limit>
/// eg.: GET /search?query=AA&type=stock&exchange=NYSE&limit=15
func Get_search(ctx *neko.Context) {
	// Parmas:
	query := ctx.Params.ByGet("query")
	query = strings.Replace(query, ":", "/", -1)
	_type := ctx.Params.ByGet("type")
	exchange := ctx.Params.ByGet("exchange")
	_ = ctx.Params.ByGet("limit")

	// Response:
	if marketsRef.GetMatchEngineExistence(query) && _type == "bitcoin" && exchange == "ultraex.io" {
		symInfo := neko.JSON{
			"symbol":      query,
			"full_name":   query,
			"description": "Cryptocurrenty: " + query,
			"exchange":    "ultraex.io",
			"ticker":      query,
			"type":        "bitcoin",
		}
		ctx.SetHeader("Access-Control-Allow-Origin", "*")
		ctx.Json(symInfo)
	}

}

/// Request:GET /history?symbol=<ticker_name>&from=<unix_timestamp>&to=<unix_timestamp>&resolution=<resolution>
/// eg.: GET /history?symbol=BEAM~0&resolution=D&from=1386493512&to=1395133512
func Get_history(ctx *neko.Context) {
	// Parmas:
	symbol := ctx.Params.ByGet("symbol")
	sym := strings.Replace(symbol, ":", "/", -1)
	from := ctx.Params.ByGet("from")
	to := ctx.Params.ByGet("to")
	resolution := ctx.Params.ByGet("resolution")
	tickerType := GetTickerTypeFromResolution(resolution)

	// Response:
	matchEng, _ := marketsRef.GetMatchEngine(sym)

	if matchEng != nil && tickerType != TickerType_Invalid {
		tp := matchEng.GetTickersEngine()

		fromInt64, _ := strconv.ParseInt(from, 10, 64)
		toInt64, _ := strconv.ParseInt(to, 10, 64)
		kLine, err := tp.GetTickerDuration(tickerType, fromInt64*int64(time.Second), toInt64*int64(time.Second))
		if err != nil {
			ctx.SetHeader("Access-Control-Allow-Origin", "*")
			ctx.Json(neko.JSON{
				"s":      "error",
				"errmsg": err.Error(),
			})
		} else {
			var (
				tArray                                 []int64
				cArray, oArray, hArray, lArray, vArray []float64
			)
			for _, v := range kLine {
				tArray = append(tArray, v.From/int64(time.Second))
				cArray = append(cArray, v.ClosePrice)
				oArray = append(oArray, v.OpenPrice)
				hArray = append(hArray, v.HightPrice)
				lArray = append(lArray, v.LowPrice)
				vArray = append(vArray, v.Volume)
			}
			ctx.SetHeader("Access-Control-Allow-Origin", "*")
			ctx.Json(neko.JSON{
				"s": "ok",
				"t": tArray,
				"c": cArray,
				"o": oArray,
				"h": hArray,
				"l": lArray,
				"v": vArray,
			})
		}

	}

}

/// Request:GET /quotes?symbols=<ticker_name_1>,<ticker_name_2>,...,<ticker_name_n>
/// eg.: GET /quotes?symbols=NYSE%3AAA%2CNYSE%3AF%2CNasdaqNM%3AAAPL(decoded:/quotes?symbols=NYSE:AA,NYSE:F,NasdaqNM:AAPL)
func Get_quotes(ctx *neko.Context) {
	var (
		dArray []neko.JSON
		err    error
		quote  *QuoteInfo
	)
	// Parmas:
	symbols := ctx.Params.ByGet("symbols")
	syms := strings.Split(symbols, ",")

	// Response:
	for _, sym := range syms {
		sym = strings.Replace(sym, ":", "/", -1)
		matchEng, _ := marketsRef.GetMatchEngine(sym)

		if matchEng != nil {
			tp := matchEng.GetTickersEngine()

			quote, err = tp.GetQuote()
			if err != nil {
				continue
			} else {
				v := neko.JSON{
					"ch":               quote.ChangePrice,
					"chp":              quote.ChangePriceRate,
					"short_name":       sym,
					"exchange":         "ultraex.io",
					"description":      "Cryptocurrency: " + sym,
					"lp":               quote.LatestTradePrice,
					"ask":              quote.Ask1stPrice,
					"bid":              quote.Bid1stPrice,
					"open_price":       quote.DayOpenPrice,
					"high_price":       quote.DayHighPrice,
					"low_price":        quote.DayLowPrice,
					"prev_close_price": quote.PreDayClosePrice,
					"volume":           quote.DayVolume,
				}
				d := neko.JSON{
					"s": "ok",
					"n": sym,
					"v": v,
				}
				dArray = append(dArray, d)
			}
		}
	}

	if len(dArray) == 0 {
		ctx.SetHeader("Access-Control-Allow-Origin", "*")
		ctx.Json(neko.JSON{
			"s":      "error",
			"errmsg": err.Error(),
		})
	} else {
		o := neko.JSON{
			"s": "ok",
			"d": dArray,
		}
		ctx.SetHeader("Access-Control-Allow-Origin", "*")
		ctx.Json(o)
	}
}

/// Request:GET /quotes_ex?symbols=<ticker_name_1>,<ticker_name_2>,...,<ticker_name_n>
/// eg.: GET /quotes_ex?symbols=NYSE%3AAA%2CNYSE%3AF%2CNasdaqNM%3AAAPL(decoded:/quotes?symbols=NYSE:AA,NYSE:F,NasdaqNM:AAPL)
func Get_quotes_ex(ctx *neko.Context) {
	var (
		dArray []neko.JSON
		err    error
		quote  *QuoteInfo
	)
	// Parmas:
	symbols := ctx.Params.ByGet("symbols")
	syms := strings.Split(symbols, ",")

	// Response:
	for _, sym := range syms {
		sym = strings.Replace(sym, ":", "/", -1)
		matchEng, _ := marketsRef.GetMatchEngine(sym)

		if matchEng != nil {
			tp := matchEng.GetTickersEngine()

			quote, err = tp.GetQuote()
			if err != nil {
				d := neko.JSON{
					"s":      "error",
					"n":      sym,
					"v":      "",
					"errmsg": "no data",
				}
				dArray = append(dArray, d)
			} else {
				v := neko.JSON{
					"ch":               quote.ChangePrice,
					"chp":              quote.ChangePriceRate,
					"chp_7d":           quote.ChangePriceRate_7D,
					"short_name":       sym,
					"exchange":         "ultraex.io",
					"description":      "Cryptocurrency: " + sym,
					"lp":               quote.LatestTradePrice,
					"ask":              quote.Ask1stPrice,
					"bid":              quote.Bid1stPrice,
					"open_price":       quote.DayOpenPrice,
					"high_price":       quote.DayHighPrice,
					"low_price":        quote.DayLowPrice,
					"prev_close_price": quote.PreDayClosePrice,
					"volume":           quote.DayVolume,
					"amount":           quote.DayAmount,
				}
				d := neko.JSON{
					"s": "ok",
					"n": sym,
					"v": v,
				}
				dArray = append(dArray, d)
			}
		}
	}

	if len(dArray) == 0 {
		ctx.SetHeader("Access-Control-Allow-Origin", "*")
		ctx.Json(neko.JSON{
			"s":      "error",
			"errmsg": "no quotes info",
		})
	} else {
		o := neko.JSON{
			"s": "ok",
			"d": dArray,
		}
		ctx.SetHeader("Access-Control-Allow-Origin", "*")
		ctx.Json(o)
	}
}

/// Request:GET /latest_trades?symbols=<ticker_name_1>,<ticker_name_2>,...,<ticker_name_n>&limit=<count>
/// eg.: GET /latest_trades?symbols=NYSE%3AAA%2CNYSE%3AF%2CNasdaqNM%3AAAPL&limit=20
func Get_latest_trades(ctx *neko.Context) {
	var (
		symbols        string
		syms           []string
		dArray, vArray []neko.JSON
		err            error
		limitInt       int64
		trades         []*Trade
	)
	// Parmas:
	symbols = ctx.Params.ByGet("symbols")
	syms = strings.Split(symbols, ",")
	limit := ctx.Params.ByGet("limit")
	limitInt, err = strconv.ParseInt(limit, 10, 64)

	// Response:
	if err != nil {
		return
	}

	for _, sym := range syms {
		sym = strings.Replace(sym, ":", "/", -1)
		matchEng, _ := marketsRef.GetMatchEngine(sym)

		if matchEng != nil {
			tp := matchEng.GetTickersEngine()

			trades, err = tp.GetLatestTradeLimit(limitInt)
			if err != nil {
				continue
			} else {
				vArray = make([]neko.JSON, 0)
				for _, trade := range trades {
					v := neko.JSON{
						"trade_time": trade.TradeTime / int64(time.Second),
						"type":       trade.AorB.String(),
						"price":      trade.Price,
						"volume":     trade.Volume,
					}
					vArray = append(vArray, v)
				}

				d := neko.JSON{
					"s": "ok",
					"n": sym,
					"v": vArray,
				}

				dArray = append(dArray, d)
			}
		}
	}

	if len(dArray) == 0 {
		ctx.SetHeader("Access-Control-Allow-Origin", "*")
		ctx.Json(neko.JSON{
			"s":      "error",
			"errmsg": err.Error(),
		})
	} else {
		o := neko.JSON{
			"s": "ok",
			"d": dArray,
		}
		ctx.SetHeader("Access-Control-Allow-Origin", "*")
		ctx.Json(o)
	}
}

/// Request:GET /top_orders?symbols=<ticker_name_1>,<ticker_name_2>,...,<ticker_name_n>&limit=<count>
/// eg.: GET /top_orders?symbols=NYSE%3AAA%2CNYSE%3AF%2CNasdaqNM%3AAAPL&limit=20
func Get_top_orders(ctx *neko.Context) {
	var (
		symbols                      string
		syms                         []string
		dArray, vAskArray, vBidArray []neko.JSON
		err, errAsk, errBid          error
		limitInt                     int64
		askOrders, bidOrders         []*Order
	)
	// Parmas:
	symbols = ctx.Params.ByGet("symbols")
	syms = strings.Split(symbols, ",")
	limit := ctx.Params.ByGet("limit")
	limitInt, err = strconv.ParseInt(limit, 10, 64)

	// Response:
	if err != nil {
		return
	}

	for _, sym := range syms {
		sym = strings.Replace(sym, ":", "/", -1)
		matchEng, _ := marketsRef.GetMatchEngine(sym)

		if matchEng != nil {
			tp := matchEng.GetTickersEngine()

			askOrders, errAsk = tp.GetAskLevelOrders(limitInt)
			bidOrders, errBid = tp.GetBidLevelOrders(limitInt)
			if errAsk != nil || errBid != nil {
				continue
			} else {
				vAskArray = make([]neko.JSON, 0)
				vBidArray = make([]neko.JSON, 0)
				for _, order := range askOrders {
					v := neko.JSON{
						"order_time": order.Timestamp / int64(time.Second),
						"price":      order.Price,
						"volume":     order.Volume,
					}
					vAskArray = append(vAskArray, v)
				}
				for _, order := range bidOrders {
					v := neko.JSON{
						"order_time": order.Timestamp / int64(time.Second),
						"price":      order.Price,
						"volume":     order.Volume,
					}
					vBidArray = append(vBidArray, v)
				}
				dAsk := neko.JSON{
					"type": "ask",
					"v":    vAskArray,
				}
				dBid := neko.JSON{
					"type": "bid",
					"v":    vBidArray,
				}
				d := neko.JSON{
					"s":    "ok",
					"n":    sym,
					"vAsk": dAsk,
					"vBid": dBid,
				}

				dArray = append(dArray, d)
			}
		}
	}

	if len(dArray) == 0 {
		ctx.SetHeader("Access-Control-Allow-Origin", "*")
		ctx.Json(neko.JSON{
			"s":      "error",
			"errmsg": "params error",
		})
	} else {
		o := neko.JSON{
			"s": "ok",
			"d": dArray,
		}
		ctx.SetHeader("Access-Control-Allow-Origin", "*")
		ctx.Json(o)
	}
}

/// Request:GET /level_orders?symbols=<ticker_name_1>,<ticker_name_2>,...,<ticker_name_n>&limit=<count>
/// Example: GET /level_orders?symbols=NYSE%3AAA%2CNYSE%3AF%2CNasdaqNM%3AAAPL&limit=20
func Get_level_orders(ctx *neko.Context) {
	var (
		symbols                      string
		syms                         []string
		dArray, vAskArray, vBidArray []neko.JSON
		err, errAsk, errBid          error
		limitInt                     int64
		askLevels, bidLevels         []OrderLevel
	)
	// Parmas:
	symbols = ctx.Params.ByGet("symbols")
	syms = strings.Split(symbols, ",")
	limit := ctx.Params.ByGet("limit")
	limitInt, err = strconv.ParseInt(limit, 10, 64)

	// Response:
	if err != nil {
		return
	}

	for _, sym := range syms {
		sym = strings.Replace(sym, ":", "/", -1)
		matchEng, _ := marketsRef.GetMatchEngine(sym)

		if matchEng != nil {
			tp := matchEng.GetTickersEngine()

			askLevels, errAsk = tp.GetAskLevelsGroupByPrice(limitInt)
			bidLevels, errBid = tp.GetBidLevelsGroupByPrice(limitInt)
			if errAsk != nil || errBid != nil {
				continue
			} else {
				vAskArray = make([]neko.JSON, 0)
				vBidArray = make([]neko.JSON, 0)
				for _, level := range askLevels {
					v := neko.JSON{
						"price":       level.Price,
						"volume":      level.Volume,
						"totalvolume": level.TotalVolume,
					}
					vAskArray = append(vAskArray, v)
				}
				for _, level := range bidLevels {
					v := neko.JSON{
						"price":       level.Price,
						"volume":      level.Volume,
						"totalvolume": level.TotalVolume,
					}
					vBidArray = append(vBidArray, v)
				}
				dAsk := neko.JSON{
					"type": "ask",
					"v":    vAskArray,
				}
				dBid := neko.JSON{
					"type": "bid",
					"v":    vBidArray,
				}
				d := neko.JSON{
					"s":    "ok",
					"n":    sym,
					"vAsk": dAsk,
					"vBid": dBid,
				}

				dArray = append(dArray, d)
			}
		}
	}

	if len(dArray) == 0 {
		ctx.SetHeader("Access-Control-Allow-Origin", "*")
		ctx.Json(neko.JSON{
			"s":      "error",
			"errmsg": "params error",
		})
	} else {
		o := neko.JSON{
			"s": "ok",
			"d": dArray,
		}
		ctx.SetHeader("Access-Control-Allow-Origin", "*")
		ctx.Json(o)
	}
}

func RESTfulMain() {
	app := neko.Classic()

	app.GET("/", func(ctx *neko.Context) {
		ctx.Text("Nice to meet you!")
	})

	app.GET("/k_line", Get_KLine)
	app.GET("/k_line_limit", Get_k_line_limit)
	app.GET("/symbol_info", Get_symbol_info)
	app.GET("/time", Get_time)
	app.GET("/config", Get_config)
	app.GET("/symbols", Get_symbols)
	app.GET("/search", Get_search)
	app.GET("/history", Get_history)
	app.GET("/quotes", Get_quotes)
	app.GET("/quotes_ex", Get_quotes_ex)

	app.GET("/latest_trades", Get_latest_trades)
	app.GET("/level_orders", Get_level_orders)

	/// server run:
	/// app.Run(":7933")
	app.Run(config.GetMEConfig().RESTfulIPPort)

}
