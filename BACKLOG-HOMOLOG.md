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
