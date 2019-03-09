// conf
package use_mysql

import (
	"strconv"

	"../../config"
)

const (
	MODULE_NAME string = "[Mysql]: "
)

const (
	ME_DB_NAME    string = "ubtdb"
	TE_DB_NAME    string = "ultraex_tickers"
	TABLE_FINANCE string = "ubt_finance"
	TABLE_ORDER   string = "ubt_orders"
	TABLE_TRADE   string = "ubt_trade"
	TABLE_MONEY   string = "ubt_currency_user"

	TABLE_TICKERS string = "ubt_tickers"
)

var (
	DB_IP_ME   string = config.GetMEConfig().MySQL_IP_ME                          /// Default:"127.0.0.1"
	DB_PORT_ME string = strconv.FormatInt(config.GetMEConfig().MySQL_PORT_ME, 10) /// Default:"3306"
	DB_USER_ME string = config.GetMEConfig().MySQL_User_ME                        /// Default:"root"
	DB_PWD_ME  string = config.GetMEConfig().MySQL_Pwd_ME                         /// Default:"root"

	DB_IP_TE   string = config.GetMEConfig().MySQL_IP_TE                          /// Default:"127.0.0.1"
	DB_PORT_TE string = strconv.FormatInt(config.GetMEConfig().MySQL_PORT_TE, 10) /// Default:"3306"
	DB_USER_TE string = config.GetMEConfig().MySQL_User_TE                        /// Default:"root"
	DB_PWD_TE  string = config.GetMEConfig().MySQL_Pwd_TE                         /// Default:"root"
)

///"localhost:"+strconv.FormatInt(config.GetMEConfig().ManagerRPCPort, 10)

//CREATE TABLE IF NOT EXISTS `ubt_orders` (
//  `orders_id` int(11) NOT NULL AUTO_INCREMENT,
//  `member_id` int(11) NOT NULL,
//  `currency_id` int(10) NOT NULL COMMENT '主币种ID',
//  `currency_trade_id` int(10) NOT NULL COMMENT '对应交易币种ID',
//  `price` decimal(20,4) NOT NULL DEFAULT '0.0000',
//  `num` decimal(20,4) NOT NULL DEFAULT '0.0000' COMMENT '挂单数量',
//  `trade_num` decimal(20,4) NOT NULL COMMENT '成交数量',
//  `fee` decimal(20,4) NOT NULL DEFAULT '0.0000' COMMENT '记录的是比例',
//  `type` char(4) NOT NULL DEFAULT '0' COMMENT 'buy sell',
//  `add_time` int(10) NOT NULL,
//  `trade_time` int(10) NOT NULL COMMENT '成交时间',
//  `status` tinyint(4) NOT NULL DEFAULT '0' COMMENT '0是挂单，1是部分成交,2成交， -1撤销',
//  `is_robot` tinyint(4) NOT NULL COMMENT '是否机器单',
//  PRIMARY KEY (`orders_id`),
//  KEY `add_time` (`add_time`),
//  KEY `cid` (`currency_id`),
//  KEY `id` (`orders_id`),
//  KEY `member_id` (`member_id`),
//  KEY `trade_id` (`currency_trade_id`),
//  KEY `member_id_2` (`member_id`,`currency_id`,`currency_trade_id`,`price`,`num`,`trade_num`,`type`,`status`),
//  KEY `status` (`status`),
//  KEY `type` (`type`),
//  KEY `currency_trade_id` (`currency_trade_id`),
//  KEY `currency_id` (`currency_id`,`type`,`add_time`) USING BTREE,
//  KEY `price` (`price`)
//) ENGINE=InnoDB  DEFAULT CHARSET=utf8 AUTO_INCREMENT=22689 ;

