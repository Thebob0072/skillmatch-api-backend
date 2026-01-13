-- Seed Test Providers for SkillMatch Platform
-- ข้อมูลทดสอบสำหรับ providers หลายคน

-- ==============================================
-- 1. สร้าง Test Users (Providers)
-- ==============================================

INSERT INTO users (username, email, password_hash, gender_id, first_name, last_name, tier_id, provider_level_id, verification_status, phone_number)
VALUES
  -- Provider 1: Female, Diamond Tier
  ('bella_bangkok', 'bella@example.com', '', 2, 'Bella', 'Charm', 1, 3, 'verified', '0812345001'),
  
  -- Provider 2: Female, Premium Tier
  ('sophia_elite', 'sophia@example.com', '', 2, 'Sophia', 'Rose', 1, 4, 'verified', '0812345002'),
  
  -- Provider 3: Female, Silver Tier
  ('mia_sweetheart', 'mia@example.com', '', 2, 'Mia', 'Love', 1, 2, 'verified', '0812345003'),
  
  -- Provider 4: Male, Diamond Tier
  ('james_gentleman', 'james@example.com', '', 1, 'James', 'Knight', 1, 3, 'verified', '0812345004'),
  
  -- Provider 5: Female, Premium Tier
  ('luna_goddess', 'luna@example.com', '', 2, 'Luna', 'Divine', 1, 4, 'verified', '0812345005'),
  
  -- Provider 6: Female, Silver Tier
  ('amber_sweet', 'amber@example.com', '', 2, 'Amber', 'Joy', 1, 2, 'verified', '0812345006'),
  
  -- Provider 7: Male, Silver Tier
  ('ryan_charmer', 'ryan@example.com', '', 1, 'Ryan', 'Steel', 1, 2, 'verified', '0812345007'),
  
  -- Provider 8: Female, Diamond Tier
  ('jade_oriental', 'jade@example.com', '', 2, 'Jade', 'Pearl', 1, 3, 'verified', '0812345008'),
  
  -- Provider 9: Female, Premium Tier
  ('crystal_luxury', 'crystal@example.com', '', 2, 'Crystal', 'Diamond', 1, 4, 'verified', '0812345009'),
  
  -- Provider 10: Male, Diamond Tier
  ('alex_phoenix', 'alex@example.com', '', 1, 'Alex', 'Phoenix', 1, 3, 'verified', '0812345010')
ON CONFLICT (email) DO NOTHING;

-- ==============================================
-- 2. สร้าง User Profiles (ข้อมูลโปรไฟล์)
-- ==============================================

