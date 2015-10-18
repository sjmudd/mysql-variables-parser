# mysql-variables-parser
A program to parse http://dev.mysql.com/doc/refman/X/en/server-system-variables.html output

This program was built to try to collect a definition of the MySQL
variables from the documentation at the URL above.  It would be nice if mysqld would
actually allow you to query this information dynamically but that is not currently possible.

Example of how to use:

```
$ sh sql_generator.sh 
```

This will generate the `examples/sysvarXX.sql` files for MySQL versions 5.0 to 5.7.

More work is needed but this is a starting point.
