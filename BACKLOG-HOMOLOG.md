# Backlog de Homologação

Pontos encontrados durante a homologação do mssql-exporter em ambientes
simulando produção (AWS EC2 Windows/Linux e RDS for SQL Server) e, a partir
de 2026-10-20, em outras nuvens (Google Cloud SQL e Azure).

**Status:** `aberto` · `em andamento` · `resolvido` · `adiado` (virou item do ROADMAP) · `não será feito`
**Severidade:** `alta` (bloqueia uso) · `média` (atrapalha operação) · `baixa` (cosmético/ruído)

| ID | Título | Ambiente | Severidade | Status |
|----|--------|----------|------------|--------|
| HML-001 | Log `ERROR` quando `.env` não existe | EC2 Windows | baixa | resolvido (v0.1.1) |
| HML-002 | `agent_jobs` falha sem permissão no msdb (ponto de atenção) | EC2 Windows | média | resolvido na EC2 (repetir no RDS) |
| HML-003 | Coleta de disk I/O no Windows falha com `Invalid class` | EC2 Windows | média | resolvido (v0.1.1) |
| HML-004 | Gaps intermitentes: `/metrics` retorna HTTP 500 por série duplicada em `volumes` | EC2 Windows | alta | resolvido (v0.1.1) |
| HML-005 | Disco C: da EC2 quase cheio durante carga no `StressDB` (ambiente) | EC2 Windows | alta | resolvido |
| HML-006 | Exporter não roda como serviço Windows nativo (ponto de atenção) | EC2 Windows | média | resolvido (v0.2.0; `uninstall` validado em 2026-10-15) |
| HML-007 | Validação do coletor `availability_groups` em AG real (validação) | EC2 Windows (AG 01+02) | — | resolvido (v0.3.0) |
| HML-008 | Fila de envio some do gráfico com a réplica suspensa; faltam `is_suspended` e `secondary_lag_seconds` | EC2 Windows (AG 01+02) | média | resolvido (v0.3.0) |
| HML-009 | Cenários de falha: banco offline, SQL parado, reboot, kill, backup diferencial (validação) | EC2 Windows (AG 01+02) | — | resolvido |
| HML-010 | Timestamps de backup e Agent jobs tratam o horário local do SQL Server como UTC | EC2 Windows (UTC-3) | média | resolvido (v0.3.1) |
| HML-011 | Sem cobertura de database mirroring e log shipping | EC2 Windows (01+02) | média | resolvido (v0.4.0) |
| HML-012 | `backups` reporta no secundário os backups restaurados do primário | EC2 Windows (01+02) | média | resolvido (v0.4.0) |
| HML-013 | Scrape falha inteiro (`up 0`) quando o SQL demora: `scrape_timeout` do exemplo = timeout do exporter | EC2 Windows (01+02) | baixa | resolvido (v0.4.0) |
| HML-014 | Atualização do SQL Server 2019 RTM → CU32 + KB5128494 (validação) | EC2 Windows (01+02) | — | resolvido |
| HML-015 | `blocking` falha esporadicamente na réplica secundária do AG (erro 976) | EC2 Windows (02) | baixa | resolvido (v0.4.0) |
| HML-016 | Driver `denisenkom/go-mssqldb` arquivado: migrar para `microsoft/go-mssqldb` | EC2 Windows (01+02) | média | resolvido (v0.5.0) |
| HML-017 | Rolling upgrade SQL Server 2019 → 2022 (02) → 2025 (01, 02) com AG, mirroring e log shipping (validação) | EC2 Windows (01+02) | — | resolvido |
| HML-018 | Estado do mirroring some (sem série) quando o banco espelhado não está acessível | EC2 Windows (01+02) | média | resolvido (documentação, v0.5.0) |
| HML-019 | `volumes` estoura o timeout de query quando um banco está trocando de estado | EC2 Windows (01+02) | baixa | adiado (ROADMAP) |
| HML-020 | `blocking` falha com erro 983 logo após o SQL Server subir (réplica do AG em `RESOLVING`) | EC2 Windows (02) | baixa | resolvido (v0.5.0) |
| HML-021 | `sqlserver_sessions_by_login` com carga de vários logins (validação) | EC2 Windows (01) | — | resolvido |
| HML-022 | Coletor `active_queries`: requests longas e transações abertas (validação) | EC2 Windows (01) | — | resolvido |
| HML-023 | Coletor `top_queries` com texto ofuscado (validação) | EC2 Windows (01) | — | resolvido |
| HML-024 | Coletor `index_fragmentation` em background e `CONNECT ANY DATABASE` (validação) | EC2 Windows (01) | — | resolvido |
| HML-025 | Exporter como serviço Linux (systemd) no bastion da fase RDS (validação) | Linux (bastion) | — | resolvido (v0.9.1) |
| HML-027 | Linha de base no RDS Single-AZ: permissões, coletores, carga e TLS (validação) | RDS (SQL Server 2025 SE) | — | resolvido |
| HML-028 | Backups no RDS: snapshots e log a cada 5 min aparecem; backups de criação gravados em `NUL` contam como FULL | RDS (SQL Server 2025 SE) | média | resolvido (v0.9.3) |
| HML-029 | `sqlserver_instance` no RDS: nome genérico do host, muda a cada parada/partida e fica vazio se o exporter sobe antes do banco | RDS (SQL Server 2025 SE) | média | resolvido (v0.10.0) |
| HML-030 | Conversão Single-AZ → Multi-AZ e failover do RDS: Always On aparece sem falhas de coleta (validação) | RDS (SQL Server 2025 SE) | — | resolvido |
| HML-031 | Dashboard de demonstração para RDS: exporter + CloudWatch (validação) | RDS (SQL Server 2025 SE) | — | resolvido |
| HML-026 | `database_log_used_percent` com bancos ocultos e, no RDS, sem os bancos de sistema | RDS (SQL Server 2025 SE) | baixa | resolvido (v0.9.2) |
| HML-032 | Linha de base no Cloud SQL: permissões, coletores e carga (validação) | Cloud SQL (SQL Server 2022 SE) | — | resolvido |
| HML-033 | TLS no Cloud SQL: `encrypt=true` com a CA da instância; `encrypt=strict` não é aceito | Cloud SQL (SQL Server 2022 SE) | baixa | resolvido (README) |
| HML-034 | Scrapes concorrentes saturam o pool de 5 conexões: coletores falham com `context deadline exceeded` | Cloud SQL (SQL Server 2022 SE) | média | adiado (ROADMAP) |
| HML-035 | Dashboard para Cloud SQL: exporter + Cloud Monitoring (validação) | Cloud SQL (SQL Server 2022 SE) | — | resolvido |
| HML-036 | Conversão para HA regional e failover do Cloud SQL (validação) | Cloud SQL (SQL Server 2022 SE) | — | resolvido |
| HML-037 | Scrape inteiro perdido quando o servidor para de responder sem fechar a conexão | Cloud SQL (SQL Server 2022 SE) | média | resolvido (v0.12.1) |
| HML-038 | SQL Server Express sem Agent (validação local) | Docker local (SQL Server 2025 Express) | — | resolvido |
| HML-039 | Linha de base no SQL Managed Instance: permissões, coletores e carga (validação) | Azure SQL Managed Instance (GP) | — | resolvido |
| HML-040 | Free Trial da Azure: cota de Managed Instance só em `westus3`, sem VMs série B e 404 ao reaproveitar o nome do resource group (ambiente) | Azure SQL Managed Instance (GP) | — | resolvido |
| HML-041 | SQL pelo túnel SSH não chega ao MI: o gateway exige o nome da instância e redireciona para o nó (ambiente) | Azure SQL Managed Instance (GP) | baixa | resolvido (T-SQL pela bastion) |
| HML-042 | `server_info` falha inteiro com `MachineName` NULL no MI | Azure SQL Managed Instance (GP) | média | resolvido (v0.12.2) |
| HML-043 | `volumes` no MI: `total_bytes` é o tamanho dos arquivos, não do volume (disponível acima de 100%) | Azure SQL Managed Instance (GP) | média | resolvido (v0.13.0) |
| HML-044 | `database_log_used_percent` sem os bancos de usuário no MI (contador pelo nome físico) | Azure SQL Managed Instance (GP) | baixa | resolvido (v0.13.1) |
| HML-045 | `backups` sem nenhuma série no MI: `server_name` NULL no `backupset` | Azure SQL Managed Instance (GP) | média | aberto |
| HML-046 | `requests_long_running` sempre com 1 em todas as faixas no MI: sessão de telemetria `TdService` | Azure SQL Managed Instance (GP) | baixa | aberto |

## Encerramento da fase EC2 (2026-10-15)

Fase EC2 Windows encerrada com a v0.9.0 rodando na 01. Coberto:

- **Versões do SQL Server:** 2019 CU32 → 2022 → 2025 RTM com rolling upgrade
  nas duas instâncias (HML-017), sem mudança de query.
- **HA/DR:** Always On AG com 2 réplicas (failover, suspend/resume, lag),
  database mirroring (failover, parceiro fora, suspend) e log shipping.
- **Conexão:** `encrypt=strict` (TDS 8) com certificado próprio, após a
  migração para `microsoft/go-mssqldb` (HML-016).
- **Operação:** serviço Windows nativo (install/start/stop/status, reboot,
  restart pelo SCM após kill, uninstall/reinstall), Event Log, cenários de
  falha (HML-009), fuso UTC-3 (HML-010), timeout de scrape (HML-013).
- **Fase 6 do ROADMAP com carga do `hack/loadgen`:** sessões por login,
  `active_queries`, `top_queries` e `index_fragmentation` (HML-021 a HML-024).

Fica para depois:

- **Fase RDS:** repetir as permissões no `msdb` (HML-002) e testar se o
  usuário master consegue conceder `CONNECT ANY DATABASE` (HML-024).
- **ROADMAP (Later):** filtro de estado na query de `volumes` (HML-019) e
  correção de `os_disk_*_queue_length` no Windows (pendência do HML-003).
- **Limitações do laboratório:** mirroring com witness (sem terceira
  instância); conta virtual de serviço no lugar de `LocalSystem` (HML-006)
  segue como ponto a considerar.

---

## Encerramento da fase RDS (2026-10-19)

Fase RDS for SQL Server encerrada com a v0.10.0 rodando no bastion Linux.
Ambiente: RDS SQL Server 2025 Standard (`17.0.5005.3`, RTM-CU9),
`db.t3.xlarge`, subnets privadas, fuso UTC-3; exporter no bastion Ubuntu
26.04 via systemd, lido pelo Prometheus local por túnel SSH. Coberto:

- **Instalação Linux:** unit systemd com sandbox entregue na release
  (v0.9.1); instalação, atualização só trocando o binário, reinício do host
  antes do banco (HML-025).
- **Permissões:** os mesmos grants da EC2 funcionam dados pelo usuário
  master, inclusive os do `msdb` (HML-002); `CONNECT ANY DATABASE` não pode
  ser concedido no RDS — usuário por banco no lugar (HML-024, README).
- **Coletores:** os 18 com `success 1` desde a primeira coleta, sob carga do
  `hack/loadgen` e durante snapshot, conversão Multi-AZ e failover
  (HML-027, HML-030).
- **Correções que saíram do RDS:** `log_used_percent` sem bancos ocultos
  (v0.9.2, HML-026), backups em `NUL` ignorados (v0.9.3, HML-028) e
  `instance_label` para fixar o rótulo `sqlserver_instance` (v0.10.0,
  HML-029).
- **Conexão:** `encrypt=true` e `encrypt=strict` (TDS 8.0) com o bundle de CA
  da AWS, sem configuração no servidor; CA errada recusada (HML-027).
- **Backups:** snapshot manual e automático (FULL), LOG a cada 5 min com a
  retenção ligada, backup nativo pelas procs (FULL e DIFFERENTIAL), tudo em
  UTC correto com o servidor em UTC-3 (HML-028).
- **Alta disponibilidade:** conversão para Multi-AZ (Basic AGs, um por banco)
  e failover com ≈45 s de `sqlserver_up 0` e volta sozinho; rótulo estável
  com `instance_label` (HML-029, HML-030).
- **Grafana:** dashboard de demonstração exporter + CloudWatch para RDS
  (HML-031).

Fica para depois:

- **ROADMAP (Later):** métricas do `sys.dm_db_log_stats` (HML-026), rótulo
  automático resolvido na primeira conexão (HML-029) e filtro de esperas
  ociosas no `wait_stats` (HML-031).
- **Repositório privado:** o roteiro de instalação do README baixa a release
  do GitHub; enquanto o repositório for privado, os binários vão pelo S3
  (HML-025).
- **Não testado:** RDS Enterprise/Express/Web, réplicas de leitura,
  `rds.force_ssl = 1` e Enhanced Monitoring/Database Insights.

---

## Abertura da fase GCP — Cloud SQL for SQL Server (2026-10-20)

Segunda nuvem gerenciada, com o mesmo roteiro da linha de base do RDS
(HML-027). O objetivo principal é a documentação de permissões por
plataforma; o Cloud SQL difere do RDS porque não há `sysadmin` e todo `GRANT`
exige `AS CustomerDbRootRole`. Depois do GCP vem a Azure (abaixo).

**Ambiente planejado**

- Projeto no free trial do GCP (crédito válido até 2027-01-13). Recursos
  criados e destruídos com `gcloud`, com os comandos registrados nos HML;
  alerta de orçamento antes do primeiro recurso.
- Região `us-central1`. Cloud SQL for SQL Server 2022 Standard (com Agent) e
  uma passada curta no Express. IP privado, fuso UTC-3 definido na criação.
- Bastion Linux na mesma VPC com o exporter via systemd, lido pelo Prometheus
  local por túnel SSH, como no RDS.
- Cloud SQL parado entre as sessões de teste (só o disco é cobrado).

**Roteiro** (cada item vira um HML ao ser executado)

1. Linha de base: login mínimo com `VIEW SERVER STATE` e `VIEW ANY DEFINITION`
   concedidos `AS CustomerDbRootRole`, e a alternativa com
   `##MS_ServerStateReader##` / `##MS_DefinitionReader##`; coletores padrão
   com `success 1` sob carga do `hack/loadgen`.
2. `msdb`: `backups` e `agent_jobs` com o login mínimo. O `msdb` é protegido
   e só usuários criados pelo console recebem a `CustomerDbRootRole`;
   documentar a receita ou o limite.
3. Coletores por banco (`index_fragmentation`, `query_store`):
   `CONNECT ANY DATABASE` ou `##MS_DatabaseConnector##` no lugar do usuário
   por banco.
4. TLS: `encrypt=true` e `encrypt=strict` com a CA do servidor do Cloud SQL;
   o certificado não tem o IP no SAN (`hostNameInCertificate`).
5. Plataforma: `@@SERVERNAME`, `sys.dm_os_host_info`, `volumes`, fuso UTC-3
   nos timestamps (HML-010) e `instance_label`.
6. HA regional e failover: `sqlserver_up` e retorno sozinho, como no RDS
   (HML-030); conferir que nenhum AG aparece, pois o HA é por disco regional.
7. Backups automáticos e PITR do Cloud SQL no coletor `backups`.
8. Express: `msdb` sem Agent, sem custo de licença.
9. Concorrência do pool (pendência da revisão): `hey -c 5` contra o
   `/metrics` da bastion, acompanhando `scrape_collector_duration_seconds`.

**Fora do escopo**

- Réplicas de leitura (Read Scale AG, só na edição Enterprise): custo de
  licença; fica como não testado, com o risco de o coletor de AG mostrar um
  AG que o usuário não administra.
- Exporter no GKE com Helm chart: depois da release pública (ROADMAP).

**Próxima: Azure**

- SQL Managed Instance pelo free offer (General Purpose, 12 meses, endpoint
  público): permissões, storage remoto em `volumes`/`io_file_stats`, fuso
  escolhido na criação (UTC-3 para repetir o HML-010).
- SQL Database pelo free offer, só como reconhecimento: um exporter por
  banco, coletores incompatíveis desligados, para dimensionar o modo novo
  (ROADMAP).

**Pendências da revisão pós-homologação (2026-10-20)**

- `volumes`: filtro `sys.databases.state IN (0, 1)` (HML-019, ROADMAP).
  Só `ONLINE` faria sumir o volume das secundárias em `RESTORING`.
- Mirroring com witness: segue como limitação do laboratório.
- Antes da release pública: varrer o histórico com `gitleaks`; atualizar o
  `CLAUDE.md` versionado; revisar o `prometheus.yml` versionado; trocar a
  porta 9187 (alocada ao postgres_exporter) e registrar a nova na lista de
  portas do Prometheus depois de o repositório ficar público; CHANGELOG
  opcional.
- Já atendidos: `ContinueOnError` no `/metrics` (HML-004, v0.1.1), testes de
  todos os coletores por um registry pedante e dashboards com
  `${DS_PROMETHEUS}` / `${DS_CLOUDWATCH}`.

---

## Encerramento da fase GCP (2026-10-26)

Fase Google Cloud SQL for SQL Server encerrada com a v0.12.1 rodando na
bastion. Ambiente: Cloud SQL SQL Server 2022 Standard (`16.0.4262.2`) em
Linux, `db-custom-2-7680`, só IP privado, `ENCRYPTED_ONLY`, fuso UTC-3, depois
regional (HA); exporter na bastion Ubuntu 26.04 via systemd, SSH pelo IAP,
lido pelo Prometheus local por túnel. Recursos em Terraform, fora do
repositório. Coberto:

- **Permissões:** sem `sysadmin`; `VIEW SERVER STATE` e `VIEW ANY
  DEFINITION` só com `AS [CustomerDbRootRole]`; roles `##MS_...##`,
  `CONNECT ANY DATABASE` e usuários no `msdb` não são permitidos; `backups`
  funciona sem grant e `agent_jobs` fica desligado (HML-032, README).
- **Coletores:** 16 de 17 padrão e os opt-in `index_fragmentation` e
  `query_store` com `success 1`, sob carga do `hack/loadgen`; primeiro
  ambiente com SQL Server em Linux, sem mudança de query (HML-032).
- **Conexão:** `encrypt=true` com a CA da instância (IP privado no SAN);
  `encrypt=strict` não é aceito pelo Cloud SQL (HML-033, README).
- **Alta disponibilidade:** conversão para regional (≈5 min 45 s fora do ar)
  e dois failovers com carga (≈77 s e ≈60 s de `sqlserver_up 0`);
  `@@SERVERNAME` estável entre zonas (HML-036).
- **Correção que saiu do GCP:** scrape inteiro perdido quando o servidor para
  de responder sem fechar a conexão — o exporter agora responde dentro do
  prazo (v0.12.1, HML-037).
- **Concorrência:** até 2 scrapes simultâneos sem falha; 5 ou mais saturam o
  pool (HML-034).
- **Grafana:** dashboard exporter + Cloud Monitoring, com autenticação sem
  chave por impersonação (HML-035).
- **Express:** validado localmente em Docker, não no Cloud SQL (HML-038).

Fica para depois:

- **ROADMAP (Later):** singleflight para scrapes simultâneos e pool
  configurável (HML-034); rótulo de fallback no `volumes` para SQL Server em
  Linux (HML-032, HML-038).
- **Não testado:** Cloud SQL Enterprise/Web, réplicas de leitura (read-scale
  AG), restore por PITR, Cloud SQL Auth Proxy e logins criados pela API do
  Cloud SQL lendo os jobs do Agent.
- **Próxima fase:** Azure (SQL Managed Instance e reconhecimento do SQL
  Database), como planejado na abertura.

---

## Abertura da fase Azure — SQL Managed Instance (2026-10-27)

Terceira nuvem gerenciada, com o mesmo roteiro das anteriores. O Managed
Instance é o primeiro ambiente com `EngineEdition` 8: o exporter vê uma
instância completa (Agent, `msdb`, DMVs de servidor), mas com storage remoto
e nomes físicos diferentes dos de um SQL Server instalado.

**Ambiente**

- Assinatura Free Trial (US$ 200 de crédito, 30 dias). Recursos em Terraform
  (`azurerm` + `azapi`), fora do repositório, com plan/apply feitos pelo
  operador; orçamento de US$ 50/mês com alertas.
