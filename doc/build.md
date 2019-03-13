
# Dependence:
1.	Thrift:
	go get github.com/apache/thrift/lib/go/thrift
2.	go-redis:
	go get -u github.com/go-redis/redis
	import "github.com/go-redis/redis"
3.	redisgo:
	go get github.com/gomodule/redigo/redis
4.	go-sql-driver:
	https://github.com/go-sql-driver/mysql.git
	go get -u github.com/go-sql-driver/mysql
	import "database/sql"
	import _ "github.com/go-sql-driver/mysql"
5.	godaemon:
	https://github.com/VividCortex/godaemon.git
	go get -u github.com/VividCortex/godaemon
		import "github.com/VividCortex/godaemon"
		func main() {
			godaemon.MakeDaemon(&godaemon.DaemonAttr{})
		}
6.	jrpc2:
	https://github.com/bitwurx/jrpc2		
	go get github.com/bitwurx/jrpc2
	In this project, jrpc2 is import as a inneral module rather than a common module.

7.	RESTful framework:	
	go get github.com/rocwong/neko
	go get github.com/julienschmidt/httprouter
	go get github.com/rocwong/neko/render
	go get github.com/gorilla/sessions	

# Thrift:
* thrift -r --gen go order.thrift
* thrift -r --gen php order.thrift

# Run:
run as daemon mode use:
	./ME -daemon
	
run as a normal program use:
	./ME 

Client tool:
	./ME-Cli
	
ME Core Daemon Manager Tool:
	./StartME	
	./StartME -daemon	

# Test:
## test http json rpc:
	curl -X POST --data '{"jsonrpc":"2.0","method":"add","params":[1,2],"id":67}' localhost:7392/api/v1/rpc 
	curl -X POST --data '{"jsonrpc":"2.0","method":"addcoin","params":["XXX",229],"id":67}' localhost:7392/api/v1/rpc 
	curl -X POST --data '{"jsonrpc":"2.0","method":"addcoin","params":["BBB",230],"id":67}' localhost:7392/api/v1/rpc
	curl -X POST --data '{"jsonrpc":"2.0","method":"addmarket","params":{"Sym":{"BaseCoin":"XXX","QuoteCoin":"BBB"}, "Market_Human":true, "Market_Robot":true, "Market_MixHR":false, "RobotList":[4,5,387,388], "NoneFinanceList":[387,388]},"id":67}' localhost:7392/api/v1/rpc

	curl -X POST --data '{"jsonrpc":"2.0","method":"removecoin","params":["XXX"],"id":67}' localhost:7392/api/v1/rpc 
	curl -X POST --data '{"jsonrpc":"2.0","method":"removecoin","params":["BBB"],"id":67}' localhost:7392/api/v1/rpc 
	curl -X POST --data '{"jsonrpc":"2.0","method":"removemarket","params":{"Sym":{"BaseCoin":"XXX","QuoteCoin":"BBB"}},"id":67}' localhost:7392/api/v1/rpc

## test RESTful:
* GET /time
	http://localhost:7933/time
* GET /config
	http://localhost:7933/config
* GET /symbols?symbol=<symbol>
	http://localhost:7933/symbols?symbol=UBT:BTC
* GET /k_line_limit?symbol=UBT-BTC&resolution=1&limit=<limit>
	http://localhost:7933/k_line_limit?symbol=UBT:BTC&resolution=1&limit=60
* GET /k_line?symbol=UBT-BTC&resolution=1
	http://localhost:7933/k_line?symbol=UBT:BTC&resolution=1
* GET /search?query=AA&type=stock&exchange=NYSE&limit=15
	http://localhost:7933/search?query=UBT:BTC&type=bitcoin&exchange=ultraex.io&limit=15
* GET /history?symbol=UBT-BTC&resolution=1&from=1386493512&to=1545183967
	http://localhost:7933/history?symbol=UBT:BTC&resolution=1&from=1386493512&to=1545183967
* GET /quotes?symbols=<ticker_name_1>,<ticker_name_2>,...,<ticker_name_n>
	http://localhost:7933/quotes?symbols=UBT:BTC,ETH:BTC
* GET /quotes_ex?symbols=<ticker_name_1>,<ticker_name_2>,...,<ticker_name_n>
	http://localhost:7933/quotes_ex?symbols=UBT:BTC,ETH:BTC
* GET /latest_trades?symbols=<ticker_name_1>,<ticker_name_2>,...,<ticker_name_n>&limit=<count>
	http://localhost:7933/latest_trades?symbols=UBT:BTC,ETH:BTC&limit=20
* GET /level_orders?symbols=<ticker_name_1>,<ticker_name_2>,...,<ticker_name_n>&limit=<count>
	http://localhost:7933//level_orders?symbols=UBT:BTC,ETH:BTC&limit=20

## Examples:
* GET /time
	Require:
		http://localhost:7933/time
	Response:
		1545894411
		
* GET /config
	Require:
		http://localhost:7933/config
	Response:
		{"supported_resolutions":["1","5","15","30","60","1D","1W","1M"],"supports_group_request":false,"supports_marks":false,"supports_search":true,"supports_time":true}
	
