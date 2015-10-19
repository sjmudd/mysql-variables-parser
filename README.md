# mysql-variables-parser
A program to parse http://dev.mysql.com/doc/refman/5.X/en/server-system-variables.html output

This program, built in go, was built to try to collect a definition of the MySQL
variables from the documentation at the URL above.  It would be nice if mysqld would
actually allow you to query this information dynamically but that is not currently possible.

Downloading can be done by doing this:

```
$ go get -u github.com/sjmudd/mysql-variables-parser
```

Example of how to use:

```
$ sh sql_generator.sh 
```

This will collect the different web pages from Oracle's site and
generate the `examples/sysvarXX.sql` files for MySQL versions 5.0
to 5.7.

More work is needed but this is a starting point.