- Região `westus3` (HML-040). SQL Managed Instance pelo free offer: General
  Purpose Gen5, 4 vCores, 64 GB, formato `SQLServer2022`, só endpoint
  privado (VNet), fuso UTC-3 (`E. South America Standard Time`), agenda de
  start/stop seg–sex 13h–21h para caber nas 720 vCore-hora do mês.
- Bastion Ubuntu 26.04 (`Standard_D2s_v4`) na mesma VNet, com o exporter via
  systemd, desligada às 21h10; Prometheus local pelo túnel SSH, como no RDS e
  no Cloud SQL. T-SQL ad hoc roda na bastion com `go-sqlcmd` (HML-041).

**Roteiro** (cada item vira um HML ao ser executado)

1. Linha de base: login mínimo, coletores padrão com carga do
   `hack/loadgen` (HML-039).
2. Diferenças do MI nos coletores: `server_info`, `volumes`,
   `log_used_percent`, `backups` e `active_queries` (HML-042 a HML-046).
3. Coletores por banco (`index_fragmentation`, `query_store`) com
   `CONNECT ANY DATABASE`.
4. TLS: `encrypt=true` (em uso) e `encrypt=strict`.
5. Autenticação Microsoft Entra ID no lugar do login SQL.
6. Failover gerenciado do MI (`az sql mi failover`): `sqlserver_up` e volta
   sozinho, como no RDS (HML-030) e no Cloud SQL (HML-036).
7. Dashboard exporter + Azure Monitor.
8. SQL Database pelo free offer, só como reconhecimento: um exporter por
   banco, coletores incompatíveis desligados, para dimensionar o modo novo
   (ROADMAP).

---

## HML-001 — Log `ERROR` quando `.env` não existe

- **Ambiente:** EC2 Windows, release v0.1.0 (`mssql-exporter_0.1.0_windows_amd64.zip`)
- **Encontrado em:** 2026-10-08
- **Severidade:** baixa
- **Status:** resolvido (v0.1.1)

**Reprodução**

```powershell
.\mssql-exporter.exe --config.file config.yml
```

**Observado**

```
ERROR Error loading .env file error="open .env: The system cannot find the file specified."
```

O exporter segue funcionando normalmente (conecta, faz ping, expõe `:9187/metrics`).

**Causa**

`main.go` chama `godotenv.Load()` incondicionalmente e loga qualquer falha em
nível `ERROR`. O `.env` é uma conveniência de desenvolvimento local; em
produção a configuração vem de `--config.file` ou variáveis de ambiente, então
a ausência do arquivo é o caso normal. Rodando como serviço, o diretório de
trabalho é `C:\Windows\System32` e a mensagem aparece em toda inicialização.

**Impacto**

Falso positivo em alertas baseados em nível de log (CloudWatch Logs, SIEM) e
confusão durante a operação.

**Correção proposta**

Ignorar `fs.ErrNotExist` e rebaixar os demais erros para `WARN`:

```go
if err := godotenv.Load(); err != nil && !errors.Is(err, fs.ErrNotExist) {
    slog.Warn("failed to load .env file", slog.Any("error", err))
}
```

Branch sugerida: `fix/dotenv-missing-log` (label `release:patch`).

**Correção (2026-10-08)** — branch `fix/homolog-windows-findings`, build de teste
`0.1.1-homolog`. Validado localmente: sem `.env` nada é logado; `.env`
malformado gera `WARN`. Falta validar na EC2.

---

## HML-002 — `agent_jobs` falha sem permissão no msdb (ponto de atenção)

- **Ambiente:** EC2 Windows, release v0.1.0
- **Encontrado em:** 2026-10-08
- **Severidade:** média
- **Status:** resolvido na EC2 (v0.1.1; repetir a validação no RDS)

**Observado**

```
sqlserver_scrape_collector_success{collector="agent_jobs"} 0
... The SELECT permission was denied on the object 'sysjobhistory', database 'msdb', schema 'dbo'.
```

Os demais collectors não são afetados.

**Causa**

O collector lê diretamente `msdb.dbo.sysjobs` e `msdb.dbo.sysjobhistory`, e o
login de monitoramento não tem acesso a essas tabelas.

**Opções de permissão**

1. **Grants somente nas duas tabelas** (menor privilégio, já documentado no README):

   ```sql
   USE msdb;
   CREATE USER [exporter] FOR LOGIN [exporter];
   GRANT SELECT ON dbo.sysjobs TO [exporter];
   GRANT SELECT ON dbo.sysjobhistory TO [exporter];
   ```

2. **Role `SQLAgentReaderRole`**:

   ```sql
   USE msdb;
   CREATE USER [exporter] FOR LOGIN [exporter];
   ALTER ROLE SQLAgentReaderRole ADD MEMBER [exporter];
   ```

   **Atenção:** as roles do SQL Agent dão acesso por meio de procedures e views
   (`sp_help_job`, `sysjobs_view`), não por `SELECT` nas tabelas base que o
   collector consulta. É provável que só a role **não** resolva. Validar na
   homologação antes de recomendar.

**Resultado da validação (2026-10-08, EC2 Windows)**

Os **grants somente em `dbo.sysjobs` e `dbo.sysjobhistory`** (opção 1)
resolveram: `sqlserver_scrape_collector_success{collector="agent_jobs"}` passou
para `1` e `sqlserver_agent_job_enabled` começou a ser emitida. As métricas
`sqlserver_agent_job_last_run_*` ainda não aparecem porque nenhum job
concluiu uma execução no ambiente.

A opção 2 (`SQLAgentReaderRole`) não foi testada. **Recomendação: opção 1**,
que é a de menor privilégio e é a que o README já documenta.

**Privilégio mínimo (2026-10-13)**

O login da homologação (`usr_exporter`) tem exatamente os grants do README:
`VIEW SERVER STATE` e `VIEW ANY DEFINITION` no servidor, e no msdb um usuário
com `SELECT` em `dbo.sysjobs` e `dbo.sysjobhistory`. Com ele, todos os 14
coletores SQL ficaram com `success 1` nas duas réplicas, inclusive com AG,
failover e suspensão. Privilégio mínimo validado; não é preciso outro login.

**Pontos a resolver**

- Comentário em `internal/collector/agent_jobs.go` cita `SQLAgentReaderRole`,
  enquanto o README documenta os grants nas tabelas. Atualizar o comentário
  para os grants validados.
- Executar um job do Agent até o fim para validar as métricas
  `sqlserver_agent_job_last_run_*`.
- Em ambientes sem SQL Agent (Express) ou sem interesse em jobs, desabilitar
  com `--collector.agent_jobs=false` ou `collectors.agent_jobs: false` no YAML.
- Repetir a validação no RDS for SQL Server, onde o acesso ao msdb tem
  restrições próprias.

---

## HML-003 — Coleta de disk I/O no Windows falha com `Invalid class`

- **Ambiente:** EC2 Windows, release v0.1.0, `collectors.os_collectors: true`
- **Encontrado em:** 2026-10-08
- **Severidade:** média
- **Status:** resolvido (v0.1.1)

**Observado** (a cada scrape)

```
ERROR failed to query disk performance counters error="Exception occurred. (Invalid class )"
```

Nenhuma métrica `os_disk_*` de I/O é emitida no Windows.

**Causa (confirmada no código)**

`internal/oscollector/disk_io_windows.go` monta a consulta com
`wmi.CreateQuery(&disks, "")`. Sem o terceiro argumento, a biblioteca
`yusufpapurcu/wmi` usa o **nome do tipo Go** como classe WMI. A consulta gerada
fica `SELECT ... FROM win32RawDiskPhysicalDisk`, que não existe. A classe
correta é `Win32_PerfRawData_PerfDisk_PhysicalDisk`.

**Correção proposta**

Passar a classe explicitamente:

```go
query := wmi.CreateQuery(&disks, "", "Win32_PerfRawData_PerfDisk_PhysicalDisk")
```

**Verificar junto com a correção**

- `PercentDiskTime` é do tipo `PERF_PRECISION_100NS_TIMER` (unidade de 100 ns).
  Dividir por `Frequency_PerfTime` só dá o valor certo quando a frequência é
  10 MHz. Conferir `os_disk_io_time_ms_total` contra o Performance Monitor.
- `AvgDiskReadQueueLength` e `AvgDiskWriteQueueLength` são acumuladores brutos
  na classe `PerfRawData`, mas estão sendo expostos como gauge. Os valores vão
  crescer sem parar.
- Os OS collectors não expõem métrica de sucesso por scrape (diferente de
  `sqlserver_scrape_collector_success`), então a falha só aparece no log.
- Não há teste para os collectors de `internal/oscollector`.

Branch sugerida: `fix/windows-disk-io-wmi-class` (label `release:patch`).

**Correção (2026-10-08)** — branch `fix/homolog-windows-findings`, build de teste
`0.1.1-homolog`:

- Classe WMI passada explicitamente (`Win32_PerfRawData_PerfDisk_PhysicalDisk`).
- `os_disk_io_time_ms_total` agora converte `PercentDiskTime` de ticks de
  100 ns para ms, sem depender de `Frequency_PerfTime`.
- **Pendente:** `os_disk_read_queue_length`/`os_disk_write_queue_length`
  continuam como gauge sobre acumuladores brutos. Corrigir exige mudar tipo e
  nome da métrica (e alinhar com o Linux); tratar em item próprio.

Validado na EC2 em 2026-10-08 com o build `0.1.1-homolog`: métricas `os_disk_*` de I/O aparecendo nos gráficos.

---

## HML-004 — Gaps intermitentes: `/metrics` retorna HTTP 500 por série duplicada em `volumes`

- **Ambiente:** EC2 Windows, build `0.1.1-homolog`, com carga no `StressDB`
- **Encontrado em:** 2026-10-08
- **Severidade:** alta (o scrape inteiro é perdido)
- **Status:** resolvido (v0.1.1)

**Observado**

Gráficos de gauges (PLE, buffer cache hit ratio, sessões, conexões) com gaps
intermitentes, enquanto gráficos de `rate()` (batch requests, compilations)
parecem contínuos: o `rate()` sobre a janela "pula" amostras ausentes e
esconde o problema. Nada aparece no log do exporter.

No Prometheus, os gaps coincidem com `up == 0` e
`scrape_samples_scraped == 0`; o target reporta
`server returned HTTP status 500 Internal Server Error`. Resposta do exporter:

```
An error has occurred while serving metrics:
2 error(s) occurred:
* collected metric "sqlserver_volume_total_bytes" { ... volume_mount_point="C:\\" ... } was collected before with the same name and label values
* collected metric "sqlserver_volume_available_bytes" { ... volume_mount_point="C:\\" ... } was collected before with the same name and label values
```

**Causa (duas falhas combinadas)**