* GET /symbols?symbol=<symbol>
	Require:
		http://localhost:7933/symbols?symbol=UBT:BTC
	Response:
		{"description":"trade pair: UBT/BTC","exchange":"ultraex.io","has_intraday":true,"name":"UBT:BTC","session":"24x7","ticker":"UBT:BTC","timezone":"Asia/Shanghai","type":"bitcoin"}
	
* GET /k_line_limit?symbol=UBT-BTC&resolution=1&limit=<limit>
	Require:
		http://localhost:7933/k_line_limit?symbol=UBT:BTC&resolution=1&limit=10
	Response:
		{"d":[{"From":1545881321,"End":1545881326,"OpenPrice":1.2,"ClosePrice":1.2,"LowPrice":1,"HightPrice":1.2,"Volume":165,"Amount":187.3},{"From":1545881220,"End":1545881263,"OpenPrice":1.1,"ClosePrice":1.1,"LowPrice":1,"HightPrice":1.2,"Volume":70,"Amount":79.3},{"From":1545881160,"End":1545881200,"OpenPrice":1.2,"ClosePrice":1,"LowPrice":1,"HightPrice":1.2,"Volume":78,"Amount":89.3},{"From":1545881100,"End":1545881136,"OpenPrice":1.2,"ClosePrice":1.2,"LowPrice":1,"HightPrice":1.2,"Volume":140,"Amount":156.7},{"From":1545881040,"End":1545881073,"OpenPrice":1.2,"ClosePrice":1.1,"LowPrice":1.1,"HightPrice":1.2,"Volume":88,"Amount":99.5},{"From":1545880980,"End":1545881010,"OpenPrice":1,"ClosePrice":1.2,"LowPrice":1,"HightPrice":1.2,"Volume":152,"Amount":170.1},{"From":1545880920,"End":1545880947,"OpenPrice":1.2,"ClosePrice":1,"LowPrice":1,"HightPrice":1.2,"Volume":85,"Amount":94.1},{"From":1545880860,"End":1545880881,"OpenPrice":1.2,"ClosePrice":1.1,"LowPrice":1,"HightPrice":1.2,"Volume":197,"Amount":216.7},{"From":1545880800,"End":1545880812,"OpenPrice":1.2,"ClosePrice":1.1,"LowPrice":1,"HightPrice":1.2,"Volume":201,"Amount":225.7},{"From":1545880740,"End":1545880743,"OpenPrice":1.1,"ClosePrice":1.2,"LowPrice":1,"HightPrice":1.2,"Volume":188,"Amount":200.4}],"s":"ok"}
	
* GET /k_line?symbol=UBT-BTC&resolution=1
	Require:
		http://localhost:7933/k_line?symbol=UBT:BTC&resolution=1
	Response:
		d	[…]
		s	"ok"
	
* GET /search?query=AA&type=stock&exchange=NYSE&limit=15
	Require:
		http://localhost:7933/search?query=UBT:BTC&type=bitcoin&exchange=ultraex.io&limit=15
	Response:
		{"description":"Cryptocurrenty: UBT/BTC","exchange":"ultraex.io","full_name":"UBT/BTC","symbol":"UBT/BTC","ticker":"UBT/BTC","type":"bitcoin"}
		
* GET /history?symbol=UBT-BTC&resolution=1&from=1386493512&to=1545183967
	Require:
		http://localhost:7933/history?symbol=UBT:BTC&resolution=1&from=1386493512&to=1545183967
	Response:	
		{"c":null,"h":null,"l":null,"o":null,"s":"ok","t":null,"v":null}
	
* GET /quotes?symbols=<ticker_name_1>,<ticker_name_2>,...,<ticker_name_n>
	Require:
		http://localhost:7933/quotes?symbols=UBT:BTC,ETH:BTC
	Response:
		{"d":[{"n":"UBT/BTC","s":"ok","v":{"ask":1.2,"bid":1.1,"ch":0.19999999999999996,"chp":0.19999999999999996,"description":"Cryptocurrency: UBT/BTC","exchange":"ultraex.io","high_price":1.2,"low_price":1,"lp":1.2,"open_price":1,"prev_close_price":1.1,"short_name":"UBT/BTC","volume":195536}},{"n":"ETH/BTC","s":"ok","v":{"ask":1.2,"bid":1,"ch":0.19999999999999996,"chp":0.19999999999999996,"description":"Cryptocurrency: ETH/BTC","exchange":"ultraex.io","high_price":1.2,"low_price":1,"lp":1.2,"open_price":1,"prev_close_price":1,"short_name":"ETH/BTC","volume":3771}}],"s":"ok"}
	
