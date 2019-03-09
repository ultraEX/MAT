#!/bin/bash
PATH=/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin:~/bin
export PATH
#+------------------------------------
#+ 宝塔释放内存脚本
#+------------------------------------

endDate=`date +"%Y-%m-%d %H:%M:%S"`
log="开始释放内存!"
echo "★[$endDate] $log"
echo '----------------------------------------------------------------------------'

sync

if [ -f "/etc/init.d/php-fpm-52" ];then
	/etc/init.d/php-fpm-52 reload
	echo "/etc/init.d/php-fpm-52: php-fpm-52 reload complete."
fi

if [ -f "/etc/init.d/php-fpm-53" ];then
	/etc/init.d/php-fpm-53 reload
	echo "/etc/init.d/php-fpm-53: php-fpm-53 reload complete."
fi

if [ -f "/etc/init.d/php-fpm-54" ];then
	/etc/init.d/php-fpm-54 reload
	echo "/etc/init.d/php-fpm-54: php-fpm-54 reload complete."
fi

if [ -f "/etc/init.d/php-fpm-55" ];then
	/etc/init.d/php-fpm-55 reload
	echo "/etc/init.d/php-fpm-55: php-fpm-55 reload complete."
fi

if [ -f "/etc/init.d/php-fpm-56" ];then
	/etc/init.d/php-fpm-56 reload
	echo "/etc/init.d/php-fpm-56: php-fpm-56 reload complete."
fi

if [ -f "/etc/init.d/php-fpm-70" ];then
	/etc/init.d/php-fpm-70 reload
	echo "/etc/init.d/php-fpm-70: php-fpm-70 reload complete."
fi

if [ -f "/etc/init.d/php-fpm-71" ];then
	/etc/init.d/php-fpm-71 reload
	echo "/etc/init.d/php-fpm-71: php-fpm-71 reload complete."
fi

if [ -f "/etc/init.d/mysqld" ];then
	/etc/init.d/mysqld reload
	echo "/etc/init.d/mysqld: mysqld reload complete."
fi

if [ -f "/etc/init.d/nginx" ];then
	/etc/init.d/nginx reload
	echo "/etc/init.d/nginx: nginx reload complete."
fi

if [ -f "/etc/init.d/httpd" ];then
	/etc/init.d/httpd graceful
	echo "/etc/init.d/httpd: httpd graceful complete."
fi

if [ -f "/etc/init.d/pure-ftpd" ];then
	pkill -9 pure-ftpd
	sleep 0.3
	/etc/init.d/pure-ftpd start 2>/dev/null
	echo "/etc/init.d/pure-ftpd: pkill -9 pure-ftpd and start pure-ftpd complete."
fi

sync
sleep 2
sync
echo 3 > /proc/sys/vm/drop_caches

endDate=`date +"%Y-%m-%d %H:%M:%S"`
log="释放内存完成."
echo "★[$endDate] $log"
echo '----------------------------------------------------------------------------'