1. **`volumes` emite séries duplicadas.** A query usa `SELECT DISTINCT` sobre
   `(mount_point, total_bytes, available_bytes)`. `sys.dm_os_volume_stats` é
   avaliada por arquivo de banco; sob escrita, arquivos do mesmo volume
   retornam `available_bytes` diferentes na mesma execução, o `DISTINCT` não
   colapsa as linhas e o mesmo `C:\` sai duas vezes.
2. **Uma métrica inválida derruba o scrape inteiro.** `promhttp.Handler()` usa
   `HTTPErrorOnError` e não tem `ErrorLog`: qualquer erro de coleta vira
   HTTP 500 para **todas** as métricas, sem nenhuma linha de log.

**Correção** — branch `fix/homolog-windows-findings`, build de teste `0.1.2-homolog`:

- `internal/collector/volumes.go`: `GROUP BY` por mount point com
  `MAX(total_bytes)`/`MIN(available_bytes)` (uma linha por volume, valor de
  espaço livre conservador).
- `main.go`: handler com `promhttp.ContinueOnError` e `ErrorLog` via `slog`.
  Uma série inválida é descartada e logada em `ERROR`; o restante é servido.

**Validação**

- Na EC2, com carga: `up` deve ficar em `1` contínuo e os gaps devem sumir.
- Query para acompanhar: `min_over_time(up{job="sqlserver_exporter_homolog"}[1h])`.
- **Validado em 2026-10-08** com `0.1.2-homolog` sob carga: 41 scrapes em
  10 min, `up` mínimo 1, nenhum HTTP 500, todos os coletores com `success 1`
  e séries de gauge contínuas (40/40 amostras de PLE). Antes: ~7% dos scrapes
  com HTTP 500.

**Pontos a considerar**

- Adicionar um painel `up` / `scrape_samples_scraped` ao dashboard de
  cobertura: o `sqlserver_up` não detecta esse tipo de falha, pois ele também
  some no 500.
- `backups` pode, em tese, gerar duplicata se dois backups do mesmo banco e
  tipo terminarem no mesmo `backup_finish_date`. Com o `ContinueOnError` isso
  não derruba mais o scrape, mas vale revisar a query.

---

## HML-005 — Disco C: da EC2 quase cheio durante carga no `StressDB` (ambiente)

- **Ambiente:** EC2 Windows (`AWSDBSERVER01`), volume C: de 59,5 GiB
- **Encontrado em:** 2026-10-08
- **Severidade:** alta (risco de falha de autogrowth no `StressDB`/`tempdb`)
- **Status:** resolvido (espaço liberado manualmente)
- **Tipo:** ambiente de homologação, não é defeito do exporter

**Observado**

Durante a validação do HML-004, `os_filesystem_free_bytes{mountpoint="C:"}` e
`sqlserver_volume_available_bytes{volume_mount_point="C:\\"}` caíram para
~9 MB livres (eram ~1,7 GiB por volta das 13h50). O `StressDB` estava com
7,7 GiB de dados e crescendo sob o SqlQueryStress.

**Ação**

Arquivos removidos da EC2 para liberar espaço.

**Pontos a considerar**

- Positivo para a homologação: o exporter detectou a situação pelos dois
  caminhos (`volumes` via SQL Server e `disk_usage` via OS collectors).
- Para próximas rodadas de carga, separar dados/log em um volume EBS dedicado
  ou limitar o crescimento do `StressDB` (`MAXSIZE`).
- Dashboard de exemplo: considerar uma regra de alerta de espaço livre
  (ex.: `sqlserver_volume_available_bytes / sqlserver_volume_total_bytes < 0.1`).

---

## HML-006 — Exporter não roda como serviço Windows nativo (ponto de atenção)

- **Ambiente:** EC2 Windows (instâncias 01 e 02)
- **Encontrado em:** 2026-10-08
- **Severidade:** média
- **Status:** resolvido (v0.2.0; `uninstall` validado em 2026-10-15)

**Observado**

Na homologação o exporter roda em primeiro plano num console do PowerShell.
As EC2 são paradas/iniciadas com frequência (economia de créditos), e a cada
start o exporter precisa ser iniciado manualmente; se o console for fechado o
processo morre e o target fica `down` sem nenhum aviso.

**Causa**

O binário não implementa o protocolo do Service Control Manager
(`golang.org/x/sys/windows/svc`). Registrar o `.exe` direto com `sc.exe create`
falha com erro 1053 após ~30 s; hoje a única opção é um wrapper (WinSW/NSSM),
com configuração manual (XML, recovery, logs).

**Correção** — branch `feat/windows-service` (sobre `fix/homolog-windows-findings`),
build de teste `0.2.0-homolog`:

- Subcomando `mssql-exporter service install|uninstall|start|stop|status`
  (`--config.file`, `--name`).
- `install` valida o config file, registra com início automático (atrasado)
  e recovery (reinicia em 5 s, 30 s, 60 s; também em saída com erro) e
  registra a origem no Event Log.
- Em execução como serviço: logs no Event Log (Application), diretório de
  trabalho = pasta do config file, parada limpa (`Stop`/`Shutdown`).
- Fora do serviço: Ctrl+C/SIGTERM agora também fazem shutdown limpo.

**Validação na EC2 (2026-10-12, release v0.2.0)**

- [x] `service install` / `start` / `status` nas instâncias 01 e 02: instalado
  de primeira, `Automatic (Delayed Start)`, rodando como `LocalSystem`.
- [x] Prometheus: os dois targets `up 1`, `sqlserver_up 1`, versão `0.2.0`,
  nenhum coletor com falha.
- [x] Reboot da EC2 → target volta sozinho (02, 2026-10-12; detalhes no HML-009).
- [x] `Stop-Process -Force` → SCM reinicia (5 s na 1ª falha, 30 s na 2ª; detalhes no HML-009).
- [x] `service stop` / `start`: usados nas atualizações para 0.2.1-homolog e
  0.3.0 nas duas réplicas, sem processo órfão. `uninstall` fica para o fim
  da homologação.
- [x] Event Log com as entradas do exporter (detalhes no HML-009, item 4).

Roteiro:

```powershell
.\mssql-exporter.exe service install --config.file C:\mssql-exporter\config.yml
.\mssql-exporter.exe service start
.\mssql-exporter.exe service status
Get-EventLog -LogName Application -Source mssql-exporter -Newest 10
```

- Reiniciar a EC2 e confirmar que o target volta sozinho no Prometheus.
- Matar o processo (`Stop-Process -Name mssql-exporter -Force`) e confirmar
  que o SCM reinicia em ~5 s.
- `service stop` / `service uninstall` sem deixar processo órfão.

**Validação do `uninstall` (2026-10-15, v0.9.0, EC2 01)**

- [x] `service uninstall` com o serviço rodando: parou e removeu
  (`Service "mssql-exporter" removed.`); `Get-Service` não encontra o serviço,
  nenhum processo `mssql-exporter`, porta 9187 livre e a origem do Event Log
  removida do registro.
- [x] `service install` + `start` de novo: `automatic delayed start, restarts
  on failure`, rodando. No Prometheus, `up 0` só entre 18:01:10 e 18:01:25;
  de volta às 18:01:40 com os 19 coletores em `success 1`.

**Pontos a considerar**

- Conta do serviço: `LocalSystem`. Avaliar conta virtual
  (`NT SERVICE\mssql-exporter`) para menor privilégio — exige ACL no config
  file e validar acesso aos contadores WMI dos OS collectors.
- Regra de firewall continua manual.

---

## HML-007 — Validação do coletor `availability_groups` em AG real (validação)

- **Ambiente:** AG `MSSQL_AG` entre `AWSDBSERVER01` e `AWSDBSERVER02`
  (SQL Server 2019, EC2 Windows), banco `AlwaysOnDB`, exporter v0.2.0 como
  serviço nas duas réplicas
- **Iniciado em:** 2026-10-12
- **Tipo:** validação (primeira execução contra um Always On real)
- **Status:** resolvido (v0.3.0)

**Resultado inicial (AG recém-criado, sem carga)**

| Métrica | Visto pela 01 (primária) | Visto pela 02 (secundária) |
|---|---|---|
| `sqlserver_is_hadr_enabled` | 1 | 1 |
| `sqlserver_hadr_replica_role` | 01 = 1 (primary), 02 = 2 | 02 = 2 |
| `sqlserver_hadr_replica_synchronization_health` | 2 / 2 | 2 |
| `sqlserver_hadr_database_synchronization_state` | 2 / 2 (synchronized) | 2 |
| `sqlserver_hadr_database_log_send_queue_kb` | 0 (só série da 02) | 0 |
| `sqlserver_hadr_database_redo_queue_kb` | 0 (só série da 02) | 0 |

`sqlserver_scrape_collector_success{collector="availability_groups"} 1` nas duas.

**Comportamentos esperados (não são defeitos)**

- A secundária enxerga apenas a própria réplica nas DMVs `sys.dm_hadr_*`;
  a primária enxerga todas. A mesma série (`replica_server="AWSDBSERVER02"`)
  vem dos dois exporters, diferenciada por `sqlserver_instance`. Em painéis e
  alertas: `max by (ag_name, replica_server, db_name)`.
- Filas só existem para réplicas secundárias; os valores nulos da primária
  são descartados pelo coletor.

**Roteiro restante**

- [x] Carga de escrita no `AlwaysOnDB` (SqlQueryStress, `INSERT dbo.EventLog`
  com `REPLICATE('x', 200)`, 32 threads × 10k iterações, sem delay).
  19:44–19:46, pico de ~13,3k batch requests/s na primária:
  - `log_send_queue_kb` estável em ~60 KB; `redo_queue_kb` entre 296 e 432 KB.
  - `synchronization_state` 2 e `synchronization_health` 2 o tempo todo.
  - Coerente com commit síncrono: cada commit espera o hardening na
    secundária, então a fila de envio não acumula.
  - Pico anterior às 19:37, **sem carga** (~2 batch/s): `log_send_queue_kb`
    até ~349 MB e `redo_queue_kb` até ~33 MB, zerando em ~1 min. Provável
    sincronização inicial do AG logo após a configuração. Confirmar.
  - Valores lidos direto da DMV (já em KB), sem conversão no coletor.
- [x] Carga em `ASYNCHRONOUS_COMMIT` (mesmo script). Às 20:32:13
  `synchronization_state` passou de 2 (synchronized) para **1
  (synchronizing)**, como esperado em modo assíncrono; `health` seguiu 2.
  Throughput subiu para ~31,7k batch/s (o dobro do síncrono: o commit não
  espera mais a secundária). Filas continuaram pequenas (send ≤ 60 KB,
  redo ≤ 548 KB): a rede e a secundária acompanharam a carga.
- [x] Failover para a 02 (~21:04–21:05). Sequência vista no Prometheus
  (amostras a cada 15 s, nenhum scrape perdido, `up`/`sqlserver_up` = 1 o
  tempo todo):
  - 21:04:00 — 02 volta para `SYNCHRONOUS_COMMIT`: health da 02 = 1
    (partially healthy) por uma amostra; 21:04:15 `synchronization_state` = 2.
  - 21:05:15 — **papéis invertidos nos dois exporters**: 02 = 1 (primary),
    01 = 2 (secondary). Refletido em uma única amostra (≤ 15 s).
  - 21:05:15 — réplica 01 (antiga primária) com `is_suspended` = 1,
    **`suspend_reason` = 1 (SUSPEND_FROM_PARTNER)**, `synchronization_state`
    = 0 e health 0, vista pela 01 e pela 02. Segundo a documentação, o motivo 1
    é a suspensão da réplica após um failover forçado — confirmado: o failover
    foi feito com o script do assistente do SSMS (modo SQLCMD), que roda
    `SET (ROLE = SECONDARY)` na 01 e `FORCE_FAILOVER_ALLOW_DATA_LOSS` na 02.
    O RESUME manual da 01 é o passo esperado depois de um failover forçado.
  - 21:07:00 — RESUME manual na 01: `is_suspended` 0, sync 2, health 2.
    Estado final: 02 primária, 01 secundária saudável.
  - A série `role` da 01 vista pela 01 passou de 1 para 2 sem gap; as séries
    que só existem na primária (ex.: filas da réplica remota) mudaram de
    exporter (antes vinham da 01, agora da 02), como esperado.
- [ ] Secundária indisponível (parar o SQL Server da 02): primária deve
  reportar `synchronization_health` < 2 e `synchronization_state` != 2.

---

## HML-008 — Fila de envio some do gráfico com a réplica suspensa; faltam `is_suspended` e `secondary_lag_seconds`

- **Ambiente:** AG `MSSQL_AG` (01 primária, 02 secundária em `ASYNCHRONOUS_COMMIT`), v0.2.0
- **Encontrado em:** 2026-10-12, durante o HML-007
- **Severidade:** média (o problema mais grave do AG aparece como "sem dados")
- **Status:** resolvido (v0.3.0)

**Observado**

Com `ALTER DATABASE AlwaysOnDB SET HADR SUSPEND` na 02 e carga na primária
(20:35–20:40), o painel `sqlserver_hadr_database_log_send_queue_kb` ficou
vazio. Visto pela primária:

| Horário | Carga | send queue | redo queue | state | health | log usado |
|---|---|---|---|---|---|---|
| 20:35–20:40 (suspenso) | até 33k batch/s | **sem série** | 0 | 0 | 0 | 27% → 98% |
| 20:41 (RESUME) | — | 2,5 GB | 11 MB | 1 | 2 | 98% |
| 20:42 | — | 1,8 GB | 72 MB | 1 | 2 | 98% |

**Causa**

Com o data movement suspenso o SQL Server retorna `NULL` em
`log_send_queue_size`; o coletor omite `NULL` de propósito (zero faria a
réplica parada parecer saudável). `synchronization_state`/`health` = 0 já
sinalizavam o problema, mas não havia métrica que dissesse *suspenso* nem o
atraso em tempo.

**Correção** — branch `feat/ag-suspend-and-lag`:

- `sqlserver_hadr_database_is_suspended` (0/1), presente mesmo com as filas `NULL`.
- `sqlserver_hadr_database_suspend_reason` (só enquanto suspenso; 0 = usuário,
  1 = parceiro, 2 = redo, ...).
- `sqlserver_hadr_database_secondary_lag_seconds` (SQL Server 2016+).
- Teste unitário reproduzindo o formato observado; README e dashboard de
  cobertura atualizados (105 métricas).

**Validação na EC2 (2026-10-12, build `0.2.1-homolog` nas duas réplicas)**

SUSPEND executado na primária (01) às ~20:55, seguido de carga com transações
grandes (50k linhas por transação, 4 threads):

| Série (vista pela 01) | Valor |
|---|---|
| `is_suspended{replica_server="AWSDBSERVER01"}` | **1** |
| `suspend_reason{replica_server="AWSDBSERVER01"}` | **0** (usuário) |
| `is_suspended{replica_server="AWSDBSERVER02"}` | 0 |
| `synchronization_state` (01 e 02) | 0 (not synchronizing) |
| `log_send_queue_kb` | ausente (`NULL`), como antes |
| `secondary_lag_seconds{replica_server="AWSDBSERVER02"}` | **0** |
| log usado do `AlwaysOnDB` na 01 | 8% → 46% |

- [x] `is_suspended` / `suspend_reason` funcionam. Com SUSPEND na primária, o
  SQL Server marca como suspensa a réplica **da primária** (a secundária fica
  `is_suspended` 0, mas `synchronization_state` 0). Alertas devem usar
  `max by (ag_name, db_name) (sqlserver_hadr_database_is_suspended) == 1`,
  sem filtrar por réplica.
- [x] **`secondary_lag_seconds` não serve para detectar suspensão:** ficou em
  0 durante toda a suspensão, mesmo com o log crescendo. Usar junto com
  `is_suspended` e `synchronization_state`, nunca sozinho. Documentar no README.
- [x] RESUME: `is_suspended` voltou a 0 nas três séries, `suspend_reason`
  sumiu, `log_send_queue_kb` reapareceu (0) e o log usado voltou a 0%.
- [x] `secondary_lag_seconds` > 0 com data movement ativo: 21:01:15–21:01:45,
  02 em assíncrono alcançando a primária após a carga com transações grandes
  — lag 71 → 86 → 101 s, `log_send_queue_kb` 1,7 GB → 0, `redo_queue_kb`
  até ~255 MB; zerou às 21:02. A documentação confirma que o lag fica em 0
  enquanto a réplica está suspensa.
- [x] `suspend_reason` = 1 (partner) observado após o failover (ver HML-007).
- Correção: a descrição da métrica `suspend_reason` tinha os códigos 3–8
  deslocados; corrigida no coletor, README e dashboard conforme a
  documentação de `sys.dm_hadr_database_replica_states`.

**Observação do teste**

Durante a suspensão o log do `AlwaysOnDB` foi de 27% para 98%: o log não
trunca enquanto a secundária não recebe. Em produção, réplica suspensa =
risco de log cheio (erro 9002) na primária. Bom candidato a alerta, junto
com `is_suspended`.

---

## HML-009 — Cenários de falha: banco offline, SQL parado, reboot, kill, backup diferencial (validação)

- **Ambiente:** AG `MSSQL_AG` (01 primária, 02 secundária), exporter v0.3.0
  como serviço nas duas réplicas, login `usr_exporter` com privilégio mínimo
- **Iniciado em:** 2026-10-13
- **Tipo:** validação
- **Status:** resolvido (todos os 5 cenários validados)

### 1. Banco `dbAdmin` OFFLINE na 02 — validado

- 21:30:47: `sqlserver_database_state{db_name="dbAdmin"}` passou de
  `state="ONLINE"` para `state="OFFLINE"` em um scrape (a série ONLINE para
  de ser emitida e nasce a OFFLINE; alertas: `sqlserver_database_state{state!="ONLINE"}`).
- Nenhum coletor falhou; `up` 1, sem séries duplicadas.
- Continuam emitidos: `data_size_bytes`/`log_size_bytes` (vêm de
  `sys.master_files`) e `recovery_model`.
- Deixam de ser emitidos (esperado): `log_used_percent` e as séries
  `sqlserver_io_file_*` do `dbAdmin` — as DMVs não retornam bancos offline.

### 2. SQL Server parado na secundária (02) — validado

Serviço do SQL Server da 02 parado às ~21:37:45:

| Série | Antes | Depois |
|---|---|---|
| `up{instance=<02>}` | 1 | **1** (exporter vivo) |
| `sqlserver_up{sqlserver_instance="AWSDBSERVER02"}` | 1 | **0** a partir de 21:37:51 |
| `scrape_samples_scraped` (02) | ~3.667 | ~108 (só OS collectors + self-metrics) |
| `scrape_duration_seconds` (02) | ~0,2 s | 0,16–2,15 s |
| Primária: `replica_synchronization_health{replica_server="AWSDBSERVER02"}` | 2 | **0** a partir de 21:38:06 |
| Primária: `database_synchronization_state` (02) | 2 | **0** |
| Primária: `database_is_suspended` (02) | 0 | **0** |
| Primária: `log_send_queue_kb` (02) | 0 | ausente (`NULL`) |

- O exporter não cai quando o SQL Server cai: `up` 1 + `sqlserver_up` 0
  separa "SQL fora" de "exporter fora". Scrapers são pulados; nenhum erro.
- A primária detectou a réplica fora em ≤ 15 s.
- `is_suspended` 0 com `synchronization_state` 0 distingue **réplica
  desconectada** de **réplica suspensa** (`is_suspended` 1). Útil para alertas.
- Retorno (SQL Server iniciado, exporter **não** reiniciado): às 21:41:13,
  em um único scrape, `sqlserver_up` voltou a 1, ~3.600 amostras por scrape,
  nenhum coletor falhando, e a primária voltou a reportar a 02 com
  health/sync 2 e filas 0. Reconexão automática pelo pool, label
  `sqlserver_instance="AWSDBSERVER02"` preservado (foi resolvido no start do
  exporter, quando o SQL estava no ar).

### 3. Reboot da EC2 02 — validado (também fecha item do HML-006)

| Evento | Horário |
|---|---|
| Último scrape OK antes do reboot | 21:45:49 |
| `os_boot_time_seconds` (Windows de volta) | 21:45:57 |
| Primária vê a 02 com health/sync 0 | 21:46:21 (uma amostra) |
| Primária vê a 02 saudável de novo (SQL Server já no ar) | 21:46:36 |
| `process_start_time_seconds` do exporter | **21:48:07** (boot + ~2 min 10 s) |
| Primeiro scrape OK, `sqlserver_up` 1, nenhum coletor falhando | 21:48:19 |

- O serviço subiu sozinho, sem console.
- **`sqlserver_instance="AWSDBSERVER02"` preenchido:** o início atrasado
  (`Automatic (Delayed Start)`, ~2 min após o boot) fez o exporter subir
  depois do SQL Server, então o label foi resolvido corretamente.
- Custo: ~2 min 30 s sem métricas da máquina a cada boot, quase todo por
  causa do atraso. Melhoria possível: resolver `sqlserver_instance` no
  primeiro ping bem-sucedido (e não só no start), o que permitiria início
  automático sem atraso e eliminaria a limitação conhecida do label vazio.

### 4. Kill do processo — validado (também fecha item do HML-006)

`Stop-Process -Name mssql-exporter -Force` na 02. `process_start_time_seconds`
registrou dois reinícios automáticos pelo SCM, sem intervenção:

| Novo processo | Coerente com |
|---|---|
| 21:50:20 | 1º kill → recovery action "restart após 5 s" |
| 21:51:09 | 2º kill (feito de propósito em seguida) → "restart após 30 s" (scrape das 21:50:56 com `up` 0) |

- Logo após o kill, `service status` mostra `stopped` / PID 0: esperado,
  é a janela de espera da recovery action.
- A escalada 5 s → 30 s (→ 60 s na 3ª) segue o `install`; o contador de
  falhas zera após 24 h sem falhas.
- Depois do reinício: `sqlserver_up` 1, label preenchido, sem intervenção.
- Event Log (`Application`, origem `mssql-exporter`): sequência completa de
  inicialização às 00:50:20 e 00:51:09 UTC (`starting mssql-exporter` com a
  versão 0.3.0, `successfully connected and pinged sql server`,
  `sql server instance instance=AWSDBSERVER02`, `os collectors enabled`,
  `Listening on`), todas como *Information*. Não há entrada do kill em si
  (esperado: `Stop-Process -Force` encerra sem chance de logar).

### 5. Backup diferencial — validado

Banco `diff_test` (recovery SIMPLE) na 02, backup full e diferencial com
`COMPRESSION` em `D:\BACKUP`:

| Métrica | full | differential |
|---|---|---|
| `sqlserver_backup_last_completed_timestamp_seconds` | 21:58:07 | 21:58:11 |
| `sqlserver_backup_last_size_bytes` | 433.994 (~424 KB) | 59.482 (~58 KB) |
| `sqlserver_backup_last_duration_seconds` | 0 | 0 |

- `backup_type="differential"` aparece corretamente.
- Tamanho é o comprimido (`compressed_backup_size`), coerente com `COMPRESSION`.
- Duração 0: `backupset` grava as datas com precisão de segundos e o banco
  vazio faz backup em menos de 1 s. Esperado.
- Horários corretos **porque o Windows das EC2 está em UTC** — ver HML-010.

---

## HML-010 — Timestamps de backup e Agent jobs tratam o horário local do SQL Server como UTC

- **Ambiente:** qualquer SQL Server cujo fuso horário não seja UTC (não
  reproduz nas EC2 da homologação, que estão em UTC)
- **Encontrado em:** 2026-10-12, revisando o código durante o HML-009 (item 5)
- **Severidade:** média (alertas de "backup atrasado" e "job não rodou" erram
  pelo tamanho do fuso)
- **Status:** resolvido (v0.3.1)

**Causa**

- `backups`: `msdb.dbo.backupset.backup_finish_date` é `datetime` no horário
  **local** do servidor, sem offset. O driver devolve um `time.Time` sem fuso
  (tratado como UTC) e o coletor emite `finishDate.Unix()`.
- `agent_jobs`: `run_date`/`run_time` de `sysjobhistory` também são horário
  local; `DATETIMEFROMPARTS` + `.Unix()` tem o mesmo problema.

**Impacto (exemplo)**

Servidor em `UTC-03:00`: backup às 21:00 local (00:00 UTC) vira 21:00 UTC no
timestamp, ou seja, 3 h no passado. `time() - sqlserver_backup_last_completed_timestamp_seconds`
mostra o backup 3 h mais velho do que é. Em fusos positivos (`UTC+X`) o
timestamp fica no futuro. Durações não são afetadas (diferença entre duas
datas locais).

**Correção proposta**

- `backups`: converter na query com a coluna `backupset.time_zone` (offset do
  servidor em intervalos de 15 min **no momento do backup**, correto mesmo com
  horário de verão): `DATEADD(MINUTE, -15 * bs.time_zone, bs.backup_finish_date)`;
  `time_zone = 127` significa desconhecido → cair no offset atual.
- `agent_jobs`: converter com o offset atual do servidor
  (`DATEDIFF(MINUTE, GETDATE(), GETUTCDATE())`).
- Testes unitários com fuso não-UTC.

**Correção aplicada**

- `backups`: `CROSS APPLY` calcula o offset por linha
  (`bs.time_zone * 15` quando entre -48 e 48; senão
  `DATEPART(TZOFFSET, SYSDATETIMEOFFSET())`) e as datas de início/fim são
  convertidas com `DATEADD(MINUTE, -offset, ...)`. Duração inalterada.
- `agent_jobs`: `DATEADD(MINUTE, -DATEPART(TZOFFSET, SYSDATETIMEOFFSET()), ...)`.
- As regex dos testes fixam a conversão; teste de mutação: removendo-a,
  10 testes falham.

**Validação 1 (2026-10-12, `0.3.1-homolog`, EC2 01 e 02 em UTC-3 via `tzutil`)**

- [x] Conversão correta para backups gravados antes da mudança
  (`time_zone` = 0): `diff_test` full/diff de 21:58 −03 → `00:58Z`.
- [ ] **Novo problema:** os backups feitos *depois* da mudança não apareceram,
  e o `JB_BACKUP_FULL` da 01 foi para `04:00Z` (3 h no futuro). Causa: o
  "último" registro era escolhido pelo horário **local**
  (`MAX(backup_finish_date)` / `ORDER BY run_date, run_time`). Depois de voltar
  o fuso 3 h, os registros antigos (01:00 em UTC) parecem mais novos que os
  novos (22:05 em −03) e o coletor fica preso neles por até 3 h. Afeta também
  a virada do horário de verão (fall-back).

**Correção 2 (`0.3.2-homolog`)**

- `backups`: `ROW_NUMBER() OVER (PARTITION BY database_name, type ORDER BY backup_set_id DESC)`.
- `agent_jobs`: `ORDER BY h.instance_id DESC`.
- Ambas são identities do msdb: só crescem, independentemente do fuso.
- Limitação que permanece: o histórico de jobs não guarda o offset, então
  uma execução anterior a uma mudança de fuso é convertida com o offset
  atual. Some assim que o job roda de novo.

**Validação 2 (2026-10-12 22:12, `0.3.2-homolog`, EC2 em UTC-3, Agent rodando)**

| Série | Horário (−03) | Idade | Real? |
|---|---|---|---|
| 02 `diff_test` full | 22:07:30 | 5,1 min | ✅ |
| 02 `diff_test` differential | 22:08:00 | 4,6 min | ✅ |
| 01 `AlwaysOnDB` full | 22:08:37 | 4,0 min | ✅ |
| 01 `AlwaysOnDB` log | 22:08:35 | 4,0 min | ✅ |
| 01 `JB_BACKUP_FULL` (status 1) | 22:07:44 | 4,8 min | ✅ |
| 01 `JB_BACKUP_LOG` (status 1) | 22:08:32 | 4,0 min | ✅ |

- [x] Backups e jobs feitos depois da mudança de fuso aparecem com o horário
  real; nenhum timestamp no futuro; nenhum coletor falhando.
- [x] Backup antigo (02 `AlwaysOnDB` full, 19:33 −03, gravado em UTC) segue
  correto: a conversão por linha usa o `time_zone` de cada backup.

**Como reproduzir na homologação**

Mudar o fuso da 02 (`tzutil /s "E. South America Standard Time"`), reiniciar o
SQL Server, fazer um backup e comparar
`sqlserver_backup_last_completed_timestamp_seconds` com o horário real.

---

## HML-011 — Sem cobertura de database mirroring e log shipping

- **Ambiente:** EC2 01 + 02 (SQL Server 2019), fora do AG `MSSQL_AG`
- **Encontrado em:** 2026-10-13, revisão de cobertura de HA/DR
- **Severidade:** média (ambientes legados com essas features ficam sem monitoramento de HA/DR)
- **Status:** resolvido (v0.4.0)

**Observado**

O exporter só cobre Always On. Ambientes que ainda usam database mirroring
(deprecated desde o 2012, mas presente nas versões atuais) ou log shipping
não têm métrica de estado, fila ou atraso. Já aparecia indiretamente o
último backup de log (`backups`) e o resultado dos jobs
LSBackup/LSCopy/LSRestore (`agent_jobs`), mas não o atraso do secundário.

**Correção proposta**

- Novo coletor `database_mirroring` (ligado por padrão; sem grant novo):
  - `sys.database_mirroring` (`mirroring_guid IS NOT NULL`): role, state,
    safety level, witness state e o `partner_instance` numa métrica `_info`.
  - Contadores `Database Mirroring` de `sys.dm_os_performance_counters`:
    filas de envio/redo, bytes de log enviados/recebidos/refeitos,
    Transaction Delay e Mirrored Write Transactions.
- Novo coletor `log_shipping` (**desligado por padrão**,
  `--collector.log_shipping`): `msdb.dbo.log_shipping_monitor_primary` e
  `_secondary`. As colunas `*_date_utc` já são UTC (sem o problema do
  HML-010). Precisa de `GRANT SELECT` nas duas tabelas.
- Suporte a coletor desligado por padrão (`ScraperEnabledByDefault`), com o
  default refletido na flag e no YAML.

**Plano de validação — mirroring**

1. Banco novo (ex.: `MirrorDB`, FULL recovery) espelhado entre 01 e 02 usando
   o endpoint `DATABASE_MIRRORING` que o AG já usa.
2. Conferir `role`, `state` = 4 e `safety_level` nas duas instâncias, e as
   filas com carga no principal.
3. Falhas: `SET PARTNER SUSPEND` / `RESUME`, `SAFETY OFF` (async), parar o SQL
   do espelho (state 1), failover manual (`SET PARTNER FAILOVER`, troca de role).
4. Confirmar que o `AlwaysOnDB` (do AG) **não** aparece no coletor.
5. Conferir se `transaction_delay_ms_total` e `write_transactions_total` são
   monotônicos (tipo cumulativo no DMV).

**Validação 1 (2026-10-13, `0.4.0-homolog`, `MirrorDB` 01 → 02, SAFETY FULL, sem witness)**

| Métrica | 01 (principal) | 02 (espelho) |
|---|---|---|
| role / state / safety | 1 / 4 / 2 | **sem série** |
| `mirroring_info` | partner = AWSDBSERVER02 | **sem série** |
| log sent / received (bytes) | 69632 / 0 | 0 / 69632 |
| redo bytes | 0 | 81920 |
| filas send / redo (KB) | 0 / 0 | 0 / 0 |

- [x] `database_mirroring` com `success 1` nas duas; `AlwaysOnDB` (AG) fora
  do coletor.
- [x] Contadores coerentes com o papel de cada lado.
- [ ] **Witness:** sem witness o principal reportou `witness_state 0`
  (unknown) — a documentação diz NULL. Na prática vem 0 com
  `mirroring_witness_name` vazio. **Correção (`0.4.1-homolog`):** só emitir
  quando `mirroring_witness_name <> ''`. O caso com witness não é validado
  no lab (precisaria de uma terceira instância).
- [x] **Espelho sem estado (limitação, não será corrigida):** na 02 o
  `usr_exporter` vê a linha de `MirrorDB` com todas as colunas `mirroring_*`
  NULL. A documentação confirma: valores no espelho só para **sysadmin**.
  Testado `GRANT SELECT ON sys.database_mirroring` com usuário no master:
  continua NULL. Decisão: não exigir sysadmin; o estado é da sessão e o
  principal o reporta (inclusive "disconnected"), e depois de um failover o
  novo principal assume. Documentado no README.

**Validação 2 — carga no `MirrorDB` (2026-10-13 11:50–11:54, `0.4.1-homolog`)**

| Horário | send queue 01 / 02 | redo queue 01 / 02 | write tx (acum.) | delay ms (acum.) | log usado 01 | state |
|---|---|---|---|---|---|---|
| 11:49:47 | 0 / 0 | 0 / 0 | 1 | 1 | 11% | 4 |
| 11:51:47 | 700 MB / 652 MB | 0 / 0 | 22 | 98 424 | 100% | 4 |
| 11:52:17 | 96 MB / 352 MB | 286 MB / 159 MB | 23 | 129 851 | 96% | 4 |
| 11:52:32 | 0 / 0 | 0 / 0 | 28 | 163 782 | 96% | 4 |
| 11:53:32 | 0 / 0 | 0 / 0 | 160 029 | 440 625 | 99% | 4 |

- [x] Filas, throughput (pico de ~23 MB/s enviado pela 01 e ~19 MB/s
  recebido/refeito pela 02) e transações acompanharam a carga.
- [x] `transaction_delay_ms_total` e `write_transactions_total` monotônicos
  (`resets()` = 0): `CounterValue` correto.
- [x] Delay médio coerente com o tipo de carga: ~37 s por transação na fase
  de poucas transações grandes (11:50–11:52, 700 MB de log em ~7 commits) e
  ~1,7 ms na fase de muitas transações pequenas (11:52:47 em diante).
- [x] **Ajuste de documentação:** as duas réplicas reportam as duas filas
  (cada uma pela sua visão), e não só o principal (send) / o espelho (redo)
  como estava no help e no README. Corrigido no help, README e dashboard.
- Observação de ambiente: o log do `MirrorDB` na 01 ficou em 96–100% (FULL
  recovery sem backup de log).

**Validação 3 — SUSPEND com carga (11:58) e RESUME (12:00)**

| Horário | state (01) | send queue 01 / 02 | write tx (acum.) | log usado 01 |
|---|---|---|---|---|
| 11:58:08 | 4 | 0 / 0 | 480 035 | 2% |
| 11:58:23 | **0** (suspended) | 0 / 0 | 480 035 | 2% |
| 11:59:23 | 0 | 1,4 GB / 1,7 GB | 480 035 | 54% |
| 12:00:38 | 0 | 3,3 GB / 3,2 GB | 480 035 | 99% |
| 12:00:53 | **2** (synchronizing) | 3,1 GB / 2,9 GB | 480 035 | 99% |

- [x] SUSPEND visível em 15 s (state 0) e a fila de envio crescendo nas duas
  visões. Diferente do AG (HML-008), a fila **não** some enquanto suspenso.
- [x] `write_transactions_total` e `transaction_delay_ms_total` param de
  crescer enquanto suspenso (nenhuma transação espera o espelho), então o
  painel de delay médio fica vazio nesse período — esperado.
- [x] RESUME: state 2 e a fila drenando (~9 MB/s de redo na 02).
- O log do `MirrorDB` na 01 cresceu para 3,3 GB durante o suspend (não
  trunca até o espelho receber).

- [x] Ressincronizado às 12:02:31 (state 4): ~3,3 GB drenados em ~2 min,
  com pico de ~43 MB/s de redo na 02.

**Validação 4 — SQL Server da 02 parado (12:03)**

| Horário | state (01) | `up` 02 | `sqlserver_up` 02 | `database_mirroring` 01 |
|---|---|---|---|---|
| 12:03:01 | 4 | 1 | 1 | success 1 |
| 12:03:16 | **1** (disconnected) | 1 | 1 | success 1 |
| 12:03:31 | 1 | 1 | **0** | success 1 |

- [x] O principal reporta a desconexão em até 15 s; o exporter da 02 segue
  no ar com `sqlserver_up 0` e sem as séries de mirroring.

**Validação 5 — SQL da 02 de volta (12:04) e failover manual (12:05)**

| Horário | `sqlserver_up` 02 | role / state 01 | role / state 02 | `mirroring_info` |
|---|---|---|---|---|
| 12:04:24 | 0 | 1 / **4** | — | 01 → partner 02 |
| 12:04:39 | **1** | 1 / 4 | — | 01 → partner 02 |
| 12:05:24 | 1 | 1 / 4 (último scrape antes) | **1 / 4**, safety 2 | — |
| 12:05:39 | 1 | — (vira espelho) | 1 / 4 | 02 → partner 01 |

- [x] Reconexão: state volta a 4 sem passar visivelmente por 2 (pouco log
  pendente).
- [x] `SET PARTNER FAILOVER`: a 02 passa a reportar role, state, safety e
  `_info` (partner AWSDBSERVER01); a 01 fica só com os contadores, e
  `sqlserver_database_state` mostra 02 ONLINE / 01 RESTORING.
- Por um scrape (12:05:24) as duas aparecem como principal: é só a defasagem
  entre os scrapes das duas instâncias (a 01 foi lida antes do failover).
  Em alertas de "dois principais", usar `for:` de pelo menos 1 min.

**Resultado:** coletor `database_mirroring` validado (carga, suspend/resume,
parceiro fora do ar e failover). Pendente só o caso com witness (sem
terceira instância no lab).

**Validação 6 — log shipping `LogShipDB` 01 → 02 (2026-10-14, `0.4.2-homolog`)**

- [x] Grants nas tabelas `log_shipping_monitor_*` e `log_shipping: true`:
  `success 1` nas duas.
- [x] Thresholds corretos: backup 3600 s (01), restore 2700 s (02).
- [ ] **Timestamps congelados no horário da configuração (15:43:16)** apesar
  do `LSBackup` gerar `.trn` a cada 3 min e de `LSCopy`/`LSRestore` rodarem.
  O próprio `sp_help_log_shipping_monitor` e até `log_shipping_primary_databases`
  (`last_backup_file`/`date` NULL) concordam: o exporter reflete o SQL Server.
- **Causa (ambiente):** bug do SQL Server 2019 RTM/CU1
  ([KB4537869](https://support.microsoft.com/help/4537869)) — o
  `sqllogship` não grava histórico/monitor: "Could not log history/error
  message ... Failed to convert parameter value from a SqlGuid to a String".
  As EC2 estão em 15.0.2000.5 (RTM). Corrigido no CU2. Documentado no README.

**Validação 7 — log shipping depois do CU32 (2026-10-14 16:22–16:31, `0.4.2-homolog`)**

- [x] Timestamps andando após o patch: `primary_last_backup` 16:22:00,
  `secondary_last_copied` / `last_restored` 16:18:07, latência 240 s.
- [x] **`LSRestore` desabilitado + threshold de restore em 2 min:** às 16:30,
  `last_restored` parado em 16:27:01 enquanto `last_copied` seguiu (16:30:01);
  `restore_threshold_seconds` = 120 refletiu a mudança; a expressão de alerta
  `time() - last_restored > restore_threshold` disparou (239 s).
- [x] **Grant removido em `log_shipping_monitor_primary` na 01:**
  `success{collector="log_shipping"} 0` só nesse coletor; as séries do
  primário somem e os demais coletores seguem com `success 1`.

**Plano de validação — log shipping**

1. Grants nas tabelas `log_shipping_monitor_*` para o `usr_exporter` nas duas
   instâncias e `log_shipping: true` no `config.yml`.
2. Log shipping de um terceiro banco (ex.: `LSDB`) da 01 para a 02.
3. Conferir os timestamps (UTC) contra os jobs e a latência de restore.
4. Falhas: desabilitar o job de restore e ver `time() - last_restored` passar
   do threshold; remover o grant e ver `success 0` só no `log_shipping`.

---

## HML-012 — `backups` reporta no secundário os backups restaurados do primário

- **Ambiente:** EC2 02 (secundário do log shipping `LogShipDB` e espelho do `MirrorDB`), `0.4.1-homolog`
- **Encontrado em:** 2026-10-14, durante o HML-011
- **Severidade:** média (o secundário parece ter backup recente sem nunca ter feito um)
- **Status:** resolvido (v0.4.0)

**Observado**

A 02 reportava `sqlserver_backup_last_completed_timestamp_seconds{backup_type="full"}`
para `LogShipDB` (2026-10-14 15:43:15) e `MirrorDB` (2026-10-13 11:20:40),
exatamente os mesmos segundos dos backups feitos na 01. Consulta no
`msdb.dbo.backupset` da 02:

| server_name | database_name | type | backup_finish_date |
|---|---|---|---|
| AWSDBSERVER01 | MirrorDB | D | 2026-10-13 11:20:40 |
| AWSDBSERVER01 | LogShipDB | D | 2026-10-14 15:43:15 |

**Causa**

O `RESTORE` grava o backup restaurado no `backupset` local, com o
`server_name` de origem. Com o log shipping rodando, cada restore de log
apareceria na 02 como um backup de log recente, escondendo que o secundário
não tem backup próprio (e mascarando um alerta de "sem backup" após um
failover/role change).

**Correção**

- Filtro `AND server_name = @@SERVERNAME` na subquery do `backupset`: só
  conta backup feito pela própria instância. Sem grant novo.
- Regex do teste fixa o filtro; teste de mutação: removendo-o, 5 testes falham.
- README atualizado.

**Validação (2026-10-14 15:58, `0.4.2-homolog`)**

- [x] Na 02, as séries de `backups` de `LogShipDB` e `MirrorDB` sumiram (não
  há backup local).
- [x] Na 01, nada mudou: `LogShipDB` full 15:43:15 / log 15:58:00 e
  `MirrorDB` full / log continuam.

---

## HML-013 — Scrape falha inteiro (`up 0`) quando o SQL demora: `scrape_timeout` do exemplo = timeout do exporter

- **Ambiente:** Prometheus local (`prometheus.yml` do repositório), EC2 02, durante o HML-014
- **Encontrado em:** 2026-10-14
- **Severidade:** baixa (só em instantes de lentidão do SQL; perde o scrape inteiro em vez de parte)
- **Status:** resolvido (v0.4.0)

**Observado**

Na 02, dois scrapes falharam (16:12:30 e 16:15:15) com `up 0`, 0 amostras e
duração de exatamente 5,0 s, enquanto o exporter seguia rodando (mesmo
`process_start_time_seconds`). O primeiro coincide com o SQL voltando do CU.

**Causa**

O `prometheus.yml` de exemplo tem `global.scrape_interval: 5s`; sem
`scrape_timeout` explícito, o Prometheus usa 5 s também para os jobs do
exporter (confirmado em `/api/v1/status/config`). O exporter usa 5 s
(`ConnTimeout`) como prazo do ping + coletores: se o SQL demora, a resposta
chega depois de 5 s e o Prometheus descarta tudo.

**Correção**

- `prometheus.yml` de exemplo: `scrape_timeout: 10s` no job `sqlserver_exporter`
  (aplicado também no job local de homologação).
- README: orientação de manter o `scrape_timeout` acima de 5 s.
- Possível melhoria futura: o exporter limitar o próprio prazo pelo header
  `X-Prometheus-Scrape-Timeout-Seconds`, devolvendo dados parciais em vez de
  estourar o timeout do Prometheus.

---

## HML-014 — Atualização do SQL Server 2019 RTM → CU32 + KB5128494 (validação)

- **Ambiente:** EC2 01 + 02, `0.4.2-homolog`; patch em paralelo nas duas, sem failover
- **Data:** 2026-10-14 16:06–16:22
- **Status:** resolvido

**Linha do tempo (Prometheus)**

| Horário | 01 | 02 |
|---|---|---|
| 16:10 | `sqlserver_up 0` (CU32) | `sqlserver_up 0` (CU32) |
| 16:12:45–16:13 | `sqlserver_up 1`, versão **15.0.4430.1** | `sqlserver_up 1`, versão **15.0.4430.1** |
| 16:12:30 / 16:15:15 | — | scrape falhou por timeout (HML-013) |
| 16:16:31 | reboot (`os_boot_time_seconds`), `up 0` | reboot, `up 0` |
| 16:18:42 | exporter sobe (início atrasado) | exporter sobe |
| 16:19:30–16:21:45 | `sqlserver_up 0` (KB5128494) | `sqlserver_up 0` (KB5128494) |
| 16:21:45 | versão **15.0.4500.2** | versão **15.0.4500.2** |

- [x] `sqlserver_product_version` registrou as duas trocas de versão nas duas
  instâncias (uma série por versão, sem sobreposição).
- [x] Exporter sobreviveu às paradas do SQL (sem reinício) e voltou sozinho
  após o reboot; `sqlserver_instance` resolvido corretamente (o SQL estava no
  ar na partida, às 16:18:42).
- [x] Depois do patch: nenhum coletor com `success 0`; AG `MSSQL_AG`
  saudável (sync 2, health 2); `MirrorDB` com 01 principal, state 4.
- [x] Log shipping voltou a gravar monitor (KB4537869 corrigido):
  `primary_last_backup` 16:22:00, `secondary_last_copied`/`last_restored`
  16:18:07, latência 240 s.

---

## HML-015 — `blocking` falha esporadicamente na réplica secundária do AG (erro 976)

- **Ambiente:** EC2 02 (secundária do `MSSQL_AG`, `ALLOW_CONNECTIONS = ALL`), `0.4.2-homolog`
- **Encontrado em:** 2026-10-14, durante o HML-011
- **Severidade:** baixa (≈1,4% dos scrapes da 02 com `success 0` no `blocking`; a 01 nunca falhou)
- **Status:** resolvido (v0.4.0)

**Observado**

`sqlserver_scrape_collector_success{collector="blocking"} 0` na 02 em scrapes
isolados (15:22, 15:37, 15:39, 16:05, 16:38). Event Log:

```
collector=blocking error="mssql: The target database, 'AlwaysOnDB', is participating
in an availability group and is currently not accessible for queries. ..."
```

A mensagem sem prefixo vem do `rows.Err()` de
`SELECT status, COUNT(*) FROM sys.dm_exec_requests GROUP BY status`: a DMV é
de servidor e a conexão está no `master` (`database=master` no DSN), mas o
SQL Server aborta a leitura no meio. Hipótese: checagem de visibilidade por
request que toca o `AlwaysOnDB` quando ele está momentaneamente ilegível na
secundária. O mesmo erro é conhecido em jobs de snapshot de sessão em
ambiente de DR (relato do usuário). Um loop de 300 execuções da query como
`usr_exporter` (16:40–16:45) não reproduziu; a causa exata ficou em aberto.

**Correção (decisão: waiting_tasks + retry)**

- `sqlserver_blocked_sessions` / `_max_wait_ms` passam a vir de
  `sys.dm_os_waiting_tasks` (sem contexto de banco):
  `COUNT(DISTINCT session_id)`, `MAX(wait_duration_ms)`,
  `blocking_session_id > 0 AND blocking_session_id <> session_id`.
- `sqlserver_active_requests{status}` continua em `dm_exec_requests`, com
  **uma** nova tentativa só para o erro 976 (detectado pelo
  `SQLErrorNumber()` do driver). As contagens são acumuladas e emitidas só
  no fim, para a tentativa que falhou não gerar séries duplicadas (HML-004).
  Se falhar de novo, `success 0` como antes.
- Testes: retry com sucesso, 976 duas vezes, outro erro sem retry. Teste de
  mutação: sem o retry, 2 testes falham; com retry para qualquer erro, 3.

**Validação (2026-10-14, `0.4.3-homolog`)**

- [x] Bloqueio provocado na 01 (`##blk`, 16:56–17:01): `blocked_sessions`
  0 → 1 → 0 e `blocked_max_wait_ms` subindo ~30 s a cada 30 s (7 964 →
  277 967 ms) até o `ROLLBACK`.