* GET /quotes_ex?symbols=<ticker_name_1>,<ticker_name_2>,...,<ticker_name_n>
	Require:
		http://localhost:7933/quotes_ex?symbols=UBT:BTC,ETH:BTC
	Response:	
		{"d":[{"n":"UBT/BTC","s":"ok","v":{"amount":214654.09999999992,"ask":1.2,"bid":1.1,"ch":0.19999999999999996,"chp":0.19999999999999996,"chp_7d":0.09090909090909079,"description":"Cryptocurrency: UBT/BTC","exchange":"ultraex.io","high_price":1.2,"low_price":1,"lp":1.2,"open_price":1,"prev_close_price":1.1,"short_name":"UBT/BTC","volume":195536}},{"n":"ETH/BTC","s":"ok","v":{"amount":4142.074000000001,"ask":1.2,"bid":1,"ch":0.19999999999999996,"chp":0.19999999999999996,"chp_7d":0.19999999999999996,"description":"Cryptocurrency: ETH/BTC","exchange":"ultraex.io","high_price":1.2,"low_price":1,"lp":1.2,"open_price":1,"prev_close_price":1,"short_name":"ETH/BTC","volume":3771}}],"s":"ok"}
		
* GET /latest_trades?symbols=<ticker_name_1>,<ticker_name_2>,...,<ticker_name_n>&limit=<count>
	Require:
		http://localhost:7933/latest_trades?symbols=UBT:BTC,ETH:BTC&limit=10
	Response:
		{"d":[{"n":"UBT/BTC","s":"ok","v":[{"price":1.2,"trade_time":1545881326,"type":"ASK","volume":4},{"price":1.2,"trade_time":1545881326,"type":"BID","volume":16},{"price":1.2,"trade_time":1545881326,"type":"ASK","volume":12},{"price":1.2,"trade_time":1545881326,"type":"BID","volume":13},{"price":1.1,"trade_time":1545881326,"type":"BID","volume":6},{"price":1.1,"trade_time":1545881326,"type":"BID","volume":14},{"price":1.1,"trade_time":1545881326,"type":"ASK","volume":16},{"price":1.2,"trade_time":1545881326,"type":"ASK","volume":13},{"price":1.1,"trade_time":1545881326,"type":"ASK","volume":19},{"price":1.1,"trade_time":1545881326,"type":"BID","volume":11}]},{"n":"ETH/BTC","s":"ok","v":[{"price":1.2,"trade_time":1545881327,"type":"ASK","volume":4},{"price":1.2,"trade_time":1545881327,"type":"ASK","volume":11},{"price":1.2,"trade_time":1545881327,"type":"BID","volume":18},{"price":1.2,"trade_time":1545881327,"type":"ASK","volume":10},{"price":1.2,"trade_time":1545881327,"type":"BID","volume":11},{"price":1.2,"trade_time":1545881327,"type":"BID","volume":14},{"price":1.1,"trade_time":1545881327,"type":"ASK","volume":10},{"price":1,"trade_time":1545881327,"type":"ASK","volume":15},{"price":1.1,"trade_time":1545881327,"type":"BID","volume":11},{"price":1.2,"trade_time":1545881327,"type":"ASK","volume":15}]}],"s":"ok"}
	
* GET /level_orders?symbols=<ticker_name_1>,<ticker_name_2>,...,<ticker_name_n>&limit=<count>
	Require:
		http://localhost:7933//level_orders?symbols=UBT:BTC,ETH:BTC&limit=5
	Response:	
		{"d":[{"n":"UBT/BTC","s":"ok","vAsk":{"type":"ask","v":[{"price":1.2,"volume":41786}]},"vBid":{"type":"bid","v":[{"price":1.1,"volume":11},{"price":1,"volume":36839}]}},{"n":"ETH/BTC","s":"ok","vAsk":{"type":"ask","v":[{"price":1.2,"volume":4179}]},"vBid":{"type":"bid","v":[{"price":1,"volume":3758}]}}],"s":"ok"}
			

# Tickers Recontruct from history trades
	* constructtickersfromhistorytrades 0 Iamsuretodoitpleaseaction!
	* constructtickerfromhistorytrades UBT/BTC 0 Iamsuretodoitpleaseaction!
	* constructtickersfromhistorytradeswithfilter  0  BCX/BTC Iamsuretodoitpleaseaction!
	* constructtickersfromhistorytradeswithuninitialized 0 Iamsuretodoitpleaseaction!
	* constructtickersfromhistorytradeswithfilter  0  ZEC/BTC,BTC/ETH,AION/ETH,HOT/ETH,SUB/USDT,MTL/BTC,NEO/ETH,NPXS/ETH,STORJ/USDT,LRC/BTC,RDN/ETH,POE/ETH,GVT/BTC,KNC/ETH,KCS/ETH,RHOC/ETH,AE/USDT,DGD/USDT,DROP/USDT,ODE/BTC,HOT/BTC,THETA/BTC,DASH/ETH,BTM/ETH,BRD/USDT,GVT/USDT,BLZ/USDT,TUSD/BTC,NULS/USDT,ENJ/USDT,BCH/BTC,PPT/BTC,MKR/ETH,EBT/ETH,XIN/ETH,MKR/USDT,WICC/USDT,SNT/BTC Iamsuretodoitpleaseaction!
	
# Match Core Config:
	"MatchAlgorithm": "sortslice"
	"MatchAlgorithm": "heapmap"



