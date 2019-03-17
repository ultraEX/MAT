// performance at engine simulate order test
test environment: 255symbols run at the same time on one PC with vmware ubuntu 1804, 8core with 8G Memory
ME: 
8threads enorder, cancelorder and getorders at the same time
TE: 
Get levelorders of all symbols and latest trades of all symbols and quotes of all symbols and get 1min 5min 1week k-line at the same time
Result:
ME perform: 290 trade complete rate and about 500 enorder rate(per second)
TE perform: about 12000/min complete rate




// performance at ultraex.io

Linux VPS 8core 32G with Intel(R) Xeon(R) Platinum 8175M CPU @ 2.50GHz:

Work with 399 symbols and 100coins
Memory usage: 1.02GB

Every go routine:
[Mysql]:
=======================================================================
EnOrder				UpdateTrade			
0.0021637025			0.0023327971			

[Match Core]:
=======================================================================
Quick add			Add	       
0.0000686525			0.0001302751

[Trade Engine]:
=======================================================================
0.0000001

[TE RESTful Performance]:
=======================================================================
50us~120us, average=80us=0.0008s

[Jmeter page access throughput Performance]:
one remote machine with 10M bandwith 
HOME:2700/min
ORDERS:2400/min
SYMBOL:24000/min
Resource usage:20000~30000TCP connections

Input and output use multi go routine mode and one trade pool mode
Max Match Core Performance: 0.0002
Max Mysql Performance: 0.002
Max Trade Engine Performance: 0.0000001


VM environment:
Mysql test:
3user, Mysqld: TPS=80
5000user, Mysqld: TPS=90, mysql output peak tps=180

Mysql cluster(Docker) test:
3user, 2 ndb: TPS=90
5000user, 10 ndb: TPS=100, mysql output peak tps=160

old engine with 3 index of price, time, id in one pool(bid or ask seperately):
Core TPS = 2500TPS

2019.09.15
----------------------------------------------------------
# Sortslice algorithm:
    ----------------------------------------------------------
    db 3t input: 
    =====================[Trade Statics]=====================
    Trade Complete Rate	: 162.333333
    Trade UserInput Rate	: 162.083333
    min=0.000010072, max=0.001722706, ave=0.000032230
    ----------------------------------------------------------
    no db 3t input: 
    =====================[Trade Statics]=====================
    Trade Complete Rate	: 1813.444444
    Trade UserInput Rate	: 1813.472222
    min=0.000006834, max=0.010230328, ave=0.000025026
    ----------------------------------------------------------
    no db no input: 
    =====================[Trade Statics]=====================
    Trade Complete Rate	: 4032.117647
    min=0.000015058, max=0.003351299, ave=0.000024192


# HeapMap algorith:
    ----------------------------------------------------------
    db 3t input:
    =====================[Trade Statics]=====================
    Trade Complete Rate	: 488.577320
    Trade UserInput Rate	: 488.079038
    min=0.000007725, max=0.012695990, ave=0.000031050
    ----------------------------------------------------------
    no db 3t input: 
    =====================[Trade Statics]=====================
    Trade Complete Rate	: 1804.695652
    Trade UserInput Rate	: 1804.782609
    min=0.000006757, max=0.005633696, ave=0.000025010
    ----------------------------------------------------------
    no db; no input:
    =====================[Trade Statics]=====================
    Trade Complete Rate	: 4064.787879
    Trade UserInput Rate	: 0.000000
    min=0.000013287, max=0.000626086, ave=0.000021118
    ----------------------------------------------------------
    db; no input:
    Trade Complete Rate	: 869.875000
    Trade UserInput Rate	: 0.000000
    min=0.000009847, max=0.245967610, ave=0.002060195
    ----------------------------------------------------------


Redis:
# write conn pool 50t and read long connect:
    ----------------------------------------------------------
    db 50t input:
    =====================[Trade Statics]=====================
    Trade Complete Rate	: 421.803279
    Trade UserInput Rate	: 424.098361
    ----------------------------------------------------------
    no db; 50t input
    =====================[Trade Statics]=====================
    Trade Complete Rate	: 177.000000
    Trade UserInput Rate	: 2348.000000

    read and write conn pool 50t : with echo 1 > /proc/sys/net/ipv4/tcp_tw_reuse to resolve
    ----------------------------------------------------------
    db 50t input:
    =====================[Trade Statics]=====================
    Trade Complete Rate	: 435.397260
    Trade UserInput Rate	: 435.301370
    no core block, input block
    ----------------------------------------------------------
    no db; 50t input
    =====================[Trade Statics]=====================
    Trade Complete Rate	: 134.210526
    Trade UserInput Rate	: 2330.763158
    core block, input block, tcp exhaust:
    ----------------------------------------------------------
    no db; 50t no input
    =====================[Trade Statics]=====================
    Trade Complete Rate	: 1045.714286
    Trade UserInput Rate	: 0.000000
    core block


# read and write conn pool size = 20 long conn 50t input:
    ----------------------------------------------------------
    db 50t input:
    =====================[Trade Statics]=====================
    Trade Complete Rate	: 435.397260
    Trade UserInput Rate	: 435.301370
    min=0.000319990, max=0.464930839, ave=0.002173594
    ----------------------------------------------------------
    DB no input:
    =====================[Trade Statics]=====================
    Trade Complete Rate	: 841.272727
    min=0.000392272, max=0.186604125, ave=0.002121815
    ----------------------------------------------------------
    no db; 50t input
    =====================[Trade Statics]=====================
    Trade Complete Rate	: 281.595960
    Trade UserInput Rate	: 2593.606061
    min=0.000298820, max=0.140903118, ave=0.004596609
    ----------------------------------------------------------
    no db;  no input
    =====================[Trade Statics]=====================
    Trade Complete Rate	: 1950.875000
    min=0.000370619, max=0.011265291, ave=0.000852587
    ----------------------------------------------------------
    db 1t input:
    =====================[Trade Statics]=====================
    Trade Complete Rate	: 482.392857
    Trade UserInput Rate	: 482.410714
    min=0.000319973, max=0.008891241, ave=0.000904471
    ----------------------------------------------------------
    no db; 5t input
    =====================[Trade Statics]=====================
    Trade Complete Rate	: 1502.738462
    Trade UserInput Rate	: 2073.415385
    min=0.000316392, max=0.011029822, ave=0.000680217
