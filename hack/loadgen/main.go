// Command loadgen generates a mixed SQL Server workload from several logins,
// to exercise the exporter during homologation (sessions per login, active
// and long-running requests, blocking). It is not part of the exporter.
//
// Create the database and logins with setup.sql, then:
//
//	LOADGEN_PASSWORD=... go run ./hack/loadgen -server 10.0.0.1 -duration 10m
package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand/v2"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	_ "github.com/microsoft/go-mssqldb"
)

// workload is one kind of statement, picked at random by weight.
type workload struct {
	name   string
	weight int
	run    func(ctx context.Context, db *sql.DB) error
}

// hotRows are updated by "hot_update" and held by "blocker", so the two
// collide and produce real blocking.
const hotRows = 5

var workloads = []workload{
	{"point_lookup", 50, func(ctx context.Context, db *sql.DB) error {
		return queryDiscard(ctx, db, `SELECT id, amount, status FROM dbo.Orders WHERE id = @p1`, rand.IntN(200000)+1)
	}},
	{"customer_orders", 20, func(ctx context.Context, db *sql.DB) error {
		return queryDiscard(ctx, db, `SELECT TOP (50) id, amount FROM dbo.Orders WHERE customer_id = @p1 ORDER BY created_at DESC`, rand.IntN(5000))
	}},
	{"status_report", 10, func(ctx context.Context, db *sql.DB) error {
		// Clustered index scan: the "heavier" query of the mix.
		return queryDiscard(ctx, db, `SELECT status, COUNT(*), SUM(amount) FROM dbo.Orders WHERE created_at >= DATEADD(DAY, -@p1, SYSUTCDATETIME()) GROUP BY status`, rand.IntN(365)+1)
	}},
	{"update_order", 10, func(ctx context.Context, db *sql.DB) error {
		_, err := db.ExecContext(ctx, `UPDATE dbo.Orders SET amount = amount + 1 WHERE id = @p1`, rand.IntN(200000-hotRows)+hotRows+1)
		return err
	}},
	{"hot_update", 5, func(ctx context.Context, db *sql.DB) error {
		_, err := db.ExecContext(ctx, `UPDATE dbo.Orders SET status = (status + 1) % 5 WHERE id = @p1`, rand.IntN(hotRows)+1)
		return err
	}},
	{"long_wait", 3, func(ctx context.Context, db *sql.DB) error {
		// A request that stays active for a while without using resources.
		_, err := db.ExecContext(ctx, `WAITFOR DELAY '00:00:30'`)
		return err
	}},
	{"blocker", 2, func(ctx context.Context, db *sql.DB) error {
		// Holds a hot row locked for 20s; hot_update sessions queue behind it.
		_, err := db.ExecContext(ctx, `
			BEGIN TRANSACTION;
			UPDATE dbo.Orders SET note = 'blocked by loadgen' WHERE id = @p1;
			WAITFOR DELAY '00:00:20';
			COMMIT;`, rand.IntN(hotRows)+1)
		return err
	}},
}

func queryDiscard(ctx context.Context, db *sql.DB, query string, args ...any) error {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
	}
	return rows.Err()
}

func pick() workload {
	total := 0
	for _, w := range workloads {
		total += w.weight
	}
	n := rand.IntN(total)
	for _, w := range workloads {
		if n < w.weight {
			return w
		}
		n -= w.weight
	}
	return workloads[0]
}

type stats struct {
	ok, failed sync.Map // workload name -> *atomic.Int64
}

func (s *stats) add(m *sync.Map, name string) {
	v, _ := m.LoadOrStore(name, new(atomic.Int64))
	v.(*atomic.Int64).Add(1)
}

func (s *stats) report() {
	line := ""
	for _, w := range workloads {
		var ok, failed int64
		if v, found := s.ok.Load(w.name); found {
			ok = v.(*atomic.Int64).Load()
		}
		if v, found := s.failed.Load(w.name); found {
			failed = v.(*atomic.Int64).Load()
		}
		line += fmt.Sprintf(" %s=%d", w.name, ok)
		if failed > 0 {
			line += fmt.Sprintf("(!%d)", failed)
		}
	}
	log.Printf("executed:%s", line)
}

func main() {
	server := flag.String("server", "localhost", "SQL Server host")
	port := flag.Int("port", 1433, "SQL Server port")
	database := flag.String("database", "LoadTestDB", "database created by setup.sql")
	logins := flag.Int("logins", 10, "number of loadgen_NN logins to use")
	workers := flag.Int("workers", 2, "concurrent workers per login")
	duration := flag.Duration("duration", 10*time.Minute, "how long to run (0 = until Ctrl+C)")
	pause := flag.Duration("pause", 200*time.Millisecond, "maximum random pause between statements")
	params := flag.String("params", "", "extra DSN query parameters, e.g. encrypt=true&trustservercertificate=true")
	flag.Parse()

	password := os.Getenv("LOADGEN_PASSWORD")
	if password == "" {
		log.Fatal("set LOADGEN_PASSWORD to the password used in setup.sql")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if *duration > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *duration)
		defer cancel()
	}

	extra, err := url.ParseQuery(*params)
	if err != nil {
		log.Fatalf("invalid -params: %v", err)
	}

	var st stats
	var wg sync.WaitGroup
	for i := 1; i <= *logins; i++ {
		login := fmt.Sprintf("loadgen_%02d", i)
		q := url.Values{"database": {*database}, "app name": {"loadgen"}}
		for k, v := range extra {
			q[k] = v
		}
		dsn := (&url.URL{
			Scheme:   "sqlserver",
			User:     url.UserPassword(login, password),
			Host:     fmt.Sprintf("%s:%d", *server, *port),
			RawQuery: q.Encode(),
		}).String()

		db, err := sql.Open("sqlserver", dsn)
		if err != nil {
			log.Fatalf("%s: %v", login, err)
		}
		defer func() { _ = db.Close() }()
		db.SetMaxOpenConns(*workers)

		pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		err = db.PingContext(pingCtx)
		cancel()
		if err != nil {
			log.Fatalf("%s: cannot connect: %v", login, err)
		}

		for range *workers {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for ctx.Err() == nil {
					w := pick()
					if err := w.run(ctx, db); err != nil {
						if ctx.Err() == nil && !errors.Is(err, context.Canceled) {
							st.add(&st.failed, w.name)
							log.Printf("%s %s: %v", login, w.name, err)
						}
					} else {
						st.add(&st.ok, w.name)
					}
					select {
					case <-ctx.Done():
					case <-time.After(time.Duration(rand.Int64N(int64(*pause) + 1))):
					}
				}
			}()
		}
	}
	log.Printf("running %d logins x %d workers against %s:%d/%s", *logins, *workers, *server, *port, *database)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	for {
		select {
		case <-ticker.C:
			st.report()
		case <-done:
			st.report()
			log.Print("done")
			return
		}
	}
}
