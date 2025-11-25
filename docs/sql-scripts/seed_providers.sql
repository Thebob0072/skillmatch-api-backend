-- Mock Data: Diverse Service Providers
-- Gender IDs: 1=Male, 2=Female, 3=LGBTQ+

-- ============================
-- USERS (Base accounts)
-- ============================

-- ผู้หญิง (Female Providers) - 5 คน
INSERT INTO users (username, email, password_hash, gender_id, tier_id, verification_status, google_profile_picture)
VALUES 
('bella_bangkok', 'bella@example.com', '$2a$10$dummyhash1', 2, 2, 'approved', 'https://i.pravatar.cc/300?img=1'),
('sophia_silom', 'sophia@example.com', '$2a$10$dummyhash2', 2, 3, 'approved', 'https://i.pravatar.cc/300?img=5'),
('maya_massage', 'maya@example.com', '$2a$10$dummyhash3', 2, 2, 'approved', 'https://i.pravatar.cc/300?img=9'),
('luna_therapy', 'luna@example.com', '$2a$10$dummyhash4', 2, 2, 'approved', 'https://i.pravatar.cc/300?img=10'),
('nina_wellness', 'nina@example.com', '$2a$10$dummyhash5', 2, 1, 'approved', 'https://i.pravatar.cc/300?img=16');

-- ผู้ชาย (Male Providers) - 5 คน
INSERT INTO users (username, email, password_hash, gender_id, tier_id, verification_status, google_profile_picture)
VALUES 
('marco_thai', 'marco@example.com', '$2a$10$dummyhash6', 1, 2, 'approved', 'https://i.pravatar.cc/300?img=12'),
('alex_sports', 'alex@example.com', '$2a$10$dummyhash7', 1, 3, 'approved', 'https://i.pravatar.cc/300?img=13'),
('david_fitness', 'david@example.com', '$2a$10$dummyhash8', 1, 2, 'approved', 'https://i.pravatar.cc/300?img=14'),
('ryan_wellness', 'ryan@example.com', '$2a$10$dummyhash9', 1, 1, 'approved', 'https://i.pravatar.cc/300?img=15'),
('jason_therapy', 'jason@example.com', '$2a$10$dummyhash10', 1, 2, 'approved', 'https://i.pravatar.cc/300?img=17');

-- Ladyboy Providers - 4 คน
INSERT INTO users (username, email, password_hash, gender_id, tier_id, verification_status, google_profile_picture)
VALUES 
('kim_beauty', 'kim@example.com', '$2a$10$dummyhash11', 3, 3, 'approved', 'https://i.pravatar.cc/300?img=20'),
('rose_glamour', 'rose@example.com', '$2a$10$dummyhash12', 3, 2, 'approved', 'https://i.pravatar.cc/300?img=21'),
('mimi_style', 'mimi@example.com', '$2a$10$dummyhash13', 3, 2, 'approved', 'https://i.pravatar.cc/300?img=22'),
('angel_paradise', 'angel@example.com', '$2a$10$dummyhash14', 3, 1, 'approved', 'https://i.pravatar.cc/300?img=23');

-- Gay Providers - 3 คน
INSERT INTO users (username, email, password_hash, gender_id, tier_id, verification_status, google_profile_picture)
VALUES 
('tony_pride', 'tony@example.com', '$2a$10$dummyhash15', 3, 2, 'approved', 'https://i.pravatar.cc/300?img=24'),
('kevin_rainbow', 'kevin@example.com', '$2a$10$dummyhash16', 3, 3, 'approved', 'https://i.pravatar.cc/300?img=25'),
('sam_fabulous', 'sam@example.com', '$2a$10$dummyhash17', 3, 2, 'approved', 'https://i.pravatar.cc/300?img=26');

-- ============================
-- USER PROFILES (Detailed info)
-- ============================

-- ผู้หญิง Profiles
INSERT INTO user_profiles (user_id, bio, age, height, weight, ethnicity, languages, working_hours, is_available, service_type, skills, province, district, sub_district, postal_code, address_line1, latitude, longitude)
VALUES 
((SELECT user_id FROM users WHERE username = 'bella_bangkok'), 'Professional massage therapist with 5 years experience. Specializing in Thai traditional and aromatherapy massage. 🌸', 25, 165, 50, 'Thai', ARRAY['Thai', 'English'], '10:00-22:00', true, 'both', ARRAY['Thai Massage', 'Aromatherapy', 'Spa'], 'กรุงเทพมหานคร', 'บางรัก', 'สีลม', '10500', '123 Silom Road', 13.7278, 100.5311),

