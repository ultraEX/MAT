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

class ReturnInfo
{
    static public $isValidate = false;

    static public $_TSPEC = array(
        1 => array(
            'var' => 'Status',
            'isRequired' => true,
            'type' => TType::I32,
        ),
        2 => array(
            'var' => 'Info',
            'isRequired' => true,
            'type' => TType::STRING,
        ),
        3 => array(
            'var' => 'Order',
            'isRequired' => false,
            'type' => TType::STRUCT,
            'class' => '\rpc_order\Order',
        ),
    );

    /**
     * @var int
     */
    public $Status = null;
    /**
     * @var string
     */
    public $Info = null;
    /**
     * @var \rpc_order\Order
     */
    public $Order = null;

    public function __construct($vals = null)
    {
        if (is_array($vals)) {
            if (isset($vals['Status'])) {
                $this->Status = $vals['Status'];
            }
            if (isset($vals['Info'])) {
                $this->Info = $vals['Info'];
            }
            if (isset($vals['Order'])) {
                $this->Order = $vals['Order'];
            }
        }
    }

    public function getName()
    {
        return 'ReturnInfo';
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
                        $xfer += $input->readI32($this->Status);
                    } else {
                        $xfer += $input->skip($ftype);
                    }
                    break;
                case 2:
                    if ($ftype == TType::STRING) {
                        $xfer += $input->readString($this->Info);
                    } else {
                        $xfer += $input->skip($ftype);
                    }
                    break;
                case 3:
                    if ($ftype == TType::STRUCT) {
                        $this->Order = new \rpc_order\Order();
                        $xfer += $this->Order->read($input);
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
        $xfer += $output->writeStructBegin('ReturnInfo');
        if ($this->Status !== null) {
            $xfer += $output->writeFieldBegin('Status', TType::I32, 1);
            $xfer += $output->writeI32($this->Status);
            $xfer += $output->writeFieldEnd();
        }
        if ($this->Info !== null) {
            $xfer += $output->writeFieldBegin('Info', TType::STRING, 2);
            $xfer += $output->writeString($this->Info);
            $xfer += $output->writeFieldEnd();
        }
        if ($this->Order !== null) {
            if (!is_object($this->Order)) {
                throw new TProtocolException('Bad type in structure.', TProtocolException::INVALID_DATA);
            }
            $xfer += $output->writeFieldBegin('Order', TType::STRUCT, 3);
            $xfer += $this->Order->write($output);
            $xfer += $output->writeFieldEnd();
        }
        $xfer += $output->writeFieldStop();
        $xfer += $output->writeStructEnd();
        return $xfer;
    }
}
