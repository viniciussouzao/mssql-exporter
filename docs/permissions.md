# Permissions

For **all default collectors** to succeed, the monitoring login needs the
following complete setup:

```sql
-- Server-level: required by the DMV-based collectors.
-- VIEW ANY DEFINITION is needed by io_file_stats, databases (sizes) and
-- volumes: without it sys.master_files silently returns zero rows.
GRANT VIEW SERVER STATE TO [your_login];
GRANT VIEW ANY DEFINITION TO [your_login];

-- msdb: required by the agent_jobs collector (enabled by default).
USE msdb;
CREATE USER [your_login] FOR LOGIN [your_login];
GRANT SELECT ON dbo.sysjobs TO [your_login];
GRANT SELECT ON dbo.sysjobhistory TO [your_login];
```

Without the msdb grants, `agent_jobs` reports
`sqlserver_scrape_collector_success{collector="agent_jobs"} 0` on every
scrape (the other collectors are unaffected); either apply the grants above
or disable it with `--collector.agent_jobs=false` if you don't monitor Agent
jobs (e.g. environments where SQL Agent is disabled).

The `log_shipping` collector is **disabled by default**. To enable it
(`--collector.log_shipping` or `collectors: {log_shipping: true}`), also grant
read access to the monitor tables on the primary and secondary instances:

```sql
USE msdb;
GRANT SELECT ON dbo.log_shipping_monitor_primary TO [your_login];
GRANT SELECT ON dbo.log_shipping_monitor_secondary TO [your_login];
```

The `index_fragmentation` and `query_store` collectors are **disabled by
default**. They read every online user database the login can access
(`HAS_DBACCESS`), and a database where the login has no user is silently
skipped; to cover all of them without creating a user in each:

```sql
GRANT CONNECT ANY DATABASE TO [your_login];
```

`CONNECT ANY DATABASE` exists since SQL Server 2014; on older versions create
a user for the login in each database. A skipped database is not an error:
without the grant the collection succeeds with
`sqlserver_index_fragmentation_last_collection_databases 0` (or
`sqlserver_query_store_last_collection_databases 0`), so compare that
metric with the number of databases you expect (or alert on `== 0`). Besides
database access, `query_store` needs no grant beyond `VIEW SERVER STATE`
(validated on SQL Server 2025).

The `missing_indexes` collector is **disabled by default**. It needs only
`VIEW SERVER STATE`, but the schema and table names come from each
database's catalog: without access to a database (`CONNECT ANY DATABASE` or
a user there) its suggestions are still exposed, with empty `schema` and
`table` labels.

The `replication` collector is **disabled by default** and only emits on a
**Distributor** (an instance with a distribution database; elsewhere it
succeeds with no series). It reads the agent tables of each distribution
database the login can access; the `replmonitor` role is not enough (it
only grants the monitoring procedures), so grant SELECT on them:

```sql
USE distribution;
CREATE USER [your_login] FOR LOGIN [your_login];
GRANT SELECT ON dbo.MSdistribution_agents TO [your_login];
GRANT SELECT ON dbo.MSdistribution_history TO [your_login];
GRANT SELECT ON dbo.MSdistribution_status TO [your_login];
GRANT SELECT ON dbo.MSlogreader_agents TO [your_login];
GRANT SELECT ON dbo.MSlogreader_history TO [your_login];
GRANT SELECT ON dbo.MSreplservers TO [your_login];
```

On **Amazon RDS** the master user cannot grant `CONNECT ANY DATABASE`
(`Grantor does not have GRANT permission`). Create a user for the login in
each database to cover instead (it grants access to the database, not to its
data):

```sql
USE [your_database];
CREATE USER [your_login] FOR LOGIN [your_login];
```

On **Google Cloud SQL** there is no `sysadmin`. The built-in `sqlserver` user
grants server permissions on behalf of its role, `CustomerDbRootRole`, which
must be named as the grantor (without `AS` the grant fails with
`Grantor does not have GRANT permission`); run it from `master`:

```sql
USE master;
GRANT VIEW SERVER STATE TO [your_login] AS [CustomerDbRootRole];
GRANT VIEW ANY DEFINITION TO [your_login] AS [CustomerDbRootRole];
```

- `CONNECT ANY DATABASE` cannot be granted and the `##MS_...##` server roles
  do not accept new members: create a user per database, as on RDS.
- `msdb` is protected: no user can be created there for your login, so
  `msdb.dbo.sysjobs` is not readable and `agent_jobs` fails. Disable it
  (`--collector.agent_jobs=false`). Logins created through Cloud SQL (console
  or API) get `CustomerDbRootRole` and can read the jobs, but that role can
  also manage logins and create databases. Cloud SQL publishes job outcomes
  itself in Cloud Monitoring (`sqlserver/sql_agent/jobs`).
- `backups` works with no grant (`public` can read `backupset`).

On **Azure SQL Managed Instance** the SQL admin of the instance is
`sysadmin`, so every grant above works as on a self-managed instance,
including `CONNECT ANY DATABASE` and the `msdb` user for `agent_jobs`
(validated on a General Purpose instance).

Notes:

- The `backups` collector reads `msdb.dbo.backupset` and
  `msdb.dbo.backupmediafamily`, which are readable by `public` on default
  installations (and on Amazon RDS) — no extra grant is usually needed.
- On SQL Server 2022+, `VIEW SERVER STATE` implies the more granular
  `VIEW SERVER PERFORMANCE STATE` / `VIEW SERVER SECURITY STATE` permissions
  that Microsoft now documents as the minimum for individual DMVs. You can go
  even more restrictive and grant only `VIEW SERVER PERFORMANCE STATE` if you
  disable the collectors that need more (e.g. `server_info` reads
  security-related server properties).

The exporter reads from (among others):

- `sys.dm_io_virtual_file_stats`
- `sys.master_files`
- `sys.dm_exec_sessions`
- `sys.dm_os_wait_stats`
- `sys.dm_os_sys_info`
- `sys.dm_os_sys_memory`
- `sys.dm_os_memory_clerks`
- `sys.dm_os_performance_counters`
- `sys.dm_os_waiting_tasks`
- `sys.databases`
- `sys.dm_db_log_stats` (transaction_log), `sys.dm_hadr_database_replica_states`
- `msdb.dbo.backupset`, `msdb.dbo.backupmediafamily`
- `sys.database_mirroring`
- `msdb.dbo.log_shipping_monitor_primary` / `log_shipping_monitor_secondary` (log_shipping only)
- `sys.dm_exec_query_stats`, `sys.dm_exec_plan_attributes`, `sys.dm_exec_sql_text` (top_queries only)
- `sys.dm_db_index_physical_stats` and each database's `sys.indexes`/`sys.objects`/`sys.schemas` (index_fragmentation only)
- Each database's `sys.database_query_store_options` and `sys.query_store_plan` (query_store only)
- `sys.dm_db_missing_index_details`, `_groups`, `_group_stats` (missing_indexes only)
- Each distribution database's `MSdistribution_agents`/`_history`/`_status`, `MSlogreader_agents`/`_history` and `MSreplservers` (replication only)
- `SERVERPROPERTY(...)`