((SELECT user_id FROM users WHERE username = 'sophia_silom'), 'VIP escort & companion. Educated, elegant, speaks 3 languages. Available for dinner dates and travel companionship. 💎', 28, 170, 52, 'Thai-Chinese', ARRAY['Thai', 'English', 'Chinese'], '18:00-02:00', true, 'outcall', ARRAY['Companion', 'Travel Partner', 'Events'], 'กรุงเทพมหานคร', 'ปทุมวัน', 'ลุมพินี', '10330', '456 Wireless Road', 13.7401, 100.5453),

((SELECT user_id FROM users WHERE username = 'maya_massage'), 'Certified spa therapist. Oil massage, body scrub, and relaxation specialist. Hotel outcall available. ✨', 23, 162, 48, 'Thai', ARRAY['Thai', 'English'], '12:00-00:00', true, 'both', ARRAY['Oil Massage', 'Body Scrub', 'Facial'], 'กรุงเทพมหานคร', 'วัฒนา', 'คลองเตย', '10110', '789 Sukhumvit Soi 11', 13.7378, 100.5569),

((SELECT user_id FROM users WHERE username = 'luna_therapy'), 'Holistic wellness provider. Yoga, meditation, and therapeutic massage. Perfect for stress relief. 🧘‍♀️', 26, 168, 54, 'Thai', ARRAY['Thai', 'English'], '09:00-21:00', true, 'both', ARRAY['Yoga', 'Meditation', 'Wellness Coaching'], 'กรุงเทพมหานคร', 'สาทร', 'ยานนาวา', '10120', '321 Sathorn Road', 13.7194, 100.5250),

((SELECT user_id FROM users WHERE username = 'nina_wellness'), 'Friendly massage therapist. Traditional Thai massage and foot reflexology. Great for relaxation after a long day. 💆‍♀️', 24, 160, 47, 'Thai', ARRAY['Thai'], '14:00-23:00', true, 'incall', ARRAY['Thai Massage', 'Foot Reflexology'], 'กรุงเทพมหานคร', 'ราชเทวี', 'ทุ่งพญาไท', '10400', '555 Phayathai Road', 13.7563, 100.5318);

-- ผู้ชาย Profiles
INSERT INTO user_profiles (user_id, bio, age, height, weight, ethnicity, languages, working_hours, is_available, service_type, skills, province, district, sub_district, postal_code, address_line1, latitude, longitude)
VALUES 
((SELECT user_id FROM users WHERE username = 'marco_thai'), 'Athletic personal trainer & massage therapist. Sports massage and deep tissue specialist. Perfect for athletes. 💪', 29, 178, 75, 'Thai', ARRAY['Thai', 'English'], '08:00-20:00', true, 'both', ARRAY['Sports Massage', 'Deep Tissue', 'Stretching'], 'กรุงเทพมหานคร', 'คลองเตย', 'คลองตัน', '10110', '100 Sukhumvit Soi 16', 13.7308, 100.5614),

((SELECT user_id FROM users WHERE username = 'alex_sports'), 'Professional male escort & companion. Gym enthusiast, great conversation, available for events and travel. 🏋️‍♂️', 31, 182, 80, 'Thai-Western', ARRAY['Thai', 'English'], '16:00-02:00', true, 'outcall', ARRAY['Companion', 'Events', 'Travel'], 'กรุงเทพมหานคร', 'วัฒนา', 'พระโขนง', '10110', '777 Sukhumvit Soi 42', 13.7098, 100.5881),

((SELECT user_id FROM users WHERE username = 'david_fitness'), 'Certified fitness coach and wellness provider. Specializing in physical therapy and body conditioning. 🏃‍♂️', 27, 175, 72, 'Thai', ARRAY['Thai', 'English'], '10:00-22:00', true, 'both', ARRAY['Physical Therapy', 'Fitness Training', 'Massage'], 'กรุงเทพมหานคร', 'บางกอกน้อย', 'บางขุนนนท์', '10700', '888 Charansanitwong Road', 13.7699, 100.4867),

((SELECT user_id FROM users WHERE username = 'ryan_wellness'), 'Male massage therapist. Traditional and oil massage. Relaxing and professional service. 🙏', 25, 172, 68, 'Thai', ARRAY['Thai'], '12:00-00:00', true, 'incall', ARRAY['Thai Massage', 'Oil Massage'], 'กรุงเทพมหานคร', 'ห้วยขวาง', 'สามเสนนอก', '10310', '222 Pracha Rat Sai 1', 13.7838, 100.5731),

