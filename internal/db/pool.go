package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Open(ctx context.Context) (*pgxpool.Pool, error) {
	primary := strings.TrimSpace(os.Getenv("SUPABASE_CONNECTION_STRING"))
	fallback := strings.TrimSpace(os.Getenv("SUPABASE_POOLER_STRING"))
	if primary == "" && fallback == "" {
		return nil, fmt.Errorf("SUPABASE_CONNECTION_STRING or SUPABASE_POOLER_STRING is required")
	}

	if primary != "" {
		pool, err := connect(ctx, primary)
		if err == nil {
			return pool, nil
		}
		if fallback == "" {
			return nil, fmt.Errorf("database ping: %w", err)
		}
		log.Printf("direct connection failed, trying pooler: %v", err)
	}

	pool, err := connect(ctx, fallback)
	if err != nil {
		return nil, fmt.Errorf("database ping: %w", err)
	}
	return pool, nil
}

func connect(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := New(pool).Ping(pingCtx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}
