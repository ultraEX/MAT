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

class Order
{
    static public $isValidate = false;

    static public $_TSPEC = array(
        1 => array(
            'var' => 'aorb',
            'isRequired' => true,
            'type' => TType::I32,
        ),
        2 => array(
            'var' => 'who',
            'isRequired' => true,
            'type' => TType::STRING,
        ),
        3 => array(
            'var' => 'symbol',
            'isRequired' => true,
            'type' => TType::STRING,
        ),
        4 => array(
            'var' => 'price',
            'isRequired' => true,
            'type' => TType::DOUBLE,
        ),
        5 => array(
            'var' => 'volume',
            'isRequired' => true,
            'type' => TType::DOUBLE,
        ),
        6 => array(
            'var' => 'fee',
            'isRequired' => true,
            'type' => TType::DOUBLE,
        ),
        7 => array(
            'var' => 'id',
            'isRequired' => false,
            'type' => TType::I64,
        ),
        8 => array(
            'var' => 'timestamp',
            'isRequired' => false,
            'type' => TType::I64,
        ),
        9 => array(
            'var' => 'status',
            'isRequired' => false,
            'type' => TType::I32,
        ),
        10 => array(
            'var' => 'tradedVolume',
            'isRequired' => false,
            'type' => TType::DOUBLE,
        ),
        11 => array(
            'var' => 'ipAddr',
            'isRequired' => true,
            'type' => TType::STRING,
        ),
    );

    /**
     * @var int
     */
    public $aorb = null;
    /**
     * @var string
     */
    public $who = null;
    /**
     * @var string
     */
    public $symbol = null;
    /**
     * @var double
     */
    public $price = null;
    /**
     * @var double
     */
    public $volume = null;
    /**
     * @var double
     */
    public $fee = null;
    /**
     * @var int
     */
    public $id = null;
    /**
     * @var int
     */
    public $timestamp = null;
    /**
     * @var int
     */
    public $status = null;
    /**
     * @var double
     */
    public $tradedVolume = null;
    /**
     * @var string
     */
    public $ipAddr = null;

    public function __construct($vals = null)
    {
        if (is_array($vals)) {
            if (isset($vals['aorb'])) {
                $this->aorb = $vals['aorb'];
            }
            if (isset($vals['who'])) {
                $this->who = $vals['who'];
            }
            if (isset($vals['symbol'])) {
                $this->symbol = $vals['symbol'];
            }
            if (isset($vals['price'])) {
                $this->price = $vals['price'];
            }
            if (isset($vals['volume'])) {
                $this->volume = $vals['volume'];
            }
            if (isset($vals['fee'])) {
                $this->fee = $vals['fee'];
            }
            if (isset($vals['id'])) {
                $this->id = $vals['id'];
            }
            if (isset($vals['timestamp'])) {
                $this->timestamp = $vals['timestamp'];
            }
            if (isset($vals['status'])) {
                $this->status = $vals['status'];
            }
            if (isset($vals['tradedVolume'])) {
                $this->tradedVolume = $vals['tradedVolume'];
            }
            if (isset($vals['ipAddr'])) {
                $this->ipAddr = $vals['ipAddr'];
            }
        }
    }

