package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// --- Handler: GET /browse/search ---
// Advanced search with all filters (location, rating, tier, category, service_type, sort)
func browseSearchHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse query parameters
		location := c.Query("location")            // Province/District name
		ratingStr := c.Query("rating")             // Min rating: "3", "4", "4.5"
		tierStr := c.Query("tier")                 // Provider level: "1"-"4"
		categoryStr := c.Query("category")         // Category ID
		serviceType := c.Query("service_type")     // "Incall", "Outcall", "Both"
		sortBy := c.DefaultQuery("sort", "rating") // "rating", "reviews", "price"
		province := c.Query("province")            // Specific province
		district := c.Query("district")            // Specific district

		// Pagination
		pageStr := c.DefaultQuery("page", "1")
		limitStr := c.DefaultQuery("limit", "20")

		page, _ := strconv.Atoi(pageStr)
		limit, _ := strconv.Atoi(limitStr)
		if page < 1 {
			page = 1
		}
		if limit < 1 || limit > 50 {
			limit = 20
		}
		offset := (page - 1) * limit

		// Build base query
		query := `
			SELECT DISTINCT
				u.user_id,
				u.username,
				u.profile_picture_url,
				p.bio,
				u.provider_level_id,
				COALESCE(pl.name, 'General') as provider_level_name,
				COALESCE(AVG(r.rating), 0) as rating_avg,
				COUNT(DISTINCT r.review_id) as review_count,
				COALESCE(p.service_type, 'Both') as service_type,
				p.location,
				p.profile_image_url,
				MIN(sp.price) as min_price
			FROM users u
			LEFT JOIN tiers pl ON u.provider_level_id = pl.tier_id
			LEFT JOIN user_profiles p ON u.user_id = p.user_id
			LEFT JOIN reviews r ON u.user_id = r.provider_id
			LEFT JOIN service_packages sp ON u.user_id = sp.provider_id AND sp.is_active = true
			WHERE u.verification_status IN ('approved', 'verified')
		`

		args := []interface{}{}
		argPos := 1

		// Apply filters
		if location != "" || province != "" || district != "" {
			// Combine all location filters into one ILIKE search on p.location
			searchTerm := location
			if searchTerm == "" {
				searchTerm = province
			}
			if searchTerm == "" {
				searchTerm = district
			}

			query += fmt.Sprintf(" AND p.location ILIKE $%d", argPos)
			args = append(args, "%"+searchTerm+"%")
			argPos++
		}

		if ratingStr != "" {
			minRating, err := strconv.ParseFloat(ratingStr, 64)
			if err == nil {
				// Will filter in HAVING clause after GROUP BY
				query += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM reviews r2 WHERE r2.provider_id = u.user_id GROUP BY r2.provider_id HAVING AVG(r2.rating) >= $%d)", argPos)
				args = append(args, minRating)
				argPos++
			}
		}

		if tierStr != "" {
			tierID, err := strconv.Atoi(tierStr)
			if err == nil && tierID >= 1 && tierID <= 4 {
				query += fmt.Sprintf(" AND u.provider_level_id = $%d", argPos)
				args = append(args, tierID)
				argPos++
			}
		}

		if categoryStr != "" {
			categoryID, err := strconv.Atoi(categoryStr)
			if err == nil {
				query += fmt.Sprintf(` AND EXISTS (
					SELECT 1 FROM provider_categories pc 
					WHERE pc.provider_id = u.user_id 
					AND pc.category_id = $%d
				)`, argPos)
				args = append(args, categoryID)
				argPos++
			}
		}

		if serviceType != "" && serviceType != "All" {
			// Match exact service_type OR "Both"
			query += fmt.Sprintf(" AND (p.service_type = $%d OR p.service_type = 'Both')", argPos)
			args = append(args, serviceType)
			argPos++
		}

		// GROUP BY
		query += `
			GROUP BY 
				u.user_id, u.username, u.profile_picture_url, p.bio,
				u.provider_level_id, pl.name, p.service_type, 
				p.location, p.profile_image_url
		`

		// Apply sorting
		switch sortBy {
		case "reviews":
			query += " ORDER BY review_count DESC, rating_avg DESC"
		case "price":
			query += " ORDER BY min_price ASC NULLS LAST, rating_avg DESC"
		default: // rating
			query += " ORDER BY rating_avg DESC, review_count DESC"
		}

		// Add pagination
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argPos, argPos+1)
		args = append(args, limit, offset)

		// Execute query
		rows, err := dbPool.Query(ctx, query, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to fetch providers",
				"details": err.Error(),
			})
			return
		}
		defer rows.Close()

		providers := []map[string]interface{}{}
		for rows.Next() {
			var (
				userID            int
				username          string
				profilePictureURL *string
				bio               *string
				providerLevelID   int
				providerLevelName string
				ratingAvg         float64
				reviewCount       int
				svcType           string
				loc               *string
				profileImageURL   *string
				minPrice          *float64
			)

			err := rows.Scan(
				&userID, &username, &profilePictureURL, &bio,
				&providerLevelID, &providerLevelName,
				&ratingAvg, &reviewCount, &svcType,
				&loc, &profileImageURL, &minPrice,
			)

			if err != nil {
				continue
			}

			// Prefer profile_picture_url over profile_image_url
			displayPicture := profilePictureURL
			if displayPicture == nil || *displayPicture == "" {
				displayPicture = profileImageURL
			}

			providers = append(providers, map[string]interface{}{
				"user_id":             userID,
				"username":            username,
				"profile_picture_url": displayPicture,
				"bio":                 bio,
				"provider_level_id":   providerLevelID,
				"provider_level_name": providerLevelName,
				"rating_avg":          ratingAvg,
				"review_count":        reviewCount,
				"service_type":        svcType,
				"location":            loc,
				"min_price":           minPrice,
			})
		}

		// Get total count (without pagination)
		countQuery := `
			SELECT COUNT(DISTINCT u.user_id) 
			FROM users u
			LEFT JOIN user_profiles p ON u.user_id = p.user_id
			WHERE u.verification_status IN ('approved', 'verified')
		`

		countArgs := []interface{}{}
		countArgPos := 1

		// Apply same filters for count
		if location != "" || province != "" || district != "" {
			searchTerm := location
			if searchTerm == "" {
				searchTerm = province
			}
			if searchTerm == "" {
				searchTerm = district
			}

			countQuery += fmt.Sprintf(" AND p.location ILIKE $%d", countArgPos)
			countArgs = append(countArgs, "%"+searchTerm+"%")
			countArgPos++
		}

		if categoryStr != "" {
			categoryID, _ := strconv.Atoi(categoryStr)
			countQuery += fmt.Sprintf(` AND EXISTS (
				SELECT 1 FROM provider_categories pc 
				WHERE pc.provider_id = u.user_id 
				AND pc.category_id = $%d
			)`, countArgPos)
			countArgs = append(countArgs, categoryID)
			countArgPos++
		}

		if serviceType != "" && serviceType != "All" {
			countQuery += fmt.Sprintf(" AND (p.service_type = $%d OR p.service_type = 'Both')", countArgPos)
			countArgs = append(countArgs, serviceType)
			countArgPos++
		}

		var total int
		err = dbPool.QueryRow(ctx, countQuery, countArgs...).Scan(&total)
		if err != nil {
			total = len(providers)
		}

		c.JSON(http.StatusOK, gin.H{
			"providers": providers,
			"pagination": gin.H{
				"page":        page,
				"limit":       limit,
				"total":       total,
				"total_pages": (total + limit - 1) / limit,
			},
			"filters_applied": gin.H{
				"location":     location,
				"province":     province,
				"district":     district,
				"rating":       ratingStr,
				"tier":         tierStr,
				"category":     categoryStr,
				"service_type": serviceType,
				"sort":         sortBy,
			},
		})
	}
}
