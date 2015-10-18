# mysql-variables-parser
A program to parse http://dev.mysql.com/doc/refman/X/en/server-system-variables.html output

This program was built to try to collect a definition of the MySQL
variables from the documentation at the URL above.  It would be nice if mysqld would
actually allow you to query this information dynamically but that is not currently possible.

Example of how to use:

```
$ cd examples
$ wget -qO server-system-variables.html.50 http://dev.mysql.com/doc/refman/5.0/en/server-system-variables.html
$ wget -qO server-system-variables.html.51 http://dev.mysql.com/doc/refman/5.1/en/server-system-variables.html
$ wget -qO server-system-variables.html.55 http://dev.mysql.com/doc/refman/5.5/en/server-system-variables.html
$ wget -qO server-system-variables.html.56 http://dev.mysql.com/doc/refman/5.6/en/server-system-variables.html
$ wget -qO server-system-variables.html.57 http://dev.mysql.com/doc/refman/5.7/en/server-system-variables.html

$ ../mysql-variables-parser server-system-variables.html.50 sysvar50 > sysvar50.sql
$ ../mysql-variables-parser server-system-variables.html.51 sysvar51 > sysvar51.sql
$ ../mysql-variables-parser server-system-variables.html.55 sysvar55 > sysvar55.sql
$ ../mysql-variables-parser server-system-variables.html.56 sysvar56 > sysvar56.sql
$ ../mysql-variables-parser server-system-variables.html.57 sysvar57 > sysvar57.sql
```

More work is needed but this is a starting point.