- [x] 02 sem `success 0` no `blocking` de 17:03 a 17:49 (184 scrapes; antes
  da correção seriam ~2–3 falhas no período). Janela encerrada por decisão do
  usuário.
- Obs.: depois da atualização, o serviço da 02 ficou parado (16:55–17:02)
  até ser iniciado manualmente — o `service start` não chegou a subir o
  processo; sem indício de crash do exporter.

---

## HML-016 — Driver `denisenkom/go-mssqldb` arquivado: migrar para `microsoft/go-mssqldb`

- **Ambiente:** EC2 01 + 02 (SQL Server 2019 CU32), antes do upgrade para 2022/2025
- **Encontrado em:** 2026-10-14, planejamento do upgrade de versão
- **Severidade:** média (driver sem manutenção; sem TDS 8.0 / `encrypt=strict`)
- **Status:** resolvido (v0.5.0)

**Contexto**

O `denisenkom/go-mssqldb` v0.12.3 está arquivado; o `microsoft/go-mssqldb` é
o fork mantido e o driver oficial para Go. Antes de homologar o SQL Server
2022 e 2025, o exporter passa a usar o driver novo.

**Mudança**

- Import e `go.mod`: `github.com/microsoft/go-mssqldb` v1.11.2 (Go 1.25+, o
  projeto já está em 1.25.0).
