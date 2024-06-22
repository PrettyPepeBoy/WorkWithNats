package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/sirupsen/logrus"
)

type Storage struct {
	db               *pgx.Conn
	putInTableStmt   *pgconn.StatementDescription
	getFromTableStmt *pgconn.StatementDescription
	checkInTableStmt *pgconn.StatementDescription
}

var (
	ErrRowNotExist = errors.New("row with such id do not exist")
)

func MustConnectDB(ctx context.Context, username, password, host, database string, port int) *Storage {
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", username, password, host, port, database)
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		logrus.Fatalf("failed to connect to postgresservice, error: %v", err)
	}
	if err = conn.Ping(ctx); err != nil {
		logrus.Fatalf("database is not pinging, error: %v", err)
	}

	putInTableStmt, err := conn.Prepare(ctx, "Put", `INSERT INTO products(name, json_data) VALUES ($1, $2) RETURNING id`)
	if err != nil {
		logrus.Errorf("failed to prepare createTableStmt, error: %v", err)
	}

	getFromTableStmt, err := conn.Prepare(ctx, "Get", `SELECT json_data FROM products WHERE id = $1`)
	if err != nil {
		logrus.Errorf("failed to prepare getFromTableStmt, error: %v", err)
	}

	checkInTableStmt, err := conn.Prepare(ctx, "Check", `SELECT json_data->>'name' as name FROM products`)
	if err != nil {
		logrus.Errorf("failed to prepare getFromTableStmt, error: %v", err)
	}
	return &Storage{db: conn,
		putInTableStmt:   putInTableStmt,
		getFromTableStmt: getFromTableStmt,
		checkInTableStmt: checkInTableStmt,
	}
}

func (s *Storage) PutInTable(name string, data []byte) (int, error) {
	var id int
	if err := s.db.QueryRow(context.Background(), s.putInTableStmt.Name, name, data).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (s *Storage) GetFromTable(id int) ([]byte, error) {
	var data []byte
	if err := s.db.QueryRow(context.Background(), s.getFromTableStmt.Name, id).Scan(&data); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRowNotExist
		}
		return nil, err
	}
	return data, nil
}

func (s *Storage) CheckInTable() ([]interface{}, error) {
	rows, err := s.db.Query(context.Background(), s.checkInTableStmt.Name)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}
