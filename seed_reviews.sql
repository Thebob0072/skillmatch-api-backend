-- Add Reviews for Test Providers

DO $$
DECLARE
  admin_id INT := 1;
  provider_rec RECORD;
  package_rec RECORD;
  new_booking_id INT;
BEGIN
  -- Loop through all test providers
  FOR provider_rec IN 
    SELECT user_id, username 
    FROM users 
    WHERE email LIKE '%@example.com' AND provider_level_id > 1
  LOOP
    -- Get first package of this provider
    FOR package_rec IN
      SELECT package_id, price
      FROM service_packages 
      WHERE provider_id = provider_rec.user_id
      LIMIT 1
    LOOP
      -- Insert booking
      INSERT INTO bookings (
        client_id, provider_id, package_id, 
        booking_date, start_time, end_time, 
        total_price, status, location
      ) VALUES (
        admin_id, 
        provider_rec.user_id, 
        package_rec.package_id,
        CURRENT_DATE - (random() * 30)::int,
        NOW() - INTERVAL '7 days',
        NOW() - INTERVAL '7 days' + INTERVAL '2 hours',
        package_rec.price,
        'completed',
        'Bangkok'
      )
      RETURNING booking_id INTO new_booking_id;

      -- Insert review
      INSERT INTO reviews (
        booking_id, client_id, provider_id,
        rating, comment, is_verified
      ) VALUES (
        new_booking_id,
        admin_id, 
        provider_rec.user_id,
        (random() * 2 + 3)::int, -- Rating 3-5
        CASE (random() * 5)::int
          WHEN 0 THEN 'Excellent service! Highly recommended.'
          WHEN 1 THEN 'Very professional and friendly. Will book again!'
          WHEN 2 THEN 'Great experience. Worth every penny.'
          WHEN 3 THEN 'Amazing personality and very attentive.'
          ELSE 'Wonderful time! Perfect companion.'
        END,
        true
      );
      
      RAISE NOTICE 'Created review for provider: %', provider_rec.username;
    END LOOP;
  END LOOP;
END $$;

-- Summary
SELECT 
  'Total Bookings' as metric,
  COUNT(*) as count 
FROM bookings;

SELECT 
  'Total Reviews' as metric,
  COUNT(*) as count 
FROM reviews;

SELECT 
  u.username,
  u.provider_level_id as tier,
  COALESCE(AVG(r.rating), 0) as avg_rating,
  COUNT(r.review_id) as review_count
FROM users u
LEFT JOIN reviews r ON u.user_id = r.provider_id
WHERE u.email LIKE '%@example.com'
GROUP BY u.user_id, u.username, u.provider_level_id
ORDER BY u.provider_level_id DESC, avg_rating DESC;

COMMIT;
