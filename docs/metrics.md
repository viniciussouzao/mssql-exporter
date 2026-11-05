# Metrics

## Collectors

The exporter currently includes these collectors:

- Server info: basic instance information and memory/CPU facts.
- Sessions: number of sessions grouped by state and user sessions per login.
- IO file stats: per-database file IO counters and stalls.
- Wait stats: wait time and counts by wait type.
- Memory clerks: memory usage by clerk type (`pages_kb` + `virtual_memory_committed_kb`).
- Backups: last backup timestamp (UTC, correct for servers in any time zone), duration and size per database, counting only backups taken by the instance itself (restores recorded in `backupset`, e.g. on a log shipping secondary, are ignored) and skipping backups written to the `NUL` device, which are discarded (Amazon RDS takes such backups at instance creation). On Amazon RDS, automated backups show up as daily snapshot FULL backups (duration ~1s, size = database size) and log backups every 5 minutes. On Azure SQL Managed Instance, the automatic backups of user databases are reported (recorded with a NULL `server_name`); system databases have no backup history there.
- TempDB: space usage and temp table activity.
- Performance counters: batch requests, compilations, page life expectancy, buffer cache hit ratio, deadlocks, lock waits, user connections and more.
- Databases: state, recovery model, data/log sizes and log usage per database.
- Transaction log: VLF count, active log and log generated since the last log backup and checkpoint, per database (`sys.dm_db_log_stats`, SQL Server 2016 SP2 and later).
- Blocking: blocked sessions, longest current block and requests by status.
- Active queries: long-running user requests and open user transactions, aggregated (no per-query labels).
- Availability Groups: replica role, synchronization health and send/redo queues (emits nothing when Always On is disabled).
- SQL Agent jobs: enabled state, last run outcome, timestamp and duration.
- Volumes: total/available disk space per volume hosting database files, as seen by SQL Server (works for remote monitoring); reserved/used instance storage on Azure SQL Managed Instance.
- Schedulers: runnable tasks (CPU pressure signal), active workers and work queue.
- Database mirroring: role, state, safety level, witness state and send/redo queues per mirrored database (emits nothing when no database is mirrored).
- Log shipping (opt-in): last backup/copy/restore timestamps, restore latency and alert thresholds from the log shipping monitor tables.
- Top queries (opt-in): top N queries by CPU from the plan cache, per `query_hash`, with optional statement text (literals obfuscated).
- Index fragmentation (opt-in): logical fragmentation and size per index, collected in the background on its own interval (default 6h).
- Missing indexes (opt-in): the top N indexes the optimizer reported as missing (`sys.dm_db_missing_index_*`), by estimated improvement, with their columns.
- Replication (opt-in, on the Distributor): status, last activity and delivery latency of the transactional replication Log Reader and Distribution agents, and the commands still waiting in the distribution database.
- Query Store health (opt-in): actual vs desired state (catches the silent switch to `READ_ONLY` when the store fills up), storage used and forced plans that fail, per database, collected in the background (default 5m).

## Availability & Self-Metrics

The exporter starts even if SQL Server is unreachable and retries on every
scrape — alert on `sqlserver_up` instead of watching the process:

- `sqlserver_up 0|1` — whether the instance is reachable (checked per scrape).
- `sqlserver_scrape_collector_duration_seconds{collector="wait_stats"}` — duration of each collector.
- `sqlserver_scrape_collector_success{collector="wait_stats"} 0|1` — whether each collector succeeded.

When `sqlserver_up` is 0, individual collectors are skipped for that scrape.

Every metric carries a `sqlserver_instance` label. By default it is
`@@SERVERNAME`, read at startup. If SQL Server is unreachable at that moment
the label is empty at first and the exporter retries every 15 seconds; once
SQL Server answers, the label is set without a restart (the series with the
empty label stop and new ones start). The name is not read again after
that.

Set `instance_label` (or `--sqlserver.instance-label`, or
`SQL_SERVER_INSTANCE_LABEL`) to fix the label to a name of your choice. The
exporter then uses it as is and does not read `@@SERVERNAME`, so the label is
also set when SQL Server is down at startup. Recommended on **Amazon RDS**:
there `@@SERVERNAME` is a generic host name (`EC2AMAZ-...`) that says nothing
about which instance it is, and it changes when the instance is stopped and
started or fails over. Use the DB instance identifier, e.g.
`instance_label: "orders-prod"`. On **Google Cloud SQL** `@@SERVERNAME` is a
generated host name as well; use the instance name. The
`sqlserver_servername` metric still reports the current `@@SERVERNAME`.