INSERT INTO user_profiles (
  user_id, bio, location, skills, profile_image_url, 
  age, height, weight, ethnicity, languages, 
  working_hours, is_available, service_type
)
VALUES
  -- Provider 1: Bella Bangkok
  (
    (SELECT user_id FROM users WHERE email = 'bella@example.com'),
    'Elegant companion for sophisticated gentlemen. Fluent in English & Thai. Available for dinner dates and social events.',
    'Sukhumvit, Bangkok',
    ARRAY['Dinner companion', 'Social events', 'Travel companion'],
    'https://images.unsplash.com/photo-1494790108377-be9c29b29330',
    25, 165, 52, 'Thai', ARRAY['Thai', 'English'],
    'Evening 6PM-2AM', true, 'Both'
  ),
  
  -- Provider 2: Sophia Elite
  (
    (SELECT user_id FROM users WHERE email = 'sophia@example.com'),
    'Premium escort service. Exclusive companionship for VIP clients. Multilingual and well-traveled.',
    'Sathorn, Bangkok',
    ARRAY['VIP escort', 'International travel', 'Business events'],
    'https://images.unsplash.com/photo-1534528741775-53994a69daeb',
    28, 170, 55, 'Thai-European', ARRAY['Thai', 'English', 'Japanese'],
    '24/7 available', true, 'Outcall'
  ),
  
  -- Provider 3: Mia Sweetheart
  (
    (SELECT user_id FROM users WHERE email = 'mia@example.com'),
    'Sweet and friendly companion. Perfect for casual dates and relaxing evenings.',
    'Thonglor, Bangkok',
    ARRAY['Casual dating', 'Massage', 'Movie companion'],
    'https://images.unsplash.com/photo-1524504388940-b1c1722653e1',
    22, 160, 48, 'Thai', ARRAY['Thai', 'English'],
    'Afternoon 2PM-10PM', true, 'Incall'
  ),
  
  -- Provider 4: James Gentleman
  (
    (SELECT user_id FROM users WHERE email = 'james@example.com'),
    'Professional male escort. Sophisticated companion for ladies and couples.',
    'Asok, Bangkok',
    ARRAY['Gentleman escort', 'Dinner dates', 'Travel companion'],
    'https://images.unsplash.com/photo-1506794778202-cad84cf45f1d',
    30, 182, 78, 'Thai', ARRAY['Thai', 'English'],
    'Evening 7PM-1AM', true, 'Both'
  ),
  
  -- Provider 5: Luna Goddess
  (
    (SELECT user_id FROM users WHERE email = 'luna@example.com'),
    'Elite companion with model experience. Stunning beauty and engaging personality.',
    'Siam, Bangkok',
    ARRAY['Fashion events', 'High-end escort', 'Photo shoots'],
    'https://images.unsplash.com/photo-1529626455594-4ff0802cfb7e',
    26, 175, 56, 'Thai-Korean', ARRAY['Thai', 'English', 'Korean'],
    '24/7 by appointment', true, 'Outcall'
  ),
  
  -- Provider 6: Amber Sweet
  (
    (SELECT user_id FROM users WHERE email = 'amber@example.com'),
    'Warm and caring companion. Specializing in relaxing massages and friendly dates.',
    'Ekkamai, Bangkok',
    ARRAY['Thai massage', 'Aromatherapy', 'Casual companion'],
    'https://images.unsplash.com/photo-1438761681033-6461ffad8d80',
    24, 162, 50, 'Thai', ARRAY['Thai'],
    'Daily 1PM-11PM', true, 'Incall'
  ),
  
  -- Provider 7: Ryan Charmer
  (
    (SELECT user_id FROM users WHERE email = 'ryan@example.com'),
    'Friendly male companion. Great conversation and fun personality.',
    'Ratchada, Bangkok',
    ARRAY['Social events', 'Casual dating', 'Sports companion'],
    'https://images.unsplash.com/photo-1500648767791-00dcc994a43e',
    27, 178, 75, 'Thai', ARRAY['Thai', 'English'],
    'Evening 6PM-12AM', true, 'Both'
  ),
  
  -- Provider 8: Jade Oriental
  (
    (SELECT user_id FROM users WHERE email = 'jade@example.com'),
    'Graceful and sophisticated. Traditional Thai beauty with modern elegance.',
    'Silom, Bangkok',
    ARRAY['Traditional Thai massage', 'Dinner companion', 'Cultural guide'],
    'https://images.unsplash.com/photo-1517841905240-472988babdf9',
    23, 163, 51, 'Thai', ARRAY['Thai', 'English', 'Chinese'],
    'Afternoon 3PM-11PM', true, 'Both'
  ),
  
  -- Provider 9: Crystal Luxury
  (
    (SELECT user_id FROM users WHERE email = 'crystal@example.com'),
    'Ultra-premium companionship. For discerning gentlemen seeking perfection.',
    'Chidlom, Bangkok',
    ARRAY['Luxury travel', 'Fine dining', 'Yacht parties'],
    'https://images.unsplash.com/photo-1488426862026-3ee34a7d66df',
    29, 168, 54, 'Thai-Japanese', ARRAY['Thai', 'English', 'Japanese', 'Chinese'],
    'By appointment only', true, 'Outcall'
  ),
  
  -- Provider 10: Alex Phoenix
  (
    (SELECT user_id FROM users WHERE email = 'alex@example.com'),
    'Distinguished male escort. Perfect companion for elegant ladies.',
    'Ploenchit, Bangkok',
    ARRAY['VIP escort', 'Business events', 'International travel'],
    'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d',
    32, 185, 82, 'Thai', ARRAY['Thai', 'English', 'French'],
    'Evening 7PM-2AM', true, 'Both'
  )
ON CONFLICT (user_id) DO NOTHING;

-- ==============================================
-- 3. เพิ่ม Provider Categories
-- ==============================================