- Mesmo nome de driver (`sqlserver`) e mesmos formatos de DSN; mesmo padrão
  de criptografia (só o pacote de login). Novo: `encrypt=strict` (TDS 8.0).
- Binário Windows +0,5 MB (15,06 → 15,58 MB); nada do SDK da Azure entra no
  binário.

**Validação**

- [x] `0.5.0-homolog` nas duas EC2 (2019 CU32, 2026-10-14 17:55):
  `sqlserver_up 1`, os 16 coletores com `success 1`, mesmos nomes de
  métrica que a `0.4.3-homolog` (= v0.4.0): 01 156 nomes / 3 779 séries
  (idêntico); 02 153 nomes / 3 802 séries (+5: o ISO do SQL 2022 montado em
  `E:` (CDFS) apareceu nas `os_filesystem_*`). Duração média do scrape
  0,29 s → 0,31 s (dentro do ruído).
- [x] SQL Server 2022 RTM e 2025 RTM (HML-017): os 16 coletores com
  `success 1` nas duas versões, nenhuma métrica sumiu.
- [x] `encrypt=strict` (TDS 8.0) na 01 (2025), 2026-10-14 19:45:
  - Sem certificado configurado no SQL Server o handshake TDS 8 é recusado:
    `openssl s_client -alpn tds/8.0` recebe EOF com 0 bytes (TLS 1.2 e 1.3) e
    o exporter fica com `sqlserver_up 0` (processo de pé, `/metrics` respondendo).
  - O driver **ignora `trustservercertificate=true` em modo strict**
    (`msdsn/conn_str.go`, `parseTLS`: `if encrypt == "strict" { trustServerCert = false }`,
    mesmo comportamento do Microsoft.Data.SqlClient): continua
    `x509: certificate signed by unknown authority`.
  - Com certificado autoassinado `CN=AWSDBSERVER01` configurado na instância
    e `certificate=C:/.../sql01.pem` no DSN: `sqlserver_up 1`, 16/16 coletores,
    scrape 0,30 s. No SQL, as 5 conexões do pool com `encrypt_option TRUE` e
    `protocol_version 134217728` (0x08000000 = TDS 8.0).
  - Testado do Mac com login falso (`Login failed` = TLS ok):
    `certificate=<pem>` falha conectando por IP (sem IP SAN) e funciona com
    `hostnameincertificate=AWSDBSERVER01`; `servercertificate=<pem>`
    (comparação de bytes) funciona por IP.
  - Documentado no README ("Encrypted connections").

---

## HML-017 — Rolling upgrade SQL Server 2019 → 2022 (02) → 2025 (01, 02) com AG, mirroring e log shipping (validação)

- **Ambiente:** EC2 01 + 02, exporter `0.5.0-homolog` (driver `microsoft/go-mssqldb`)
- **Encontrado em:** 2026-10-14
- **Severidade:** — (validação)
- **Status:** em andamento

**Plano**

Rolling upgrade a partir do 2019 CU32 (`15.0.4500.2`): 02 (secundária do AG,
espelho do mirroring, secundária do log shipping) para 2022; failover do AG e
do mirroring para a 02; 01 direto para 2025 (decisão de ganhar tempo, sem
equalizar as duas em 2022); depois 02 para 2025.

**Etapa 1 — upgrade da 02 para 2022 RTM (`16.0.1000.6`), 2026-10-14**

| Horário | Evento (visto no Prometheus) |
|---------|------------------------------|
| 18:03:37 | reboot do Windows (`os_boot_time_seconds`); 01 vê `mirroring_state` 1 (disconnected) |
| 18:05:50 | serviço do exporter sobe sozinho após o boot (`process_start_time_seconds`) |
| 18:08–18:10, 18:15, 18:19–18:22 | restarts do SQL durante o setup: 02 `sqlserver_up 0`, 01 mirroring disconnected |
| 18:22:06 | `product_version` `16.0.1000.6`, `sqlserver_up 1` |
| 18:22:13 | mirroring volta a 4 (synchronized) |

- O processo do exporter **não precisou ser reiniciado**: reconectou sozinho a
  cada restart do SQL; `sqlserver_instance` continuou correto.
- Único erro: `volumes` com `success 0` às 18:22:02, primeiro scrape após o SQL
  subir (bancos em recovery). Levou 4,98 s (timeout de query de 5 s); o scrape
  inteiro levou 5,17 s, abaixo do `scrape_timeout: 10s` (HML-013), então o
  Prometheus recebeu os dados parciais. Próximo scrape já `success 1`.
- `AlwaysOnDB` na 02 passou de `ONLINE` para `RECOVERING`: secundária em versão
  maior não é legível até o fim do upgrade. É o cenário do erro 976; `blocking`
  não falhou nenhuma vez depois do upgrade (correção do HML-015 confirmada).
- AG synchronized/healthy, filas zeradas; mirroring synchronized; log shipping
  2019 (01) → 2022 (02) restaurando normalmente.
- Diferenças nas métricas da 02 (3 802 → 4 532 séries), nenhuma métrica sumiu:
  - wait types 1 027 → 1 266 (+239 novos no 2022) em `sqlserver_wait_time_ms`,
    `sqlserver_max_wait_time_ms` e `sqlserver_waiting_tasks`;
  - `sqlserver_database_log_used_percent` ganhou `model_msdb` e
    `model_replicatedmaster` (bancos-modelo internos do 2022 para Contained AG,
    só nos performance counters, como o `mssqlsystemresource`). Cardinalidade
    baixa, sem ação;
  - `sqlserver_memory_clerk_used_kb` +1 clerk.
- Pré-existente, não relacionado ao upgrade: `LSAlert_AWSDBSERVER02` com
  `last_run_status 0` desde antes (threshold de restore ainda em 2 min do teste
  do HML-011).

**Etapa 2 — failover do AG e do mirroring para a 02 (2022), 18:39:33**

- AG: `replica_role` invertido (02 primária); a `AlwaysOnDB` foi atualizada na
  02 (`ONLINE`) e ficou na 01 em `RECOVERING` (Not Synchronizing). As duas
  instâncias reportam a réplica da 01 com `is_suspended 1`,
  `suspend_reason 7` (revalidation), `synchronization_state 0` e
  `synchronization_health 0`; o alerta do README
  (`max by (ag_name, db_name) (sqlserver_hadr_database_is_suspended) == 1`)
  dispararia.
- Mirroring: `MirrorDB` ficou `RECOVERING` na 02 (nova principal) e
  `RESTORING` na 01 — a principal no 2022 não sai do recovery com o espelho no
  2019. As séries de estado sumiram: ver HML-018.
- Nenhum coletor com `success 0`.

**Etapa 3 — upgrade da 01 para 2025 RTM (`17.0.1000.7`), 18:52–18:56**

- 18:47:06 reboot do Windows; 18:49:18 serviço do exporter sobe sozinho;
  restarts do SQL 18:52–18:53 e 18:54–18:55; 18:55:50 versão nova,
  `sqlserver_up 1`, 16/16 coletores. Nenhum coletor falhou.
- Métricas da 01 (3 777 → 5 143 séries): wait types 1 027 → 1 475 (+209 sobre
  o 2022), +15 memory clerks, os mesmos `model_msdb`/`model_replicatedmaster`.
- AG: a `AlwaysOnDB` da 01 seguiu suspensa até o `SET HADR RESUME` manual
  (18:59:00), depois synchronized; continua `RECOVERING` na 01 (secundária em
  versão maior que a primária não é legível).
- Log shipping quebrou como esperado: 19:00:15 `LSRestore` da 02 (2022) com
  `last_run_status 0` e o tempo desde o último restore, que oscilava 0–180 s
  (ciclo de 3 min), passou a crescer sem parar, acima do threshold.
- Mirroring: o `MirrorDB` da 02 ficou preso no recovery. ERRORLOG da 02:
  18:39:56 troca MIRROR → PRINCIPAL, "Converting database 'MirrorDB' from
  version 904 to the current version 957" e parada no primeiro passo
  (904 → 905): o passo gera log que precisa ser confirmado pelo espelho em
  modo síncrono, e o espelho no 2019 não aplicava log do 2022. Threads
  `DB MIRROR` ociosas em `DBMIRROR_EVENTS_QUEUE`; a 01 (sysadmin) via
  `MIRROR` / `SUSPENDED`. `SET PARTNER RESUME` na 01 levou a sessão a
  `SYNCHRONIZING`, mas a 02 só destravou no restart da etapa 4. Para um
  rolling upgrade real: não fazer failover do mirroring para o parceiro
  atualizado enquanto o outro estiver numa versão menor, ou suspender/
  remover a sessão antes.

**Etapa 4 — upgrade da 02 para 2025, failovers e failbacks**

| Horário | Evento |
|---------|--------|
| 19:08:21 | failover do AG para a 01 (2025 como primária); `AlwaysOnDB` da 02 (2022) suspensa |
| 19:17:51–19:21:51 | restarts do SQL na 02 durante o setup (sem reboot, sem restart do exporter) |
| 19:18:41 | `blocking` com erro 983 na 02 (HML-020) |
| 19:18:51 | `MirrorDB` da 02 em `RECOVERY_PENDING` |
| 19:21:51 | 02 em `17.0.1000.7`; `MirrorDB` `ONLINE` como principal na 02, séries de estado de volta (`role 1`, `state 4`) |
| 19:23:06 | failover do AG para a 02; `volumes` da 01 com timeout (HML-019) |
| 19:24:21 | `LSRestore` volta a `last_run_status 1` (secundária do log shipping no 2025) |
| 19:24:36 | `AlwaysOnDB` da 01 suspensa com `suspend_reason 1` (vinda do parceiro, típico de failover forçado) |
| 19:24:51 | failback do mirroring para a 01: principal 2025, synchronized |
| 19:28:10 | `SET HADR RESUME` na 01, `is_suspended 0` nas duas |

- As duas em 2025: 01 5 129 / 02 5 136 séries, 1 475 wait types cada, scrape
  0,29 s.

**Cobertura final**

| Papel | 2019 CU32 | 2022 RTM | 2025 RTM |
|-------|-----------|----------|----------|
| AG primária | ✓ | ✓ | ✓ |
| AG secundária | ✓ | ✓ | ✓ |
| Mirroring principal (`ONLINE`) | ✓ | — (preso no recovery) | ✓ |
| Mirroring espelho | ✓ | ✓ | ✓ |
| Log shipping primária | ✓ | — | ✓ |
| Log shipping secundária | ✓ | ✓ | ✓ |

O exporter nunca precisou ser reiniciado por causa de um upgrade e reconectou
sozinho a cada restart do SQL; o serviço sobe sozinho após reboot. Falhas de
coletor vistas: só as de HML-019 e HML-020, uma coleta cada.

---

## HML-018 — Estado do mirroring some (sem série) quando o banco espelhado não está acessível

- **Ambiente:** EC2 01 + 02, durante o HML-017 (principal 2022, espelho 2019)
- **Encontrado em:** 2026-10-14 18:39
- **Severidade:** média (alerta em `sqlserver_mirroring_state` não dispara)
- **Status:** resolvido (documentação, v0.5.0)

**Sintoma**

Depois do failover do mirroring para a 02 (2022), `sqlserver_mirroring_info`,
`_role` e `_state` deixaram de existir nas duas instâncias. Só os performance
counters (`_log_send_queue_kb`, `_redo_queue_kb`, `*_total`) continuaram.

**Causa**

Com o `MirrorDB` em `RECOVERING` na 02, `sys.database_mirroring` devolve
`mirroring_guid`, `mirroring_role_desc` e `mirroring_state_desc` NULL **inclusive
para sysadmin** (validado na 02 como sysadmin e com
`EXECUTE AS LOGIN = 'usr_exporter'`: mesmo resultado). O coletor filtra
`mirroring_guid IS NOT NULL`, então a linha some. Na 01 (espelho) o login sem
sysadmin já recebia NULL (limitação do HML-011). Limitação do SQL Server, não
de permissão: não há outra fonte para o estado da sessão.

**Impacto**

Um alerta `sqlserver_mirroring_state != 4` não dispara: a série desaparece em
vez de mudar de valor.

**Proposta**

Documentar no README um alerta por ausência, usando os performance counters
(que continuam presentes para todo banco espelhado) como referência, e lembrar
que `sqlserver_database_state{state="RECOVERING"}` também cobre o caso.

**Evidência adicional (etapa 3 do HML-017)**

Com a sessão `SUSPENDED`, só um sysadmin no espelho (01) via o estado; para o
`usr_exporter` a 02 não retornava linha (banco em recovery) e a 01 retornava
NULL (espelho sem sysadmin). Na falha mais séria, o estado sumiu nas duas
pontas. Quando o `MirrorDB` voltou a `ONLINE` na 02 (19:21:51), as séries
voltaram sozinhas.

**Resolução**

Documentado no README (seção Database mirroring), com a expressão testada no
Prometheus (retornou o `MirrorDB` durante todo o período sem estado):
`count by (db_name) (sqlserver_mirroring_log_send_queue_kb) unless on (db_name) count by (db_name) (sqlserver_mirroring_state)`.
Não emitir um estado "desconhecido" no coletor: a expressão cobre o caso sem
inventar um valor que o SQL Server não informa.

---

## HML-019 — `volumes` estoura o timeout de query quando um banco está trocando de estado

- **Ambiente:** EC2 01 + 02, durante o HML-017
- **Encontrado em:** 2026-10-14
- **Severidade:** baixa (uma coleta parcial por evento)
- **Status:** adiado (ROADMAP, Later)

**Observado**

`volumes` com `success 0` duas vezes, sempre com ≈5 s de duração (timeout de
query de 5 s):

- 18:22:02 na 02, primeira coleta após o SQL subir no 2022 (bancos em recovery);
- 19:23:06 na 01, no failover do AG (a `AlwaysOnDB` da 01 em transição).

`sys.dm_os_volume_stats(database_id, file_id)` é aplicado a todos os arquivos
de `sys.master_files` e espera enquanto algum banco está mudando de estado. O
`scrape_timeout: 10s` (HML-013) absorveu os dois casos: o scrape levou ≈5,2 s,
o Prometheus recebeu os dados parciais e a coleta seguinte já teve `success 1`.
Bancos estáveis em `RECOVERING` (secundária ilegível) ou `RESTORING` não causaram
falha.

**Opções**

Filtrar a query por `sys.databases.state IN (0, 1)` (ONLINE/RESTORING), o que
reduz a janela mas não elimina a corrida se a transição começar durante a
query. Ganho pequeno; decidir depois.

---

## HML-020 — `blocking` falha com erro 983 logo após o SQL Server subir (réplica do AG em `RESOLVING`)

- **Ambiente:** EC2 02, durante o HML-017 (restart do setup do 2025)
- **Encontrado em:** 2026-10-14 19:18:41
- **Severidade:** baixa (uma coleta)
- **Status:** resolvido (v0.5.0)

**Observado**

```
collector=blocking error="mssql: Unable to access availability database 'AlwaysOnDB'
because the database replica is not in the PRIMARY or SECONDARY role. ... (983)"
```

Mesma query do HML-015 (`sys.dm_exec_requests`), com a réplica ainda
resolvendo o papel logo após o SQL subir. O retry do HML-015 só reconhecia o 976.

**Correção**

`isAGDatabaseNotAccessible` reconhece 976 e 983; o teste de retry roda para os
dois códigos. Mutação (remover o 983 do `case`) quebra
`TestBlockingScrapeRequestsRetriesOnceOnAGNotAccessible/983`. O `RESOLVING`
dura alguns segundos, então o retry imediato nem sempre o supera; reduz o
ruído sem mudar o comportamento essencial.

**Validação**

Por teste unitário, sem reprodução nas EC2 (decisão do usuário): a janela do
`RESOLVING` é de segundos e o caminho de retry é o mesmo já validado ao vivo no
HML-015. Build `0.5.1-homolog` publicado no S3.

---

## HML-021 — `sqlserver_sessions_by_login` com carga de vários logins (validação)

- **Ambiente:** EC2 01 (SQL Server 2025 RTM), `0.6.0-homolog`; AG, mirroring e
  log shipping desmontados, 02 desligada
- **Encontrado em:** 2026-10-15
- **Severidade:** — (validação do item 1 da Fase 6 do roadmap)
- **Status:** resolvido

**Cenário**

`hack/loadgen` rodando do Mac contra a 01: 10 logins `loadgen_01..10`
(`setup.sql`), 2 workers cada, mistura de leituras, scans, updates, esperas
longas e bloqueio proposital.

**Resultado**

- `sqlserver_sessions_by_login` = 2 por `loadgen_NN`, mais
  `AWSDBSERVER01\usr_dbsvc` (SSMS), `usr_exporter` (pool do exporter, 5) e
  `NT SERVICE\SQLTELEMETRY`; a soma bate com `sqlserver_user_connections`.
- Um primeiro teste mostrou 4 sessões por login: eram dois processos do
  loadgen ao mesmo tempo (um disparado pela UI). Reproduzido num SQL Server
  2025 local em container com um processo: exatamente 2 sessões, sempre nas
  mesmas 2 portas TCP.
- Sem RCSI no `LoadTestDB`, o scan do `status_report` esperava o `blocker` e
  38 sessões ficaram bloqueadas (o `blocking` reportou corretamente, espera
  máxima de 54,7 s); com RCSI e pesos rebalanceados: ~42 batch requests/s,
  bloqueio só em rajadas e version store do tempdb em uso.
- 16/16 coletores com `success 1` durante toda a carga; scrape máximo 1,06 s.

---

## HML-022 — Coletor `active_queries`: requests longas e transações abertas (validação)

- **Ambiente:** EC2 01 (SQL Server 2025 RTM), `0.7.0-homolog`
- **Encontrado em:** 2026-10-15
- **Severidade:** — (validação do item 2 da Fase 6 do roadmap)
- **Status:** resolvido

**Cenário**

`hack/loadgen` (10 logins × 2 workers, a partir de 16:32) mais dois testes no
SSMS por volta de 16:32: um `BEGIN TRAN` + `UPDATE` sem `COMMIT` (sessão ociosa
com transação aberta) e um `WAITFOR DELAY '00:02:00'` (request longa).
`ROLLBACK` às 16:44.

**Resultado**

| Métrica | Observado |
|---------|-----------|
| `sqlserver_request_max_elapsed_seconds{command="WAITFOR"}` | 27 → 117 s, +15 s por scrape; depois do fim do `WAITFOR` do SSMS, só os `long_wait` do loadgen (até ~30 s) |
| `sqlserver_requests_long_running{over_seconds="60"}` | 0 → 1 às 16:33:20 (o `WAITFOR` passou de 57 para 72 s), volta a 0 às 16:34:20 |
| `sqlserver_requests_long_running{over_seconds="10"}` | 1–2 (o `WAITFOR` do SSMS + `long_wait`/`blocker` do loadgen) |
| `sqlserver_sleeping_sessions_with_open_transaction` | 1 durante todo o teste, 0 às 16:45:11 |
| `sqlserver_oldest_open_transaction_seconds` | 32 → 767 s em linha reta; 17 s às 16:45:11 (só os `blocker` do loadgen) |
| `sqlserver_request_max_elapsed_seconds` de `SELECT`/`UPDATE` | 0 s (queries do loadgen são rápidas) |

- `active_queries` sempre com `success 1`, ≈27 ms por scrape.

---

## HML-023 — Coletor `top_queries` com texto ofuscado (validação)

- **Ambiente:** EC2 01 (SQL Server 2025 RTM), `0.8.0-homolog`, com
  `collectors: {top_queries: true}` e `top_queries: {include_text: true}`
- **Encontrado em:** 2026-10-15
- **Severidade:** — (validação do item 3 da Fase 6 do roadmap)
- **Status:** resolvido

**Antes da EC2**

Query e exporter completos validados contra um SQL Server 2025 local em
container: `query_hash` em hexadecimal, `db_name` vindo dos atributos do plano,
texto por statement (não o batch inteiro), literais ofuscados e truncamento.

**Resultado (carga do `hack/loadgen` a partir de 17:01)**