//CREATE TABLE IF NOT EXISTS `ubt_trade` (
//  `trade_id` int(32) NOT NULL AUTO_INCREMENT COMMENT '交易表 交易表的id',
//  `trade_no` varchar(32) NOT NULL COMMENT '订单号',
//  `member_id` int(10) NOT NULL COMMENT '买家uid即member_id',
//  `currency_id` int(10) NOT NULL COMMENT '货币id',
//  `currency_trade_id` int(10) NOT NULL,
//  `price` decimal(20,4) NOT NULL COMMENT '价格',
//  `num` decimal(20,4) NOT NULL COMMENT '数量',
//  `money` decimal(20,4) NOT NULL,
//  `fee` decimal(20,4) NOT NULL COMMENT '手续费',
//  `type` char(10) NOT NULL COMMENT 'buy 或sell',
//  `add_time` int(10) NOT NULL COMMENT '成交时间 （添加表的时间）',
//  `status` tinyint(4) NOT NULL,
//  PRIMARY KEY (`trade_id`),
//  KEY `type` (`type`),
//  KEY `id` (`trade_id`),
//  KEY `member_id` (`member_id`),
//  KEY `currency_id` (`currency_id`),
//  KEY `currency_trade_id` (`currency_trade_id`)
//) ENGINE=MyISAM  DEFAULT CHARSET=utf8 AUTO_INCREMENT=32775 ;

//CREATE TABLE IF NOT EXISTS `ubt_currency_user` (
//  `cu_id` int(32) NOT NULL AUTO_INCREMENT,
//  `member_id` int(32) NOT NULL COMMENT '用户id',
//  `currency_id` int(32) NOT NULL COMMENT '货币id',
//  `num` decimal(40,4) NOT NULL COMMENT '数量',
//  `forzen_num` decimal(40,4) NOT NULL COMMENT '冻结数量',
//  `status` tinyint(4) NOT NULL,
//  `chongzhi_url` varchar(128) NOT NULL DEFAULT '' COMMENT '钱包充值地址',
//  PRIMARY KEY (`cu_id`),
//  KEY `member_id_2` (`member_id`,`currency_id`),
//  KEY `cu_id` (`cu_id`,`member_id`,`currency_id`,`num`,`forzen_num`,`status`)
//) ENGINE=InnoDB  DEFAULT CHARSET=utf8 ROW_FORMAT=DYNAMIC AUTO_INCREMENT=954 ;

//CREATE TABLE IF NOT EXISTS `ubt_finance` (
//  `finance_id` int(32) NOT NULL AUTO_INCREMENT COMMENT '财务日志表',
//  `member_id` int(32) NOT NULL COMMENT '用户id',
//  `type` tinyint(4) NOT NULL COMMENT '类型',
//  `content` text NOT NULL COMMENT '内容',
//  `money_type` tinyint(4) NOT NULL COMMENT '收入=1/支出=2',
//  `money` decimal(10,2) NOT NULL COMMENT '价格',
//  `add_time` int(10) NOT NULL COMMENT '添加时间',
//  `currency_id` int(10) NOT NULL COMMENT '币种',
//  `ip` varchar(64) NOT NULL,
//  PRIMARY KEY (`finance_id`),
//  KEY `种类` (`type`)
//) ENGINE=MyISAM  DEFAULT CHARSET=utf8 AUTO_INCREMENT=32832 ;

//CREATE TABLE IF NOT EXISTS `ubt_tickers_` (
//  `id` bigint(32) NOT NULL AUTO_INCREMENT,
//  `period_type` int(10) NOT NULL COMMENT 'period type: 1,5,15,30min and 1day 1week or 1month',
//  `timefrom` bigint(20) NOT NULL DEFAULT '0' COMMENT 'ticker time from(use unix nano second unit with 20 bit)',
//  `openprice` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT 'ticker: open price in this period',
//  `closeprice` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT 'ticker: close price in this period',
//  `lowprice` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT 'ticker: lowest price in this period',
//  `highprice` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT 'ticker: highest price in this period',
//  `volume` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT 'ticker: trade volume in this period',
//  `amount` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT 'ticker: trade amount in this period',
//  PRIMARY KEY (`id`),
//  UNIQUE KEY `pt` (`period_type`,`timefrom`)
//) ENGINE=InnoDB  DEFAULT CHARSET=utf8 AUTO_INCREMENT=0 ;
