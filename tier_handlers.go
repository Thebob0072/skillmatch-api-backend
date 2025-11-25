package main

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// --- Handler: GET /tiers (ดึงข้อมูล Tiers ทั้งหมด) ---
func getTiersHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tiers []Tier // (Tier struct มาจาก models.go)

		rows, err := dbPool.Query(ctx, `
			SELECT tier_id, name, access_level, price_monthly
			FROM tiers
			ORDER BY access_level ASC
		`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed", "details": err.Error()})
			return
		}
		defer rows.Close()

		for rows.Next() {
			var t Tier
			if err := rows.Scan(
				&t.TierID, &t.Name, &t.AccessLevel, &t.PriceMonthly,
			); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan tier row"})
				return
			}
			tiers = append(tiers, t)
		}

		c.JSON(http.StatusOK, tiers)
	}
}