## Exposed Metrics

Metric names are stable and follow Prometheus best practices. Selected examples:

- Server info
  - `sqlserver_hostname{hostname="MACHINE"} 1`
  - `sqlserver_servername{servername="HOST\\INSTANCE"} 1`
  - `sqlserver_edition{edition="Developer Edition (64-bit)"} 1`
  - `sqlserver_product_version{product_version="16.0.4095.4"} 1`
  - `sqlserver_engine_edition 3` (numeric code; 8 on Azure SQL Managed Instance)
  - `sqlserver_is_hadr_enabled 0|1`
  - `sqlserver_visible_cpus 8`
  - `sqlserver_total_physical_memory_kb 67108864`
  - `sqlserver_available_physical_memory_kb 43008000`
  - A property that `SERVERPROPERTY` returns as NULL omits only its own
    metric: Azure SQL Managed Instance has no `MachineName`, so it exposes no
    `sqlserver_hostname`.

- Sessions
  - `sqlserver_sessions_total{session_state="running"} 12`
  - `sqlserver_sessions_total{session_state="sleeping"} 84`
  - `sqlserver_sessions_by_login{login_name="app_user"} 40` (user sessions
    only, `is_user_process = 1`; one series per login with open sessions,
    so expect many series where each person connects with their own
    Windows login)

- IO file stats (per file)
  - `sqlserver_io_file_reads_total{db_name="master",file_name="master",file_type="ROWS",physical_name="/var/opt/mssql/data/master.mdf"} 12345`
  - `sqlserver_io_file_writes_total{...} ...`
  - `sqlserver_io_file_read_bytes_total{...} ...`
  - `sqlserver_io_file_write_bytes_total{...} ...`
  - `sqlserver_io_file_stall_read_ms_total{...} ...`
  - `sqlserver_io_file_stall_write_ms_total{...} ...`

