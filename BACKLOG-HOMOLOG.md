# Backlog de Homologação

Pontos encontrados durante a homologação do mssql-exporter em ambientes
simulando produção (AWS EC2 Windows/Linux e RDS for SQL Server).

**Status:** `aberto` · `em andamento` · `resolvido` · `não será feito`
**Severidade:** `alta` (bloqueia uso) · `média` (atrapalha operação) · `baixa` (cosmético/ruído)

| ID | Título | Ambiente | Severidade | Status |
|----|--------|----------|------------|--------|
| HML-001 | Log `ERROR` quando `.env` não existe | EC2 Windows | baixa | resolvido (v0.1.1) |
| HML-002 | `agent_jobs` falha sem permissão no msdb (ponto de atenção) | EC2 Windows | média | resolvido na EC2 (repetir no RDS) |
| HML-003 | Coleta de disk I/O no Windows falha com `Invalid class` | EC2 Windows | média | resolvido (v0.1.1) |
| HML-004 | Gaps intermitentes: `/metrics` retorna HTTP 500 por série duplicada em `volumes` | EC2 Windows | alta | resolvido (v0.1.1) |
| HML-005 | Disco C: da EC2 quase cheio durante carga no `StressDB` (ambiente) | EC2 Windows | alta | resolvido |
| HML-006 | Exporter não roda como serviço Windows nativo (ponto de atenção) | EC2 Windows | média | resolvido (v0.2.0; `uninstall` no fim da homologação) |
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
| HML-019 | `volumes` estoura o timeout de query quando um banco está trocando de estado | EC2 Windows (01+02) | baixa | aberto |
| HML-020 | `blocking` falha com erro 983 logo após o SQL Server subir (réplica do AG em `RESOLVING`) | EC2 Windows (02) | baixa | resolvido (v0.5.0) |

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
- **Status:** resolvido (v0.2.0; `uninstall` no fim da homologação)

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
- **Status:** aberto

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