((SELECT user_id FROM users WHERE username = 'jason_therapy'), 'Holistic health practitioner. Offering therapeutic massage and wellness consultation. Great for stress management. 🌿', 30, 180, 78, 'Thai', ARRAY['Thai', 'English'], '11:00-21:00', true, 'both', ARRAY['Therapeutic Massage', 'Wellness', 'Consultation'], 'กรุงเทพมหานคร', 'จตุจักร', 'ลาดยาว', '10900', '999 Phahonyothin Road', 13.8169, 100.5614);

-- Ladyboy Profiles
INSERT INTO user_profiles (user_id, bio, age, height, weight, ethnicity, languages, working_hours, is_available, service_type, skills, province, district, sub_district, postal_code, address_line1, latitude, longitude)
VALUES 
((SELECT user_id FROM users WHERE username = 'kim_beauty'), 'Beautiful ladyboy escort. Glamorous, feminine, and sophisticated. Available for companionship and special occasions. 💋', 26, 172, 58, 'Thai', ARRAY['Thai', 'English', 'Japanese'], '18:00-03:00', true, 'outcall', ARRAY['Companion', 'Entertainment', 'Events'], 'กรุงเทพมหานคร', 'ปทุมวัน', 'รองเมือง', '10330', '111 Rama 1 Road', 13.7469, 100.5348),

((SELECT user_id FROM users WHERE username = 'rose_glamour'), 'Stunning transgender beauty specialist. Makeup, styling, and beauty consultation. Also available for companionship. 💅', 24, 168, 55, 'Thai', ARRAY['Thai', 'English'], '14:00-00:00', true, 'both', ARRAY['Beauty Consultation', 'Styling', 'Companion'], 'กรุงเทพมหานคร', 'บางรัก', 'สุริยวงศ์', '10500', '333 Surawong Road', 13.7233, 100.5289),

((SELECT user_id FROM users WHERE username = 'mimi_style'), 'Fabulous ladyboy performer & companion. Fun, energetic, perfect for parties and events. Love to entertain! 🎭', 23, 170, 56, 'Thai', ARRAY['Thai', 'English'], '19:00-02:00', true, 'outcall', ARRAY['Entertainment', 'Performance', 'Companion'], 'กรุงเทพมหานคร', 'ดินแดง', 'ดินแดง', '10400', '444 Ratchadaphisek Road', 13.7649, 100.5583),

((SELECT user_id FROM users WHERE username = 'angel_paradise'), 'Sweet and caring transgender massage therapist. Gentle touch and understanding service. 🌺', 25, 166, 53, 'Thai', ARRAY['Thai'], '13:00-23:00', true, 'both', ARRAY['Thai Massage', 'Oil Massage', 'Aromatherapy'], 'กรุงเทพมหานคร', 'พระนคร', 'สำราญราษฎร์', '10200', '666 Khao San Road', 13.7588, 100.4975);

-- Gay Profiles
INSERT INTO user_profiles (user_id, bio, age, height, weight, ethnicity, languages, working_hours, is_available, service_type, skills, province, district, sub_district, postal_code, address_line1, latitude, longitude)
VALUES 
((SELECT user_id FROM users WHERE username = 'tony_pride'), 'Professional gay escort. Masculine, charming, great for dinner dates and travel. LGBTQ+ friendly always. 🏳️‍🌈', 28, 176, 73, 'Thai', ARRAY['Thai', 'English'], '17:00-02:00', true, 'outcall', ARRAY['Companion', 'Travel Partner', 'Events'], 'กรุงเทพมหานคร', 'วัฒนา', 'คลองเตยเหนือ', '10110', '123 Asoke Tower', 13.7362, 100.5601),

((SELECT user_id FROM users WHERE username = 'kevin_rainbow'), 'Fit gay companion & wellness coach. Gym buddy, travel companion, and life coach. Positive vibes only! ✨', 30, 180, 76, 'Thai-Western', ARRAY['Thai', 'English'], '10:00-22:00', true, 'both', ARRAY['Fitness Coach', 'Companion', 'Wellness'], 'กรุงเทพมหานคร', 'สาทร', 'ทุ่งมหาเมฆ', '10120', '789 Sathorn Unique Tower', 13.7194, 100.5345),

