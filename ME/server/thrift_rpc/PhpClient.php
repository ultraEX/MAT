#!/usr/bin/env php
<?php
namespace thrift_rpc;
require_once("ThriftRPC.php");
error_reporting(E_ALL);

$count = 0;
$startTime = time();
$lastTime = $startTime;

$thriftRpc = new ThriftRPC();
try {
//    /// To testItf
//    TestSymbolItf("BTC/ETH", $thriftRpc);
//    /// To cancelOrder
//    $res = $thriftRpc->cancelOrder('387', 1539658023601600002, 'BTC/ETH');
//    print_r($res);
//    /// To enOrder
//    $order = new \rpc_order\Order(array(
//        'aorb' => \rpc_order\TradeType::BID,
//        'who' => "2",
//        'symbol' => 'BTC/ETH',
//        'price' => 1.2,
//        'volume' => 5,
//        'fee' => 0.001,
//        'ipAddr' => "phpClient-IP:localhost"
//    ));
//    $res = $thriftRpc->enOrder($order);
//    print_r($res);

    while(true){
        TestSymbolItf("BTC/ETH", $thriftRpc);
        //TestSymbolItf("ETH/BTC", $thriftRpc);
    }

    while (true) {
        print("########################################################### Round $count ###########################################################\n");

        TestSymbolItf("UBT/BTC", $thriftRpc);
        TestSymbolItf("BCX/BTC", $thriftRpc);
        TestSymbolItf("LTC/BTC", $thriftRpc);
        TestSymbolItf("BCH/BTC", $thriftRpc);
        TestSymbolItf("BTG/BTC", $thriftRpc);
        TestSymbolItf("ETH/BTC", $thriftRpc);
        TestSymbolItf("USDT/BTC", $thriftRpc);
        TestSymbolItf("NEO/BTC", $thriftRpc);
        TestSymbolItf("UEX/BTC", $thriftRpc);
        TestSymbolItf("EOS/BTC", $thriftRpc);
        TestSymbolItf("XMR/BTC", $thriftRpc);
        TestSymbolItf("DASH/BTC", $thriftRpc);
        TestSymbolItf("ZEC/BTC", $thriftRpc);
        TestSymbolItf("ETC/BTC", $thriftRpc);
        TestSymbolItf("UFN1/BTC", $thriftRpc);
        TestSymbolItf("OMG/BTC", $thriftRpc);
        TestSymbolItf("ZRX/BTC", $thriftRpc);
        TestSymbolItf("ZIL/BTC", $thriftRpc);
        TestSymbolItf("AE/BTC", $thriftRpc);
        TestSymbolItf("MKR/BTC", $thriftRpc);
        TestSymbolItf("BTM/BTC", $thriftRpc);
        TestSymbolItf("WTC/BTC", $thriftRpc);
        TestSymbolItf("MANA/BTC", $thriftRpc);
        TestSymbolItf("NULS/BTC", $thriftRpc);
        TestSymbolItf("ENG/BTC", $thriftRpc);
        TestSymbolItf("PMC/BTC", $thriftRpc);
        TestSymbolItf("WICC/BTC", $thriftRpc);
        TestSymbolItf("SNT/BTC", $thriftRpc);
        TestSymbolItf("PMCCC/BTC", $thriftRpc);
        TestSymbolItf("GNT/BTC", $thriftRpc);
        TestSymbolItf("PPT/BTC", $thriftRpc);
        TestSymbolItf("MITH/BTC", $thriftRpc);
        TestSymbolItf("DGD/BTC", $thriftRpc);
        TestSymbolItf("IOST/BTC", $thriftRpc);
        TestSymbolItf("LRC/BTC", $thriftRpc);
        TestSymbolItf("ELF/BTC", $thriftRpc);
        TestSymbolItf("FUN/BTC", $thriftRpc);
        TestSymbolItf("KNC/BTC", $thriftRpc);
        TestSymbolItf("LOOM/BTC", $thriftRpc);
        TestSymbolItf("KIN/BTC", $thriftRpc);
        TestSymbolItf("WAX/BTC", $thriftRpc);
        TestSymbolItf("DCN/BTC", $thriftRpc);
        TestSymbolItf("MCO/BTC", $thriftRpc);
        TestSymbolItf("BNT/BTC", $thriftRpc);
        TestSymbolItf("POWR/BTC", $thriftRpc);
        TestSymbolItf("ETHOS/BTC", $thriftRpc);
        TestSymbolItf("BAT/BTC", $thriftRpc);
        TestSymbolItf("DROP/BTC", $thriftRpc);
        TestSymbolItf("TUSD/BTC", $thriftRpc);
        TestSymbolItf("REP/BTC", $thriftRpc);
        TestSymbolItf("EBT/BTC", $thriftRpc);
        TestSymbolItf("VET/BTC", $thriftRpc);
        TestSymbolItf("NPXS/BTC", $thriftRpc);
        TestSymbolItf("KCS/BTC", $thriftRpc);
        TestSymbolItf("RHOC/BTC", $thriftRpc);
        TestSymbolItf("XIN/BTC", $thriftRpc);
        TestSymbolItf("AION/BTC", $thriftRpc);
        TestSymbolItf("QASH/BTC", $thriftRpc);
        TestSymbolItf("ODE/BTC", $thriftRpc);
        TestSymbolItf("POLY/BTC", $thriftRpc);
        TestSymbolItf("GTO/BTC", $thriftRpc);
        TestSymbolItf("CTXC/BTC", $thriftRpc);
        TestSymbolItf("VERI/BTC", $thriftRpc);
        TestSymbolItf("DENT/BTC", $thriftRpc);
        TestSymbolItf("CVC/BTC", $thriftRpc);
        TestSymbolItf("SUB/BTC", $thriftRpc);
        TestSymbolItf("ICX/BTC", $thriftRpc);
        TestSymbolItf("HOT/BTC", $thriftRpc);
        TestSymbolItf("GLOEX/BTC", $thriftRpc);
        TestSymbolItf("LINK/BTC", $thriftRpc);
        TestSymbolItf("NAS/BTC", $thriftRpc);
        TestSymbolItf("THETA/BTC", $thriftRpc);
        TestSymbolItf("ICN/BTC", $thriftRpc);
        TestSymbolItf("STORJ/BTC", $thriftRpc);
        TestSymbolItf("STORM/BTC", $thriftRpc);
        TestSymbolItf("SALT/BTC", $thriftRpc);
        TestSymbolItf("SAN/BTC", $thriftRpc);
        TestSymbolItf("ABT/BTC", $thriftRpc);
        TestSymbolItf("CMT/BTC", $thriftRpc);
        TestSymbolItf("BRD/BTC", $thriftRpc);
        TestSymbolItf("RLC/BTC", $thriftRpc);
        TestSymbolItf("GVT/BTC", $thriftRpc);
        TestSymbolItf("BTC/ETH", $thriftRpc);
        TestSymbolItf("UBT/ETH", $thriftRpc);
        TestSymbolItf("BCX/ETH", $thriftRpc);
        TestSymbolItf("LTC/ETH", $thriftRpc);
        TestSymbolItf("BCH/ETH", $thriftRpc);
        TestSymbolItf("BTG/ETH", $thriftRpc);
        TestSymbolItf("USDT/ETH", $thriftRpc);
        TestSymbolItf("NEO/ETH", $thriftRpc);
        TestSymbolItf("UEX/ETH", $thriftRpc);
        TestSymbolItf("EOS/ETH", $thriftRpc);
        TestSymbolItf("XMR/ETH", $thriftRpc);
        TestSymbolItf("DASH/ETH", $thriftRpc);
        TestSymbolItf("ZEC/ETH", $thriftRpc);
        TestSymbolItf("ETC/ETH", $thriftRpc);
        TestSymbolItf("UFN1/ETH", $thriftRpc);
        TestSymbolItf("OMG/ETH", $thriftRpc);
        TestSymbolItf("ZRX/ETH", $thriftRpc);
        TestSymbolItf("ZIL/ETH", $thriftRpc);
        TestSymbolItf("AE/ETH", $thriftRpc);
        TestSymbolItf("MKR/ETH", $thriftRpc);
        TestSymbolItf("BTM/ETH", $thriftRpc);
        TestSymbolItf("WTC/ETH", $thriftRpc);
        TestSymbolItf("MANA/ETH", $thriftRpc);
        TestSymbolItf("NULS/ETH", $thriftRpc);
        TestSymbolItf("ENG/ETH", $thriftRpc);
        TestSymbolItf("PMC/ETH", $thriftRpc);
        TestSymbolItf("WICC/ETH", $thriftRpc);
        TestSymbolItf("SNT/ETH", $thriftRpc);
        TestSymbolItf("PMCCC/ETH", $thriftRpc);
        TestSymbolItf("GNT/ETH", $thriftRpc);
        TestSymbolItf("PPT/ETH", $thriftRpc);
        TestSymbolItf("MITH/ETH", $thriftRpc);
        TestSymbolItf("DGD/ETH", $thriftRpc);
        TestSymbolItf("IOST/ETH", $thriftRpc);
        TestSymbolItf("LRC/ETH", $thriftRpc);
        TestSymbolItf("ELF/ETH", $thriftRpc);
        TestSymbolItf("FUN/ETH", $thriftRpc);
        TestSymbolItf("KNC/ETH", $thriftRpc);
        TestSymbolItf("LOOM/ETH", $thriftRpc);
        TestSymbolItf("KIN/ETH", $thriftRpc);
        TestSymbolItf("WAX/ETH", $thriftRpc);
        TestSymbolItf("DCN/ETH", $thriftRpc);
        TestSymbolItf("MCO/ETH", $thriftRpc);
        TestSymbolItf("BNT/ETH", $thriftRpc);
        TestSymbolItf("POWR/ETH", $thriftRpc);
        TestSymbolItf("ETHOS/ETH", $thriftRpc);
        TestSymbolItf("BAT/ETH", $thriftRpc);
        TestSymbolItf("DROP/ETH", $thriftRpc);
        TestSymbolItf("TUSD/ETH", $thriftRpc);
        TestSymbolItf("REP/ETH", $thriftRpc);
        TestSymbolItf("EBT/ETH", $thriftRpc);
        TestSymbolItf("VET/ETH", $thriftRpc);
        TestSymbolItf("NPXS/ETH", $thriftRpc);
        TestSymbolItf("KCS/ETH", $thriftRpc);
        TestSymbolItf("RHOC/ETH", $thriftRpc);
        TestSymbolItf("XIN/ETH", $thriftRpc);
        TestSymbolItf("AION/ETH", $thriftRpc);
        TestSymbolItf("QASH/ETH", $thriftRpc);
        TestSymbolItf("ODE/ETH", $thriftRpc);
        TestSymbolItf("POLY/ETH", $thriftRpc);
        TestSymbolItf("GTO/ETH", $thriftRpc);
        TestSymbolItf("CTXC/ETH", $thriftRpc);
        TestSymbolItf("VERI/ETH", $thriftRpc);
        TestSymbolItf("DENT/ETH", $thriftRpc);
        TestSymbolItf("CVC/ETH", $thriftRpc);
        TestSymbolItf("SUB/ETH", $thriftRpc);
        TestSymbolItf("ICX/ETH", $thriftRpc);
        TestSymbolItf("HOT/ETH", $thriftRpc);
        TestSymbolItf("GLOEX/ETH", $thriftRpc);
        TestSymbolItf("LINK/ETH", $thriftRpc);
        TestSymbolItf("NAS/ETH", $thriftRpc);
        TestSymbolItf("THETA/ETH", $thriftRpc);
        TestSymbolItf("ICN/ETH", $thriftRpc);
        TestSymbolItf("STORJ/ETH", $thriftRpc);
        TestSymbolItf("STORM/ETH", $thriftRpc);
        TestSymbolItf("SALT/ETH", $thriftRpc);
        TestSymbolItf("SAN/ETH", $thriftRpc);
        TestSymbolItf("ABT/ETH", $thriftRpc);
        TestSymbolItf("CMT/ETH", $thriftRpc);
        TestSymbolItf("BRD/ETH", $thriftRpc);
        TestSymbolItf("RLC/ETH", $thriftRpc);
        TestSymbolItf("GVT/ETH", $thriftRpc);
        TestSymbolItf("BTC/USDT", $thriftRpc);
        TestSymbolItf("UBT/USDT", $thriftRpc);
        TestSymbolItf("BCX/USDT", $thriftRpc);
        TestSymbolItf("LTC/USDT", $thriftRpc);
        TestSymbolItf("BCH/USDT", $thriftRpc);
        TestSymbolItf("BTG/USDT", $thriftRpc);
        TestSymbolItf("ETH/USDT", $thriftRpc);
        TestSymbolItf("NEO/USDT", $thriftRpc);
        TestSymbolItf("UEX/USDT", $thriftRpc);
        TestSymbolItf("EOS/USDT", $thriftRpc);
        TestSymbolItf("XMR/USDT", $thriftRpc);
        TestSymbolItf("DASH/USDT", $thriftRpc);
        TestSymbolItf("ZEC/USDT", $thriftRpc);
        TestSymbolItf("ETC/USDT", $thriftRpc);
        TestSymbolItf("UFN1/USDT", $thriftRpc);
        TestSymbolItf("OMG/USDT", $thriftRpc);
        TestSymbolItf("ZRX/USDT", $thriftRpc);
        TestSymbolItf("ZIL/USDT", $thriftRpc);
        TestSymbolItf("AE/USDT", $thriftRpc);
        TestSymbolItf("MKR/USDT", $thriftRpc);
        TestSymbolItf("BTM/USDT", $thriftRpc);
        TestSymbolItf("WTC/USDT", $thriftRpc);
        TestSymbolItf("MANA/USDT", $thriftRpc);
        TestSymbolItf("NULS/USDT", $thriftRpc);
        TestSymbolItf("ENG/USDT", $thriftRpc);
        TestSymbolItf("PMC/USDT", $thriftRpc);
        TestSymbolItf("WICC/USDT", $thriftRpc);
        TestSymbolItf("SNT/USDT", $thriftRpc);
        TestSymbolItf("PMCCC/USDT", $thriftRpc);
        TestSymbolItf("GNT/USDT", $thriftRpc);
        TestSymbolItf("PPT/USDT", $thriftRpc);
        TestSymbolItf("MITH/USDT", $thriftRpc);
        TestSymbolItf("DGD/USDT", $thriftRpc);
        TestSymbolItf("IOST/USDT", $thriftRpc);
        TestSymbolItf("LRC/USDT", $thriftRpc);
        TestSymbolItf("ELF/USDT", $thriftRpc);
        TestSymbolItf("FUN/USDT", $thriftRpc);
        TestSymbolItf("KNC/USDT", $thriftRpc);
        TestSymbolItf("LOOM/USDT", $thriftRpc);
        TestSymbolItf("KIN/USDT", $thriftRpc);
        TestSymbolItf("WAX/USDT", $thriftRpc);
        TestSymbolItf("DCN/USDT", $thriftRpc);
        TestSymbolItf("MCO/USDT", $thriftRpc);
        TestSymbolItf("BNT/USDT", $thriftRpc);
        TestSymbolItf("POWR/USDT", $thriftRpc);
        TestSymbolItf("ETHOS/USDT", $thriftRpc);
        TestSymbolItf("BAT/USDT", $thriftRpc);
        TestSymbolItf("DROP/USDT", $thriftRpc);
        TestSymbolItf("TUSD/USDT", $thriftRpc);
        TestSymbolItf("REP/USDT", $thriftRpc);
        TestSymbolItf("EBT/USDT", $thriftRpc);
        TestSymbolItf("VET/USDT", $thriftRpc);
        TestSymbolItf("NPXS/USDT", $thriftRpc);
        TestSymbolItf("KCS/USDT", $thriftRpc);
        TestSymbolItf("RHOC/USDT", $thriftRpc);
        TestSymbolItf("XIN/USDT", $thriftRpc);
        TestSymbolItf("AION/USDT", $thriftRpc);
        TestSymbolItf("QASH/USDT", $thriftRpc);
        TestSymbolItf("ODE/USDT", $thriftRpc);
        TestSymbolItf("POLY/USDT", $thriftRpc);
        TestSymbolItf("GTO/USDT", $thriftRpc);
        TestSymbolItf("CTXC/USDT", $thriftRpc);
        TestSymbolItf("VERI/USDT", $thriftRpc);
        TestSymbolItf("DENT/USDT", $thriftRpc);
        TestSymbolItf("CVC/USDT", $thriftRpc);
        TestSymbolItf("SUB/USDT", $thriftRpc);
        TestSymbolItf("ICX/USDT", $thriftRpc);
        TestSymbolItf("HOT/USDT", $thriftRpc);
        TestSymbolItf("GLOEX/USDT", $thriftRpc);
        TestSymbolItf("LINK/USDT", $thriftRpc);
        TestSymbolItf("NAS/USDT", $thriftRpc);
        TestSymbolItf("THETA/USDT", $thriftRpc);
        TestSymbolItf("ICN/USDT", $thriftRpc);
        TestSymbolItf("STORJ/USDT", $thriftRpc);
        TestSymbolItf("STORM/USDT", $thriftRpc);
        TestSymbolItf("SALT/USDT", $thriftRpc);
        TestSymbolItf("SAN/USDT", $thriftRpc);
        TestSymbolItf("ABT/USDT", $thriftRpc);
        TestSymbolItf("CMT/USDT", $thriftRpc);
        TestSymbolItf("BRD/USDT", $thriftRpc);
        TestSymbolItf("RLC/USDT", $thriftRpc);
        TestSymbolItf("GVT/USDT", $thriftRpc);

        if ((time() - $lastTime) > 120) {
            TestCancelRobotOvertimeOrders($thriftRpc, 120);
            $lastTime = time();
        }
    }
} catch (TException $tx) {
    print 'TException: ' . $tx->getMessage() . "\n";
    //print '[' . date('y-m-d h:i:s', time()) . ']' . 'bid operate out=====t:\n' . 'ask test in=====' . 'count=' . $count . '\n';
    sleep(1);
    print "\n" . "\n" . "\n";
    print "================================================Test resault:================================================\n";
    print 'ENORDER====count=' . $count . '; total second=' . ($totalTime = time() - $startTime) . '; freqIndex=' . floatval($count) / floatval($totalTime) . "\n";
    print "================================================Test resault:================================================\n";
}

