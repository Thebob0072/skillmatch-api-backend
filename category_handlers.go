package main

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ServiceCategory Model
type ServiceCategory struct {
	CategoryID   int     `json:"category_id"`
	Name         string  `json:"name"`
	NameThai     string  `json:"name_thai"`
	Description  *string `json:"description"`
	Icon         *string `json:"icon"`
	IsAdult      bool    `json:"is_adult"`
	DisplayOrder int     `json:"display_order"`
	IsActive     bool    `json:"is_active"`
}

// --- List All Service Categories ---
func listServiceCategoriesHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		includeAdult := c.DefaultQuery("include_adult", "true") == "true"

		sqlStatement := `
			SELECT category_id, name, name_thai, description, icon, 
			       is_adult, display_order, is_active
			FROM service_categories
			WHERE is_active = true
		`

		if !includeAdult {
			sqlStatement += " AND is_adult = false"
		}

		sqlStatement += " ORDER BY display_order ASC"

		rows, err := dbPool.Query(ctx, sqlStatement)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to fetch categories",
				"details": err.Error(),
			})
			return
		}
		defer rows.Close()

		categories := make([]ServiceCategory, 0)
		for rows.Next() {
			var cat ServiceCategory
			err := rows.Scan(
				&cat.CategoryID, &cat.Name, &cat.NameThai, &cat.Description,
				&cat.Icon, &cat.IsAdult, &cat.DisplayOrder, &cat.IsActive,
			)
			if err != nil {
				continue
			}
			categories = append(categories, cat)
		}

		c.JSON(http.StatusOK, gin.H{
			"categories": categories,
			"total":      len(categories),
		})
	}
}

// --- Get Provider's Categories ---
func getProviderCategoriesHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		providerIDStr := c.Param("userId")
		providerID, err := strconv.Atoi(providerIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider ID"})
			return
		}

		rows, err := dbPool.Query(ctx, `
			SELECT sc.category_id, sc.name, sc.name_thai, sc.description,
			       sc.icon, sc.is_adult, sc.display_order
			FROM provider_categories pc
			JOIN service_categories sc ON pc.category_id = sc.category_id
			WHERE pc.provider_id = $1 AND sc.is_active = true
			ORDER BY sc.display_order ASC
		`, providerID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch provider categories",
			})
			return
		}
		defer rows.Close()

		categories := make([]ServiceCategory, 0)
		for rows.Next() {
			var cat ServiceCategory
			err := rows.Scan(
				&cat.CategoryID, &cat.Name, &cat.NameThai, &cat.Description,
				&cat.Icon, &cat.IsAdult, &cat.DisplayOrder,
			)
			if err != nil {
				continue
			}
			categories = append(categories, cat)
		}

		c.JSON(http.StatusOK, gin.H{
			"provider_id": providerID,
			"categories":  categories,
			"total":       len(categories),
		})
	}
}

// --- Update Provider's Categories ---
func updateProviderCategoriesHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")

		var req struct {
			CategoryIDs []int `json:"category_ids" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate: max 5 categories
		if len(req.CategoryIDs) > 5 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Cannot select more than 5 categories",
			})
			return
		}

		// Start transaction
		tx, err := dbPool.Begin(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to start transaction",
			})
			return
		}
		defer tx.Rollback(ctx)

		// Delete existing categories
		_, err = tx.Exec(ctx,
			"DELETE FROM provider_categories WHERE provider_id = $1",
			userID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to clear existing categories",
			})
			return
		}

		// Insert new categories
		for _, categoryID := range req.CategoryIDs {
			_, err = tx.Exec(ctx, `
				INSERT INTO provider_categories (provider_id, category_id)
				VALUES ($1, $2)
				ON CONFLICT (provider_id, category_id) DO NOTHING
			`, userID, categoryID)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to add category",
					"details": err.Error(),
				})
				return
			}
		}

		// Commit transaction
		if err = tx.Commit(ctx); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to commit transaction",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":      "Categories updated successfully",
			"category_ids": req.CategoryIDs,
			"total":        len(req.CategoryIDs),
		})
	}
}

// --- Browse Providers by Category ---
func browseProvidersByCategoryHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		categoryIDStr := c.Param("category_id")
		categoryID, err := strconv.Atoi(categoryIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
			return
		}

		// Pagination
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		if page < 1 {
			page = 1
		}
		if limit < 1 || limit > 50 {
			limit = 20
		}
		offset := (page - 1) * limit

		// Get providers in this category
		rows, err := dbPool.Query(ctx, `
			SELECT 
				u.user_id, u.username, u.gender_id,
				p.age, p.profile_image_url, u.google_profile_picture,
				p.province, p.district, p.sub_district,
				COALESCE(AVG(r.rating), 0) as avg_rating,
				COUNT(DISTINCT r.review_id) as review_count,
				MIN(sp.price) as min_price
			FROM provider_categories pc
			JOIN users u ON pc.provider_id = u.user_id
			LEFT JOIN user_profiles p ON u.user_id = p.user_id
			LEFT JOIN reviews r ON u.user_id = r.provider_id
			LEFT JOIN service_packages sp ON u.user_id = sp.provider_id AND sp.is_active = true
			WHERE pc.category_id = $1
			  AND u.verification_status IN ('verified', 'approved')
			GROUP BY u.user_id, u.username, u.gender_id, p.age, p.profile_image_url,
			         u.google_profile_picture, p.province, p.district, p.sub_district
			ORDER BY avg_rating DESC, review_count DESC
			LIMIT $2 OFFSET $3
		`, categoryID, limit, offset)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to fetch providers",
				"details": err.Error(),
			})
			return
		}
		defer rows.Close()

		providers := make([]map[string]interface{}, 0)
		for rows.Next() {
			var (
				userID, genderID                      int
				username                              string
				age                                   *int
				profileImageURL, googleProfilePicture *string
				province, district, subDistrict       *string
				avgRating                             float64
				reviewCount                           int
				minPrice                              *float64
			)

			err := rows.Scan(
				&userID, &username, &genderID, &age, &profileImageURL,
				&googleProfilePicture, &province, &district, &subDistrict,
				&avgRating, &reviewCount, &minPrice,
			)
			if err != nil {
				continue
			}

			providers = append(providers, map[string]interface{}{
				"user_id":                userID,
				"username":               username,
				"gender_id":              genderID,
				"age":                    age,
				"profile_image_url":      profileImageURL,
				"google_profile_picture": googleProfilePicture,
				"province":               province,
				"district":               district,
				"sub_district":           subDistrict,
				"average_rating":         avgRating,
				"review_count":           reviewCount,
				"min_price":              minPrice,
			})
		}

		// Get total count
		var totalCount int
		err = dbPool.QueryRow(ctx, `
			SELECT COUNT(DISTINCT pc.provider_id)
			FROM provider_categories pc
			JOIN users u ON pc.provider_id = u.user_id
			WHERE pc.category_id = $1
			  AND u.verification_status IN ('verified', 'approved')
		`, categoryID).Scan(&totalCount)

		if err != nil {
			totalCount = 0
		}

		c.JSON(http.StatusOK, gin.H{
			"category_id": categoryID,
			"providers":   providers,
			"total":       totalCount,
			"page":        page,
			"limit":       limit,
			"total_pages": (totalCount + limit - 1) / limit,
		})
	}
}
