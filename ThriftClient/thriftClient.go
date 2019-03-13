// thriftClient
package main

/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"../ME/config"
	"../ME/server/thrift_rpc/gen-go/rpc_order"
	"github.com/apache/thrift/lib/go/thrift"
)

const (
	MAX_TEST_CONCURRENCY int = 1
)

var defaultCtx = context.Background()
var wg *sync.WaitGroup = &sync.WaitGroup{}
var qch chan bool = make(chan bool)
var startID, endID int = 0, 0

func testEnOrder(client *rpc_order.IOrderClient) {
	wg.Add(1)
	defer wg.Done()

	fac := 0
	for {
		fac++
		for userID := startID + fac%2; userID < endID; userID += 2 {
			select {
			case <-qch:
				fmt.Println("exTestEnOrder routine exiting...")
				return
			default:
			}
			{
				/// enorder: user buy
				order := rpc_order.Order{
					Aorb:   rpc_order.TradeType_BID,
					Who:    strconv.FormatInt(int64(userID), 10),
					Symbol: "BTC/ETH",
					Price:  1.2,
					Volume: 3.5,
					Fee:    0.001,
					IpAddr: "thrift go test",
				}

				fmt.Printf("[timetick: %d]: Order{%s, %d, %f, %f} pre enorder\n", time.Now().UnixNano(), order.Who, order.Aorb, order.Price, order.Volume)
				res, e := client.EnOrder(defaultCtx, &order)
				if e != nil {
					fmt.Println(e.Error())
					continue
				}
				fmt.Printf("[timetick: %d]: Enorder respond: %s\n", time.Now().UnixNano(), res.Info)
				fmt.Printf("[timetick: %d]: Order: Who=%s, Type=%s, Price=%f, Volume=%f, ID=%d, Timestamp=%d, TradedVolume=%f, Status=%s, Ipaddr=%s\n",
					time.Now().UnixNano(),
					res.Order.Who,
					res.Order.Aorb,
					res.Order.Price,
					res.Order.Volume,
					*res.Order.ID,
					res.Order.Timestamp,
					*res.Order.TradedVolume,
					res.Order.Status,
					res.Order.IpAddr,
				)
			}

			{
				/// enorder: user sell
				order := rpc_order.Order{
					Aorb:   rpc_order.TradeType_ASK,
					Who:    strconv.FormatInt(int64(userID+1), 10),
					Symbol: "BTC/ETH",
					Price:  1.2,
					Volume: 3.5,
					Fee:    0.001,
					IpAddr: "thrift go test",
				}

				fmt.Printf("[timetick: %d]: Order{%s, %d, %f, %f} pre enorder\n", time.Now().UnixNano(), order.Who, order.Aorb, order.Price, order.Volume)
				res, e := client.EnOrder(defaultCtx, &order)
				if e != nil {
					fmt.Println(e.Error())
					continue
				}
				fmt.Printf("[timetick: %d]: Enorder respond: %s\n", time.Now().UnixNano(), res.Info)
				fmt.Printf("[timetick: %d]: Order: Who=%s, Type=%s, Price=%f, Volume=%f, ID=%d, Timestamp=%d, TradedVolume=%f, Status=%s, Ipaddr=%s\n",
					time.Now().UnixNano(),
					res.Order.Who,
					res.Order.Aorb,
					res.Order.Price,
					res.Order.Volume,
					*res.Order.ID,
					res.Order.Timestamp,
					*res.Order.TradedVolume,
					res.Order.Status,
					res.Order.IpAddr,
				)
			}
		}
	}
}
func handleClient(client *rpc_order.IOrderClient) (err error) {

	for i := 0; i < MAX_TEST_CONCURRENCY; i++ {
		go testEnOrder(client)
	}

	// fmt.Println(<-qch)
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println(<-ch)
	close(qch)

	wg.Wait()
	fmt.Printf("[timetick: %d]: handleClient complete to exit\n", time.Now().UnixNano())

	return nil
}

func runClient(transportFactory thrift.TTransportFactory, protocolFactory thrift.TProtocolFactory, addr string, secure bool) error {
	var transport thrift.TTransport
	var err error
	if secure {
		cfg := new(tls.Config)
		cfg.InsecureSkipVerify = true
		transport, err = thrift.NewTSSLSocket(addr, cfg)
	} else {
		transport, err = thrift.NewTSocket(addr)
	}
	if err != nil {
		fmt.Println("Error opening socket:", err)
		return err
	}
	transport, err = transportFactory.GetTransport(transport)
	if err != nil {
		return err
	}
	defer transport.Close()
	if err := transport.Open(); err != nil {
		return err
	}
	iprot := protocolFactory.GetProtocol(transport)
	oprot := protocolFactory.GetProtocol(transport)
	return handleClient(rpc_order.NewIOrderClient(thrift.NewTStandardClient(iprot, oprot)))
}

func Usage() {
	fmt.Fprint(os.Stderr, "Usage of ", os.Args[0], ":\n")
	flag.PrintDefaults()
	fmt.Fprint(os.Stderr, "\n")
}

func main() {
	flag.Usage = Usage
	protocol := flag.String("P", "binary", "Specify the protocol (binary, compact, json, simplejson)")
	framed := flag.Bool("framed", false, "Use framed transport")
	buffered := flag.Bool("buffered", false, "Use buffered transport")
	addr := flag.String("addr", config.GetMEConfig().CommandRPCIPPort, "Address to listen to")
	secure := flag.Bool("secure", false, "Use tls secure transport")
	flag.IntVar(&startID, "startid", 1, "Testing start user ID")
	flag.IntVar(&endID, "endid", 10, "Testing end user ID")

	flag.Parse()

	var protocolFactory thrift.TProtocolFactory
	switch *protocol {
	case "compact":
		protocolFactory = thrift.NewTCompactProtocolFactory()
	case "simplejson":
		protocolFactory = thrift.NewTSimpleJSONProtocolFactory()
	case "json":
		protocolFactory = thrift.NewTJSONProtocolFactory()
	case "binary", "":
		protocolFactory = thrift.NewTBinaryProtocolFactoryDefault()
	default:
		fmt.Fprint(os.Stderr, "Invalid protocol specified", protocol, "\n")
		Usage()
		os.Exit(1)
	}

	var transportFactory thrift.TTransportFactory
	if *buffered {
		transportFactory = thrift.NewTBufferedTransportFactory(8192)
	} else {
		transportFactory = thrift.NewTTransportFactory()
	}

	if *framed {
		transportFactory = thrift.NewTFramedTransportFactory(transportFactory)
	}

	if err := runClient(transportFactory, protocolFactory, *addr, *secure); err != nil {
		fmt.Println("error running client:", err)
	}
}
