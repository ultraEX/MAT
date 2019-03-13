[Control Type]
	Debug Info Switch:
	setinfo [true/false]
	setlog	[true/false]

	Engine Work Status:
	trade stop
	trade pause
	trade resume

	Pool dump:
	dump true/false
	dumpcm
	dumpch
	beatheart
	faulty
	
	Exit:
	stop
	exit
	
	Power:
	exitme
	restartme
	
	Version:
	version
	
[Moduler Type]
	[Match Engine]
		Cancel Routine:
		test cancel user id symbol

		Statics Info:
		statics
		markets

		Pool dump:
		dump true/false

	[Redis]
		Redis Order Function:
		redis order add
		redis order rm user id symbol
		redis order get user id symbol
		redis order all
		redis order ones user symbol

		Redis Trade Function:
		redis trade add
		redis trade rm user id symbol
		redis trade get user id symbol
		redis trade all
		redis trade ones user symbol

	[Mysql]
		Mysql Order Function:
		mysql order add
		mysql order update
		mysql order rm user id symbol
		mysql order rm2
		mysql order get
		mysql order all
		mysql order ones

		Mysql Trade Function:
		mysql trade add
		mysql trade add2
		mysql trade rm user id symbol
		mysql trade get
		mysql trade all
		mysql trade ones

		Mysql Fund Function:
		mysql fund get user
		mysql fund freeze user aorb(1/2) price volume
		mysql fund unfreeze user aorb(1/2) price volume
		mysql fund settletx user aorb(1/2) amount
		mysql fund settlequick user aorb(1/2) amount
		mysql fund settle biduser askuser amount
		
		Mysql Finance Function:
		mysql finance get user id symbol
		mysql finance add2 bidOrderID askOrderID

		mysql tickers init sym
		mysql tickers add sym type
		mysql tickers get sym type
		mysql tickers getlimit sym type size