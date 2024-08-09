package database

import (
	"Go_Food_Delivery/pkg/database/models/restaurant"
	"Go_Food_Delivery/pkg/database/models/user"
	"context"
	"database/sql"
	"fmt"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/driver/sqliteshim"
	"log"
	"log/slog"
	"os"
	"strconv"
	"time"
)

type Database interface {
	Db() *bun.DB
	Migrate() error
	HealthCheck() bool
	Close() error
	Insert(ctx context.Context, model any) (sql.Result, error)
	Delete(ctx context.Context, tableName string, filter Filter) (sql.Result, error)
	Select(ctx context.Context, model any, columnName string, parameter any) error
	Count(ctx context.Context, tableName string, ColumnExpression string, columnName string, parameter any) (int64, error)
}

type Filter map[string]any

type DB struct {
	db *bun.DB
}

func (d *DB) Db() *bun.DB {
	return d.db
}

func (d *DB) Insert(ctx context.Context, model any) (sql.Result, error) {
	modelInfo := d.loadModel(model, "INSERT").(*bun.InsertQuery)
	result, err := modelInfo.Exec(ctx)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (d *DB) Delete(ctx context.Context, tableName string, filter Filter) (sql.Result, error) {
	result, err := d.db.NewDelete().Table(tableName).Where(d.whereCondition(filter)).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return result, err
}

func (d *DB) Select(ctx context.Context, model any, columnName string, parameter any) error {
	modelInfo := d.loadModel(model, "SELECT").(bun.SelectQuery)
	err := modelInfo.Where(fmt.Sprintf("%s = ?", columnName), parameter).Scan(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) Count(ctx context.Context, tableName string, ColumnExpression string, columnName string, parameter any) (int64, error) {
	var count int64
	err := d.db.NewSelect().Table(tableName).ColumnExpr(ColumnExpression).
		Where(fmt.Sprintf("%s = ?", columnName), parameter).Scan(ctx, &count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (d *DB) HealthCheck() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := d.db.PingContext(ctx)
	if err != nil {
		slog.Error("DB::error", err)
		return false
	}
	return true
}

func (d *DB) Close() error {
	slog.Info("DB::Closing database connection")
	return d.db.Close()
}

func New() Database {
	dbHost := os.Getenv("DB_HOST")
	dbUsername := os.Getenv("DB_USERNAME")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")
	databasePort, err := strconv.Atoi(dbPort)
	if err != nil {
		log.Fatal("Invalid DB Port")
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", dbUsername, dbPassword, dbHost, databasePort, dbName)
	database := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(database, pgdialect.New())
	return &DB{db: db}

}

// NewTestDB creates a new in-memory test database.
func NewTestDB() Database {
	database, err := sql.Open(sqliteshim.ShimName, "file::memory:?cache=shared")
	if err != nil {
		panic(err)
	}
	db := bun.NewDB(database, sqlitedialect.New())
	return &DB{db}
}

func (d *DB) Migrate() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	models := []interface{}{
		(*user.User)(nil),
		(*restaurant.Restaurant)(nil),
		(*restaurant.MenuItem)(nil),
	}

	for _, model := range models {
		if _, err := d.db.NewCreateTable().Model(model).IfNotExists().Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}
