
echo off

for /l %%i in (1,1,50) do (
	echo %%i
	set a=160
	set /a sID=%%i*%a% 
	set /a e=%%i+1
	set /a eID=%e%*%a% 
	echo %sID% %eID%
	start /b Tcli.exe -startid=%sID% -endid=%eID%
)

pause
echo "Task complete, enjoy it..."