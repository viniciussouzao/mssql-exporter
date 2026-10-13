# Backlog de Homologação

Pontos encontrados durante a homologação do mssql-exporter em ambientes
simulando produção (AWS EC2 Windows/Linux e RDS for SQL Server).

**Status:** `aberto` · `em andamento` · `resolvido` · `não será feito`
**Severidade:** `alta` (bloqueia uso) · `média` (atrapalha operação) · `baixa` (cosmético/ruído)

| ID | Título | Ambiente | Severidade | Status |
|----|--------|----------|------------|--------|
| HML-001 | Log `ERROR` quando `.env` não existe | EC2 Windows | baixa | em andamento |
| HML-002 | `agent_jobs` falha sem permissão no msdb (ponto de atenção) | EC2 Windows | média | em andamento |
| HML-003 | Coleta de disk I/O no Windows falha com `Invalid class` | EC2 Windows | média | em andamento |
| HML-004 | Gaps intermitentes: `/metrics` retorna HTTP 500 por série duplicada em `volumes` | EC2 Windows | alta | em andamento (validado na EC2, aguardando merge) |
| HML-005 | Disco C: da EC2 quase cheio durante carga no `StressDB` (ambiente) | EC2 Windows | alta | resolvido |
| HML-006 | Exporter não roda como serviço Windows nativo (ponto de atenção) | EC2 Windows | média | em andamento (v0.2.0 instalada; falta reboot/recovery) |
| HML-007 | Validação do coletor `availability_groups` em AG real (validação) | EC2 Windows (AG 01+02) | — | em andamento |
| HML-008 | Fila de envio some do gráfico com a réplica suspensa; faltam `is_suspended` e `secondary_lag_seconds` | EC2 Windows (AG 01+02) | média | em andamento |

---

## HML-001 — Log `ERROR` quando `.env` não existe

- **Ambiente:** EC2 Windows, release v0.1.0 (`mssql-exporter_0.1.0_windows_amd64.zip`)
- **Encontrado em:** 2026-10-08
- **Severidade:** baixa
- **Status:** aberto

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
- **Status:** em andamento (resolvido no ambiente; pendências abaixo)

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
- **Status:** aberto

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
- **Status:** em andamento

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
- **Status:** em andamento

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
- [ ] Reboot da EC2 → target volta sozinho.
- [ ] `Stop-Process -Force` → SCM reinicia em ~5 s.
- [ ] `service stop` / `uninstall` sem processo órfão.
- [ ] Event Log com as entradas do exporter.

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
- **Status:** em andamento

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
- **Status:** em andamento

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
