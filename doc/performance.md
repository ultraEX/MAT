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
    db 50t input:
    =====================[Trade Statics]=====================
    Trade Complete Rate	: 516.125000
    Trade UserInput Rate	: 517.687500
    min=0.000007848, max=0.003499953, ave=0.000047240
    ----------------------------------------------------------
    db 3t input:
    =====================[Trade Statics]=====================
    Trade Complete Rate	: 519.000000
    Trade UserInput Rate	: 519.375000
    min=0.000007348, max=0.006515375, ave=0.000022775
    ----------------------------------------------------------
    no db;  no input
    =====================[Trade Statics]=====================
    Trade Complete Rate	: 4104.285714
    min=0.000010371, max=0.000222119, ave=0.000018568
    ----------------------------------------------------------
    no db; 10t input
    =====================[Trade Statics]=====================
    Trade Complete Rate	: 2803.052632
    Trade UserInput Rate	: 2777.421053
    min=0.000005767, max=0.007608442, ave=0.000032099

# read and write conn pool size = 20 long conn  ZSET:
    ----------------------------------------------------------
    db 50t input:
    =====================[Trade Statics]=====================
    Trade Complete Rate	: 502.473684
    Trade Output Rate	: 502.473684
    min=0.000292373, max=0.007700843, ave=0.000860140
    ----------------------------------------------------------
    db 1t input:
    =====================[Trade Statics]=====================
    Trade Complete Rate	: 512.533333
    Trade UserInput Rate : 512.700000
    min=0.000295650, max=0.007213568, ave=0.000864988
    ----------------------------------------------------------
    no db;  no input
    =====================[Trade Statics]=====================
    Trade Complete Rate	: 1248.166667
    min=0.000644140, max=0.011659677, ave=0.000813154
    ----------------------------------------------------------
    no db; 3t input
    =====================[Trade Statics]=====================
    Trade Complete Rate	: 1467.142857
    Trade UserInput Rate	: 1728.666667
    min=0.000284877, max=0.017973812, ave=0.000681658


# read and write conn pool size = 20 long conn ZSET cluster:
    ----------------------------------------------------------
    db 50t input:
    =====================[Trade Statics]=====================
    Trade Complete Rate	: 496.040000
    Trade UserInput Rate	: 496.070000
    min=0.000295701, max=0.012430970, ave=0.000876474
    ----------------------------------------------------------
    db 3t input:
    =====================[Trade Statics]=====================
    Trade Complete Rate	: 500.771186
    Trade UserInput Rate	: 501.258475
    min=0.000292328, max=0.008344563, ave=0.000863759
    ----------------------------------------------------------
    no db;  no input
    =====================[Trade Statics]=====================
    Trade Complete Rate	: 1101.250000
    min=0.000457682, max=0.011416975, ave=0.000907607
    ----------------------------------------------------------
    no db; 3t input
    =====================[Trade Statics]=====================
    Trade Complete Rate	: 1404.000000
    Trade UserInput Rate	: 1664.717391
    min=0.000285335, max=0.015078506, ave=0.000727047


# MEEXCore 1level lazy skiplist algorithm: input 10 channel,out put 10 channel, bypass mysql:
    ============[Market BTC/ETH-MixHR Market MEEXCore Trade Info]===========
    ----------------------------------------------------------
    Ask Pool Scale	:	213665
    Bid Pool Scale	:	213774
    Newest Price	:	2.000000
    ----------------------------------------------------------
    =====================[Trade Statics]=====================
    Trade Complete Rate	: 2334.581028
    Trade Output Rate	: 2334.581028
    Trade UserInput Rate	: 2853.841897
    ----------------------------------------------------------
    Match core performance(second/round):
    min=0.000005527, max=0.384575573, ave=0.000268804

# MEEXCore 1level lazy skiplist algorithm: input 10 channel,out put 10 channel, bypass mysql:
    ============[Market BTC/ETH-MixHR Market MEEXCore Trade Info]===========
    ----------------------------------------------------------
    Ask Pool Scale	:	241727
    Bid Pool Scale	:	241885
    Newest Price	:	1.600000
    ----------------------------------------------------------
    =====================[Trade Statics]=====================
    Trade Complete Rate	: 2532.507692
    Trade Output Rate	: 2532.507692
    Trade UserInput Rate	: 3089.892308
    ----------------------------------------------------------
    Match core performance(second/round):
    min=0.000002139, max=0.428364849, ave=0.000038247
