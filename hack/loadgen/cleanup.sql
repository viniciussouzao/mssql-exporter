-- Removes what setup.sql created: the loadgen_NN logins and LoadTestDB.
-- Run as sysadmin after stopping loadgen. Review before running: it drops
-- the LoadTestDB database and closes any session still using it.

SET NOCOUNT ON;

DECLARE @name sysname, @sql nvarchar(max);

DECLARE logins CURSOR LOCAL FAST_FORWARD FOR
    SELECT name FROM sys.server_principals
    WHERE type = 'S' AND name LIKE N'loadgen[_][0-9][0-9]';
OPEN logins;
FETCH NEXT FROM logins INTO @name;
WHILE @@FETCH_STATUS = 0
BEGIN
    SET @sql = N'';
    SELECT @sql += N'KILL ' + CAST(session_id AS nvarchar(10)) + N';'
    FROM sys.dm_exec_sessions WHERE login_name = @name;
    EXEC (@sql);

    SET @sql = N'DROP LOGIN ' + QUOTENAME(@name) + N';';
    EXEC (@sql);
    FETCH NEXT FROM logins INTO @name;
END
CLOSE logins;
DEALLOCATE logins;

IF DB_ID(N'LoadTestDB') IS NOT NULL
BEGIN
    ALTER DATABASE LoadTestDB SET SINGLE_USER WITH ROLLBACK IMMEDIATE;
    DROP DATABASE LoadTestDB;
END
