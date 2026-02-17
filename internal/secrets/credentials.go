package secrets

import "encoding/json"

// DBCredentials holds database connection data as returned by Secrets Manager.
// NOTE: never log this struct directly, to avoid leaking secrets.
type DBCredentials struct {
	dbName   string
	host     string
	password string
	port     int
	username string
}

// dbCredentialsWire is the JSON representation of the secret.
// It is kept internal to avoid exposing raw fields outside this package.
type dbCredentialsWire struct {
	DBName   string `json:"dbname"`
	Host     string `json:"host"`
	Password string `json:"password"`
	Port     int    `json:"port"`
	Username string `json:"username"`
}

// UnmarshalJSON parses the JSON secret string into the internal fields.
func (c *DBCredentials) UnmarshalJSON(data []byte) error {
	var wire dbCredentialsWire
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	c.dbName = wire.DBName
	c.host = wire.Host
	c.password = wire.Password
	c.port = wire.Port
	c.username = wire.Username
	return nil
}

func (c *DBCredentials) DBName() string   { return c.dbName }
func (c *DBCredentials) Host() string     { return c.host }
func (c *DBCredentials) Password() string { return c.password }
func (c *DBCredentials) Port() int        { return c.port }
func (c *DBCredentials) Username() string { return c.username }