INSERT INTO provider_categories (provider_id, category_id)
VALUES
  -- Bella: Escort + Dinner Date
  ((SELECT user_id FROM users WHERE email = 'bella@example.com'), 1),
  ((SELECT user_id FROM users WHERE email = 'bella@example.com'), 4),
  
  -- Sophia: Escort + Entertainment + Model
  ((SELECT user_id FROM users WHERE email = 'sophia@example.com'), 1),
  ((SELECT user_id FROM users WHERE email = 'sophia@example.com'), 3),
  ((SELECT user_id FROM users WHERE email = 'sophia@example.com'), 6),
  
  -- Mia: Massage + Dinner Date
  ((SELECT user_id FROM users WHERE email = 'mia@example.com'), 2),
  ((SELECT user_id FROM users WHERE email = 'mia@example.com'), 4),
  
  -- James: Escort + Tour Guide
  ((SELECT user_id FROM users WHERE email = 'james@example.com'), 1),
  ((SELECT user_id FROM users WHERE email = 'james@example.com'), 5),
  
  -- Luna: Escort + Model + Entertainment
  ((SELECT user_id FROM users WHERE email = 'luna@example.com'), 1),
  ((SELECT user_id FROM users WHERE email = 'luna@example.com'), 3),
  ((SELECT user_id FROM users WHERE email = 'luna@example.com'), 6),
  
  -- Amber: Massage + Dinner Date
  ((SELECT user_id FROM users WHERE email = 'amber@example.com'), 2),
  ((SELECT user_id FROM users WHERE email = 'amber@example.com'), 4),
  
  -- Ryan: Entertainment + Tour Guide
  ((SELECT user_id FROM users WHERE email = 'ryan@example.com'), 3),
  ((SELECT user_id FROM users WHERE email = 'ryan@example.com'), 5),
  
  -- Jade: Massage + Tour Guide + Escort
  ((SELECT user_id FROM users WHERE email = 'jade@example.com'), 1),
  ((SELECT user_id FROM users WHERE email = 'jade@example.com'), 2),
  ((SELECT user_id FROM users WHERE email = 'jade@example.com'), 5),
  
  -- Crystal: Escort + Model
  ((SELECT user_id FROM users WHERE email = 'crystal@example.com'), 1),
  ((SELECT user_id FROM users WHERE email = 'crystal@example.com'), 6),
  
  -- Alex: Escort + Entertainment
  ((SELECT user_id FROM users WHERE email = 'alex@example.com'), 1),
  ((SELECT user_id FROM users WHERE email = 'alex@example.com'), 3)
ON CONFLICT DO NOTHING;

-- ==============================================
-- 4. เพิ่ม Service Packages
-- ==============================================

INSERT INTO service_packages (provider_id, package_name, description, duration, price, is_active)
VALUES
  -- Bella's packages
  ((SELECT user_id FROM users WHERE email = 'bella@example.com'), '1 Hour Dinner Date', 'Elegant dinner companion', 60, 3000.00, true),
  ((SELECT user_id FROM users WHERE email = 'bella@example.com'), '3 Hours Evening', 'Extended evening companion', 180, 7500.00, true),
  ((SELECT user_id FROM users WHERE email = 'bella@example.com'), 'Overnight Package', 'Full night companionship', 720, 20000.00, true),
  
  -- Sophia's packages
  ((SELECT user_id FROM users WHERE email = 'sophia@example.com'), '2 Hours VIP', 'Premium companionship', 120, 8000.00, true),
  ((SELECT user_id FROM users WHERE email = 'sophia@example.com'), 'Half Day Exclusive', 'Exclusive 6-hour package', 360, 25000.00, true),
  ((SELECT user_id FROM users WHERE email = 'sophia@example.com'), 'Overnight Elite', 'Elite overnight service', 720, 50000.00, true),
  
  -- Mia's packages
  ((SELECT user_id FROM users WHERE email = 'mia@example.com'), '1 Hour Massage', 'Relaxing Thai massage', 60, 1500.00, true),
  ((SELECT user_id FROM users WHERE email = 'mia@example.com'), '2 Hours Date', 'Casual date package', 120, 2500.00, true),
  
  -- James's packages
  ((SELECT user_id FROM users WHERE email = 'james@example.com'), '2 Hours Gentleman', 'Professional escort service', 120, 4000.00, true),
  ((SELECT user_id FROM users WHERE email = 'james@example.com'), 'Evening Package', 'Full evening companion', 240, 9000.00, true),
  
  -- Luna's packages
  ((SELECT user_id FROM users WHERE email = 'luna@example.com'), '2 Hours Elite', 'Elite model companion', 120, 10000.00, true),
  ((SELECT user_id FROM users WHERE email = 'luna@example.com'), 'Full Day VIP', 'Full day exclusive', 480, 40000.00, true),
  
  -- Amber's packages
  ((SELECT user_id FROM users WHERE email = 'amber@example.com'), '1 Hour Thai Massage', 'Traditional Thai massage', 60, 1200.00, true),
  ((SELECT user_id FROM users WHERE email = 'amber@example.com'), '2 Hours Relaxation', 'Massage + companionship', 120, 2000.00, true),
  
  -- Ryan's packages
  ((SELECT user_id FROM users WHERE email = 'ryan@example.com'), '2 Hours Fun', 'Fun companion service', 120, 2500.00, true),
  ((SELECT user_id FROM users WHERE email = 'ryan@example.com'), 'Evening Entertainment', 'Full evening package', 240, 5000.00, true),
  
  -- Jade's packages
  ((SELECT user_id FROM users WHERE email = 'jade@example.com'), '1 Hour Massage', 'Thai traditional massage', 60, 1800.00, true),
  ((SELECT user_id FROM users WHERE email = 'jade@example.com'), '3 Hours Package', 'Massage + dinner companion', 180, 4500.00, true),
  
  -- Crystal's packages
  ((SELECT user_id FROM users WHERE email = 'crystal@example.com'), '3 Hours Luxury', 'Luxury companionship', 180, 15000.00, true),
  ((SELECT user_id FROM users WHERE email = 'crystal@example.com'), 'Full Day Premium', 'Premium full day service', 480, 50000.00, true),
  
  -- Alex's packages
  ((SELECT user_id FROM users WHERE email = 'alex@example.com'), '2 Hours Elite', 'Elite gentleman service', 120, 5000.00, true),
  ((SELECT user_id FROM users WHERE email = 'alex@example.com'), 'Evening VIP', 'VIP evening package', 300, 12000.00, true);