- `top_queries` com `success 1` em ≈19 ms por scrape; 50 séries por métrica
  (= `limit` padrão) e 50 `sqlserver_top_query_info`.
- Top por CPU: o `status_report` do loadgen (scan da tabela; 450 s de CPU em
  7 691 execuções, ≈58 ms cada) com folga, seguido das queries do próprio
  exporter em `master` (documentado: num servidor ocioso o top é o exporter).
- Texto: `TOP (50)` → `TOP (?)`, strings das queries do exporter → `?`,
  parâmetros `@p1` mantidos, corte em 120 caracteres com `...`.
- Top 5 por CPU acumulada (hash, CPU, execuções) conferido no SSMS com
  `sys.dm_exec_query_stats` agrupado por `query_hash`: mesmos hashes, mesma
  ordem.

## HML-024 — Coletor `index_fragmentation` em background e `CONNECT ANY DATABASE` (validação)

- **Ambiente:** EC2 01 (SQL Server 2025 RTM), `0.9.0-homolog`, com
  `collectors: {index_fragmentation: true}` e
  `index_fragmentation: {interval: 2m, min_page_count: 100}`
- **Encontrado em:** 2026-10-15
- **Severidade:** — (validação do item 4 da Fase 6 do roadmap)
- **Status:** resolvido

**Antes da EC2**

Validado contra um SQL Server 2025 local em container: tabela com chave
`NEWID()` e 20 000 inserts linha a linha levou o `PK_Guids` de 0,05% a 99,03%
no ciclo seguinte (intervalo de 1 min); banco chamado `Odd]Name` lido sem erro
(nome entre colchetes); scrape em microssegundos.

**Resultado**

- Primeira subida (17:31) sem o coletor: a seção `index_fragmentation:` com as
  opções estava no arquivo, mas faltava `index_fragmentation: true` dentro de
  `collectors:` — a seção de opções não liga o coletor (mesma regra do
  `top_queries`). `config.example.yml` passou a dizer isso explicitamente.
- 17:37, **sem grant novo**: `last_collection_success 1`,
  `last_collection_databases 0` em 0,13 s — o `usr_exporter` não tem usuário
  nos bancos e o filtro `HAS_DBACCESS` os ignora sem erro. Como `success 1`
  com zero bancos esconde que nada é coletado, o README documenta o alerta
  `last_collection_databases == 0`.
- Após `GRANT CONNECT ANY DATABASE TO usr_exporter` (≈17:38), o ciclo
  seguinte leu **3 bancos** em 0,36 s: `LoadTestDB.dbo.Guids.PK_Guids` com
  **98,9%** (1 614 páginas, criado pelo script de inserts aleatórios),
  `Orders` com 0,02% / 0,86%; `MirrorDB` com índices de 277 778 e 286 838
  páginas (~4,4 GB) e fragmentação < 0,3%. Modo `LIMITED` lê só o nível acima
  da folha: barato mesmo em índices de GB.
- O scrape do coletor reporta duração 0 (serve o cache);
  `scrape_collector_success 1` o tempo todo.

**Fase RDS (2026-10-18):** o usuário master do RDS **não** consegue conceder
o grant (`Grantor does not have GRANT permission`). Alternativa documentada
no README e validada no HML-026: criar um usuário para o login em cada banco
a ser lido. O
`rdsadmin` já é acessível a qualquer login, então mesmo sem nada o coletor lê
1 banco no RDS (sem índices acima de 100 páginas).

---

## HML-025 — Exporter como serviço Linux (systemd) no bastion da fase RDS (validação)

- **Ambiente:** bastion Linux da fase RDS (o exporter roda nele e acessa o RDS
  pela VPC; o Prometheus local lê o exporter por túnel SSH)
- **Encontrado em:** 2026-10-15
- **Severidade:** — (até aqui o Linux não tinha forma documentada de rodar
  como serviço; o subcomando `service` é só do Windows)
- **Status:** resolvido (v0.9.1)

**Entrega**

- Unit `contrib/systemd/mssql-exporter.service`, incluída nos arquivos de
  release a partir da v0.9.1 junto com `config.example.yml`: usuário de
  sistema `mssql-exporter` sem shell, config em `/etc/mssql-exporter`
  (`0640 root:mssql-exporter`), `Restart=on-failure`, sandbox do systemd.
- README: seção "Running as a Linux service (systemd)" com o roteiro de
  instalação a partir da release.

**Validação local (2026-10-15, container Debian 12 com systemd, build snapshot)**

- [x] Roteiro do README a partir do tarball: checksum OK, binário, usuário,
  config e unit instalados com as permissões esperadas.
- [x] `systemctl enable --now`: serviço ativo como `mssql-exporter`,
  `sqlserver_up 1` e os 16 coletores padrão + `index_fragmentation` com
  `success 1` contra o SQL Server 2025 local.
- [x] Com o sandbox, os coletores de SO (CPU, memória, disco, rede, boot)
  funcionam; `os_filesystem_*` e `os_process_*` não aparecem no container nem
  sem o sandbox (sem partição física nem processo do SQL Server) — conferir
  `os_filesystem_*` numa VM.
- [x] `kill -9` → reiniciado em 5 s (`NRestarts=1`); `systemctl stop` →
  `shutting down` no journal, status 0; restart do container → serviço volta
  sozinho (`enabled`).
- [x] `systemd-analyze security`: exposição 3.1 ("OK").

**Validação no bastion (2026-10-18, Ubuntu 26.04 `amd64`, release v0.9.1)**

- [x] Roteiro do README seguido no bastion, com uma diferença: o repositório
  ainda é privado, então o `curl` da release no GitHub não funciona; o
  tarball e o `.sha256` foram para o S3 e baixados com `aws s3 cp`. Depois de
  público, o roteiro do README vale como está.
- [x] `systemctl enable --now` às 23:14:49 UTC: ativo de primeira, sem erro
  no journal; `web.listen_address: "127.0.0.1:9187"` e túnel SSH
  (`-L 19187:127.0.0.1:9187`) até o Prometheus local, alvo `up` às 23:15.
- [x] Atualizações para `0.9.2-homolog` e `0.9.2` trocando só o binário
  (`install` + `systemctl restart`), sem mexer em config ou unit.
- `cd /etc/mssql-exporter` como `ubuntu` dá *Permission denied* (pasta
  `0750 root:mssql-exporter`, como planejado); editar com `sudo`.

---

## HML-026 — `database_log_used_percent` com bancos ocultos e, no RDS, sem os bancos de sistema

- **Ambiente:** RDS for SQL Server 2025 Standard (`17.0.5005.3`, Single-AZ),
  exporter `0.9.1` no bastion; reproduzido no SQL Server 2025 local
- **Encontrado em:** 2026-10-18
- **Severidade:** baixa
- **Status:** resolvido (v0.9.2)

**Observado**

No RDS, `sqlserver_database_log_used_percent` vinha para `model_msdb`,
`model_replicatedmaster`, `mssqlsystemresource`, `tempdb` e o banco de usuário
`StressDB`, mas não para `master`, `model`, `msdb` e `rdsadmin`. Nenhuma outra
métrica conhece os três primeiros: não estão em `sys.databases`.

**Causa**

A métrica repassa todas as instâncias do contador `Percent Log Used`
(`sys.dm_os_performance_counters`). O contador tem instâncias para bancos
ocultos — `mssqlsystemresource` em qualquer versão e, no SQL Server 2025,
`model_msdb`/`model_replicatedmaster` (o container local mostra os mesmos) —
e, só no RDS, não tem instância para os bancos de sistema. Não é permissão:
`usr_exporter` e `usr_dba` veem a mesma lista.

Fontes alternativas testadas no RDS como `usr_exporter`:

| Fonte | Bancos | Faltam |
|---|---|---|
| Contador `Percent Log Used` | StressDB, tempdb + 3 ocultos | master, model, msdb, rdsadmin |
| `DBCC SQLPERF(LOGSPACE)` | master, tempdb, msdb, rdsadmin + 2 ocultos | **model e StressDB** |
| `sys.dm_db_log_stats` | os 6 de `sys.databases` | — |

`dm_db_log_stats` é completo, mas `active_log_size_mb` mede outra coisa (VLFs
ativos): master 10,6% contra 44% do DBCC, msdb 0,5% contra 10,9%. Misturar
as fontes faria a mesma métrica significar coisas diferentes por ambiente.

**Correção** — branch `fix/log-used-percent-ghost-databases`

- Contador cruzado com `sys.databases` (`COLLATE DATABASE_DEFAULT` no nome,
  para não depender da collation do DMV): some a série dos bancos ocultos em
  qualquer ambiente.
- README documenta que no RDS os bancos de sistema não têm a métrica (os de
  usuário têm).
- ROADMAP (Later): métricas próprias do `sys.dm_db_log_stats` (VLFs, motivo
  de não truncar o log, log desde o último backup), que funcionam no RDS para
  todos os bancos.

Validado no SQL Server 2025 local: antes 9 séries (com `model_msdb`,
`model_replicatedmaster` e `mssqlsystemresource`), depois só os 6 bancos de
`sys.databases`.

**Regressão em outras plataformas (2026-10-18)**

- Histórico da EC2 01 (2025) no Prometheus: `log_used_percent` tinha os 10
  bancos reais **+** `model_msdb`, `model_replicatedmaster` e
  `mssqlsystemresource`; com o join ficam exatamente os 10 (bancos
  `RESTORING` de mirroring/log shipping continuam, estão em `sys.databases`).
  Em 2019/2022 só o `mssqlsystemresource` some.
- Nomes difíceis, SQL Server 2025 local (Linux): 11 bancos de teste —
  `]`, `/`, `\`, `#`, `%`, `'`, `.`, `(`/`)`, espaço, acentos e um nome de 128
  caracteres. Instância do contador idêntica ao `sys.databases` em todos; o
  exporter com a correção emitiu as 11 séries. Não verificado no Windows (o
  DMV lê os contadores do próprio SQL Server, não do PerfMon, então não se
  espera troca de caracteres; os bancos da EC2 não tinham nomes especiais).

**Validação no RDS (2026-10-18, `0.9.2-homolog` no bastion, reiniciado 23:32 UTC)**

- `log_used_percent` só para `StressDB` (17) e `tempdb` (9); os três bancos
  ocultos sumiram. `databases` com `success 1` em 22 ms; os 18 coletores com
  `success 1`.
- Alternativa ao `CONNECT ANY DATABASE` (HML-024): com
  `CREATE USER usr_exporter FOR LOGIN usr_exporter` no `StressDB`, o ciclo
  seguinte do `index_fragmentation` leu 2 bancos (`rdsadmin` + `StressDB`) em
  0,46 s: `PK_Orders` com 277 778 páginas (0,01%) e `cx_Numbers` (0%).

---

## HML-027 — Linha de base no RDS Single-AZ: permissões, coletores, carga e TLS (validação)

- **Ambiente:** RDS for SQL Server 2025 Standard (`17.0.5005.3`, RTM-CU9),
  `db.t3.xlarge` Single-AZ, `mssql-exporter-hml`, fuso UTC-3, subnets
  privadas; exporter `0.9.2` no bastion (HML-025)
- **Encontrado em:** 2026-10-18
- **Severidade:** — (validação)
- **Status:** resolvido

**Permissões**

Os mesmos grants da EC2, dados pelo usuário master (`usr_dba`):
`VIEW SERVER STATE`, `VIEW ANY DEFINITION` e `SELECT` em
`msdb.dbo.sysjobs`/`sysjobhistory` (HML-002 repetido no RDS: OK).
`CONNECT ANY DATABASE` **não** pode ser concedido (HML-024); usuário por banco
no lugar.

**Coletores (padrão + `top_queries` e `index_fragmentation`)**

- Os 18 coletores com `success 1` desde a primeira coleta; medianas por
  coletor iguais às da EC2 (`perf_counters` ≈ 0,13 s, demais < 0,09 s).
- `server_info`: Standard Edition, 4 CPUs, 16 GB, `is_hadr_enabled 1` mesmo
  em Single-AZ.