function TestCancelRobotOvertimeOrders($thriftRpc, $durations)
{
    print("=======================To CancelRobotOverTimeOrder(ETH/BTC, 10)========================\n");
    $res = $thriftRpc->CancelRobotOverTimeOrder("ETH/BTC", $durations);
    printf("CancelRobotOverTimeOrder: %s\n", $res->Info);

    print("=======================To CancelRobotOverTimeOrder(LTC/BTC, 10)========================\n");
    $res = $thriftRpc->CancelRobotOverTimeOrder("LTC/BTC", $durations);
    printf("CancelRobotOverTimeOrder: %s\n", $res->Info);
}

function TestSymbolItf($symbol, $thriftRpc)
{
    global $count;

    /// ==============================================================================================
    /// construct ASK Order to ME
    $price = 1 + mt_rand(0, 20) / 100.0;
    $volume = 10 + 10 * rand(0, 9) / 10.0;
    $order = new \rpc_order\Order(array(
        'aorb' => \rpc_order\TradeType::ASK,
        'who' => "2"/*"John-PHP-Client"*/,
        'symbol' => $symbol,
        'price' => $price,
        'volume' => $volume,
        'fee' => 0.001,
        'ipAddr' => "phpClient-IP:localhost"
        /*, 'id'=>, 'timestamp'=>, 'status'=>, 'tradedVolume'=> */
    ));
    print("=======================To EnOrder($order->symbol, $order->who)========================\n");
    $res = $thriftRpc->enOrder($order);
    $user2AskID = $res->Order->id;
    DebugLog_Order($res);
    printf("EnOrder result: %s\n", $res->Info);

    $order->who = "3";
    $order->price = 1 + mt_rand(0, 2) / 10.0;
    $order->volume = 10 + 10 * rand(0, 9) / 10.0;
    print("=======================To EnOrder($order->symbol, $order->who)========================\n");
    $res = $thriftRpc->enOrder($order);
    $user3AskID = $res->Order->id;
    DebugLog_Order($res);
    printf("EnOrder result: %s\n", $res->Info);

    $order->who = "387";
    $order->price = 1 + mt_rand(0, 2) / 10.0;
    $order->volume = 10 + 10 * rand(0, 9) / 10.0;
    print("=======================To EnOrder($order->symbol, $order->who)========================\n");
    $res = $thriftRpc->enOrder($order);
    $user387AskID = $res->Order->id;
    DebugLog_Order($res);
    printf("EnOrder result: %s\n", $res->Info);

    $order->who = "388";
    $order->price = 1 + mt_rand(0, 2) / 10.0;
    $order->volume = 10 + 10 * rand(0, 9) / 10.0;
    print("=======================To EnOrder($order->symbol, $order->who)========================\n");
    $res = $thriftRpc->enOrder($order);
    $user388AskID = $res->Order->id;
    DebugLog_Order($res);
    printf("EnOrder result: %s\n", $res->Info);

    /// ==============================================================================================
    /// construct BID Order to ME
    $price = 1 + mt_rand(0, 20) / 100.0;
    $volume = 10 + 10 * rand(0, 9) / 10.0;
    $order = new \rpc_order\Order(array(
        'aorb' => \rpc_order\TradeType::BID,
        'who' => "2"/*"John-PHP-Client"*/,
        'symbol' => $symbol,
        'price' => $price,
        'volume' => $volume,
        'fee' => 0.001,
        'ipAddr' => "phpClient-IP:localhost"
        /*, 'id'=>, 'timestamp'=>, 'status'=>, 'tradedVolume'=> */
    ));
    print("=======================To EnOrder($order->symbol, $order->who)========================\n");
    $res = $thriftRpc->enOrder($order);
    $user2BidID = $res->Order->id;
    DebugLog_Order($res);
    printf("EnOrder result: %s\n", $res->Info);

    $order->who = "3";
    $order->price = 1 + mt_rand(0, 2) / 10.0;
    $order->volume = 10 + 10 * rand(0, 9) / 10.0;
    print("=======================To EnOrder($order->symbol, $order->who)========================\n");
    $res = $thriftRpc->enOrder($order);
    $user3BidID = $res->Order->id;
    DebugLog_Order($res);
    printf("EnOrder result: %s\n", $res->Info);

    $order->who = "387";
    $order->price = 1 + mt_rand(0, 2) / 10.0;
    $order->volume = 10 + 10 * rand(0, 9) / 10.0;
    print("=======================To EnOrder($order->symbol, $order->who)========================\n");
    $res = $thriftRpc->enOrder($order);
    $user387BidID = $res->Order->id;
    DebugLog_Order($res);
    printf("EnOrder result: %s\n", $res->Info);

    $order->who = "388";
    $order->price = 1 + mt_rand(0, 2) / 10.0;
    $order->volume = 10 + 10 * rand(0, 9) / 10.0;
    print("=======================To EnOrder($order->symbol, $order->who)========================\n");
    $res = $thriftRpc->enOrder($order);
    $user388BidID = $res->Order->id;
    DebugLog_Order($res);
    printf("EnOrder result: %s\n", $res->Info);

    /// ==============================================================================================
    /// to getOrder from ME
    print("=============To GetOrder($order->symbol, $user387AskID, 387)==============\n");
    $res = $thriftRpc->getOrder("387", $user387AskID, $order->symbol);
    DebugLog_Order($res);
    printf("GetOrder: %s\n", $res->Info);
    print("=============To GetOrder($order->symbol, $user387AskID, 388)==============\n");
    $res = $thriftRpc->getOrder("388", $user388BidID, $order->symbol);
    DebugLog_Order($res);
    printf("GetOrder: %s\n", $res->Info);

    /// ==============================================================================================
    /// to cancel the Order from ME
    $ids = array($user2AskID, $user3AskID, $user387AskID, $user388AskID, $user2BidID, $user3BidID, $user387BidID, $user388BidID);
    $idUser = array($user2AskID => "2", $user3AskID => "3", $user387AskID => "387", $user388AskID => "388", $user2BidID => "2", $user3BidID => "3", $user387BidID => "387", $user388BidID => "388");
    $idS = $ids[mt_rand(0, count($ids) - 1)];
    print("============To cancelOrder($order->symbol, $idS, $idUser[$idS])============\n");
    $res = $thriftRpc->cancelOrder($idUser[$idS], $idS, $order->symbol);
    DebugLog_Order($res);
    printf("cancelOrder: %s\n", $res->Info);
}

function DebugLog_Order($res)
{
    global $count;
    print
        "Operate Status:" . $res->Status .
        "\nOperate Info:" . $res->Info . "\n";
    if ($res->Status == \rpc_order\RetunStatus::SUCC) {
        print
            "\nOrder: " .
            '[count-' . $count++ . ']' .
            'Order===type:' . $res->Order->aorb .
            '; who:' . $res->Order->who .
            '; symbol:' . $res->Order->symbol .
            '; price:' . $res->Order->price .
            '; volume:' . $res->Order->volume .
            '; fee:' . $res->Order->fee .
            ';;; id:' . $res->Order->id .
            '; timestamp:' . $res->Order->timestamp .
            '; status:' . $res->Order->status .
            '; tradedVolume:' . $res->Order->tradedVolume .
            '; ipAddr:' . $res->Order->ipAddr;
        print "\n";
    }
}


?>
