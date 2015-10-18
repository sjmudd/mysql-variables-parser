# mysql-variables-parser
A program to parse http://dev.mysql.com/doc/refman/X/en/server-system-variables.html output

This program was built to try to collect a definition of the MySQL
variables from the documentation at the URL above.  It would be nice if mysqld would
actually allow you to query this information dynamically but that is not currently possible.

Example of how to use:

```
$ ./mysql-variables-parser 5.0/server-system-variables.html sysvar50 > 5.0/sysvar50.sql 
$ ./mysql-variables-parser 5.1/server-system-variables.html sysvar51 > 5.1/sysvar51.sql 
$ ./mysql-variables-parser 5.5/server-system-variables.html sysvar55 > 5.5/sysvar51.sql 
$ ./mysql-variables-parser 5.5/server-system-variables.html sysvar55 > 5.5/sysvar55.sql 
$ ./mysql-variables-parser 5.6/server-system-variables.html sysvar56 > 5.6/sysvar56.sql 
$ ./mysql-variables-parser 5.7/server-system-variables.html sysvar57 > 5.7/sysvar57.sql 
```

More work is needed but this is a starting point.