((SELECT user_id FROM users WHERE username = 'sam_fabulous'), 'Fabulous gay massage therapist & entertainer. Fun personality, great hands, unforgettable experience. 💆‍♂️', 26, 174, 70, 'Thai', ARRAY['Thai', 'English'], '15:00-01:00', true, 'both', ARRAY['Massage', 'Entertainment', 'Companion'], 'กรุงเทพมหานคร', 'บางรัก', 'สีลม', '10500', '456 Silom Soi 4', 13.7291, 100.5323);

-- ============================
-- SERVICE PACKAGES
-- ============================

-- Bella's packages
INSERT INTO service_packages (provider_id, package_name, description, duration, price)
VALUES 
((SELECT user_id FROM users WHERE username = 'bella_bangkok'), '1 Hour Thai Massage', 'Traditional Thai massage with stretching', 60, 1200),
((SELECT user_id FROM users WHERE username = 'bella_bangkok'), '2 Hours Aromatherapy', 'Full body aromatherapy massage with essential oils', 120, 2500),
((SELECT user_id FROM users WHERE username = 'bella_bangkok'), 'Spa Package (3 Hours)', 'Thai massage + aromatherapy + body scrub', 180, 3500);

-- Sophia's packages
INSERT INTO service_packages (provider_id, package_name, description, duration, price)
VALUES 
((SELECT user_id FROM users WHERE username = 'sophia_silom'), 'Dinner Date (3 Hours)', 'Elegant companion for dinner and conversation', 180, 8000),
((SELECT user_id FROM users WHERE username = 'sophia_silom'), 'Overnight Companion', 'Full night companionship (8 hours)', 480, 25000),
((SELECT user_id FROM users WHERE username = 'sophia_silom'), 'Weekend Travel', 'Travel companion for weekend trip', 2880, 80000);

-- Marco's packages
INSERT INTO service_packages (provider_id, package_name, description, duration, price)
VALUES 
((SELECT user_id FROM users WHERE username = 'marco_thai'), 'Sports Massage (90 min)', 'Deep tissue sports massage for athletes', 90, 1800),
((SELECT user_id FROM users WHERE username = 'marco_thai'), 'Personal Training + Massage', '1 hour training + 1 hour massage', 120, 3000);

-- Kim's packages
INSERT INTO service_packages (provider_id, package_name, description, duration, price)
VALUES 
((SELECT user_id FROM users WHERE username = 'kim_beauty'), 'Companion Evening (4 Hours)', 'Glamorous companion for events or dinner', 240, 10000),
((SELECT user_id FROM users WHERE username = 'kim_beauty'), 'Overnight Experience', 'Full night companionship', 480, 28000);

-- Tony's packages
INSERT INTO service_packages (provider_id, package_name, description, duration, price)
VALUES 
((SELECT user_id FROM users WHERE username = 'tony_pride'), 'Dinner Companion (3 Hours)', 'Professional companion for social events', 180, 7000),
((SELECT user_id FROM users WHERE username = 'tony_pride'), 'Weekend Trip', 'Travel companion for weekend getaway', 2880, 75000);

-- Additional packages for other providers
INSERT INTO service_packages (provider_id, package_name, description, duration, price)
VALUES 
((SELECT user_id FROM users WHERE username = 'maya_massage'), '1 Hour Oil Massage', 'Relaxing oil massage', 60, 1500),
((SELECT user_id FROM users WHERE username = 'luna_therapy'), 'Yoga + Meditation (90 min)', 'Private yoga and meditation session', 90, 2000),
((SELECT user_id FROM users WHERE username = 'alex_sports'), '4 Hour Companion', 'Companion for events or dinner', 240, 9000),
((SELECT user_id FROM users WHERE username = 'rose_glamour'), 'Beauty Makeover', 'Full makeover and styling', 120, 3500),
((SELECT user_id FROM users WHERE username = 'kevin_rainbow'), 'Fitness + Wellness Session', 'Personal training and wellness coaching', 120, 2800);

-- ============================
-- Summary
-- ============================
-- ✅ 5 Female providers (ผู้หญิง)
-- ✅ 5 Male providers (ผู้ชาย)  
-- ✅ 4 Ladyboy providers
-- ✅ 3 Gay providers
-- Total: 17 diverse service providers
-- All with complete profiles and packages!
-- Note: Reviews require bookings first (ต้องสร้าง bookings ก่อนถึงจะรีวิวได้)
