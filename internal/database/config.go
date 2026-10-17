package database

import (
	"net"
	"net/url"

	"github.com/jackc/pgx/v5"
)

// Local coursework settings. Change these if PostgreSQL runs elsewhere.
const (
	host     = "127.0.0.1"
	port     = "5432"
	username = "postgres"
	password = ""
)

func localConfig(databaseName string) (*pgx.ConnConfig, error) {

	connection := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(username, password),
		Host:     net.JoinHostPort(host, port),
		Path:     databaseName,
		RawQuery: "sslmode=disable",
	}

	config, err := pgx.ParseConfig(connection.String())
	if err != nil {
		return nil, err
	}
	
	config.Password = password
	return config, nil
}