-- ==============================================
-- 5. เพิ่ม Sample Reviews (สร้าง fake bookings ก่อน)
-- ==============================================

-- สร้าง fake bookings จาก admin user (user_id = 1)
DO $$
DECLARE
  admin_id INT := 1;
  provider_ids INT[];
  provider_id INT;
  package_id INT;
BEGIN
  -- Get all provider IDs
  SELECT ARRAY_AGG(user_id) INTO provider_ids 
  FROM users 
  WHERE email LIKE '%@example.com' AND provider_level_id > 1;

  -- สร้าง bookings และ reviews สำหรับแต่ละ provider
  FOREACH provider_id IN ARRAY provider_ids
  LOOP
    -- Get first package of this provider
    SELECT sp.package_id INTO package_id
    FROM service_packages sp
    WHERE sp.provider_id = provider_id
    LIMIT 1;

    IF package_id IS NOT NULL THEN
      -- Insert booking
      INSERT INTO bookings (
        client_id, provider_id, package_id, 
        booking_date, start_time, end_time, 
        total_price, status, location
      ) VALUES (
        admin_id, provider_id, package_id,
        CURRENT_DATE - (random() * 30)::int,
        NOW() - INTERVAL '7 days',
        NOW() - INTERVAL '7 days' + INTERVAL '2 hours',
        (random() * 10000 + 2000)::decimal(10,2),
        'completed',
        'Bangkok'
      );

      -- Insert review
      INSERT INTO reviews (
        booking_id, client_id, provider_id,
        rating, comment, is_verified
      ) VALUES (
        currval('bookings_booking_id_seq'),
        admin_id, provider_id,
        (random() * 2 + 3)::int, -- Rating 3-5
        'Great service! Very professional and friendly.',
        true
      );
    END IF;
  END LOOP;
END $$;

-- ==============================================
-- 6. Summary
-- ==============================================

SELECT 
  'Providers Created' as action,
  COUNT(*) as count 
FROM users 
WHERE email LIKE '%@example.com';

SELECT 
  'Profiles Created' as action,
  COUNT(*) as count 
FROM user_profiles 
WHERE user_id IN (SELECT user_id FROM users WHERE email LIKE '%@example.com');

SELECT 
  'Service Packages' as action,
  COUNT(*) as count 
FROM service_packages 
WHERE provider_id IN (SELECT user_id FROM users WHERE email LIKE '%@example.com');

SELECT 
  'Reviews Created' as action,
  COUNT(*) as count 
FROM reviews 
WHERE provider_id IN (SELECT user_id FROM users WHERE email LIKE '%@example.com');

COMMIT;
