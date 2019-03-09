<?php
namespace rpc_order;

/**
 * Autogenerated by Thrift Compiler (1.0.0-dev)
 *
 * DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING
 *  @generated
 */
use Thrift\Base\TBase;
use Thrift\Type\TType;
use Thrift\Type\TMessageType;
use Thrift\Exception\TException;
use Thrift\Exception\TProtocolException;
use Thrift\Protocol\TProtocol;
use Thrift\Protocol\TBinaryProtocolAccelerated;
use Thrift\Exception\TApplicationException;

class IOrder_cancelRobotOverTimeOrder_args
{
    static public $isValidate = false;

    static public $_TSPEC = array(
        1 => array(
            'var' => 'symbol',
            'isRequired' => false,
            'type' => TType::STRING,
        ),
        2 => array(
            'var' => 'durations',
            'isRequired' => false,
            'type' => TType::I64,
        ),
    );

    /**
     * @var string
     */
    public $symbol = null;
    /**
     * @var int
     */
    public $durations = null;

    public function __construct($vals = null)
    {
        if (is_array($vals)) {
            if (isset($vals['symbol'])) {
                $this->symbol = $vals['symbol'];
            }
            if (isset($vals['durations'])) {
                $this->durations = $vals['durations'];
            }
        }
    }

    public function getName()
    {
        return 'IOrder_cancelRobotOverTimeOrder_args';
    }


    public function read($input)
    {
        $xfer = 0;
        $fname = null;
        $ftype = 0;
        $fid = 0;
        $xfer += $input->readStructBegin($fname);
        while (true) {
            $xfer += $input->readFieldBegin($fname, $ftype, $fid);
            if ($ftype == TType::STOP) {
                break;
            }
            switch ($fid) {
                case 1:
                    if ($ftype == TType::STRING) {
                        $xfer += $input->readString($this->symbol);
                    } else {
                        $xfer += $input->skip($ftype);
                    }
                    break;
                case 2:
                    if ($ftype == TType::I64) {
                        $xfer += $input->readI64($this->durations);
                    } else {
                        $xfer += $input->skip($ftype);
                    }
                    break;
                default:
                    $xfer += $input->skip($ftype);
                    break;
            }
            $xfer += $input->readFieldEnd();
        }
        $xfer += $input->readStructEnd();
        return $xfer;
    }

    public function write($output)
    {
        $xfer = 0;
        $xfer += $output->writeStructBegin('IOrder_cancelRobotOverTimeOrder_args');
        if ($this->symbol !== null) {
            $xfer += $output->writeFieldBegin('symbol', TType::STRING, 1);
            $xfer += $output->writeString($this->symbol);
            $xfer += $output->writeFieldEnd();
        }
        if ($this->durations !== null) {
            $xfer += $output->writeFieldBegin('durations', TType::I64, 2);
            $xfer += $output->writeI64($this->durations);
            $xfer += $output->writeFieldEnd();
        }
        $xfer += $output->writeFieldStop();
        $xfer += $output->writeStructEnd();
        return $xfer;
    }
}
