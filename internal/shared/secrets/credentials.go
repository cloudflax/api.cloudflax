package secrets

import (
	"encoding/json"
	"fmt"
	"strconv"
)

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
// Port uses json.RawMessage because AWS RDS secrets store it as a string ("5432"),
// while other sources may use a number (5432).
type dbCredentialsWire struct {
	DBName   string          `json:"dbname"`
	Host     string          `json:"host"`
	Password string          `json:"password"`
	Port     json.RawMessage `json:"port"`
	Username string          `json:"username"`
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
	c.username = wire.Username

	port, err := parsePort(wire.Port)
	if err != nil {
		return fmt.Errorf("parse port: %w", err)
	}
	c.port = port

	return nil
}

// parsePort handles port as either a JSON number (5432) or a JSON string ("5432").
func parsePort(raw json.RawMessage) (int, error) {
	var portInt int
	if err := json.Unmarshal(raw, &portInt); err == nil {
		return portInt, nil
	}

	var portStr string
	if err := json.Unmarshal(raw, &portStr); err == nil {
		return strconv.Atoi(portStr)
	}

	return 0, fmt.Errorf("unexpected port value: %s", string(raw))
}

func (c *DBCredentials) DBName() string   { return c.dbName }
func (c *DBCredentials) Host() string     { return c.host }
func (c *DBCredentials) Password() string { return c.password }
func (c *DBCredentials) Port() int        { return c.port }
func (c *DBCredentials) Username() string { return c.username }
