package main

import (
	"math"
)

// Haversine Formula - คำนวณระยะทางระหว่างพิกัด GPS 2 จุด (ในหน่วย กิโลเมตร)
// lat1, lon1: พิกัดจุดแรก (user's location)
// lat2, lon2: พิกัดจุดที่สอง (provider's location)
func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusKm = 6371.0 // รัศมีโลก (กิโลเมตร)

	// แปลงองศาเป็น radian
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	// Haversine formula
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	distance := earthRadiusKm * c
	return math.Round(distance*100) / 100 // ปัดเศษ 2 ตำแหน่ง
}

// Helper สำหรับตรวจสอบว่าพิกัดครบหรือไม่
func hasValidCoordinates(lat, lon *float64) bool {
	return lat != nil && lon != nil && *lat != 0 && *lon != 0
}
