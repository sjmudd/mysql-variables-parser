#!/bin/sh

cd examples
for v in 5.{0,1,5,6,7}; do
	dotless=$(echo "$v" | sed -e 's/\.//')
	wget -q -O- http://dev.mysql.com/doc/refman/$v/en/server-system-variables.html |\
		../mysql-variables-parser - sysvar$dotless > sysvar$dotless.sql
done
