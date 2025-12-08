package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func runMigrations(dbPool *pgxpool.Pool, ctx context.Context) {
	// --- 1. Genders Table ---
	_, err := dbPool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS genders (
			gender_id SERIAL PRIMARY KEY,
			gender_name VARCHAR(50) NOT NULL UNIQUE
		);
	`)
	if err != nil {
		log.Fatalf("ไม่สามารถสร้างตาราง genders: %v\n", err)
	}

	_, err = dbPool.Exec(ctx, `
		INSERT INTO genders (gender_id, gender_name) VALUES
		(1, 'Male'), (2, 'Female'), (3, 'Other'), (4, 'Prefer not to say')
		ON CONFLICT (gender_id) DO NOTHING; 
	`)
	if err != nil {
		log.Printf("ไม่สามารถเพิ่มข้อมูลเริ่มต้นให้ genders: %v\n", err)
	}

	// --- 2. Tiers Table (รวม GOD Tier) ---
	_, err = dbPool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS tiers (
			tier_id SERIAL PRIMARY KEY,
			name VARCHAR(50) NOT NULL UNIQUE,
			access_level INT NOT NULL UNIQUE,
			price_monthly DECIMAL(10, 2) NOT NULL DEFAULT 0.00
		);
	`)
	if err != nil {
		log.Fatalf("ไม่สามารถสร้างตาราง tiers: %v\n", err)
	}

	_, err = dbPool.Exec(ctx, `
		INSERT INTO tiers (tier_id, name, access_level, price_monthly) VALUES
		(1, 'General', 0, 0.00),
		(2, 'Silver', 1, 9.99),
		(3, 'Diamond', 2, 29.99),
		(4, 'Premium', 3, 99.99),
		(5, 'GOD', 999, 9999.99)
		ON CONFLICT (name) DO NOTHING; 
	`)
	if err != nil {
		log.Printf("ไม่สามารถเพิ่มข้อมูล tiers: %v\n", err)
	}

	// --- 3. ตาราง Users (สร้าง) ---
	_, err = dbPool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			user_id SERIAL PRIMARY KEY,
			username VARCHAR(100) NOT NULL,
			email VARCHAR(255) NOT NULL UNIQUE,
			password_hash TEXT, 
			gender_id INT NOT NULL REFERENCES genders(gender_id) DEFAULT 4,
			first_name VARCHAR(100),
			last_name VARCHAR(100),
			registration_date TIMESTAMPTZ DEFAULT NOW(),
			google_id TEXT UNIQUE,
			google_profile_picture TEXT,
			tier_id INT REFERENCES tiers(tier_id) DEFAULT 1,
			phone_number VARCHAR(20),
			verification_status VARCHAR(20) NOT NULL DEFAULT 'unverified',
			is_admin BOOLEAN NOT NULL DEFAULT false,
			provider_level_id INT REFERENCES tiers(tier_id) DEFAULT 1
		);
	`)
	if err != nil {
		log.Fatalf("ไม่สามารถสร้างตาราง users: %v\n", err)
	}

	// (Index)
	_, err = dbPool.Exec(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS email_idx ON users (email);`)
	if err != nil {
		log.Printf("ไม่สามารถสร้าง Index (email_idx): %v\n", err)
	}
	_, err = dbPool.Exec(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS google_id_idx ON users (google_id);`)
	if err != nil {
		log.Printf("ไม่สามารถสร้าง Index (google_id_idx): %v\n", err)
	}

	// --- 4. ตาราง User_Photos (Gallery) ---
	_, err = dbPool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS user_photos (
			photo_id SERIAL PRIMARY KEY,
			user_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			photo_url TEXT NOT NULL,
			sort_order INT NOT NULL DEFAULT 0,
			uploaded_at TIMESTAMPTZ DEFAULT NOW()
		);
	`)
	if err != nil {
		log.Fatalf("ไม่สามารถสร้างตาราง user_photos: %v\n", err)
	}

	// --- 5. ตาราง User_Verifications (KYC) ---
	_, err = dbPool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS user_verifications (
			verification_id SERIAL PRIMARY KEY,
			user_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE UNIQUE,
			national_id_url TEXT,
			health_cert_url TEXT,
			face_scan_url TEXT,
			profile_photos JSONB,
			submitted_at TIMESTAMPTZ
		);
	`)
	if err != nil {
		log.Fatalf("ไม่สามารถสร้างตาราง user_verifications: %v\n", err)
	}

	// Add profile_photos column if not exists (for existing tables)
	_, _ = dbPool.Exec(ctx, `ALTER TABLE user_verifications ADD COLUMN IF NOT EXISTS profile_photos JSONB;`)

	// --- 6. ตาราง User_Profiles (ข้อมูลที่ผู้ใช้กรอกเอง) ---
	_, err = dbPool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS user_profiles (
			user_id INT PRIMARY KEY REFERENCES users(user_id) ON DELETE CASCADE,
			bio TEXT,
			location VARCHAR(255),
			skills TEXT[],
			profile_image_url TEXT,
			updated_at TIMESTAMPTZ DEFAULT NOW()
		);
	`)
	if err != nil {
		log.Fatalf("ไม่สามารถสร้างตาราง user_profiles: %v\n", err)
	}

	// --- 7. ตาราง Service_Packages (แพ็คเกจบริการ) ---
	_, err = dbPool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS service_packages (
			package_id SERIAL PRIMARY KEY,
			provider_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			package_name VARCHAR(100) NOT NULL,
			description TEXT,
			duration INT NOT NULL,
			price DECIMAL(10, 2) NOT NULL,
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);
	`)
	if err != nil {
		log.Fatalf("ไม่สามารถสร้างตาราง service_packages: %v\n", err)
	}

	// --- 8. ตาราง Bookings (การจอง) ---
	_, err = dbPool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS bookings (
			booking_id SERIAL PRIMARY KEY,
			client_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			provider_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			package_id INT NOT NULL REFERENCES service_packages(package_id),
			booking_date DATE NOT NULL,
			start_time TIMESTAMPTZ NOT NULL,
			end_time TIMESTAMPTZ NOT NULL,
			total_price DECIMAL(10, 2) NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			location TEXT,
			special_notes TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			completed_at TIMESTAMPTZ,
			cancelled_at TIMESTAMPTZ,
			cancellation_reason TEXT
		);
	`)
	if err != nil {
		log.Fatalf("ไม่สามารถสร้างตาราง bookings: %v\n", err)
	}

	// --- 9. ตาราง Reviews (รีวิว) ---
	_, err = dbPool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS reviews (
			review_id SERIAL PRIMARY KEY,
			booking_id INT NOT NULL REFERENCES bookings(booking_id) ON DELETE CASCADE UNIQUE,
			client_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			provider_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			rating INT NOT NULL CHECK (rating >= 1 AND rating <= 5),
			comment TEXT,
			is_verified BOOLEAN DEFAULT true,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);
	`)
	if err != nil {
		log.Fatalf("ไม่สามารถสร้างตาราง reviews: %v\n", err)
	}

	// --- 10. ตาราง Provider_Availability (ช่วงเวลาว่าง) ---
	_, err = dbPool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS provider_availability (
			availability_id SERIAL PRIMARY KEY,
			provider_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			day_of_week INT NOT NULL CHECK (day_of_week >= 0 AND day_of_week <= 6),
			start_time TIME NOT NULL,
			end_time TIME NOT NULL,
			is_active BOOLEAN DEFAULT true,
			UNIQUE(provider_id, day_of_week, start_time, end_time)
		);
	`)
	if err != nil {
		log.Fatalf("ไม่สามารถสร้างตาราง provider_availability: %v\n", err)
	}

	// --- 11. ตาราง Favorites (รายการโปรด) ---
	_, err = dbPool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS favorites (
			favorite_id SERIAL PRIMARY KEY,
			client_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			provider_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			UNIQUE(client_id, provider_id)
		);
	`)
	if err != nil {
		log.Fatalf("ไม่สามารถสร้างตาราง favorites: %v\n", err)
	}

	// --- 12. เพิ่มคอลัมน์ใน user_profiles สำหรับข้อมูลเพิ่มเติม ---
	_, err = dbPool.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'user_profiles' AND column_name = 'age') THEN
				ALTER TABLE user_profiles ADD COLUMN age INT;
			END IF;
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'user_profiles' AND column_name = 'height') THEN
				ALTER TABLE user_profiles ADD COLUMN height INT;
			END IF;
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'user_profiles' AND column_name = 'weight') THEN
				ALTER TABLE user_profiles ADD COLUMN weight INT;
			END IF;
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'user_profiles' AND column_name = 'ethnicity') THEN
				ALTER TABLE user_profiles ADD COLUMN ethnicity VARCHAR(50);
			END IF;
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'user_profiles' AND column_name = 'languages') THEN
				ALTER TABLE user_profiles ADD COLUMN languages TEXT[];
			END IF;
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'user_profiles' AND column_name = 'working_hours') THEN
				ALTER TABLE user_profiles ADD COLUMN working_hours VARCHAR(100);
			END IF;
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'user_profiles' AND column_name = 'is_available') THEN
				ALTER TABLE user_profiles ADD COLUMN is_available BOOLEAN DEFAULT false;
			END IF;
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'user_profiles' AND column_name = 'service_type') THEN
				ALTER TABLE user_profiles ADD COLUMN service_type VARCHAR(20);
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning: Could not add profile columns: %v\n", err)
	}

	// --- 13. สร้าง Indexes ---
	_, err = dbPool.Exec(ctx, `
		CREATE INDEX IF NOT EXISTS idx_bookings_provider ON bookings(provider_id);
		CREATE INDEX IF NOT EXISTS idx_bookings_client ON bookings(client_id);
		CREATE INDEX IF NOT EXISTS idx_bookings_status ON bookings(status);
		CREATE INDEX IF NOT EXISTS idx_bookings_date ON bookings(booking_date);
		CREATE INDEX IF NOT EXISTS idx_reviews_provider ON reviews(provider_id);
		CREATE INDEX IF NOT EXISTS idx_favorites_client ON favorites(client_id);
		CREATE INDEX IF NOT EXISTS idx_favorites_provider ON favorites(provider_id);
	`)
	if err != nil {
		log.Printf("Warning: Could not create indexes: %v\n", err)
	}

	// --- Migration 016: Platform Bank Account Tracking ---
	fmt.Println("🔄 Running Migration 016: Platform Bank Account Tracking...")

	_, err = dbPool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS platform_bank_accounts (
			platform_bank_id SERIAL PRIMARY KEY,
			bank_name VARCHAR(100) NOT NULL,
			bank_code VARCHAR(10),
			account_number VARCHAR(50) NOT NULL UNIQUE,
			account_name VARCHAR(200) NOT NULL,
			account_type VARCHAR(20) DEFAULT 'current',
			branch_name VARCHAR(100),
			account_holder VARCHAR(200),
			account_holder_id_card VARCHAR(50),
			current_balance DECIMAL(12, 2) DEFAULT 0.00,
			total_inflow DECIMAL(12, 2) DEFAULT 0.00,
			total_outflow DECIMAL(12, 2) DEFAULT 0.00,
			is_active BOOLEAN DEFAULT true,
			is_default BOOLEAN DEFAULT false,
			owned_by INTEGER REFERENCES users(user_id),
			notes TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		log.Printf("Warning: platform_bank_accounts table may already exist: %v\n", err)
	}

	_, err = dbPool.Exec(ctx, `
		ALTER TABLE withdrawals 
			ADD COLUMN IF NOT EXISTS platform_bank_account_id INTEGER REFERENCES platform_bank_accounts(platform_bank_id),
			ADD COLUMN IF NOT EXISTS platform_transfer_timestamp TIMESTAMP,
			ADD COLUMN IF NOT EXISTS platform_transfer_by INTEGER REFERENCES users(user_id);
	`)
	if err != nil {
		log.Printf("Warning: Could not add platform columns to withdrawals: %v\n", err)
	}

	_, err = dbPool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS withdrawal_transfer_logs (
			log_id SERIAL PRIMARY KEY,
			withdrawal_id INTEGER NOT NULL REFERENCES withdrawals(withdrawal_id),
			platform_bank_account_id INTEGER NOT NULL REFERENCES platform_bank_accounts(platform_bank_id),
			platform_account_number VARCHAR(50) NOT NULL,
			platform_account_name VARCHAR(200) NOT NULL,
			provider_account_number VARCHAR(50) NOT NULL,
			provider_account_name VARCHAR(200) NOT NULL,
			provider_bank_name VARCHAR(100) NOT NULL,
			transfer_amount DECIMAL(12, 2) NOT NULL,
			transfer_timestamp TIMESTAMP NOT NULL,
			transfer_reference VARCHAR(100),
			transfer_slip_url TEXT,
			transferred_by INTEGER NOT NULL REFERENCES users(user_id),
			transfer_method VARCHAR(50),
			verified BOOLEAN DEFAULT false,
			verified_at TIMESTAMP,
			verified_by INTEGER REFERENCES users(user_id),
			notes TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		log.Printf("Warning: withdrawal_transfer_logs table may already exist: %v\n", err)
	}

	_, err = dbPool.Exec(ctx, `
		CREATE INDEX IF NOT EXISTS idx_platform_bank_active ON platform_bank_accounts(is_active) WHERE is_active = true;
		CREATE INDEX IF NOT EXISTS idx_platform_bank_default ON platform_bank_accounts(is_default) WHERE is_default = true;
		CREATE INDEX IF NOT EXISTS idx_withdrawals_platform_bank ON withdrawals(platform_bank_account_id);
		CREATE INDEX IF NOT EXISTS idx_withdrawal_transfer_logs_withdrawal ON withdrawal_transfer_logs(withdrawal_id);
	`)
	if err != nil {
		log.Printf("Warning: Could not create platform bank indexes: %v\n", err)
	}

	// Insert default platform bank account (GOD)
	_, err = dbPool.Exec(ctx, `
		INSERT INTO platform_bank_accounts (
			bank_name, bank_code, account_number, account_name, account_type,
			branch_name, account_holder, is_active, is_default, owned_by, notes
		) VALUES (
			'ธนาคารกสิกรไทย', 'KBANK', 'XXX-X-XXXXX-X', 'บริษัท SkillMatch จำกัด', 'current',
			'สาขาสีลม', 'นาย GOD Master', true, true, 1,
			'บัญชีธนาคารหลักของแพลตฟอร์ม ใช้สำหรับโอนเงินให้ providers ทั้งหมด'
		) ON CONFLICT (account_number) DO NOTHING;
	`)
	if err != nil {
		log.Printf("Warning: Could not insert default platform bank account: %v\n", err)
	}

	fmt.Println("✅ Migration 016 completed!")

	// --- Migration 017: GOD Commission Tracking ---
	fmt.Println("🔄 Running Migration 017: GOD Commission Tracking...")

	_, err = dbPool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS god_commission_balance (
			balance_id SERIAL PRIMARY KEY,
			god_user_id INTEGER NOT NULL REFERENCES users(user_id),
			platform_bank_account_id INTEGER NOT NULL REFERENCES platform_bank_accounts(platform_bank_id),
			total_commission_collected DECIMAL(12, 2) DEFAULT 0.00 NOT NULL,
			total_transferred DECIMAL(12, 2) DEFAULT 0.00 NOT NULL,
			current_balance DECIMAL(12, 2) DEFAULT 0.00 NOT NULL,
			total_withdrawals_processed INTEGER DEFAULT 0,
			average_withdrawal_amount DECIMAL(12, 2) DEFAULT 0.00,
			last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(god_user_id, platform_bank_account_id),
			CONSTRAINT positive_balances CHECK (
				total_commission_collected >= 0 AND 
				total_transferred >= 0 AND 
				current_balance >= 0
			)
		);
	`)
	if err != nil {
		log.Printf("Warning: god_commission_balance table may already exist: %v\n", err)
	}

	_, err = dbPool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS commission_transactions (
			commission_txn_id SERIAL PRIMARY KEY,
			booking_id INTEGER REFERENCES bookings(booking_id),
			transaction_id INTEGER REFERENCES transactions(transaction_id),
			booking_amount DECIMAL(12, 2) NOT NULL,
			commission_rate DECIMAL(5, 4) DEFAULT 0.1000,
			commission_amount DECIMAL(12, 2) NOT NULL,
			provider_amount DECIMAL(12, 2) NOT NULL,
			provider_id INTEGER NOT NULL REFERENCES users(user_id),
			platform_bank_account_id INTEGER REFERENCES platform_bank_accounts(platform_bank_id),
			status VARCHAR(20) DEFAULT 'collected',
			collected_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			refunded_at TIMESTAMP,
			refund_reason TEXT,
			notes TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		log.Printf("Warning: commission_transactions table may already exist: %v\n", err)
	}

	_, err = dbPool.Exec(ctx, `
		ALTER TABLE withdrawals 
			ADD COLUMN IF NOT EXISTS original_slip_url TEXT,
			ADD COLUMN IF NOT EXISTS commission_withheld DECIMAL(12, 2) DEFAULT 0.00,
			ADD COLUMN IF NOT EXISTS notification_sent BOOLEAN DEFAULT false,
			ADD COLUMN IF NOT EXISTS email_sent BOOLEAN DEFAULT false;
	`)
	if err != nil {
		log.Printf("Warning: Could not add withdrawal notification columns: %v\n", err)
	}

	_, err = dbPool.Exec(ctx, `
		CREATE INDEX IF NOT EXISTS idx_god_commission_balance_user ON god_commission_balance(god_user_id);
		CREATE INDEX IF NOT EXISTS idx_commission_transactions_booking ON commission_transactions(booking_id);
		CREATE INDEX IF NOT EXISTS idx_commission_transactions_provider ON commission_transactions(provider_id);
	`)
	if err != nil {
		log.Printf("Warning: Could not create commission indexes: %v\n", err)
	}

	// Initialize GOD commission balance
	_, err = dbPool.Exec(ctx, `
		INSERT INTO god_commission_balance (
			god_user_id, platform_bank_account_id,
			total_commission_collected, total_transferred, current_balance
		)
		SELECT 1, platform_bank_id, 0.00, 0.00, 0.00
		FROM platform_bank_accounts
		WHERE is_default = true AND is_active = true
		LIMIT 1
		ON CONFLICT (god_user_id, platform_bank_account_id) DO NOTHING;
	`)
	if err != nil {
		log.Printf("Warning: Could not initialize GOD commission balance: %v\n", err)
	}

	fmt.Println("✅ Migration 017 completed!")

	// --- Migration 018: Update Fee Structure 12.75% ---
	fmt.Println("🔄 Running Migration 018: Fee Structure 12.75%...")

	_, err = dbPool.Exec(ctx, `
		ALTER TABLE commission_rules
			ADD COLUMN IF NOT EXISTS total_rate DECIMAL(5, 4) 
				GENERATED ALWAYS AS (platform_rate + payment_gateway_rate) STORED;
	`)
	if err != nil {
		log.Printf("Warning: Could not add total_rate column: %v\n", err)
	}

	_, err = dbPool.Exec(ctx, `
		UPDATE commission_rules
		SET 
			platform_rate = 0.1000,
			payment_gateway_rate = 0.0275,
			description = 'Total fee: 12.75% (Platform 10% + Payment Gateway 2.75%) - Provider receives 87.25%',
			name = 'Default Fee Structure',
			updated_at = CURRENT_TIMESTAMP
		WHERE rule_id = 1;
	`)
	if err != nil {
		log.Printf("Warning: Could not update commission rules: %v\n", err)
	}

	_, err = dbPool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS provider_fee_notifications (
			notification_id SERIAL PRIMARY KEY,
			provider_id INTEGER NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			platform_rate DECIMAL(5, 4) NOT NULL,
			payment_gateway_rate DECIMAL(5, 4) NOT NULL,
			total_rate DECIMAL(5, 4) NOT NULL,
			notification_type VARCHAR(50) NOT NULL,
			shown_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			acknowledged BOOLEAN DEFAULT false,
			acknowledged_at TIMESTAMP,
			notification_channel VARCHAR(50),
			notes TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		log.Printf("Warning: provider_fee_notifications table may already exist: %v\n", err)
	}

	_, err = dbPool.Exec(ctx, `
		ALTER TABLE transactions
			ADD COLUMN IF NOT EXISTS stripe_fee DECIMAL(12, 2) DEFAULT 0.00,
			ADD COLUMN IF NOT EXISTS platform_commission DECIMAL(12, 2) DEFAULT 0.00,
			ADD COLUMN IF NOT EXISTS total_fee_percentage DECIMAL(5, 4) DEFAULT 0.1275;
	`)
	if err != nil {
		log.Printf("Warning: Could not add fee columns to transactions: %v\n", err)
	}

	_, err = dbPool.Exec(ctx, `
		CREATE INDEX IF NOT EXISTS idx_provider_fee_notifications_provider ON provider_fee_notifications(provider_id);
		CREATE INDEX IF NOT EXISTS idx_provider_fee_notifications_acknowledged ON provider_fee_notifications(acknowledged) 
			WHERE acknowledged = false;
	`)
	if err != nil {
		log.Printf("Warning: Could not create fee notification indexes: %v\n", err)
	}

	// Create helper function for fee calculation
	_, err = dbPool.Exec(ctx, `
		CREATE OR REPLACE FUNCTION calculate_provider_earning(booking_amount DECIMAL)
		RETURNS TABLE (
			gross_amount DECIMAL,
			stripe_fee DECIMAL,
			platform_commission DECIMAL,
			total_fee DECIMAL,
			net_amount DECIMAL,
			provider_percentage DECIMAL
		) AS $$
		BEGIN
			RETURN QUERY
			SELECT 
				booking_amount,
				ROUND(booking_amount * 0.0275, 2),
				ROUND(booking_amount * 0.1000, 2),
				ROUND(booking_amount * 0.1275, 2),
				ROUND(booking_amount * 0.8725, 2),
				87.25;
		END;
		$$ LANGUAGE plpgsql;
	`)
	if err != nil {
		log.Printf("Warning: Could not create calculate_provider_earning function: %v\n", err)
	}

	fmt.Println("✅ Migration 018 completed!")

	// --- Migration 019: Provider Schedules/Calendar System ---
	fmt.Println("🔄 Running Migration 019: Provider Schedules System...")

	_, err = dbPool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS provider_schedules (
			schedule_id SERIAL PRIMARY KEY,
			provider_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			booking_id INT REFERENCES bookings(booking_id) ON DELETE SET NULL,
			
			-- Time slot
			start_time TIMESTAMP NOT NULL,
			end_time TIMESTAMP NOT NULL,
			
			-- Status: available (free slot), booked (has booking), blocked (unavailable)
			status VARCHAR(20) NOT NULL DEFAULT 'available' CHECK (status IN ('available', 'booked', 'blocked')),
			
			-- Location details (where provider will be)
			location_type VARCHAR(20) CHECK (location_type IN ('Incall', 'Outcall', 'Both')),
			location_address TEXT,
			location_province VARCHAR(100),
			location_district VARCHAR(100),
			latitude DECIMAL(10, 8),
			longitude DECIMAL(11, 8),
			
			-- Additional info
			notes TEXT, -- Provider's notes (e.g., "At spa", "Available for outcall only")
			
			-- Admin/GOD visibility
			is_visible_to_admin BOOLEAN DEFAULT TRUE,
			
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		log.Printf("Warning: Could not create provider_schedules table: %v\n", err)
	}

	// Indexes for performance
	_, err = dbPool.Exec(ctx, `
		CREATE INDEX IF NOT EXISTS idx_schedules_provider ON provider_schedules(provider_id);
		CREATE INDEX IF NOT EXISTS idx_schedules_time ON provider_schedules(start_time, end_time);
		CREATE INDEX IF NOT EXISTS idx_schedules_status ON provider_schedules(status);
		CREATE INDEX IF NOT EXISTS idx_schedules_booking ON provider_schedules(booking_id);
	`)
	if err != nil {
		log.Printf("Warning: Could not create schedule indexes: %v\n", err)
	}

	// Trigger to auto-update updated_at
	_, err = dbPool.Exec(ctx, `
		CREATE OR REPLACE FUNCTION update_schedule_timestamp()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = CURRENT_TIMESTAMP;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;

		DROP TRIGGER IF EXISTS trigger_update_schedule_timestamp ON provider_schedules;
		CREATE TRIGGER trigger_update_schedule_timestamp
		BEFORE UPDATE ON provider_schedules
		FOR EACH ROW
		EXECUTE FUNCTION update_schedule_timestamp();
	`)
	if err != nil {
		log.Printf("Warning: Could not create schedule update trigger: %v\n", err)
	}

	fmt.Println("✅ Migration 019 completed!")

	// --- Migration 020: Face Verification System ---
	migrationSQL, err := os.ReadFile("docs/sql-migrations/020_add_face_verification.sql")
	if err != nil {
		log.Printf("⚠️  Could not read migration 020 file: %v\n", err)
	} else {
		_, err = dbPool.Exec(ctx, string(migrationSQL))
		if err != nil {
			log.Printf("Warning: Migration 020 error: %v\n", err)
		} else {
			fmt.Println("✅ Migration 020: Face Verification System completed!")
		}
	}

	// --- Migration 021: Add Passport Support for Face Verification ---
	migration021SQL, err := os.ReadFile("docs/sql-migrations/021_add_passport_support_face_verification.sql")
	if err != nil {
		log.Printf("⚠️  Could not read migration 021 file: %v\n", err)
	} else {
		_, err = dbPool.Exec(ctx, string(migration021SQL))
		if err != nil {
			log.Printf("Warning: Migration 021 error: %v\n", err)
		} else {
			fmt.Println("✅ Migration 021: Passport Support for Face Verification completed!")
		}
	}

	// --- Migration 022: User Type Separation ---
	fmt.Println("🔄 Running Migration 022: User Type Separation...")
	migration022SQL, err := os.ReadFile("docs/sql-migrations/022_add_user_type_separation.sql")
	if err != nil {
		log.Printf("⚠️  Could not read migration 022 file: %v\n", err)
	} else {
		_, err = dbPool.Exec(ctx, string(migration022SQL))
		if err != nil {
			log.Printf("Warning: Migration 022 error: %v\n", err)
		} else {
			fmt.Println("✅ Migration 022: User Type Separation completed!")
		}
	}

	// --- Migration 023: Service Categories ---
	fmt.Println("🔄 Running Migration 023: Service Categories...")
	_, err = dbPool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS service_categories (
			category_id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL UNIQUE,
			name_thai VARCHAR(100),
			icon VARCHAR(50),
			description TEXT,
			is_adult BOOLEAN DEFAULT false,
			display_order INT DEFAULT 0,
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);

		INSERT INTO service_categories (name, name_thai, icon, description, is_adult, display_order) VALUES
		('Escort', 'เด็กเอน', '💃', 'Escort and companion services', true, 1),
		('Bar Attendant', 'เด็กชงเหล้า', '🍸', 'Bar service and entertainment', true, 2),
		('Adult Services', 'บริการผู้ใหญ่', '💋', 'Adult entertainment services', true, 3),
		('Spa & Bath', 'อาบน้ำ/สปา', '🛁', 'Spa and bathing services', true, 4),
		('Dining Companion', 'ทานข้าว', '🍽️', 'Dining and meal companion services', true, 5),
		('Movie Companion', 'ดูหนัง', '🎬', 'Movie and entertainment companion', true, 6)
		ON CONFLICT (name) DO NOTHING;
	`)
	if err != nil {
		log.Printf("Warning: Migration 023 error: %v\n", err)
	} else {
		fmt.Println("✅ Migration 023: Service Categories completed!")
	}

	// --- Migration 024: Provider Categories Junction ---
	fmt.Println("🔄 Running Migration 024: Provider Categories...")
	_, err = dbPool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS provider_categories (
			provider_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			category_id INT NOT NULL REFERENCES service_categories(category_id) ON DELETE CASCADE,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			PRIMARY KEY (provider_id, category_id)
		);
	`)
	if err != nil {
		log.Printf("Warning: Migration 024 error: %v\n", err)
	} else {
		fmt.Println("✅ Migration 024: Provider Categories completed!")
	}

	// --- Migration 025: Conversations & Messages ---
	fmt.Println("🔄 Running Migration 025: Messaging System...")
	_, err = dbPool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS conversations (
			conversation_id SERIAL PRIMARY KEY,
			user1_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			user2_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			CONSTRAINT user_order CHECK (user1_id < user2_id),
			UNIQUE (user1_id, user2_id)
		);

		CREATE TABLE IF NOT EXISTS messages (
			message_id SERIAL PRIMARY KEY,
			conversation_id INT NOT NULL REFERENCES conversations(conversation_id) ON DELETE CASCADE,
			sender_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			content TEXT NOT NULL,
			is_read BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_messages_conversation ON messages(conversation_id);
		CREATE INDEX IF NOT EXISTS idx_messages_sender ON messages(sender_id);
	`)
	if err != nil {
		log.Printf("Warning: Migration 025 error: %v\n", err)
	} else {
		fmt.Println("✅ Migration 025: Messaging System completed!")
	}

	// --- Migration 026: Notifications ---
	fmt.Println("🔄 Running Migration 026: Notifications...")
	_, err = dbPool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS notifications (
			notification_id SERIAL PRIMARY KEY,
			user_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			type VARCHAR(50) NOT NULL,
			title VARCHAR(255) NOT NULL,
			message TEXT NOT NULL,
			is_read BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_notifications_user ON notifications(user_id);
		CREATE INDEX IF NOT EXISTS idx_notifications_unread ON notifications(user_id, is_read);
	`)
	if err != nil {
		log.Printf("Warning: Migration 026 error: %v\n", err)
	} else {
		fmt.Println("✅ Migration 026: Notifications completed!")
	}

	// --- Migration 027: Blocks System ---
	fmt.Println("🔄 Running Migration 027: Blocks System...")
	_, err = dbPool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS blocks (
			block_id SERIAL PRIMARY KEY,
			blocker_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			blocked_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			UNIQUE (blocker_id, blocked_id)
		);

		CREATE INDEX IF NOT EXISTS idx_blocks_blocker ON blocks(blocker_id);
		CREATE INDEX IF NOT EXISTS idx_blocks_blocked ON blocks(blocked_id);
	`)
	if err != nil {
		log.Printf("Warning: Migration 027 error: %v\n", err)
	} else {
		fmt.Println("✅ Migration 027: Blocks System completed!")
	}

	// --- Migration 028: Reports System ---
	fmt.Println("🔄 Running Migration 028: Reports System...")
	_, err = dbPool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS reports (
			report_id SERIAL PRIMARY KEY,
			reporter_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			reported_user_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			reason VARCHAR(255) NOT NULL,
			description TEXT,
			status VARCHAR(50) DEFAULT 'pending',
			created_at TIMESTAMPTZ DEFAULT NOW(),
			resolved_at TIMESTAMPTZ,
			admin_notes TEXT
		);

		CREATE INDEX IF NOT EXISTS idx_reports_status ON reports(status);
	`)
	if err != nil {
		log.Printf("Warning: Migration 028 error: %v\n", err)
	} else {
		fmt.Println("✅ Migration 028: Reports System completed!")
	}

	// --- Migration 029: Financial System ---
	fmt.Println("🔄 Running Migration 029: Financial System...")
	_, err = dbPool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS wallets (
			wallet_id SERIAL PRIMARY KEY,
			user_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE UNIQUE,
			available_balance DECIMAL(10, 2) DEFAULT 0.00,
			pending_balance DECIMAL(10, 2) DEFAULT 0.00,
			total_earned DECIMAL(10, 2) DEFAULT 0.00,
			total_withdrawn DECIMAL(10, 2) DEFAULT 0.00,
			updated_at TIMESTAMPTZ DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS transactions (
			transaction_id SERIAL PRIMARY KEY,
			wallet_id INT NOT NULL REFERENCES wallets(wallet_id) ON DELETE CASCADE,
			booking_id INT REFERENCES bookings(booking_id),
			type VARCHAR(50) NOT NULL,
			amount DECIMAL(10, 2) NOT NULL,
			status VARCHAR(50) DEFAULT 'pending',
			stripe_transaction_id VARCHAR(255),
			platform_fee DECIMAL(10, 2) DEFAULT 0.00,
			stripe_fee DECIMAL(10, 2) DEFAULT 0.00,
			net_amount DECIMAL(10, 2) NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS bank_accounts (
			bank_account_id SERIAL PRIMARY KEY,
			user_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			bank_name VARCHAR(100) NOT NULL,
			account_number VARCHAR(50) NOT NULL,
			account_holder_name VARCHAR(255) NOT NULL,
			is_default BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS withdrawal_requests (
			withdrawal_id SERIAL PRIMARY KEY,
			wallet_id INT NOT NULL REFERENCES wallets(wallet_id) ON DELETE CASCADE,
			bank_account_id INT NOT NULL REFERENCES bank_accounts(bank_account_id),
			amount DECIMAL(10, 2) NOT NULL,
			status VARCHAR(50) DEFAULT 'pending',
			requested_at TIMESTAMPTZ DEFAULT NOW(),
			approved_at TIMESTAMPTZ,
			rejected_at TIMESTAMPTZ,
			completed_at TIMESTAMPTZ,
			admin_notes TEXT,
			transfer_slip_url TEXT
		);

		CREATE INDEX IF NOT EXISTS idx_transactions_wallet ON transactions(wallet_id);
		CREATE INDEX IF NOT EXISTS idx_withdrawal_status ON withdrawal_requests(status);
	`)
	if err != nil {
		log.Printf("Warning: Migration 029 error: %v\n", err)
	} else {
		fmt.Println("✅ Migration 029: Financial System completed!")
	}

	// --- Migration 030: Provider Documents ---
	fmt.Println("🔄 Running Migration 030: Provider Documents...")
	_, err = dbPool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS provider_documents (
			document_id SERIAL PRIMARY KEY,
			user_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			document_type VARCHAR(50) NOT NULL,
			document_url TEXT NOT NULL,
			verification_status VARCHAR(50) DEFAULT 'pending',
			uploaded_at TIMESTAMPTZ DEFAULT NOW(),
			verified_at TIMESTAMPTZ,
			admin_notes TEXT
		);

		CREATE INDEX IF NOT EXISTS idx_provider_docs_user ON provider_documents(user_id);
	`)
	if err != nil {
		log.Printf("Warning: Migration 030 error: %v\n", err)
	} else {
		fmt.Println("✅ Migration 030: Provider Documents completed!")
	}

	// --- Migration 031: Provider Tier History ---
	fmt.Println("🔄 Running Migration 031: Provider Tier History...")
	_, err = dbPool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS provider_tier_history (
			history_id SERIAL PRIMARY KEY,
			provider_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			old_tier_id INT REFERENCES tiers(tier_id),
			new_tier_id INT NOT NULL REFERENCES tiers(tier_id),
			changed_at TIMESTAMPTZ DEFAULT NOW(),
			reason TEXT
		);

		CREATE INDEX IF NOT EXISTS idx_tier_history_provider ON provider_tier_history(provider_id);
	`)
	if err != nil {
		log.Printf("Warning: Migration 031 error: %v\n", err)
	} else {
		fmt.Println("✅ Migration 031: Provider Tier History completed!")
	}

	// --- Migration 032: Fix Service Categories Schema ---
	fmt.Println("🔄 Running Migration 032: Fix Service Categories Schema...")
	_, err = dbPool.Exec(ctx, `
		-- Add missing columns to service_categories
		ALTER TABLE service_categories 
			ADD COLUMN IF NOT EXISTS name_thai VARCHAR(100),
			ADD COLUMN IF NOT EXISTS is_adult BOOLEAN DEFAULT false,
			ADD COLUMN IF NOT EXISTS display_order INT DEFAULT 0,
			ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true;

		-- Update existing records with Thai names, display order, and adult flag
		UPDATE service_categories SET name_thai = 'เด็กเอน', is_adult = true, display_order = 1 WHERE name = 'Escort';
		UPDATE service_categories SET name_thai = 'เด็กชงเหล้า', is_adult = true, display_order = 2 WHERE name = 'Bar Attendant';
		UPDATE service_categories SET name_thai = 'บริการผู้ใหญ่', is_adult = true, display_order = 3 WHERE name = 'Adult Services';
		UPDATE service_categories SET name_thai = 'อาบน้ำ/สปา', is_adult = true, display_order = 4 WHERE name = 'Spa & Bath';
		UPDATE service_categories SET name_thai = 'ทานข้าว', is_adult = true, display_order = 5 WHERE name = 'Dining Companion';
		UPDATE service_categories SET name_thai = 'ดูหนัง', is_adult = true, display_order = 6 WHERE name = 'Movie Companion';
	`)
	if err != nil {
		log.Printf("Warning: Migration 032 error: %v\n", err)
	} else {
		fmt.Println("✅ Migration 032: Fix Service Categories Schema completed!")
	}

	// --- Migration 033: Add profile_picture_url to users table ---
	fmt.Println("🔄 Running Migration 033: Add profile_picture_url to users...")
	_, err = dbPool.Exec(ctx, `
		ALTER TABLE users 
			ADD COLUMN IF NOT EXISTS profile_picture_url TEXT;
		
		CREATE INDEX IF NOT EXISTS idx_users_profile_picture ON users(profile_picture_url) WHERE profile_picture_url IS NOT NULL;
	`)
	if err != nil {
		log.Printf("Warning: Migration 033 error: %v\n", err)
	} else {
		fmt.Println("✅ Migration 033: Add profile_picture_url completed!")
	}

	// --- Migration 034: Safety & Business Features ---
	fmt.Println("🔄 Running Migration 034: Safety & Business Features...")
	migration034SQL, err := os.ReadFile("docs/sql-migrations/034_safety_business_features.sql")
	if err != nil {
		log.Printf("⚠️  Could not read migration 034 file: %v\n", err)
	} else {
		_, err = dbPool.Exec(ctx, string(migration034SQL))
		if err != nil {
			log.Printf("Warning: Migration 034 error: %v\n", err)
		} else {
			fmt.Println("✅ Migration 034: Safety & Business Features completed!")
		}
	}

	fmt.Println("✅ All Database Migrations สำเร็จ!")
}
