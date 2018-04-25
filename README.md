# mackerel-plugin-postgres-table [![Build Status](https://travis-ci.org/y-kuno/mackerel-plugin-postgres-table.svg?branch=master)](https://travis-ci.org/y-kuno/mackerel-plugin-postgres-table)

PostgreSQL Table plugin for mackerel.io agent. This repository releases an artifact to Github Releases, which satisfy the format for mkr plugin installer.

## Install

```shell
mkr plugin install y-kuno/mackerel-plugin-postgres-table [-host=<host>] [-port=<port>] [-user=<user>] [-password=<password>] [-database=<databasename>] [-sslmode=<sslmode>] [-metric-key-prefix=<prefix>]
```

## Example of mackerel-agent.conf
```
[plugin.metrics.postgres-table]
command = "/path/to/mackerel-plugin-postgres-table -user=postgres -database=databasename"
```

## Documents

* [PostgreSQL Documentation (The Statistics Collector)](https://www.postgresql.org/docs/10/static/monitoring-stats.html#PG-STAT-ALL-TABLES-VIEW)