- Wait stats
  - `sqlserver_wait_time_ms_total{wait_type="PAGEIOLATCH_SH"} 123456`
  - `sqlserver_waiting_tasks_total{wait_type="PAGEIOLATCH_SH"} 789`
  - `sqlserver_max_wait_time_ms{wait_type="PAGEIOLATCH_SH"} 1024`
  - The first two are counters since SQL Server started; use `rate()`
    (renamed in v0.19.0, see [Upgrading](upgrading.md#v0190))
  - Idle and background waits (lazy writer, checkpoint, Always On and
    mirroring queues, Query Store and trace flushes, `SOS_WORK_DISPATCHER`
    and others that only accumulate while SQL Server is idle) and wait types
    that never happened are skipped: about 75 wait types instead of ~1,500.
    Waits that slow queries, such as `HADR_SYNC_COMMIT`, are kept.
    `--collector.wait_stats.include_idle` exposes everything again (only
    `SLEEP%`, `XE%` and `BROKER_%` excluded, as before v0.15.0).

- OS collectors (`--collector.os-collectors`, Linux and Windows), disk I/O
  per device
  - `os_disk_reads_total`, `os_disk_writes_total`, `os_disk_read_bytes_total`,
    `os_disk_write_bytes_total`, `os_disk_io_time_ms_total`
  - `os_disk_read_time_seconds_total{device="0 C:"}` and
    `os_disk_write_time_seconds_total`: time spent on reads/writes summed
    over concurrent requests; `rate()` is the average read/write queue
    length, e.g. `rate(os_disk_write_time_seconds_total[5m])`. Same
    semantics as node_exporter's `node_disk_*_time_seconds_total`.
  - `os_disk_queue_length`: requests in progress at the time of the scrape.
  - These replaced the Windows `os_disk_*_queue_length` gauges in v0.15.0
    (see [Upgrading](upgrading.md#v0150)).

- Performance counters
  - `sqlserver_batch_requests_total 656`
  - `sqlserver_sql_compilations_total 628`
  - `sqlserver_sql_recompilations_total 5`
  - `sqlserver_page_life_expectancy_seconds 10064`
  - `sqlserver_buffer_cache_hit_ratio 1`
  - `sqlserver_lazy_writes_total 0`
  - `sqlserver_deadlocks_total 0`
  - `sqlserver_lock_waits_total 5`
  - `sqlserver_user_connections 6`
  - `sqlserver_page_splits_total 60`

- Databases
  - `sqlserver_database_state{db_name="master",state="ONLINE"} 1`
  - `sqlserver_database_recovery_model{db_name="master",recovery_model="SIMPLE"} 1`
  - `sqlserver_database_log_reuse_wait{db_name="AppDB",reason="AVAILABILITY_REPLICA"} 1`:
    why the transaction log cannot be truncated (`log_reuse_wait_desc` in
    `sys.databases`; `NOTHING` when it can). `LOG_BACKUP` is the normal value
    of a FULL recovery database between log backups. SQL Server refreshes it
    when it tries to truncate the log (checkpoint or log backup), so it can
    lag a few minutes; alert with `for:` and together with the log usage, e.g.
    `sqlserver_database_log_reuse_wait{reason!~"NOTHING|LOG_BACKUP|CHECKPOINT"} == 1
    and on (instance, db_name) sqlserver_database_log_used_percent > 60`.
    `AVAILABILITY_REPLICA` points to a lagging AG secondary (see the
    `sqlserver_hadr_*` queues), `REPLICATION` to CDC or transactional
    replication, `ACTIVE_TRANSACTION` to a long-open transaction
  - `sqlserver_database_data_size_bytes{db_name="master"} 4915200`
  - `sqlserver_database_log_size_bytes{db_name="master"} 2097152`
  - `sqlserver_database_log_used_percent{db_name="master"} 71`: the `Percent
    Log Used` performance counter, reported only for databases listed in
    `sys.databases` (the counter also has instances for hidden databases such
    as `mssqlsystemresource`). On Amazon RDS the counter has no instance for
    `master`, `model`, `msdb` and `rdsadmin`, so they have no series there;
    user databases are reported normally. On Azure SQL Managed Instance the
    counter instance of a user database is its `physical_database_name` (a
    GUID); the series still carries the database name

- Transaction log (`sys.dm_db_log_stats`, SQL Server 2016 SP2 and later; on
  older versions disable it with `--collector.transaction_log=false`). Every
  `ONLINE` database the login can access; local secondary replicas of an
  Availability Group are skipped (their log follows the primary's)
  - `sqlserver_database_log_vlf_count{db_name="AppDB"} 12` and
    `sqlserver_database_log_active_vlf_count{db_name="AppDB"} 8`: thousands
    of VLFs (many small autogrowths) slow down recovery, restores and redo
  - `sqlserver_database_log_active_bytes{db_name="AppDB"} 91561984`
  - `sqlserver_database_log_since_last_log_backup_bytes{db_name="AppDB"} 91533312`:
    log generated since the last log backup; absent for databases without a
    log backup chain (SIMPLE recovery, or FULL before the first full backup).
    A value that only grows means log backups stopped:
    `deriv(sqlserver_database_log_since_last_log_backup_bytes[1h]) > 0 and sqlserver_database_log_since_last_log_backup_bytes > 10e9`
  - `sqlserver_database_log_since_last_checkpoint_bytes{db_name="AppDB"} 3293184`
  - Works on Amazon RDS for the system databases too, where the `Percent Log
    Used` counter has no instance for them

- Blocking
  - `sqlserver_blocked_sessions 0` (from `sys.dm_os_waiting_tasks`; a parallel
    request waiting on its own session is not counted)
  - `sqlserver_blocked_max_wait_ms 0`
  - `sqlserver_active_requests{status="running"} 3` (on an AG replica,
    `sys.dm_exec_requests` intermittently fails while an availability
    database is momentarily not accessible: error 976 on a secondary, 983
    while the replica is still resolving its role right after SQL Server
    starts; the collector retries once before reporting the scrape as failed)

- Active queries (user sessions only; the exporter's own request is excluded,
  and so is the `TdService` telemetry session of Azure SQL Managed Instance,
  which keeps a request open from instance start)
  - `sqlserver_requests_long_running{over_seconds="60"} 2`: user requests
    running for at least 10, 60, 300 and 1800 seconds (always emitted, 0 when
    idle)
  - `sqlserver_request_max_elapsed_seconds{command="SELECT"} 75`: longest
    running user request per command, only for commands currently running.
    Exclude commands that wait by design when alerting, e.g.
    `max(sqlserver_request_max_elapsed_seconds{command!~"WAITFOR|BACKUP.*|RESTORE.*|DBCC.*"}) > 600`
    (a Service Broker `WAITFOR (RECEIVE ...)` or a polling loop can run for days)
  - `sqlserver_open_transactions 4`
  - `sqlserver_oldest_open_transaction_seconds 620` (0 when there is none):
    long open transactions hold locks, keep row versions in tempdb and
    prevent log truncation
  - `sqlserver_sleeping_sessions_with_open_transaction 1`: idle sessions
    holding a transaction, typically a `BEGIN TRAN` never committed
  - Elapsed times are computed in server local time (`GETDATE()`), so a
    daylight saving change can skew them by one hour for requests spanning it
    (negative values are reported as 0)

- Availability Groups (only when Always On is enabled)
  - `sqlserver_hadr_replica_role{ag_name="ag1",replica_server="NODE1"} 1`
  - `sqlserver_hadr_replica_synchronization_health{ag_name="ag1",replica_server="NODE1"} 2`
  - `sqlserver_hadr_database_synchronization_state{ag_name="ag1",replica_server="NODE1",db_name="app"} 2`
  - `sqlserver_hadr_database_log_send_queue_kb{...} 0`
  - `sqlserver_hadr_database_redo_queue_kb{...} 0`
  - `sqlserver_hadr_database_is_suspended{...} 0` (data movement suspended; the
    log send queue is absent while suspended, so alert on this instead)
  - `sqlserver_hadr_database_suspend_reason{...} 0` (only while suspended; 0 = user,
    1 = partner/forced failover, 2 = redo, 3 = capture, 4 = apply, 5 = restart,
    6 = undo, 7 = revalidation, 8 = synchronization point error)
  - `sqlserver_hadr_database_secondary_lag_seconds{...} 0` (SQL Server 2016+;
    stays 0 while data movement is suspended, so don't alert on it alone)
  - Suspending on the primary marks the *primary's* database replica as
    suspended: alert on `max by (ag_name, db_name) (sqlserver_hadr_database_is_suspended) == 1`.

- SQL Agent jobs
  - `sqlserver_agent_job_enabled{job_name="nightly-backup"} 1`
  - `sqlserver_agent_job_last_run_status{job_name="nightly-backup"} 1`
  - `sqlserver_agent_job_last_run_timestamp_seconds{job_name="nightly-backup"} 1789344000`
    (UTC Unix time, converted from the server's local time with its current
    offset, since `sysjobhistory` keeps none: a run from before a daylight
    saving or time zone change is off by the size of the change until the job
    runs again)
  - `sqlserver_agent_job_last_run_duration_seconds{job_name="nightly-backup"} 42`

  There is intentionally no aggregate "failed jobs" metric — derive it in
  PromQL, which also tells you *which* jobs failed:

  ```promql
  count(sqlserver_agent_job_last_run_status == 0)   # how many jobs failed last run
  sqlserver_agent_job_last_run_status == 0          # which jobs failed last run
  ```

- Database mirroring (only for mirrored databases)
  - `sqlserver_mirroring_info{db_name="app",partner_instance="NODE2"} 1`
  - `sqlserver_mirroring_role{db_name="app"} 1` (1 = principal, 2 = mirror)
  - `sqlserver_mirroring_state{db_name="app"} 4` (0 = suspended, 1 = disconnected,
    2 = synchronizing, 3 = pending failover, 4 = synchronized,
    5 = not synchronized, 6 = synchronized/failover possible)
  - `sqlserver_mirroring_safety_level{db_name="app"} 2` (1 = off/async, 2 = full/sync)
  - `sqlserver_mirroring_witness_state{db_name="app"} 1` (only with a witness;
    1 = connected, 2 = disconnected)
  - `sqlserver_mirroring_log_send_queue_kb{db_name="app"} 0`
  - `sqlserver_mirroring_redo_queue_kb{db_name="app"} 0`
    (both partners report both queues, each from its own view; the values
    can differ slightly between them)
  - `sqlserver_mirroring_log_sent_bytes_total`, `_log_received_bytes_total`,
    `_redo_bytes_total`, `_transaction_delay_ms_total`, `_write_transactions_total`
  - Role, state, safety level, witness state and `_info` come from the
    **principal** only: on the mirror, `sys.database_mirroring` returns NULL
    unless the login is sysadmin (a `GRANT SELECT` on the view doesn't
    change that). The state is a property of the session, so the principal's
    view is enough to alert on; after a failover the new principal reports
    it. The mirror still exposes the queues and throughput counters.
    Alert with `sqlserver_mirroring_state != 4` and on the principal's
    `sqlserver_up`/`up` for a principal that is down.
  - While the mirrored database is not accessible (e.g. stuck `RECOVERING`
    on the principal), `sys.database_mirroring` returns NULL even for
    sysadmin, so the state series disappear instead of changing value. The
    performance counters are still reported by both partners, so also alert
    on mirrored databases without a state:
    `count by (db_name) (sqlserver_mirroring_log_send_queue_kb) unless on (db_name) count by (db_name) (sqlserver_mirroring_state)`
    (`sqlserver_database_state{state="RECOVERING"}` covers the same case).
  - Average commit delay in high-safety mode:
    `rate(sqlserver_mirroring_transaction_delay_ms_total[5m]) / rate(sqlserver_mirroring_write_transactions_total[5m])`

- Log shipping (opt-in, `--collector.log_shipping`; timestamps are UTC)
  - `sqlserver_log_shipping_primary_last_backup_timestamp_seconds{primary_server="NODE1",primary_database="app"} 1791892800`
  - `sqlserver_log_shipping_primary_backup_threshold_seconds{...} 3600`
  - `sqlserver_log_shipping_secondary_last_copied_timestamp_seconds{secondary_server="NODE2",secondary_database="app",primary_server="NODE1",primary_database="app"} 1791892860`
  - `sqlserver_log_shipping_secondary_last_restored_timestamp_seconds{...} 1791892920`
  - `sqlserver_log_shipping_secondary_last_restored_latency_seconds{...} 120` (minute precision)
  - `sqlserver_log_shipping_secondary_restore_threshold_seconds{...} 2700`
  - Restore falling behind, using the threshold configured in SQL Server:

    ```promql
    time() - sqlserver_log_shipping_secondary_last_restored_timestamp_seconds
      > sqlserver_log_shipping_secondary_restore_threshold_seconds
    ```

  - On SQL Server 2019 RTM and CU1 the log shipping jobs run but never
    update the monitor tables ([KB4537869](https://support.microsoft.com/help/4537869),
    job history shows "Failed to convert parameter value from a SqlGuid to a
    String"), so these timestamps stay frozen — the same values
    `sp_help_log_shipping_monitor` reports. Fixed in CU2.

- Top queries (opt-in, `--collector.top_queries`; no extra grants beyond
  `VIEW SERVER STATE`)
  - The top N queries (`--collector.top_queries.limit`, default 50) by total
    CPU time in the plan cache (`sys.dm_exec_query_stats`), aggregated by
    `query_hash` and `db_name`, the database the statement ran in: a query
    that runs in several databases (one per tenant) has one series per
    database. The N queries with the most CPU are picked first and the top N
    of their (query, database) pairs is exposed, so the series stay bounded
    by the limit
  - `sqlserver_top_query_executions_total{db_name="app",query_hash="0x1A2B3C4D5E6F7081"} 72600`
  - `sqlserver_top_query_cpu_seconds_total{...}`,
    `sqlserver_top_query_elapsed_seconds_total{...}`,
    `sqlserver_top_query_logical_reads_total{...}`,
    `sqlserver_top_query_physical_reads_total{...}`,
    `sqlserver_top_query_logical_writes_total{...}`
  - Average duration:
    `rate(sqlserver_top_query_elapsed_seconds_total[5m]) / rate(sqlserver_top_query_executions_total[5m])`
  - `sqlserver_top_query_info{db_name="app",query_hash="0x1A2B...",query_text="SELECT p.parcela FROM dbo.t p WHERE p.id = ?"} 1`
    only with `--collector.top_queries.include_text`: the statement text with
    string, numeric, hex and money (`$12.50`) literals replaced by `?`
    (`"..."` too when the plan was compiled with `QUOTED_IDENTIFIER OFF`,
    where it is a string), comments removed and
    cut to `--collector.top_queries.text_length` characters. Ad hoc
    statements carry real values (customer data, IDs) in their text, so it
    is off by default; join it in queries with
    `* on (query_hash) group_left(query_text) sqlserver_top_query_info`
  - Caveats:
    - Series come and go as queries enter and leave the top N, so `rate()`
      has gaps; and the counters only cover what is still in the plan cache:
      when a plan is evicted (memory pressure, recompile, `DBCC
      FREEPROCCACHE`) a query's counters drop, which Prometheus treats as a
      counter reset
    - The plan cache is scanned on every scrape; check
      `sqlserver_scrape_collector_duration_seconds{collector="top_queries"}`
      on instances with very large caches
    - On an idle instance the top is the exporter's own DMV queries

- Index fragmentation (opt-in, `--collector.index_fragmentation`)
  - Unlike every other collector it **does not query SQL Server on
    scrape**: `sys.dm_db_index_physical_stats` (`LIMITED` mode) reads index
    pages, which takes minutes and real I/O on large databases, so it runs in
    the background every `--collector.index_fragmentation.interval` (default
    6h, first run on the first scrape, one database at a time, each run
    bounded by the interval) and scrapes serve the last result
  - `sqlserver_index_fragmentation_percent{db_name="app",schema_name="dbo",table_name="Orders",index_name="PK_Orders"} 37.5`:
    logical fragmentation of the leaf level, weighted by page count across
    partitions; `index_name=""` for heaps
  - `sqlserver_index_page_count{...} 2048` (8 KB pages, in-row data)
  - Indexes below `--collector.index_fragmentation.min_page_count` (default
    1000 pages, ~8 MB) are skipped: fragmentation of small indexes doesn't
    matter and they would dominate the series count
  - `sqlserver_index_fragmentation_last_collection_timestamp_seconds`,
    `_last_collection_duration_seconds`, `_last_collection_databases` and
    `_last_collection_success` (0 when any database failed; the error is
    logged once per collection). The scrape itself always succeeds, so
    alert on staleness and success, e.g.
    `time() - sqlserver_index_fragmentation_last_collection_timestamp_seconds > 2 * 6 * 3600`
    and `sqlserver_index_fragmentation_last_collection_databases == 0`
    (no accessible database, usually a missing `CONNECT ANY DATABASE`)
  - Example alert (large and fragmented):
    `sqlserver_index_fragmentation_percent > 30 and on (db_name, schema_name, table_name, index_name) sqlserver_index_page_count > 10000`
  - On SSD/EBS storage, logical fragmentation matters much less than page
    density (`avg_page_space_used_in_percent`), which needs the more
    expensive `SAMPLED` mode and is not collected

- Query Store health (opt-in, `--collector.query_store`)
  - Query Store is per database and its catalog views get slow on large
    stores, so like index_fragmentation it **collects in the background**
    every `--collector.query_store.interval` (default 5m, first run on the
    first scrape) and scrapes serve the last result. Requires SQL Server
    2016+ (the views don't exist before; every database fails)
  - `sqlserver_query_store_actual_state{db_name="AppDB",state="READ_ONLY"} 1`
    and `sqlserver_query_store_desired_state{db_name="AppDB",state="READ_WRITE"} 1`
    (`OFF`, `READ_ONLY`, `READ_WRITE`, `ERROR`). Query Store switches itself
    to `READ_ONLY` when it fills up and silently stops capturing; alert when
    the actual state differs from the configured one:
    `sqlserver_query_store_desired_state == 1 unless on (instance, db_name, state) sqlserver_query_store_actual_state == 1`
  - `sqlserver_query_store_readonly_reason{db_name="AppDB"} 65536`: bitmask
    of why it is `READ_ONLY` (0 when it isn't). 65536 = max storage size
    reached, 131072 = statement limit, 262144 = in-memory limit, 524288 =
    database disk size limit; 1/2/4 = database read-only, single user,
    emergency; 8 = secondary replica, which is expected on a readable AG
    secondary (exclude it from the state alert there). To alert only on
    capture stopped by a limit: `sqlserver_query_store_readonly_reason >= 65536`
  - `sqlserver_query_store_storage_used_bytes` and
    `sqlserver_query_store_storage_max_bytes` (`MAX_STORAGE_SIZE_MB`);
    alert before it fills up, e.g. `used / max > 0.9`
  - `sqlserver_query_store_forced_plans{db_name="AppDB"} 3` and
    `sqlserver_query_store_forced_plans_failing{db_name="AppDB"} 1`: forced
    plans whose forcing has failed (`force_failure_count > 0`, e.g. the
    forced plan uses an index that was dropped), so the query runs with
    another plan while everyone believes it is pinned. See
    `last_force_failure_reason_desc` in `sys.query_store_plan`
  - `sqlserver_query_store_last_collection_timestamp_seconds`,
    `_last_collection_duration_seconds`, `_last_collection_databases` and
    `_last_collection_success`, as in index_fragmentation
  - No per-query series: the Query Store UI is the tool for analysis (see
    `ROADMAP.md`)

- Missing indexes (opt-in, `--collector.missing_indexes`): the top
  `--collector.missing_indexes.limit` (default 20) suggestions of the query
  optimizer by estimated improvement, labeled by `db_name`, `schema`,
  `table` and `index_handle`
  - `sqlserver_missing_index_improvement_measure{db_name="AppDB",schema="dbo",table="Orders",index_handle="3"} 93.4`:
    `avg_total_user_cost * avg_user_impact / 100 * (user_seeks + user_scans)`,
    the usual ranking; compare suggestions with each other, not with a fixed
    threshold
  - `sqlserver_missing_index_user_seeks_total`, `_user_scans_total`,
    `sqlserver_missing_index_avg_user_impact_percent` (estimated cost
    reduction) and `sqlserver_missing_index_avg_total_user_cost`
  - `sqlserver_missing_index_info{...,equality_columns="[CustomerID]",inequality_columns="",included_columns=""} 1`:
    the suggested key and included columns
  - The suggestions are reset when SQL Server restarts, dropped for a table
    when an index is created on it or its schema changes, and capped at 600
    by SQL Server, so series come and go. Review them before creating
    indexes: they overlap and ignore the cost of maintaining the index

- Replication (opt-in, `--collector.replication`; run it against the
  **Distributor**; transactional replication; merge and snapshot agents
  are not covered). Labels: `distribution_db`, `agent_type` (`logreader` or
  `distribution`), `agent_name`, `publisher`, `publisher_db`, `publication`,
  `subscriber`, `subscriber_db` (the last three are empty for the Log
  Reader). Virtual subscriptions of `immediate_sync` publications are skipped
  - `sqlserver_replication_agent_run_status{agent_type="distribution",...} 3`:
    status of the last history entry: 1 started, 2 succeeded, 3 in
    progress, 4 idle, 5 retrying, 6 failed. A stopped agent keeps its last
    status (2 after a clean stop), so also alert on the two metrics below
  - `sqlserver_replication_agent_last_history_timestamp_seconds{...}` (UTC,
    converted from the Distributor's local time with its current offset: an
    entry from before a daylight saving or time zone change is off by the size
    of the change until the agent logs again, which can move the stale history
    alert by that much)
    and `sqlserver_replication_agent_delivery_latency_seconds{...}` from the
    same entry
  - `sqlserver_replication_pending_commands{agent_type="distribution",...} 300`:
    commands in the distribution database not yet delivered to the
    subscription (`MSdistribution_status`). Alert when they keep growing:
    `sqlserver_replication_pending_commands > 10000 and deriv(sqlserver_replication_pending_commands[15m]) > 0`.
    On busy distributors that view can take seconds; keep an eye on
    `sqlserver_scrape_collector_duration_seconds{collector="replication"}`

- Volumes
  - `sqlserver_volume_total_bytes{volume_mount_point="C:\\"} 1962694664192`
  - `sqlserver_volume_available_bytes{volume_mount_point="C:\\"} 1846100799488`
  - SQL Server on Linux (Google Cloud SQL, the Docker images) reports no
    mount point, so the label is the directory of the database files, e.g.
    `volume_mount_point="/var/opt/mssql/data/"`. Directories on the same
    file system show up as separate series with the same values.
  - Only databases that are `ONLINE` or `RESTORING` (AG and log shipping
    secondaries) count: a database changing state could hold the query
    until the scrape timeout.
  - Azure SQL Managed Instance (`sqlserver_engine_edition 8`): no volume
    series. There `dm_os_volume_stats` reports in `total_bytes` the size of
    the files on the volume, not its capacity (smaller than
    `available_bytes`), so the collector reports the instance storage from
    `sys.server_resource_stats` instead (latest 15-second sample, published a
    few minutes late):
    - `sqlserver_instance_storage_reserved_bytes 68719476736` (the storage
      size of the instance)
    - `sqlserver_instance_storage_used_bytes 352321536`

- Schedulers (the `VISIBLE ONLINE` ones, which run user work; schedulers kept
  offline by affinity or licensing limits are left out)
  - `sqlserver_schedulers_runnable_tasks 1`
  - `sqlserver_schedulers_current_tasks 53`
  - `sqlserver_schedulers_active_workers 49`
  - `sqlserver_schedulers_work_queue 0`
  - `sqlserver_schedulers_online 8`
