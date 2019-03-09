# Thrift Order
# 
#
# Before running this file, you will need to have installed the thrift compiler
# into /usr/local/bin.

/**
 * The first thing to know about are types. The available types in Thrift are:
 *
 *  bool        Boolean, one byte
 *  i8 (byte)   Signed 8-bit integer
 *  i16         Signed 16-bit integer
 *  i32         Signed 32-bit integer
 *  i64         Signed 64-bit integer
 *  double      64-bit floating point value
 *  string      String
 *  binary      Blob (byte array)
 *  map<t1,t2>  Map from one type to another
 *  list<t1>    Ordered list of one type
 *  set<t1>     Set of unique elements of one type
 *
 * Did you also notice that Thrift supports C style comments?
 */

// Just in case you were wondering... yes. We support simple C comments too.

/**
 * Thrift files can reference other Thrift files to include common struct
 * and service definitions. These are found using the current path, or by
 * searching relative to any paths specified with the -I compiler flag.
 *
 * Included objects are accessed using the name of the .thrift file as a
 * prefix. i.e. shared.SharedObject
 * i.e.: include "shared.thrift"
 */


/**
 * Thrift files can namespace, package, or prefix their output in various
 * target languages.
 *
 * i.e.:
 * namespace cl tutorial
 * namespace cpp tutorial
 * namespace d tutorial
 * namespace dart tutorial
 * namespace java tutorial
 * namespace php tutorial
 * namespace perl tutorial
 * namespace haxe tutorial
 * namespace netcore tutorial
 */


/**
 * Thrift lets you do typedefs to get pretty names for your types. Standard
 * C style here.
 * i.e.: 
	typedef i32 MyInteger
 */

/**
 * Thrift also lets you define constants for use across languages. Complex
 * types and structs are specified using JSON notation.
 * i.e.: 
 * const i32 INT32CONSTANT = 9853
 * const map<string,string> MAPCONSTANT = {'hello':'world', 'goodnight':'moon'}
 */


/**
 * You can define enums, which are just 32 bit integers. Values are optional
 * and start at 1 if not supplied, C style again.
 * i.e.:
 * enum Operation {
 *   ADD = 1,      
 *   SUBTRACT = 2, 
 *   MULTIPLY = 3, 
 *   DIVIDE = 4    
 * }               
 */


/**
 * Structs are the basic complex data structures. They are comprised of fields
 * which each have an integer identifier, a type, a symbolic name, and an
 * optional default value.
 *
 * Fields can be declared "optional", which ensures they will not be included
 * in the serialized output if they aren't set.  Note that this requires some
 * manual management in some languages.
 * ie:
 * struct Work {
 *  1: i32 num1 = 0,
 *  2: i32 num2,
 *  3: Operation op,
 *  4: optional string comment,
 * }
 */


/**
 * Structs can also be exceptions, if they are nasty.
 * ie:
 * exception InvalidOperation {
 *   1: i32 whatOp,
 *   2: string why
 * }
 */

/**
 * Ahh, now onto the cool part, defining a service. Services just need a name
 * and can optionally inherit from another service using the extends keyword.
 * ie:
 * service Calculator extends shared.SharedService {
 * 
 *    * A method definition looks like C code. It has a return type, arguments,
 *    * and optionally a list of exceptions that it may throw. Note that argument
 *    * lists and exception lists are specified using the exact same syntax as
 *    * field lists in struct or exception definitions.
 * 
 *    void ping(),
 * 
 *    i32 add(1:i32 num1, 2:i32 num2),
 * 
 *    i32 calculate(1:i32 logid, 2:Work w) throws (1:InvalidOperation ouch),
 * 
 *     * This method has a oneway modifier. That means the client only makes
 *     * a request and does not listen for any response at all. Oneway methods
 *     * must be void.
 *    oneway void zip()
 * 
 * }
 */

/**
 * That just about covers the basics. Take a look in the test/ folder for more
 * detailed examples. After you run this file, your generated code shows up
 * in folders with names gen-<language>. The generated code isn't too scary
 * to look at. It even has pretty indentation.
 *
 * command:
 * thrift -r --gen go 	order.thrift
 * thrift -r --gen php 	order.thrift
 */


 

////=====================================================================================
////begin
namespace go 	rpc_order
namespace php 	rpc_order

const string MODULE_NAME = "[Thrift-GO]: "

////-------------------------------------------------------------------------------------
enum TradeType {
	BID = 1, 
	ASK = 2
}

enum OrderStatus {
	ORDER_SUBMIT = 1,			// 已挂单
	ORDER_FILLED = 2,			// 已成交
	ORDER_PARTIAL_FILLED = 3,	// 已部分成交
	ORDER_PARTIAL_CANCEL = 4,	// 已部分成交后取消订单
	ORDER_CANCELED = 5			// 已取消订单
	ORDER_CANCELING = 6			// 取消订单ing
	ORDER_STATUSUNKNOW = 7		// unknown status
}

struct Order {
	1:  required TradeType aorb,
	2:  required string who,
	3:  required string symbol,
	4:  required double price,
	5:  required double volume,				//to trade volume
	6:  required double fee,
	7:  optional i64  id,
	8:  optional i64  timestamp
	9:  optional OrderStatus status,
	10: optional double tradedVolume,		//had trade volume
	11: required string ipAddr,				//user ip address
	
}

////-------------------------------------------------------------------------------------
enum RetunStatus {
	FAIL	= 0,
	SUCC 	= 1, 
}

struct ReturnInfo {
	1:  required RetunStatus	Status,			/// status code
	2:  required string			Info,		/// return information
	3:  optional Order			Order		/// return data
}

////-------------------------------------------------------------------------------------
service IOrder{
	ReturnInfo enOrder(1: Order order),
	ReturnInfo cancelOrder(1: string user, 2: i64 id,  3: string symbol),
	ReturnInfo cancelRobotOverTimeOrder(1: string symbol, 2: i64 durations),
	ReturnInfo getOrder(1: string user, 2: i64 id,  3: string symbol),
	
}

