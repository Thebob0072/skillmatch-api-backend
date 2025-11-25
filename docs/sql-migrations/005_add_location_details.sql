-- Migration: เพิ่มฟิลด์ที่อยู่แบบละเอียดและพิกัด GPS ใน user_profiles

ALTER TABLE user_profiles 
ADD COLUMN IF NOT EXISTS province VARCHAR(100),
ADD COLUMN IF NOT EXISTS district VARCHAR(100),
ADD COLUMN IF NOT EXISTS sub_district VARCHAR(100),
ADD COLUMN IF NOT EXISTS postal_code VARCHAR(10),
ADD COLUMN IF NOT EXISTS address_line1 VARCHAR(255),
ADD COLUMN IF NOT EXISTS latitude DECIMAL(10, 8),
ADD COLUMN IF NOT EXISTS longitude DECIMAL(11, 8);

-- สร้าง index สำหรับการค้นหาตามจังหวัด เขต แขวง
CREATE INDEX IF NOT EXISTS idx_user_profiles_province ON user_profiles(province);
CREATE INDEX IF NOT EXISTS idx_user_profiles_district ON user_profiles(district);
CREATE INDEX IF NOT EXISTS idx_user_profiles_sub_district ON user_profiles(sub_district);

-- สร้าง index สำหรับพิกัด GPS (ช่วยในการคำนวณระยะทาง)
CREATE INDEX IF NOT EXISTS idx_user_profiles_location ON user_profiles(latitude, longitude);

-- Comment อธิบายฟิลด์
COMMENT ON COLUMN user_profiles.province IS 'จังหวัด เช่น "กรุงเทพมหานคร", "เชียงใหม่"';
COMMENT ON COLUMN user_profiles.district IS 'เขต/อำเภอ เช่น "บางรัก", "เมือง"';
COMMENT ON COLUMN user_profiles.sub_district IS 'แขวง/ตำบล เช่น "สีลม", "ช้างคลาน"';
COMMENT ON COLUMN user_profiles.postal_code IS 'รหัสไปรษณีย์ เช่น "10500"';
COMMENT ON COLUMN user_profiles.address_line1 IS 'บ้านเลขที่ ถนน ซอย (แสดงเฉพาะหลัง booking confirmed)';
COMMENT ON COLUMN user_profiles.latitude IS 'พิกัด GPS (Latitude) ช่วง -90 ถึง 90';
COMMENT ON COLUMN user_profiles.longitude IS 'พิกัด GPS (Longitude) ช่วง -180 ถึง 180';
