# Encrypted connections (`encrypt=strict`)

By default the driver encrypts only the login packet. Add `encrypt=true` for a
fully encrypted TDS 7.x connection, or `encrypt=strict` for TDS 8.0 (SQL Server
2022+), where TLS is negotiated before any TDS traffic:

```
sqlserver://user:pass@NODE1:1433?database=master&encrypt=strict&certificate=C:/mssql-exporter/node1.pem
```

- SQL Server only accepts TDS 8.0 with a certificate configured for the
  instance (SQL Server Configuration Manager → Protocols → Certificate); with
  the self-generated certificate it closes the connection during the handshake.
- The driver always validates the server certificate in strict mode:
  `trustservercertificate=true` is ignored. Point it at the certificate
  instead (PEM): `certificate=<file>` validates the chain and the host name,
  `servercertificate=<file>` only compares the certificate bytes (no host name
  check). When the DSN connects by IP or alias, set
  `hostnameincertificate=<name in the certificate>`.
- Confirm on the server with `sys.dm_exec_connections`: `encrypt_option` TRUE
  and `protocol_version` 134217728 (0x08000000, TDS 8.0).
- The certificate file must end in `.pem` or `.der`. Any other extension
  (e.g. `.crt`) makes the DSN invalid and the exporter exits at startup.
- **Amazon RDS:** download the AWS CA bundle
  (`https://truststore.pki.rds.amazonaws.com/global/global-bundle.pem`) and
  use it with the instance endpoint as the host, with `encrypt=true` or
  `encrypt=strict`. Both are validated on RDS for SQL Server 2025, and no
  server-side change is needed:
  `sqlserver://user:pass@mydb.xxxx.us-east-1.rds.amazonaws.com:1433?database=master&encrypt=strict&certificate=/etc/mssql-exporter/global-bundle.pem`
- **Google Cloud SQL:** use `encrypt=true` with the instance's server CA
  (`gcloud sql instances describe <instance> --format="value(serverCaCert.cert)"`,
  saved as `.pem`). The server certificate lists the private IP and the
  instance DNS name (`<id>.<region>.sql.goog`), so connecting by IP needs no
  `hostnameincertificate`. Cloud SQL does **not** accept `encrypt=strict`: the
  handshake fails with `EOF`, even with `trustservercertificate=true`
  (validated on Cloud SQL for SQL Server 2022):
  `sqlserver://user:pass@10.0.0.3:1433?database=master&encrypt=true&certificate=/etc/mssql-exporter/cloudsql-server-ca.pem`
- **Azure SQL Managed Instance:** the certificate is issued by a public
  Microsoft CA for `*.<dns-zone>.database.windows.net`, so `encrypt=true` and
  `encrypt=strict` work with the system CAs and the instance FQDN as the host,
  with no certificate file and no server-side change; the instance refuses
  unencrypted connections (validated on a General Purpose instance):
  `sqlserver://user:pass@<instance>.<dns-zone>.database.windows.net:1433?database=master&encrypt=strict`.
  Run the exporter inside the instance's virtual network (or a peered one):
  the gateway redirects the connection to the node's private IP, so a port
  forward to the gateway alone is not enough.
