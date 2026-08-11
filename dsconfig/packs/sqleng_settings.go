package packs

import (
	_ "embed"

	"github.com/grafana/dsconfig/dsconfig"
)

// sqleng_settings.json defines the common fields shared by the SQL engine
// (`sqleng`) used by grafana-postgresql-datasource, grafana-mysql-datasource,
// and grafana-mssql-datasource: host URL, database, username/password, TLS
// certificates, connection pool tuning, and Secure Socks Proxy settings.
// Field IDs are namespaced with the "sqleng_settings." prefix.
//
//go:embed sqleng_settings.json
var sqlengSettingsJSON []byte

func init() {
	mustLoadPack(dsconfig.PackSqlengSettings, sqlengSettingsJSON)
}
