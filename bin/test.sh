work_path=$(dirname $(readlink -f $0))

usage="\n
==================================================================================\n
Usage: $0 COMMAND Args...\n
==================================================================================\n
Commands:\n
     --help		display this help and exit.\n
\n
  tcy			test thrift concurrency.\n

==================================================================================\n
"
######### testThriftConcurrency
function testThriftConcurrency()
{
	echo -e "Begin to test Thrift Concurrency..."

    for ((i=1; i<=10; i++))
    do
        ./Tcli & 
    done

	wait

	echo -e "Thrift Concurrency test complete."
}

######### help
function helpME(){
	echo -e $usage
	exit $?
}


##################################################################################
if [ $# -eq 0 ] || ( [ $# -eq 1 ] && [ $1 = "tcy" ] )
then
	testThriftConcurrency
fi



echo -e "Task complete, enjoy it..."
exit $?