    public function getName()
    {
        return 'Order';
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
                    if ($ftype == TType::I32) {
                        $xfer += $input->readI32($this->aorb);
                    } else {
                        $xfer += $input->skip($ftype);
                    }
                    break;
                case 2:
                    if ($ftype == TType::STRING) {
                        $xfer += $input->readString($this->who);
                    } else {
                        $xfer += $input->skip($ftype);
                    }
                    break;
                case 3:
                    if ($ftype == TType::STRING) {
                        $xfer += $input->readString($this->symbol);
                    } else {
                        $xfer += $input->skip($ftype);
                    }
                    break;
                case 4:
                    if ($ftype == TType::DOUBLE) {
                        $xfer += $input->readDouble($this->price);
                    } else {
                        $xfer += $input->skip($ftype);
                    }
                    break;
                case 5:
                    if ($ftype == TType::DOUBLE) {
                        $xfer += $input->readDouble($this->volume);
                    } else {
                        $xfer += $input->skip($ftype);
                    }
                    break;
                case 6:
                    if ($ftype == TType::DOUBLE) {
                        $xfer += $input->readDouble($this->fee);
                    } else {
                        $xfer += $input->skip($ftype);
                    }
                    break;
                case 7:
                    if ($ftype == TType::I64) {
                        $xfer += $input->readI64($this->id);
                    } else {
                        $xfer += $input->skip($ftype);
                    }
                    break;
                case 8:
                    if ($ftype == TType::I64) {
                        $xfer += $input->readI64($this->timestamp);
                    } else {
                        $xfer += $input->skip($ftype);
                    }
                    break;
                case 9:
                    if ($ftype == TType::I32) {
                        $xfer += $input->readI32($this->status);
                    } else {
                        $xfer += $input->skip($ftype);
                    }
                    break;
                case 10:
                    if ($ftype == TType::DOUBLE) {
                        $xfer += $input->readDouble($this->tradedVolume);
                    } else {
                        $xfer += $input->skip($ftype);
                    }
                    break;
                case 11:
                    if ($ftype == TType::STRING) {
                        $xfer += $input->readString($this->ipAddr);
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
        $xfer += $output->writeStructBegin('Order');
        if ($this->aorb !== null) {
            $xfer += $output->writeFieldBegin('aorb', TType::I32, 1);
            $xfer += $output->writeI32($this->aorb);
            $xfer += $output->writeFieldEnd();
        }
        if ($this->who !== null) {
            $xfer += $output->writeFieldBegin('who', TType::STRING, 2);
            $xfer += $output->writeString($this->who);
            $xfer += $output->writeFieldEnd();
        }
        if ($this->symbol !== null) {
            $xfer += $output->writeFieldBegin('symbol', TType::STRING, 3);
            $xfer += $output->writeString($this->symbol);
            $xfer += $output->writeFieldEnd();
        }
        if ($this->price !== null) {
            $xfer += $output->writeFieldBegin('price', TType::DOUBLE, 4);
            $xfer += $output->writeDouble($this->price);
            $xfer += $output->writeFieldEnd();
        }
        if ($this->volume !== null) {
            $xfer += $output->writeFieldBegin('volume', TType::DOUBLE, 5);
            $xfer += $output->writeDouble($this->volume);
            $xfer += $output->writeFieldEnd();
        }
        if ($this->fee !== null) {
            $xfer += $output->writeFieldBegin('fee', TType::DOUBLE, 6);
            $xfer += $output->writeDouble($this->fee);
            $xfer += $output->writeFieldEnd();
        }
        if ($this->id !== null) {
            $xfer += $output->writeFieldBegin('id', TType::I64, 7);
            $xfer += $output->writeI64($this->id);
            $xfer += $output->writeFieldEnd();
        }
        if ($this->timestamp !== null) {
            $xfer += $output->writeFieldBegin('timestamp', TType::I64, 8);
            $xfer += $output->writeI64($this->timestamp);
            $xfer += $output->writeFieldEnd();
        }
        if ($this->status !== null) {
            $xfer += $output->writeFieldBegin('status', TType::I32, 9);
            $xfer += $output->writeI32($this->status);
            $xfer += $output->writeFieldEnd();
        }
        if ($this->tradedVolume !== null) {
            $xfer += $output->writeFieldBegin('tradedVolume', TType::DOUBLE, 10);
            $xfer += $output->writeDouble($this->tradedVolume);
            $xfer += $output->writeFieldEnd();
        }
        if ($this->ipAddr !== null) {
            $xfer += $output->writeFieldBegin('ipAddr', TType::STRING, 11);
            $xfer += $output->writeString($this->ipAddr);
            $xfer += $output->writeFieldEnd();
        }
        $xfer += $output->writeFieldStop();
        $xfer += $output->writeStructEnd();
        return $xfer;
    }
}
