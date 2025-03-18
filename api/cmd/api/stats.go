package main

import (
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type poolStats struct {
	AcquireCount            int64
	AcquireDuration         time.Duration
	AcquiredConns           int32
	CanceledAcquireCount    int64
	ConstructingConns       int32
	EmptyAcquireCount       int64
	IdleConns               int32
	MaxConns                int32
	MaxIdleDestroyCount     int64
	MaxLifetimeDestroyCount int64
	NewConnsCount           int64
	TotalConns              int32
}

func dbStats(st *pgxpool.Stat) poolStats {
	return poolStats{
		AcquireCount:            st.AcquireCount(),
		AcquireDuration:         st.AcquireDuration(),
		AcquiredConns:           st.AcquiredConns(),
		CanceledAcquireCount:    st.CanceledAcquireCount(),
		ConstructingConns:       st.ConstructingConns(),
		EmptyAcquireCount:       st.EmptyAcquireCount(),
		IdleConns:               st.IdleConns(),
		MaxConns:                st.MaxConns(),
		MaxIdleDestroyCount:     st.MaxIdleDestroyCount(),
		MaxLifetimeDestroyCount: st.MaxLifetimeDestroyCount(),
		NewConnsCount:           st.NewConnsCount(),
		TotalConns:              st.TotalConns(),
	}
}
