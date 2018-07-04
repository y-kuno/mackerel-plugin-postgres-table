package mppostgrestable

import (
	"flag"
	"os"

	"fmt"
	"github.com/jmoiron/sqlx"
	// PostgreSQL Driver
	_ "github.com/lib/pq"
	mp "github.com/mackerelio/go-mackerel-plugin"
	"strings"
)

// PostgresTablePlugin mackerel plugin
type PostgresTablePlugin struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	Option   string
	SSLmode  string
	Timeout  int
	Prefix   string
}

// StatTable postgres stat table struct
type StatTable struct {
	RelName          string `db:"relname"`
	SeqScan          uint64 `db:"seq_scan"`
	SeqTupRead       uint64 `db:"seq_tup_read"`
	IdxScan          uint64 `db:"idx_scan"`
	IdxTupFetch      uint64 `db:"idx_tup_fetch"`
	NTupIns          uint64 `db:"n_tup_ins"`
	NTupUpd          uint64 `db:"n_tup_upd"`
	NTupDel          uint64 `db:"n_tup_del"`
	NTupHotUpd       uint64 `db:"n_tup_hot_upd"`
	NLiveTup         uint64 `db:"n_live_tup"`
	NDeadTup         uint64 `db:"n_dead_tup"`
	VacuumCount      uint64 `db:"vacuum_count"`
	AutoVacuumCount  uint64 `db:"autovacuum_count"`
	AnalyzeCount     uint64 `db:"analyze_count"`
	AutoAnalyzeCount uint64 `db:"autoanalyze_count"`
}

// MetricKeyPrefix interface for PluginWithPrefix
func (p *PostgresTablePlugin) MetricKeyPrefix() string {
	if p.Prefix == "" {
		p.Prefix = "postgres"
	}
	return p.Prefix
}

// GraphDefinition interface for mackerelplugin
func (p *PostgresTablePlugin) GraphDefinition() map[string]mp.Graphs {
	labelPrefix := strings.Title(p.Prefix)
	return map[string]mp.Graphs{
		"table.scan.#": {
			Label: labelPrefix + " Table Scans",
			Unit:  "integer",
			Metrics: []mp.Metrics{
				{Name: "seq_scan", Label: "Sequential Scans", Diff: true},
				{Name: "seq_tup_read", Label: "Rows Fetched by Sequential Scan", Diff: true},
				{Name: "idx_scan", Label: "Index Scans", Diff: true},
				{Name: "idx_tup_fetch", Label: "Rows Fetched by Index Scan", Diff: true},
			},
		},
		"table.row.#": {
			Label: labelPrefix + " Table Rows",
			Unit:  "integer",
			Metrics: []mp.Metrics{
				{Name: "n_tup_ins", Label: "Inserted Rows", Diff: true},
				{Name: "n_tup_upd", Label: "Updated Rows", Diff: true},
				{Name: "n_tup_del", Label: "Deleted Rows", Diff: true},
				{Name: "n_tup_hot_upd", Label: "HOT Updated Rows", Diff: true},
				{Name: "n_live_tup", Label: "Estimated Live Rows"},
				{Name: "n_dead_tup", Label: "Estimated Dead Rows"},
			},
		},
		"table.vacuum.#": {
			Label: labelPrefix + " Table Vacuume Counts",
			Unit:  "integer",
			Metrics: []mp.Metrics{
				{Name: "vacuum_count", Label: "Vacuumd Counts", Diff: true},
				{Name: "autovacuum_count", Label: "Auto Vacuume Counts", Diff: true},
			},
		},
		"table.analyze.#": {
			Label: labelPrefix + " Table Analyze Counts",
			Unit:  "integer",
			Metrics: []mp.Metrics{
				{Name: "analyze_count", Label: "Analyze Counts", Diff: true},
				{Name: "autoanalyze_count", Label: "Auto Analyzed Counts", Diff: true},
			},
		},
	}
}

// FetchMetrics interface for mackerelplugin
func (p *PostgresTablePlugin) FetchMetrics() (map[string]float64, error) {
	db, err := sqlx.Connect("postgres",
		fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=%s connect_timeout=%d", p.User, p.Password, p.Host, p.Port, p.Database, p.SSLmode, p.Timeout))
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var query = "SELECT * FROM pg_stat_user_tables"
	if p.Option != "" {
		query = fmt.Sprintf("%s %s", query, p.Option)
	}

	db = db.Unsafe()
	rows, err := db.Queryx(query)
	if err != nil {
		return nil, err
	}

	metrics := make(map[string]float64)
	for rows.Next() {
		var stat StatTable
		if err := rows.StructScan(&stat); err != nil {
			return nil, err
		}
		// scan
		metrics["table.scan."+stat.RelName+".seq_scan"] = float64(stat.SeqScan)
		metrics["table.scan."+stat.RelName+".seq_tup_read"] = float64(stat.SeqTupRead)
		metrics["table.scan."+stat.RelName+".idx_scan"] = float64(stat.IdxScan)
		metrics["table.scan."+stat.RelName+".idx_tup_fetch"] = float64(stat.IdxTupFetch)
		// row
		metrics["table.row."+stat.RelName+".n_tup_ins"] = float64(stat.NTupIns)
		metrics["table.row."+stat.RelName+".n_tup_upd"] = float64(stat.NTupUpd)
		metrics["table.row."+stat.RelName+".n_tup_del"] = float64(stat.NTupDel)
		metrics["table.row."+stat.RelName+".n_tup_hot_upd"] = float64(stat.NTupHotUpd)
		metrics["table.row."+stat.RelName+".n_live_tup"] = float64(stat.NLiveTup)
		metrics["table.row."+stat.RelName+".n_dead_tup"] = float64(stat.NDeadTup)
		// vacuum
		metrics["table.vacuum."+stat.RelName+".vacuum_count"] = float64(stat.VacuumCount)
		metrics["table.vacuum."+stat.RelName+".autovacuum_count"] = float64(stat.AutoVacuumCount)
		// analyze
		metrics["table.analyze."+stat.RelName+".analyze_count"] = float64(stat.AnalyzeCount)
		metrics["table.analyze."+stat.RelName+".autoanalyze_count"] = float64(stat.AutoAnalyzeCount)
	}
	return metrics, nil
}

// Do the plugin
func Do() {
	optHost := flag.String("host", "localhost", "Hostname")
	optPort := flag.String("port", "5432", "Port")
	optUser := flag.String("user", "postgres", "Username")
	optPassword := flag.String("password", os.Getenv("PGPASSEORD"), "Password")
	optDatabase := flag.String("database", "", "Database")
	optOption := flag.String("option", "", "Query option")
	optSSLmode := flag.String("sslmode", "disable", "Whether or not to use SSL")
	optConnectTimeout := flag.Int("connect-timeout", 5, "Maximum wait for connection, in seconds.")
	optPrefix := flag.String("metric-key-prefix", "postgres", "Metric key prefix")
	optTempfile := flag.String("tempfile", "", "Temp file name")
	flag.Parse()

	plugin := mp.NewMackerelPlugin(&PostgresTablePlugin{
		Host:     *optHost,
		Port:     *optPort,
		User:     *optUser,
		Password: *optPassword,
		Database: *optDatabase,
		Option:   *optOption,
		SSLmode:  *optSSLmode,
		Timeout:  *optConnectTimeout,
		Prefix:   *optPrefix,
	})
	plugin.Tempfile = *optTempfile
	plugin.Run()
}