- `volumes`: `D:\` do RDS (21,4 GB). `io_file_stats`: 13 arquivos.
- `backups`: FULL de `model`, `msdb` e `rdsadmin` feitos pelo RDS na criação
  da instância (retenção de backup 0).
- `agent_jobs`: `syspolicy_purge_history`.
- `sqlserver_instance` = `EC2AMAZ-UU69199` (nome do host Windows por trás do
  RDS), não o identificador da instância — rever no failover do Multi-AZ.
- `rdsadmin` é acessível a qualquer login: entra em `databases`,
  `index_fragmentation` etc.
- `log_used_percent`: ver HML-026.

**Carga (`hack/loadgen`, 10 logins × 2 workers, 23:52–00:12 UTC, pelo túnel)**

- 55 020 statements, nenhum erro no loadgen; ≈33 batch requests/s; CPU do
  RDS 24% em média (pico 29%).
- `sessions_by_login`: 2 sessões por `loadgen_NN`, mais `usr_exporter`,
  `usr_dba` e as contas de serviço do RDS (`NT AUTHORITY\SYSTEM`,
  `NETWORK SERVICE`).
- `active_queries`: até 14 requests acima de 10 s, `UPDATE` com até 18 s e
  até 13 sessões bloqueadas — todos na fila do `blocker` (segura a linha
  quente por 20 s) atrás de `hot_update`; conferido no
  `sys.dm_exec_requests` (cadeia `LCK_M_X` até a sessão do `blocker`).
  Transação aberta mais antiga até 36 s.
- `top_queries`: 1º lugar o `status_report` do loadgen (223 s de CPU em 1 405
  execuções, ≈159 ms cada; na EC2 ≈58 ms), texto com literais ofuscados; em
  seguida as queries do próprio exporter em `master`.
- `index_fragmentation` com 3 bancos (`rdsadmin`, `StressDB`, `LoadTestDB`,
  este com usuário criado para o `usr_exporter`).
- Nenhum coletor com `success 0` durante a carga.

**TLS (bastion → RDS, 23:58–23:59 UTC)**

| DSN | Resultado |
|---|---|
| sem `encrypt` | conecta (só o login criptografado) |
| `encrypt=true&certificate=global-bundle.pem` | OK, `encrypt_option` TRUE |
| `certificate=` com uma CA que não é a da AWS (`.pem`) | `sqlserver_up 0`, `x509: certificate signed by unknown authority` (a validação acontece) |
| `certificate=/etc/ssl/certs/ca-certificates.crt` | exporter sai na hora: `certificate type .crt is not supported` (o driver só aceita `.pem`/`.der`); systemd reinicia a cada 5 s até corrigir |
| `encrypt=strict&certificate=global-bundle.pem` | OK, `protocol_version` 0x08000000 (TDS 8.0), 18 coletores com `success 1` |

No RDS, o `encrypt=strict` funciona sem nenhuma configuração no servidor (na
EC2 foi preciso instalar um certificado na instância). O bastion ficou com
`encrypt=strict`. README: seção de TLS com o caso do RDS e a restrição de
extensão do arquivo.

---

## HML-028 — Backups no RDS: snapshots e log a cada 5 min aparecem; backups de criação gravados em `NUL` contam como FULL

- **Ambiente:** RDS for SQL Server 2025 SE (`mssql-exporter-hml`), fuso UTC-3,
  exporter `0.9.2`
- **Encontrado em:** 2026-10-18/19
- **Severidade:** média (o achado do `NUL` faz uma instância sem backup
  parecer ter backup FULL)
- **Status:** resolvido (v0.9.3)

**1. Snapshot manual com retenção 0 (2026-10-19, 01:16 UTC)**

`aws rds create-db-snapshot` registrou no `msdb.backupset` um FULL de cada um
dos 6 bancos com `is_snapshot = 1`, `copy_only = 0` e dispositivo virtual
(`device_type` 7), além de FULL comuns de `model`/`msdb` logo antes. O
coletor passou a mostrar `backup_type="full"` às 01:16:47 UTC para todos — fuso
convertido certo (HML-010 vale no RDS). A suspensão de I/O do snapshot
Single-AZ não afetou nenhum coletor.

- `backup_last_duration_seconds` = 1 s (o snapshot é instantâneo).
- `backup_last_size_bytes` = tamanho do banco (`StressDB` 2,3 GB), não o que
  o snapshot ocupa no S3.

**2. Retenção 0 → 1 (2026-10-19, `apply` às 13:57 UTC, sem reinício)**

Eventos do RDS: `Enabled automated backups` 13:57:25, snapshot terminado
14:01:37, `Emergent Snapshot Request` (o RDS forçou `model`, `msdb` e
`rdsadmin` de volta para FULL e tirou outro snapshot) terminado 14:03:48.

- `databases`: modelo de recuperação dos bancos de usuário de `SIMPLE` para
  `FULL` (com retenção 0 o RDS os mantém em `SIMPLE`; `master` continua
  `SIMPLE`).
- `msdb.backupset`: FULL de `master` em disco local
  (`D:\rdsdbdata\BACKUP`), dois snapshots (10:59 e 11:02 locais), um FULL de
  verdade de `StressDB` (41 s) e `LoadTestDB` para iniciar a cadeia de log e,
  depois, **BACKUP LOG a cada 5 min** (dispositivo virtual) para todos os
  bancos em FULL.
- Coletor: `backup_type="full"` às 14:02:56 UTC (o último snapshot) e
  `backup_type="log"` às 14:07:01 e 14:12:02 UTC — o intervalo de 5 min
  aparece como esperado. Alertas de "sem backup de log há mais de N minutos"
  funcionam no RDS (sem log para `master`, em `SIMPLE`).

**3. Backups de criação em `NUL`**

Na criação da instância (retenção 0) o RDS faz `BACKUP DATABASE ... TO DISK =
'nul'` de `model`, `msdb` e `rdsadmin` (só para iniciar a cadeia; o backup é
descartado). O coletor reporta esses registros como último FULL. Com a
retenção em 0 eles seriam os únicos backups da instância para sempre: um
painel mostraria backups que não existem e um alerta de "sem FULL há mais de
25 h" ficaria silencioso no primeiro dia. O mesmo vale fora do RDS para quem
faz `BACKUP ... TO DISK = 'NUL'` (truque comum para cortar a cadeia de log).

**Proposta**

Ignorar no coletor `backups` os registros cujo `physical_device_name` é
`NUL` (sem diferenciar maiúsculas). Snapshots (`is_snapshot = 1`) continuam
contando: o RDS restaura a partir deles.

**4. Backup nativo pelas procs do RDS (2026-10-19, ≈18:14 UTC)**

Option group com `SQLSERVER_BACKUP_RESTORE` e IAM role
`rds-backup-restore-role`; `msdb.dbo.rds_backup_database` para o S3
(retenção de backup automático em 0 nesse momento — o `apply` das 18:03
desligou os backups automáticos e o RDS voltou todos os bancos para
`SIMPLE`).

| Tarefa | Banco | Tipo | S3 | `msdb` / coletor |
|---|---|---|---|---|
| 1 | `LoadTestDB` | FULL | 56,6 MB | `full` 18:14:53 UTC, 1 s |
| 2 | `StressDB` | FULL | 2,1 GB | `full` 18:17:28 UTC, **36 s**, 2,3 GB |
| 3 | `LoadTestDB` | DIFFERENTIAL | 537 KB | `differential` 18:16:52 UTC |

- As três tarefas rodaram em paralelo, todas `SUCCESS`. A tarefa espera na
  fila do RDS antes de começar (criada 18:13:55, iniciada 18:14:52).
- Ao contrário do snapshot, duração e tamanho são reais: alertas de duração
  de backup fazem sentido para o backup nativo.
- No `msdb` o backup nativo é igual ao FULL que o RDS faz por conta própria:
  dispositivo virtual (`device_type` 7) com um GUID no `physical_device_name`,
  não o caminho do S3. Não dá para separar um do outro — e não precisa: os
  dois são restauráveis.
- `rds_backup_database` **não** é `COPY_ONLY`: um FULL nativo vira a nova base
  dos diferenciais.
- Com isso os três tipos ficam validados no RDS: FULL (snapshot, automático e
  nativo), DIFFERENTIAL (nativo) e LOG (automático, a cada 5 min).

**Correção** — branch `fix/backups-ignore-nul-device`

- O coletor descarta, antes de escolher o último backup por banco/tipo, os
  backup sets com alguma família em `NUL` ou `NUL:` (`UPPER(RTRIM(...))`; um
  backup em stripes com uma parte em `NUL` também não restaura). Assim o
  backup real anterior é o reportado. `/dev/null` nem é aceito pelo SQL
  Server no Linux (erro 3634).
- `msdb.dbo.backupmediafamily` tem `SELECT` para `public`, como o
  `backupset`, localmente e no RDS: nenhum grant novo.
- README: o filtro, como os backups automáticos do RDS aparecem e a nova
  tabela lida.

**Validação**

- SQL Server 2025 local, com `BACKUP DATABASE model TO DISK = 'NUL'`
  (gravado como `NUL`, `nul` e `NUL:`): a versão anterior reportava o FULL
  em `NUL`; a corrigida não reporta nenhum. Com um backup real às 14:16:56
  seguido de um em `NUL` às 14:16:58, a anterior reporta 14:16:58 e a
  corrigida 14:16:56.
- RDS (backups automáticos ligados): as duas versões emitem as mesmas 11
  séries com os mesmos valores.
- 4 mutações na query (`EXISTS`, só `'NUL'`, sem `UPPER/RTRIM`, join errado
  com a mídia) pegas pelos testes.

---

## HML-029 — `sqlserver_instance` no RDS: nome genérico do host, muda a cada parada/partida e fica vazio se o exporter sobe antes do banco

- **Ambiente:** RDS for SQL Server 2025 SE, exporter `0.9.2`/`0.9.3` no bastion
- **Encontrado em:** 2026-10-19
- **Severidade:** média (séries trocam de nome e o rótulo não identifica a
  instância)
- **Status:** resolvido (v0.10.0)

**Observado**

- O rótulo é o `@@SERVERNAME`, que no RDS é o nome do host Windows por trás
  da instância: `EC2AMAZ-UU69199` em 2026-10-18 e `EC2AMAZ-JFK9UC6` em
  2026-10-19, só por parar e ligar o RDS (sem failover). Cada troca quebra a
  continuidade das séries e dos dashboards, e com várias instâncias RDS o
  nome não diz qual é qual.
- O bastion ligou às 13:39 UTC e o RDS só às 13:54: o exporter subiu sem
  conectar e, como documentado, ficou com `sqlserver_instance=""` até ser
  reiniciado (14:20).

**Correção** — branch `feat/instance-label`

- Opção `instance_label` (YAML), `--sqlserver.instance-label` (flag) e
  `SQL_SERVER_INSTANCE_LABEL` (env), com a precedência de sempre. Preenchida,
  o valor é usado como está e o `@@SERVERNAME` nem é lido: o rótulo fica
  estável em parada/partida e failover e existe mesmo com o banco fora no
  startup. Vazia ou ausente, o comportamento automático de antes. Valor só
  com espaços é recusado na inicialização.
- Uma opção só (e não `auto_label` + `custom_label`): preenchida significa
  fixa, sem combinação contraditória para validar.
- README (com a recomendação de usar o identificador da instância no RDS) e
  `config.example.yml`.
- ROADMAP (Later): no modo automático, resolver o nome na primeira conexão
  bem-sucedida em vez de só no startup.

**Validação local (SQL Server 2025)**

| Cenário | `sqlserver_instance` |
|---|---|
| automático, banco no ar | `mssql2025` |
| `--sqlserver.instance-label=orders-prod`, banco no ar | `orders-prod` (o `sqlserver_servername` segue `mssql2025`) |
| fixo, banco fora no startup | `orders-prod`, com `sqlserver_up 0` |
| automático, banco fora no startup | vazio, com o aviso sugerindo `instance_label` |

Testes de precedência (YAML < env < flag, flag vazia volta ao automático) e
de valor em branco; 5 mutações no `config.go` pegas.

**Validação no RDS (`0.9.4-homolog`, `instance_label: "mssql-exporter-hml"`,
15:14 UTC):** `sqlserver_instance="mssql-exporter-hml"` em todas as séries,
`sqlserver_servername` com o host real. No failover do HML-030 o
`servername` passou de `EC2AMAZ-JFK9UC6` para `EC2AMAZ-94TRL9U` e o rótulo
continuou `mssql-exporter-hml`: as séries seguiram sem quebra.

---

## HML-030 — Conversão Single-AZ → Multi-AZ e failover do RDS: Always On aparece sem falhas de coleta (validação)

- **Ambiente:** RDS for SQL Server 2025 SE (`db.t3.xlarge`), exporter
  `0.9.3` no bastion; amostras a cada 30 s
- **Encontrado em:** 2026-10-19
- **Severidade:** — (validação)
- **Status:** resolvido

**Conversão (`apply` às 14:20 UTC, `available` com Multi-AZ às 15:02)**

| UTC | RDS | Exporter |
|---|---|---|
| 14:20 | `modifying`, Multi-AZ falso | sem séries `sqlserver_hadr_*` |
| 14:55 | `modifying` | aparece o AG `RDSAG0` (banco `rdsadmin`) |
| 14:56 | Multi-AZ verdadeiro | — |
| 14:57–14:58 | `modifying` | aparecem dois AGs com nome GUID, um por banco de usuário (`StressDB`, `LoadTestDB`) |
| 14:59 | `backing-up` | — |
| 15:02 | `available`, secundária em `us-east-1b` | 33 séries `sqlserver_hadr_*` |

- `sqlserver_up 1` o tempo todo, nenhum coletor com `success 0`, o rótulo
  e o `servername` (`EC2AMAZ-JFK9UC6`) não mudaram.
- O RDS Standard usa **Basic Availability Groups**: um AG por banco, com nome
  GUID (`762b6dfb-…`, `de928b35-…`), mais o `RDSAG0` para o `rdsadmin`. Nos
  dashboards o `ag_name` é um GUID; o `db_name` é o que identifica.
- O primário enxerga as duas réplicas: `replica_role` 1 para
  `EC2AMAZ-JFK9UC6` e 2 para `EC2AMAZ-94TRL9U` (a nova), `synchronization_health`
  2 nas duas, bancos `SYNCHRONIZED` (2), sem suspensão, fila de redo 0 e fila
  de envio de 60 KB.

**Failover (`aws rds reboot-db-instance --force-failover`, 15:15 UTC)**

Eventos do RDS: reinício 15:15:31, failover iniciado 15:15:35 e concluído
15:15:40, "DNS propagation may take a few minutes" 15:16:04.

| UTC (scrapes de 15 s) | `sqlserver_up` | Observação |
|---|---|---|
| até 15:15:01 | 1 | primária `EC2AMAZ-JFK9UC6` |
| 15:15:16–15:15:46 | 0 | `connection reset by peer` para o IP antigo (10.0.1.249) nas conexões do pool |
| a partir de 15:16:01 | 1 | `servername` = `EC2AMAZ-94TRL9U`, conexões novas já no IP novo |

- Indisponível por ≈45 s (3 scrapes); voltou sozinho, sem reiniciar o
  exporter: conexões novas resolvem o DNS do endpoint de novo.
- Nenhum coletor com `success 0` com o banco no ar; nos scrapes com
  `sqlserver_up 0` os coletores são pulados, como projetado.
- Papéis invertidos nos 3 AGs: `EC2AMAZ-94TRL9U` primária (1),
  `EC2AMAZ-JFK9UC6` secundária (2), saúde 2 nas duas, os 6 estados de banco
  `SYNCHRONIZED`.
- Com `instance_label` (HML-029) o rótulo não mudou; sem ele as séries
  passariam a ter `sqlserver_instance="EC2AMAZ-94TRL9U"` só depois de um
  restart do exporter (até lá ficariam com o nome da antiga primária).
- A API do RDS seguiu mostrando `us-east-1a`/`us-east-1b` como
  primária/secundária minutos depois do failover (atraso de atualização da
  própria AWS); o `servername` do exporter é a fonte mais rápida.

---

## HML-031 — Dashboard de demonstração para RDS: exporter + CloudWatch (validação)

- **Ambiente:** Grafana 13.1 local com Prometheus (exporter `0.10.0` no
  bastion) e CloudWatch (`us-east-1`), RDS `mssql-exporter-hml`
- **Encontrado em:** 2026-10-19
- **Severidade:** — (entrega pedida na fase RDS)
- **Status:** resolvido

**Entrega**

`dashboards/mssql-exporter-rds-cloudwatch-dashboard.json`: CloudWatch
(`AWS/RDS`) para o host e o exporter para a engine, com painéis que misturam
as duas fontes (CPU × batch requests, escrita em disco × escrita nos arquivos
do SQL Server, `DatabaseConnections` × user connections). Variáveis: região,
instância RDS (dimensão `DBInstanceIdentifier`) e `sqlserver_instance`; com
`instance_label` igual ao identificador do RDS (HML-029) os dois seletores
usam o mesmo nome.

**Validação**

- Todas as consultas dos 34 painéis executadas pela API do Grafana para as
  últimas 6 h: todas com dados, exceto `SwapUsage` (o CloudWatch não publica
  essa métrica para SQL Server — removida) e os painéis de Always On (vazios
  em Single-AZ; tinham dados durante o HML-030).
- Conferido visualmente no navegador: UP, host atual `EC2AMAZ-V7Q122H`, CPU
  ≈7%, memória livre 12,6 GB, storage livre 17,1 GB, último FULL há ≈30 min.
- Ajustes depois da primeira visualização: linhas quebradas em lacunas de
  mais de 10 min (o RDS ficou parado entre 15:35 e 17:58 UTC e as séries do
  CloudWatch eram ligadas por uma reta) e "Top waits" sem as esperas ociosas
  (`LAZYWRITER_SLEEP`, `CHECKPOINT_QUEUE`, `HADR_WORK_QUEUE`, …), que
  dominavam o painel; o exporter só filtra `SLEEP%`, `XE%` e `BROKER_%`.

---

## HML-032 — Linha de base no Cloud SQL: permissões, coletores e carga (validação)

- **Ambiente:** Cloud SQL for SQL Server 2022 Standard (`16.0.4262.2`,
  RTM-CU25-GDR) **em Linux** (Ubuntu 22.04), `db-custom-2-7680` (2 vCPUs,
  7,5 GB) zonal em `us-central1-a`, só IP privado (Private Service Access),
  `ssl_mode = ENCRYPTED_ONLY`, fuso UTC-3 (`E. South America Standard Time`);
  exporter `0.12.0` via systemd na bastion Ubuntu 26.04 (`e2-small`, sem IP
  público, SSH pelo IAP), lido pelo Prometheus local por túnel. Recursos em
  Terraform, fora do repositório
- **Encontrado em:** 2026-10-26
- **Severidade:** — (validação)
- **Status:** resolvido (README: permissões no Cloud SQL)

**Permissões**

Não existe `sysadmin`. O usuário `sqlserver` é membro da
`CustomerDbRootRole`, que tem `VIEW SERVER STATE` e `VIEW ANY DEFINITION`
`WITH GRANT OPTION`. Testes feitos como `sqlserver`, cada um numa transação
desfeita com `ROLLBACK`:

| Comando | Resultado |
|---|---|
| `GRANT VIEW SERVER STATE TO [usr_exporter]` | erro 4613: `Grantor does not have GRANT permission` |
| `GRANT VIEW SERVER STATE … AS [CustomerDbRootRole]` (em `master`) | OK |
| `GRANT VIEW ANY DEFINITION … AS [CustomerDbRootRole]` | OK |
| `ALTER SERVER ROLE [##MS_ServerStateReader##]` / `##MS_DefinitionReader##` / `##MS_DatabaseConnector##` `ADD MEMBER` | erro 15151: as roles existem, mas não aceitam membros |
| `GRANT CONNECT ANY DATABASE … AS [CustomerDbRootRole]` | erro 4613: usuário por banco, como no RDS (HML-024) |
| `CREATE USER` no `msdb` | erro 15247: o `msdb` é protegido |

- `backups` funciona sem grant: `guest` habilitado no `msdb` e `SELECT` em
  `backupset`/`backupmediafamily` para `public`.
- `agent_jobs`: `SELECT` em `sysjobs`/`sysjobhistory` só para o próprio
  `sqlserver`; o login mínimo recebe o erro 229 (`success 0`). Desligado no
  config. Logins criados pela API do Cloud SQL entram na `CustomerDbRootRole`
  (documentação do Google, não testado) e leriam os jobs, mas com privilégio
  demais (`ALTER ANY LOGIN`, `CREATE ANY DATABASE`). O Cloud SQL publica
  `sqlserver/sql_agent/jobs` no Cloud Monitoring.
- O `setup.sql` do `hack/loadgen` roda como `sqlserver` sem ajuste
  (`CREATE ANY DATABASE` e `ALTER ANY LOGIN`).

**Coletores**

- Primeira coleta: 16 dos 17 coletores padrão com `success 1`, só o
  `agent_jobs` com `success 0`; scrape de 0,78 s e 4 708 amostras.
- Primeiro ambiente da homologação com SQL Server em **Linux**: nenhuma query
  precisou de mudança.
- `@@SERVERNAME` = `MachineName` = `3da04acc8e845da`, nome gerado do host:
  `instance_label` recomendado, como no RDS (HML-029).
- `volumes`: `sys.dm_os_volume_stats` devolve `NULL` em
  `volume_mount_point`, `logical_volume_name` e `file_system_type`; os bytes
  vêm certos (disco de 20 GB), mas a série sai com `volume_mount_point=""`.
  Ponto a considerar: rótulo de fallback no coletor.
- Banco interno `gcloud_cloudsqladmin` em `databases`, `backups` e
  `log_used_percent`, como o `rdsadmin` no RDS. Em `sessions_by_login`
  aparecem as contas do Cloud SQL: `CloudDbSqlRoot`, `CloudDbSqlRoot_2`,
  `CloudDbSqlAgent` e `NT AUTHORITY\NETWORK SERVICE`.
- `backups`: FULL dos bancos de usuário e LOG a cada 5 min com o PITR ligado
  (gravados em `/var/opt/mssql/data/pitr/tmptlog/…`); com o servidor em UTC-3
  o LOG mais recente aparecia com 8 s de idade (HML-010 OK).
- `index_fragmentation` e `query_store` (opt-in) com usuário por banco:
  `success 1` em 2 bancos (`LoadTestDB` e `StressDB`), Query Store em
  `READ_WRITE` com limite de 1 GB.

**Carga (`hack/loadgen` na bastion, 10 logins × 2 workers, 19:25–19:44 UTC)**

- ≈56 700 statements sem erro no loadgen; ≈61 batch requests/s em média,
  pico de ≈101/s.
- Até 14 sessões bloqueadas (espera máxima de 28 s), 10 requests acima de
  10 s e transação aberta mais antiga de 29 s, todas na fila do `blocker`.
- Mediana do `perf_counters` em 0,64 s sob carga (0,13 s no RDS). A CPU da
  instância chegou a ≈98% (Cloud Monitoring, HML-035) com o loadgen e o teste
  do HML-034 juntos.

---

## HML-033 — TLS no Cloud SQL: `encrypt=true` com a CA da instância; `encrypt=strict` não é aceito

- **Ambiente:** o do HML-032; testes pelo túnel (Mac → bastion → Cloud SQL)
  com `hostnameincertificate` no lugar do host do túnel
- **Encontrado em:** 2026-10-26
- **Severidade:** baixa (limitação da plataforma; `encrypt=true` cobre o caso)
- **Status:** resolvido (README: seção de TLS)

Certificado do servidor: `CN=<projeto>:<instância>`, emitido pela
`Cloud SQL Server CA` da instância (validade de 10 anos), com o IP privado e
`<uuid>.us-central1.sql.goog` no SAN. A preocupação de o IP não estar no
certificado não se confirma no IP privado.

| DSN | Resultado |
|---|---|
| `encrypt=true&certificate=<CA>` pelo IP ou pelo nome `sql.goog` | OK, `encrypt_option` TRUE, TDS 7.4 |
| `encrypt=true&certificate=<CA>&hostnameincertificate=wrong.example.com` | recusado (`x509: certificate is valid for …`) |
| `encrypt=true` sem a CA (CAs do sistema) | recusado |
| `encrypt=strict` com a CA, com o nome ou com `trustservercertificate=true` | **falha** no handshake com `EOF` |
| `encrypt=disable` | recusado (`ENCRYPTED_ONLY`) |

O RDS aceitava `encrypt=strict` sem configuração (HML-027); o Cloud SQL não
aceita TDS 8.0. Exporter da bastion com `encrypt=true` e a CA em
`/etc/mssql-exporter/cloudsql-server-ca.pem`.

---

## HML-034 — Scrapes concorrentes saturam o pool de 5 conexões: coletores falham com `context deadline exceeded`

- **Ambiente:** o do HML-032, durante a carga do loadgen; `hey` na bastion
  contra o `/metrics` local, com o Prometheus raspando a cada 15 s
- **Encontrado em:** 2026-10-26
- **Severidade:** média (dados parciais e silenciosos: HTTP 200 com coletores
  faltando)
- **Status:** adiado (ROADMAP)

Cada scrape roda todos os coletores em paralelo sobre o pool fixo de 5
conexões, com um timeout único de 5 s (pendência da revisão de 2026-10-20).
Escada de concorrência, 100 requests por nível:

| Concorrência | p50 | p90 | p99 | Coletas com falha (Prometheus) |
|---|---|---|---|---|
| 1 | 0,55 s | 1,79 s | 2,18 s | 0 de 85 |
| 2 | 1,12 s | 2,22 s | 3,39 s | 0 de 68 |
| 5 | 4,57 s | 5,06 s | 5,22 s | 2 de 85 |
| 10 | 5,03 s | 5,07 s | 5,43 s | 14 de 68 (21%) |

- Até 2 scrapes simultâneos (um Prometheus ou um par em HA) não há falha.
- A partir de 5 os coletores esperam conexão, a mediana encosta nos 5 s e os
  coletores falham com `context deadline exceeded`; todas as respostas são
  HTTP 200 (`ContinueOnError`), só o `scrape_collector_success` mostra.
- A instância de 2 vCPUs estava com CPU em ≈98% no período, então parte da
  espera é do próprio SQL Server.
- O Cloud SQL também fez um `BACKUP_VOLUME` das 19:31 às 19:34 UTC, no início
  da escada, o que pode ter pesado nos primeiros níveis.

Correções candidatas (ROADMAP, Later): scrapes simultâneos aguardam a coleta
em andamento e recebem o mesmo resultado (singleflight), e tamanho do pool
configurável.

---

## HML-035 — Dashboard para Cloud SQL: exporter + Cloud Monitoring (validação)

- **Ambiente:** Grafana 13.1 local com Prometheus (exporter `0.12.0` na
  bastion) e o datasource Google Cloud Monitoring
- **Encontrado em:** 2026-10-26
- **Severidade:** — (entrega)
- **Status:** resolvido (PR #35)

**Entrega**

`dashboards/mssql-exporter-cloudsql-dashboard.json`, a versão Cloud SQL do
HML-031: Cloud Monitoring (`cloudsql.googleapis.com`) para o host, o exporter
para a engine, painéis de correlação e uma linha com os contadores que o
Cloud SQL também publica (batch requests, page life expectancy e deadlocks)
nas duas fontes; mais os painéis de log e Query Store da Fase 7.

**Autenticação sem chave**

A organização do projeto aplica `iam.disableServiceAccountKeyCreation`
(padrão do Google em organizações novas). O datasource usa uma service
account só com `roles/monitoring.viewer`, por impersonação: ADC gerado com
`gcloud auth application-default login --impersonate-service-account=…`
numa pasta separada do gcloud e disponível para o processo do Grafana, com a
autenticação "GCE Default Service Account". O botão de teste do datasource
responde `400` (procura o metadata server do GCE), mas as consultas
funcionam.

**Validação**

- Todas as consultas executadas pela API do Grafana na janela da carga
  (19:20–19:50 UTC): todas com dados, exceto Top queries (`top_queries`
  desligado) e as tabelas de Always On (sem réplicas de leitura).
- As duas fontes batem entre 19:27 e 19:42 UTC: batch requests com média de
  75,3/s no exporter (`rate[5m]`) e 81,4/s no Cloud Monitoring (alinhamento
  de 1 min); PLE de 4 888 s e 4 909 s.
- Conferido no navegador: variáveis resolvidas e painéis renderizados. Uma aba
  do Grafana aberta antes da criação do datasource mostra
  `Datasource … was not found` até ser recarregada.

---

## HML-036 — Conversão para HA regional e failover do Cloud SQL (validação)

- **Ambiente:** o do HML-032, com `availability_type` `ZONAL` → `REGIONAL`
  pelo Terraform; carga do `hack/loadgen` durante os failovers
- **Encontrado em:** 2026-10-26
- **Severidade:** — (validação)
- **Status:** resolvido

**Conversão para regional (22:13–22:23 UTC)**

A instância fica indisponível por ≈5 min 45 s (22:15:48–22:21:30): `Read:
EOF`, depois `i/o timeout` e `connection refused` no IP privado. Primária em
`us-central1-a`, standby em `us-central1-f`. Diferente do RDS, em que a
conversão para Multi-AZ não derrubou a instância (HML-030). O exporter não
reiniciou e voltou sozinho; o `query_store` registrou `WARN` e voltou na
coleta seguinte.

**Failovers (`gcloud sql instances failover`, com carga)**

| Failover | Comando | `sqlserver_up 0` | Observação |
|---|---|---|---|
| 1 (a → f) | 22:39:41–22:41:01 | ≈77 s (22:39:58–22:41:15) | servidor parou de responder sem fechar as conexões: 2 scrapes perdidos (HML-037) |
| 2 (f → a) | 22:51:24–22:52:45 | ≈60 s (22:52:04–22:53:15) | conexões derrubadas com `connection reset by peer`; nenhum scrape perdido |

- `@@SERVERNAME` continua `3da04acc8e845da` nas duas zonas: no Cloud SQL o
  nome acompanha a instância, não a VM (no RDS muda a cada failover,
  HML-029). `instance_label` segue recomendado, porque o nome é gerado.
- Nenhum AG aparece: o HA do Cloud SQL é por disco regional.
- O loadgen recebe `Neither the database "LoadTestDB" … could be opened`
  enquanto o banco faz recovery na nova primária (168 erros no 1º failover,
  1 284 no 2º) e continua depois sem intervenção.

---

## HML-037 — Scrape inteiro perdido quando o servidor para de responder sem fechar a conexão

- **Ambiente:** o do HML-036, no 1º failover; reproduzido na bastion com
  `iptables` descartando o tráfego para a porta 1433 por 45 s
- **Encontrado em:** 2026-10-26
- **Severidade:** média (a janela de indisponibilidade some: nem
  `sqlserver_up 0` é exposto)
- **Status:** resolvido (v0.12.1)

**Observado**

No 1º failover, dois scrapes seguidos estouraram o `scrape_timeout` de 10 s
(`up` 0, nenhuma amostra) e o exporter registrou:

```
sql server is unreachable error="Invalid TDS stream: did not get cancellation confirmation from the server (current response: context deadline exceeded)"
```

**Causa**

O ping estoura o prazo de 5 s, o driver envia o cancelamento e continua
esperando a confirmação de um servidor que não responde mais, sem respeitar o
contexto. O `Collect` esperava o ping (e os coletores) sem limite próprio,
então a resposta passava dos 10 s do Prometheus. Um alerta em
`sqlserver_up == 0` não dispara nessa janela; só um em `up == 0`.

**Correção** — branch `fix/scrape-deadline`, build de teste `0.12.1-homolog`

O `Collect` espera o ping e cada coletor só até o prazo mais 500 ms: um ping
que não voltou expõe `sqlserver_up 0`; um coletor que não voltou expõe
`success 0` e os demais continuam expostos; o que ele enviar depois vai para
um dreno, não para o canal do registry (já fechado). Testes com ping e coletor
que ignoram o contexto; teste de mutação: sem o limite, os dois falham.

**Validação (A/B na bastion, carga do loadgen, 45 s sem resposta do servidor)**

| Versão | Scrapes no bloqueio | `up` | Log |
|---|---|---|---|
| `0.12.0` | 10 s (timeout), 0 amostras, 2 seguidos | 0 | `did not get cancellation confirmation…` |
| `0.12.1-homolog` | 5,7 s, 50 amostras | 1 (`sqlserver_up 0`) | `ping did not return before the scrape deadline` |

Nas duas versões o exporter volta sozinho ≈50 s depois de o tráfego ser
liberado, quando o driver descarta as conexões presas.

---

## HML-038 — SQL Server Express sem Agent (validação local)

- **Ambiente:** SQL Server 2025 Express (`17.0.4025.3`) em Docker no Mac
  (container `mssql-2025-express`, imagem amd64 emulada), exporter `0.12.1`
  local, Prometheus local
- **Encontrado em:** 2026-10-26
- **Severidade:** — (validação)
- **Status:** resolvido

Feito localmente em vez do Cloud SQL: as diferenças do Express (sem Agent,
limites de CPU, memória e tamanho) são da edição, não da plataforma, e o
exporter não tem lógica por edição.

- Login com os grants do README (`VIEW SERVER STATE`, `VIEW ANY DEFINITION`,
  `CONNECT ANY DATABASE`, `SELECT` em `sysjobs`/`sysjobhistory`).
- Os 19 coletores (padrão + `top_queries`, `index_fragmentation` e
  `query_store`) com `success 1` desde a primeira coleta; nenhum `ERROR` ou
  `WARN` no log do exporter.
- `agent_jobs`: `SQL Server Agent` em `Stopped` e `msdb.dbo.sysjobs` vazio;
  o coletor reporta `success 1` sem séries.
- `sqlserver_edition{edition="Express Edition (64-bit)"}`;
  `schedulers_online` = 4 com 8 CPUs no container (limite de 4 núcleos do
  Express).
- `volumes`: `volume_mount_point=""` também aqui — o `NULL` do
  `sys.dm_os_volume_stats` é do SQL Server em Linux, não do Cloud SQL
  (HML-032).
- Carga do `hack/loadgen` (10 logins × 2 workers, 10 min): ≈45 200 statements
  sem erro, média de 78 batch requests/s (pico 118), até 16 sessões
  bloqueadas e 16 requests acima de 10 s; scrape máximo de 0,23 s, nenhum
  coletor com `success 0`; `top_queries` com 50 queries.

---

## HML-039 — Linha de base no SQL Managed Instance: permissões, coletores e carga (validação)

- **Ambiente:** Azure SQL Managed Instance pelo free offer (General Purpose
  Gen5, 4 vCores, 64 GB, formato `SQLServer2022`) em `westus3`, só endpoint
  privado, fuso UTC-3; host Windows Server 2025 com 22 GB e 4 CPUs visíveis.
  Exporter via systemd na bastion Ubuntu 26.04 da mesma VNet, DSN com
  `encrypt=true`, lido pelo Prometheus local por túnel. Recursos em
  Terraform, fora do repositório
- **Encontrado em:** 2026-10-27
- **Severidade:** — (validação)
- **Status:** resolvido

**Permissões**

O admin SQL da instância (`sqladmin`) é `sysadmin`, então os grants da EC2
funcionam sem adaptação, inclusive no `msdb`:

```sql
CREATE LOGIN [usr_exporter] WITH PASSWORD = '…', CHECK_POLICY = OFF, CHECK_EXPIRATION = OFF;
GRANT VIEW SERVER STATE TO [usr_exporter];
GRANT VIEW ANY DEFINITION TO [usr_exporter];
USE msdb;
CREATE USER usr_exporter FOR LOGIN usr_exporter;
GRANT SELECT ON sysjobs TO usr_exporter;
GRANT SELECT ON sysjobhistory TO usr_exporter;
```

Sem os grants do `msdb`, o `agent_jobs` fica com `success 0`, como na EC2
(HML-002).

**Plataforma**

- `SERVERPROPERTY`: `EngineEdition` 8, `Edition` `SQL Azure`,
  `ProductVersion` `12.0.2000.8` (fixo, não segue a versão do motor);
  `@@SERVERNAME` é o FQDN da instância; `MachineName`,
  `ComputerNamePhysicalNetBIOS` e `InstanceName` são NULL (HML-042).
- Bancos de sistema e de usuário ficam no blob storage remoto
  (`physical_name` `https://…blob.core.windows.net/…` no `io_file_stats`); só
  o `tempdb` fica no SSD local (`C:\`, 12 arquivos de 16 MB). Ver HML-043.
- `sessions_by_login` mostra as contas do serviço: `NT AUTHORITY\SYSTEM` e
  `User Manager\ContainerAdministrator`.
- O `availability_groups` não mostra nada (o HA do General Purpose não aparece
  como AG) e o `agent_jobs` não tem séries (nenhum job criado).
- `wait_stats`: 1 499 tipos de espera expostos. No topo, por tempo
  acumulado, só esperas de fundo do serviço: `SOS_WORK_DISPATCHER`,
  `HADR_WORK_QUEUE`, `DISPATCHER_QUEUE_SEMAPHORE` e `LOGMGR_QUEUE`; a
  primeira espera de carga é `LCK_M_X`. Reforça o filtro de esperas ociosas
  do ROADMAP (HML-031).

**Coletores**

Na primeira coleta, 14 dos 16 coletores padrão estavam em `success 1`. As
falhas eram o `server_info` (HML-042) e o `agent_jobs`, que rodou antes dos
grants do `msdb`. Depois dos grants e das correções, os 16 com `success 1`.
`volumes`, `log_used_percent`, `backups` e `requests_long_running` precisaram
de análise própria (HML-043 a HML-046).

**Carga (`hack/loadgen` na bastion, 10 logins × 2 workers, 30 min, 14:55–15:25 UTC)**

- ≈115 300 statements sem erro no loadgen; ≈67 batch requests/s em média.
- Até 16 sessões bloqueadas (espera máxima de 28,6 s) e transação aberta mais
  antiga de 35 s, na fila do `blocker`; nenhum deadlock.
- Scrape máximo de 1,31 s; os coletores mais lentos foram o `perf_counters`
  (0,77 s) e o `tempdb` (0,46 s); nenhum coletor com `success 0`.
- Latência de escrita do `LoadTestDB`: 33 ms nos dados e 4,3 ms no log, por
  causa do storage remoto do General Purpose.
- `log_reuse_wait` do `LoadTestDB` em `OLDEST_PAGE` durante a carga.

---

## HML-040 — Free Trial da Azure: cota de Managed Instance só em `westus3`, sem VMs série B e 404 ao reaproveitar o nome do resource group (ambiente)

- **Ambiente:** assinatura Free Trial (`FreeTrial_2014-09-01`), Terraform com
  `azurerm` 5.11 e `azapi` 2.13
- **Encontrado em:** 2026-10-27
- **Severidade:** — (ambiente)
- **Status:** resolvido

- **Cota de MI:** 0 vCores em `eastus`, `eastus2`, `centralus`,
  `southcentralus`, `westus2`, `brazilsouth`, `canadacentral`,
  `northeurope` e `westeurope`; 16 em `westus3`. O ambiente foi para
  `westus3`.
- **VM da bastion:** a assinatura não oferece a série B (nem a B1s gratuita);
  a mais barata disponível foi a `Standard_D2s_v4` (US$ 0,096/h), com
  desligamento automático às 21h10.
- **Resource providers:** a primeira execução falha até Compute,
  Consumption, DevTestLab, Network e Sql serem registrados na assinatura.
- **Free offer do MI:** o `azurerm` 5.11 não expõe o free offer; o MI foi
  criado pelo `azapi` com `pricingModel = "Freemium"`
  (`Microsoft.Sql/managedInstances@2025-01-01`), junto com a agenda de
  start/stop. Criação em 6 min 14 s, com a sub-rede nova.
- **Nome do resource group:** recriar o grupo com o mesmo nome em outra
  região (de `eastus2` para `westus3`) fez o ARM responder 404 para recursos
  que acabavam de ser criados (`Root object was present, but now absent`):
  NSG, NIC e VNet sumiam do state de forma intermitente. Resolvido com a
  região no nome (`mssql-exporter-hml-westus3`). As sub-redes também
  precisaram de `replace_triggered_by` na VNet, porque o ID delas não muda
  quando a VNet é recriada.

---

## HML-041 — SQL pelo túnel SSH não chega ao MI: o gateway exige o nome da instância e redireciona para o nó (ambiente)

- **Ambiente:** o do HML-039; Mac → túnel SSH na bastion
  (`-L 14332:<fqdn>:1433`) → endpoint privado do MI
- **Encontrado em:** 2026-10-27
- **Severidade:** baixa (afeta só o acesso ad hoc; o exporter na VNet
  conecta normalmente)
- **Status:** resolvido (T-SQL pela bastion)

- Conectando em `localhost:14332`, o gateway do MI recusa o login com o
  erro 40532, porque não sabe para qual instância rotear.
- Com o login no formato `usuario@instancia`, o gateway aceita, mas
  responde com um redirecionamento para o IP do nó na sub-rede do MI
  (`10.20.1.x:1433`), que o Mac não alcança pelo túnel.
- O exporter na bastion não tem o problema: ele resolve o FQDN e segue o
  redirecionamento dentro da VNet.
- T-SQL de teste e de administração roda na bastion com `go-sqlcmd`, a
  partir de arquivos `.sql` copiados por `scp` (`sqlcmd -i`).

---

## HML-042 — `server_info` falha inteiro com `MachineName` NULL no MI

- **Ambiente:** o do HML-039, exporter `0.12.1`
- **Encontrado em:** 2026-10-27
- **Severidade:** média (perde edição, versão e `engine_edition` da
  instância)
- **Status:** resolvido (v0.12.2)

No MI, `SERVERPROPERTY('MachineName')`, `ComputerNamePhysicalNetBIOS` e
`InstanceName` são NULL. A linha de propriedades era lida em strings do Go,
e um único NULL derrubava a linha inteira:

```
ERROR collector scrape failed collector=server_info error="failed to query server properties: sql: Scan error on column index 0, name \"hostname\": converting NULL to string is unsupported"
```

`sqlserver_servername`, `sqlserver_edition`, `sqlserver_product_version`,
`sqlserver_engine_edition` e `sqlserver_is_hadr_enabled` sumiam, e o coletor
ficava em `success 0` em todas as coletas. CPU e memória continuavam, porque
vêm de outras queries.

**Correção:** propriedades lidas como `sql.NullString`/`sql.NullInt64`; só a
métrica da propriedade NULL é omitida (no MI, só `sqlserver_hostname`).
Teste com uma linha igual à do MI; README documenta o comportamento.

---

## HML-043 — `volumes` no MI: `total_bytes` é o tamanho dos arquivos, não do volume (disponível acima de 100%)

- **Ambiente:** o do HML-039, exporter `0.12.2`
- **Encontrado em:** 2026-10-27
- **Severidade:** média (painel e alerta de disco sem sentido no MI)
- **Status:** resolvido (v0.13.0)

No MI, `sys.dm_os_volume_stats` não descreve o volume: `total_bytes` é a
soma dos arquivos no volume e fica menor que `available_bytes`.

| Volume | O que é | `total_bytes` | `available_bytes` |
|---|---|---|---|
| `http://` | blob storage remoto (bancos de sistema e de usuário) | 304 MB | 63,7 GB |
| `C:\` | SSD local (só o `tempdb`) | 192 MB | 95,8 GB |

O painel "Volume Space Available (%)" mostrava ≈21 000% e ≈51 000%, e um
alerta de pouco espaço montado nessas séries nunca dispararia.

**Correção:** o `volumes` consulta `SERVERPROPERTY('EngineEdition')`; no MI
(8) não emite as séries de volume e reporta o storage da instância pela
amostra mais recente de `master.sys.server_resource_stats` (coberta pelo
`VIEW SERVER STATE`): `sqlserver_instance_storage_reserved_bytes` (64 GiB) e
`sqlserver_instance_storage_used_bytes` (336 MiB na validação; 99,49%
disponível no novo painel). A view é publicada com alguns minutos de atraso e
vem vazia logo depois da criação da instância; sem linha, nada é emitido.
Dashboard com "Instance Storage Available (%)" e "Instance Storage Used vs
Reserved". Nas outras plataformas, só uma query de `SERVERPROPERTY` a mais
por scrape (validado no SQL Server 2025 em Docker).

---

## HML-044 — `database_log_used_percent` sem os bancos de usuário no MI (contador pelo nome físico)

- **Ambiente:** o do HML-039, exporter `0.13.0`, durante a carga
- **Encontrado em:** 2026-10-27
- **Severidade:** baixa (os bancos de sistema continuam; o uso do log dos
  bancos de usuário some)
- **Status:** resolvido (v0.13.1)

No MI, a instância do contador `Percent Log Used` de um banco de usuário é o
`physical_database_name` (um GUID), não o nome: o `LoadTestDB` aparece como
`17f319d1-1099-430f-8903-7c77b206bf6b`. O join do `databases` pelo nome
descartava todos os bancos de usuário. O contador também tem instâncias para
os nomes físicos de `master` (`replicatedmaster`) e `model`
(`managed_model`), com valores diferentes dos de `master` e `model`.

**Correção:** só no MI, o contador é casado pelo nome ou pelo
`physical_database_name`, com preferência pelo nome (`CROSS APPLY … TOP (1)`),
para `master` e `model` não pegarem o valor da instância física. Nas outras
plataformas a query não muda; um teste garante que ela não usa
`physical_database_name`, que só existe a partir do SQL Server 2019. A
consulta de `EngineEdition` virou um helper compartilhado com o `volumes`.
Validado no MI (`LoadTestDB` 1, `master` 44, `model` 19) e no SQL Server
2025 em Docker (séries iguais às de antes).

---

## HML-045 — `backups` sem nenhuma série no MI: `server_name` NULL no `backupset`

- **Ambiente:** o do HML-039, exporter `0.13.1`
- **Encontrado em:** 2026-10-27
- **Severidade:** média (sem `sqlserver_backup_*` não há alerta de backup
  atrasado no MI)
- **Status:** aberto

Os backups automáticos do MI aparecem no `msdb.dbo.backupset` (FULL do
`LoadTestDB` na criação e LOG a cada ≈5 min durante a carga, depois a cada
≈1 h), com `time_zone` -12 (UTC-3) correto. Mas `server_name` e
`machine_name` vêm NULL, e `physical_device_name` também (`device_type` 9,
URL). O coletor filtra `server_name = @@SERVERNAME` para não reportar backups
restaurados de outro servidor (HML-012), então descarta todas as linhas e
fica com `success 1` sem séries. Os bancos de sistema não têm histórico no
`msdb`.

Correção candidata: só no MI (`EngineEdition` 8), aceitar `server_name` NULL,
como nas correções do HML-043 e do HML-044.

---

## HML-046 — `requests_long_running` sempre com 1 em todas as faixas no MI: sessão de telemetria `TdService`

- **Ambiente:** o do HML-039, exporter `0.13.1`, sem carga
- **Encontrado em:** 2026-10-27
- **Severidade:** baixa (um alerta em
  `sqlserver_requests_long_running{over_seconds="1800"} > 0` dispararia para
  sempre no MI)
- **Status:** aberto

O `active_queries` já exclui as sessões de sistema (`is_user_process = 0`,
98 requests no MI). Sobra uma sessão do serviço marcada como de usuário:
`program_name` `TdService`, login `NT AUTHORITY\SYSTEM`, um `SELECT` em
`suspended` com espera `XE_LIVE_TARGET_TVF` (leitura contínua de uma sessão
de Extended Events), aberto desde a subida da instância (15 292 s na
consulta). Ela conta 1 nas faixas de 10, 60, 300 e 1 800 s o tempo todo.

Correção candidata: só no MI, ignorar essa sessão (`program_name =
'TdService'` ou espera `XE_LIVE_TARGET_TVF`).
