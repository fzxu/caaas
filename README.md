# caaas
Assets(Images for now) as a Service written in Go and stores in Cassandra

The idea and funtions are exactly the same as [This one](https://github.com/arkxu/aaas)
But written in Golang.

## Create the keyspace and tables

```
CREATE KEYSPACE aaas WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };

USE aaas;

CREATE TABLE assets (
    id timeuuid PRIMARY KEY,
    binary blob,
    contenttype text,
    createdat timestamp,
    name text,
    path text
);

CREATE TABLE assetbypaths (
    path text,
    id timeuuid,
    name text,
    PRIMARY KEY (path, id)
);
```
