-- Test database and logins for hack/loadgen (homologation only).
-- Run as sysadmin. Safe to re-run: existing objects are kept.
-- Set the shared password in the last batch (@password); CHECK_POLICY is
-- OFF so short test passwords are accepted. See cleanup.sql to remove
-- everything.

SET NOCOUNT ON;

IF DB_ID(N'LoadTestDB') IS NULL
    CREATE DATABASE LoadTestDB;
GO

USE LoadTestDB;
GO

IF OBJECT_ID(N'dbo.Orders') IS NULL
BEGIN
    CREATE TABLE dbo.Orders (
        id          int IDENTITY(1, 1) NOT NULL CONSTRAINT PK_Orders PRIMARY KEY,
        customer_id int            NOT NULL,
        amount      decimal(12, 2) NOT NULL,
        status      tinyint        NOT NULL,
        created_at  datetime2(0)   NOT NULL,
        note        char(200)      NOT NULL
    );

    -- 200k rows; customer_id/status spread so range queries do real work.
    INSERT INTO dbo.Orders (customer_id, amount, status, created_at, note)
    SELECT TOP (200000)
        ABS(CHECKSUM(NEWID())) % 5000,
        ABS(CHECKSUM(NEWID())) % 100000 / 100.0,
        ABS(CHECKSUM(NEWID())) % 5,
        DATEADD(MINUTE, -(ABS(CHECKSUM(NEWID())) % 525600), SYSUTCDATETIME()),
        'loadgen'
    FROM sys.all_objects a CROSS JOIN sys.all_objects b;

    CREATE INDEX IX_Orders_customer ON dbo.Orders (customer_id);
END
GO

DECLARE @password nvarchar(128) = N'change-me';
DECLARE @logins int = 10;
DECLARE @i int = 1, @name sysname, @sql nvarchar(max);

WHILE @i <= @logins
BEGIN
    SET @name = N'loadgen_' + RIGHT(N'0' + CAST(@i AS nvarchar(2)), 2);

    IF SUSER_ID(@name) IS NULL
    BEGIN
        SET @sql = N'CREATE LOGIN ' + QUOTENAME(@name)
            + N' WITH PASSWORD = N''' + REPLACE(@password, N'''', N'''''')
            + N''', DEFAULT_DATABASE = LoadTestDB, CHECK_POLICY = OFF;';
        EXEC (@sql);
    END

    IF USER_ID(@name) IS NULL
    BEGIN
        SET @sql = N'CREATE USER ' + QUOTENAME(@name) + N' FOR LOGIN ' + QUOTENAME(@name) + N';'
            + N'ALTER ROLE db_datareader ADD MEMBER ' + QUOTENAME(@name) + N';'
            + N'ALTER ROLE db_datawriter ADD MEMBER ' + QUOTENAME(@name) + N';';
        EXEC (@sql);
    END

    SET @i += 1;
END

SELECT name FROM sys.server_principals WHERE name LIKE N'loadgen[_]%' ORDER BY name;
