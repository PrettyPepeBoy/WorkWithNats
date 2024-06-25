package product

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
)

type Table struct {
	db               *pgx.Conn
	putInTableStmt   *pgconn.StatementDescription
	getFromTableStmt *pgconn.StatementDescription
}

var (
	ErrRowNotExist = errors.New("row with such id do not exist")
)

func NewTable() (*Table, error) {
	username := viper.GetString("product.table.username")
	password := os.Getenv(viper.GetString("product.table.password"))
	host := viper.GetString("product.table.host")
	port := viper.GetInt("product.table.port")
	database := viper.GetString("product.table.database")

	conn, err := pgx.Connect(context.Background(), fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", username, password, host, port, database))
	if err != nil {
		return nil, err
	}

	if err = conn.Ping(context.Background()); err != nil {
		return nil, err
	}

	putInTableStmt, err := conn.Prepare(context.Background(), "Put", `INSERT INTO products(name, json_data) VALUES ($1, $2) RETURNING id`)
	if err != nil {
		logrus.Errorf("failed to prepare createTableStmt, error: %v", err)
	}

	getFromTableStmt, err := conn.Prepare(context.Background(), "GetById", `SELECT json_data FROM products WHERE id = $1`)
	if err != nil {
		logrus.Errorf("failed to prepare getFromTableStmt, error: %v", err)
	}

	return &Table{db: conn,
		putInTableStmt:   putInTableStmt,
		getFromTableStmt: getFromTableStmt,
	}, nil
}

func (s *Table) Put(name string, data []byte) (int, error) {
	var id int
	if err := s.db.QueryRow(context.Background(), s.putInTableStmt.Name, name, data).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (s *Table) GetById(id int) ([]byte, error) {
	var data []byte
	if err := s.db.QueryRow(context.Background(), s.getFromTableStmt.Name, id).Scan(&data); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRowNotExist
		}
		return nil, err
	}
	return data, nil
}
