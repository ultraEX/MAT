<?php

/**
 * Created by PhpStorm.
 * User: hunter
 * Date: 18-10-10
 * Time: 下午5:06
 */

namespace thrift_rpc;

use Thrift\ClassLoader\ThriftClassLoader;


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
require_once __DIR__ . '/thrift-php/lib/Thrift/ClassLoader/ThriftClassLoader.php';
use Thrift\Protocol\TBinaryProtocol;
use Thrift\Transport\TSocket;
use Thrift\Transport\THttpClient;
use Thrift\Transport\TBufferedTransport;
use Thrift\Exception\TException;

class ThriftRPC
{
    protected $client = null;
    protected $transport = null;

    public function __construct()
    {
        $GEN_DIR = realpath(dirname(__FILE__) . '') . '/gen-php';

        $loader = new ThriftClassLoader();
        $loader->registerNamespace('Thrift', __DIR__ . '/thrift-php/lib');
        $loader->registerDefinition('rpc_order', $GEN_DIR);
        $loader->register();

        try {
            global $argv;
            if (array_search('--http', $argv)) {
                $socket = new THttpClient('localhost', 8080, '/php/PhpServer.php');
            }else{
                $socket = new TSocket('localhost', 7391);
            }
           
            $this->transport = new TBufferedTransport($socket, 1024, 1024);
            $protocol = new TBinaryProtocol($this->transport);
            $this->client = new \rpc_order\IOrderClient($protocol);
            $this->transport->open();

        } catch (TException $tx) {
            print $tx->getMessage();
        }
    }

    public function enOrder($order)
    {
        return $this->client->enOrder($order);
    }

    public function cancelOrder($user, $id, $symbol)
    {
        return $this->client->cancelOrder($user, $id, $symbol);
    }

    public function cancelRobotOverTimeOrder($symbol, $durations)
    {
        return $this->client->cancelRobotOverTimeOrder($symbol, $durations);
    }

    public function getOrder($user, $id, $symbol)
    {
        return $this->client->getOrder($user, $id, $symbol);
    }

    public function __destruct()
    {
        if ($this->transport != null) {
            $this->transport->close();
        }
    }
}