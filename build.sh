
work_path=$(dirname $(readlink -f $0))
bin_path=${work_path}"/bin"
inst_path=${work_path}"/release"
config_path=${work_path}"/doc/config"
script_dir=${work_path}"/bin/script"

ME_dir="ME"
ME_exe="ME"
Cli_dir="Cli"
Cli_exe="ME-cli"
StartME_dir="StartME"
StartME_exe="StartME"
TCli_dir="ThriftClient"
TCli_exe="Tcli"
MEconfig="MEconfig.json"
MEconfig_ultraEx="ultraEX.json"


usage="\n
==================================================================================\n
Usage: $0 [COMMAND1, COMMAND2]\n
==================================================================================\n
Commands:\n
     --help		display this help and exit.\n
\n
  build			build all the ME exe to bin directory.\n
  install		install the output exe to release directory.\n
==================================================================================\n
"

function buildME()
{
	echo -e "Begin to build ME..."

	#########build ME
	echo -e "Build ${ME_exe}..."
	cd ${work_path}"/"${ME_dir}
	# go build -o ${bin_path}"/"${ME_exe} -ldflags "-s -w"
	go build -o ${bin_path}"/"${ME_exe} 
	
	
	#########build ME-cli
	echo -e "Build ${Cli_exe}..."
	cd ${work_path}"/"${Cli_dir}
	go build -o ${bin_path}"/"${Cli_exe} -ldflags "-s -w"
	
	#########build StartME
	echo -e "Build ${StartME_exe}..."
	cd ${work_path}"/"${StartME_dir}
	go build -o ${bin_path}"/"${StartME_exe} -ldflags "-s -w"

	#########build ThriftClient
	echo -e "Build ${TCli_exe}..."
	cd ${work_path}"/"${TCli_dir}
	go build -o ${bin_path}"/"${TCli_exe} -ldflags "-s -w"
	
	echo -e "Complete ME build."
}

######### install
function installME(){
	echo -e "Begin to install ME..."

	cd $inst_path
	cp ${bin_path}/$ME_exe ./
	cp ${bin_path}/$Cli_exe ./
	cp ${bin_path}/$StartME_exe ./
	cp ${config_path}/${MEconfig_ultraEx} ./$MEconfig
	cp -r ${script_dir} ./

	echo -e "Complete ME install."
}

######### help
function helpME(){
	echo -e $usage
	exit $?
}


##################################################################################
if [ $# -eq 0 ] || ( [ $# -ge 1 ] && [ $1 = "build" ] )
then
	buildME
fi

if [ $# -eq 1 ] && [ $1 = "--help" ] 
then
	helpME
fi

if  ( [ $# -eq 1 ] && [ $1 = "install" ] ) ||  ( [ $# -eq 2 ] && [ $2 = "install" ] && [ $1 = "build" ] )
then
	if [ ! -d $inst_path ]
	then
		mkdir $inst_path
	fi
	installME
fi

echo -e "Task complete, enjoy it..."
exit $?

