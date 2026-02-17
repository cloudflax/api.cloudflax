package secrets

// DBCredentials holds database connection data as returned by Secrets Manager.
// JSON format: {"dbname":"...","host":"...","password":"...","port":4510,"username":"..."}
type DBCredentials struct {
	DBName   string `json:"dbname"`
	Host     string `json:"host"`
	Password string `json:"password"`
	Port     int    `json:"port"`
	Username string `json:"username"`
}
