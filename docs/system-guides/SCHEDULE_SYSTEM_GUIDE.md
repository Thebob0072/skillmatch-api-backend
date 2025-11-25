# Provider Schedule System Guide

## Overview

The Provider Schedule System allows providers to manage their availability calendar and enables Admin/GOD to track where providers are going and what they're doing. The system automatically creates schedule entries when bookings are confirmed and prevents double-booking through database-level constraints.

## Key Features

- **Schedule Management**: Providers create, view, update, and delete their availability slots
- **Auto-Schedule Creation**: When a booking is confirmed, a schedule entry is automatically created with `status='booked'`
- **Overlap Prevention**: GIST index with EXCLUDE constraint prevents double-booking at database level
- **Admin Visibility**: Only Admin and GOD can view provider schedules with full location details
- **Status Tracking**: Three states - `available` (free), `booked` (has booking), `blocked` (unavailable)
- **Location Tracking**: Captures location type, address, province, district, and coordinates

## Database Schema

### `provider_schedules` Table

```sql
CREATE TABLE provider_schedules (
    schedule_id SERIAL PRIMARY KEY,
    provider_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    booking_id INT REFERENCES bookings(booking_id) ON DELETE SET NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'available' CHECK (status IN ('available', 'booked', 'blocked')),
    location_type VARCHAR(20) CHECK (location_type IN ('Incall', 'Outcall', 'Both')),
    location_address TEXT,
    location_province VARCHAR(100),
    location_district VARCHAR(100),
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    notes TEXT,
    is_visible_to_admin BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Prevents overlapping schedules per provider
    CONSTRAINT no_overlap EXCLUDE USING GIST (
        provider_id WITH =,
        tsrange(start_time, end_time) WITH &&
    )
);
```

**Indexes:**
- `idx_schedules_provider` - Fast lookup by provider_id
- `idx_schedules_time` - Time-based queries
- `idx_schedules_status` - Filter by status
- `idx_schedules_booking` - Join with bookings table

**Trigger:**
- `update_schedule_timestamp()` - Auto-updates `updated_at` on row changes

## API Endpoints

### 1. Create Schedule Entry (Provider)

**Endpoint:** `POST /provider/schedule`

**Auth:** Requires JWT token (provider)

**Request Body:**
```json
{
  "start_time": "2025-11-15T09:00:00Z",
  "end_time": "2025-11-15T17:00:00Z",
  "status": "available",
  "location_type": "Outcall",
  "location_address": "123 Sukhumvit Rd, Khlong Toei Nuea",
  "location_province": "Bangkok",
  "location_district": "Watthana",
  "latitude": 13.7563,
  "longitude": 100.5018,
  "notes": "Available for Outcall services in Bangkok area"
}
```

**Validation Rules:**
- `start_time` and `end_time` are **required** (ISO 8601 format)
- `end_time` must be **after** `start_time`
- `status` must be `available` or `blocked` (cannot manually create `booked` status)
- Schedule must **not overlap** with existing schedules (enforced by EXCLUDE constraint)
- `location_type` must be `Incall`, `Outcall`, or `Both` (optional)

**Response (201 Created):**
```json
{
  "message": "Schedule created successfully",
  "schedule_id": 42
}
```

**Error Responses:**
- **400 Bad Request**: Missing required fields, invalid time format, end_time <= start_time
- **409 Conflict**: Schedule overlaps with existing entry
- **500 Internal Server Error**: Database error

---

### 2. Get My Schedules (Provider)

**Endpoint:** `GET /provider/schedule/me`

**Auth:** Requires JWT token (provider)

**Query Parameters:**
- `start_date` (optional) - Filter schedules starting from this date (ISO 8601)
- `end_date` (optional) - Filter schedules ending before this date (ISO 8601)
- `status` (optional) - Filter by status: `available`, `booked`, `blocked`

**Example Request:**
```
GET /provider/schedule/me?start_date=2025-11-01&end_date=2025-11-30&status=available
```

**Response (200 OK):**
```json
{
  "schedules": [
    {
      "schedule_id": 42,
      "provider_id": 5,
      "booking_id": null,
      "start_time": "2025-11-15T09:00:00Z",
      "end_time": "2025-11-15T17:00:00Z",
      "status": "available",
      "location_type": "Outcall",
      "location_address": "123 Sukhumvit Rd, Khlong Toei Nuea",
      "location_province": "Bangkok",
      "location_district": "Watthana",
      "latitude": 13.7563,
      "longitude": 100.5018,
      "notes": "Available for Outcall services in Bangkok area",
      "is_visible_to_admin": true,
      "created_at": "2025-11-14T10:30:00Z",
      "updated_at": "2025-11-14T10:30:00Z"
    },
    {
      "schedule_id": 45,
      "provider_id": 5,
      "booking_id": 128,
      "start_time": "2025-11-16T14:00:00Z",
      "end_time": "2025-11-16T18:00:00Z",
      "status": "booked",
      "location_type": "Incall",
      "location_address": "456 Rama IV Rd",
      "location_province": "Bangkok",
      "location_district": "Pathum Wan",
      "latitude": 13.7440,
      "longitude": 100.5332,
      "notes": "Auto-created from booking confirmation",
      "is_visible_to_admin": true,
      "created_at": "2025-11-14T11:00:00Z",
      "updated_at": "2025-11-14T11:00:00Z"
    }
  ]
}
```

---

### 3. Update Schedule Entry (Provider)

**Endpoint:** `PATCH /provider/schedule/:scheduleId`

**Auth:** Requires JWT token (provider, owner only)

**Request Body (all fields optional):**
```json
{
  "start_time": "2025-11-15T10:00:00Z",
  "end_time": "2025-11-15T18:00:00Z",
  "status": "blocked",
  "location_type": "Both",
  "location_address": "Updated address",
  "location_province": "Chiang Mai",
  "location_district": "Mueang Chiang Mai",
  "latitude": 18.7883,
  "longitude": 98.9853,
  "notes": "Updated availability notes"
}
```

**Validation Rules:**
- Only **owner** can update their schedule
- Cannot update schedules with `status='booked'` (system-managed)
- Cannot manually change status **to** or **from** `booked`
- If updating times, new schedule must **not overlap** with other schedules
- Dynamic updates: only provided fields are modified

**Response (200 OK):**
```json
{
  "message": "Schedule updated successfully"
}
```

**Error Responses:**
- **400 Bad Request**: Trying to update booked schedule, invalid data
- **403 Forbidden**: Not the owner of this schedule
- **404 Not Found**: Schedule does not exist
- **409 Conflict**: Updated schedule would overlap with another

---

### 4. Delete Schedule Entry (Provider)

**Endpoint:** `DELETE /provider/schedule/:scheduleId`

**Auth:** Requires JWT token (provider, owner only)

**Validation Rules:**
- Only **owner** can delete their schedule
- Cannot delete schedules with `status='booked'` (must cancel booking first)

**Response (200 OK):**
```json
{
  "message": "Schedule deleted successfully"
}
```

**Error Responses:**
- **400 Bad Request**: Trying to delete booked schedule
- **403 Forbidden**: Not the owner of this schedule
- **404 Not Found**: Schedule does not exist

---

### 5. Get Provider Schedule (Admin)

**Endpoint:** `GET /admin/schedules/provider/:providerId`

**Auth:** Requires JWT token + Admin permission

**Query Parameters:**
- `start_date` (optional) - Filter schedules from this date
- `end_date` (optional) - Filter schedules until this date
- `status` (optional) - Filter by status

**Example Request:**
```
GET /admin/schedules/provider/5?start_date=2025-11-01&end_date=2025-11-30
```

**Response (200 OK):**
```json
{
  "schedules": [
    {
      "schedule_id": 45,
      "provider_id": 5,
      "provider_username": "john_provider",
      "provider_phone": "0812345678",
      "booking_id": 128,
      "client_id": 12,
      "client_username": "jane_client",
      "start_time": "2025-11-16T14:00:00Z",
      "end_time": "2025-11-16T18:00:00Z",
      "status": "booked",
      "location_type": "Incall",
      "location_address": "456 Rama IV Rd",
      "location_province": "Bangkok",
      "location_district": "Pathum Wan",
      "latitude": 13.7440,
      "longitude": 100.5332,
      "notes": "Auto-created from booking confirmation",
      "is_visible_to_admin": true,
      "created_at": "2025-11-14T11:00:00Z",
      "updated_at": "2025-11-14T11:00:00Z"
    }
  ]
}
```

**Additional Fields (Admin View):**
- `provider_username` - Provider's display name
- `provider_phone` - Provider's contact number
- `client_id` - Client who booked (if status='booked')
- `client_username` - Client's display name (if booked)

**Error Responses:**
- **403 Forbidden**: Not an admin
- **404 Not Found**: Provider does not exist

---

### 6. Get All Providers' Schedules (Admin)

**Endpoint:** `GET /admin/schedules/all`

**Auth:** Requires JWT token + Admin permission

**Query Parameters:**
- `start_date` (optional) - Filter schedules from this date
- `end_date` (optional) - Filter schedules until this date
- `status` (optional) - Filter by status
- `province` (optional) - Filter by location province

**Example Request:**
```
GET /admin/schedules/all?start_date=2025-11-15&status=booked&province=Bangkok
```

**Use Case:**
- GOD dashboard overview of all provider activities
- Track where providers are working and when
- Monitor booking patterns across regions

**Response (200 OK):**
```json
{
  "schedules": [
    {
      "schedule_id": 45,
      "provider_id": 5,
      "provider_username": "john_provider",
      "provider_phone": "0812345678",
      "booking_id": 128,
      "client_id": 12,
      "client_username": "jane_client",
      "start_time": "2025-11-16T14:00:00Z",
      "end_time": "2025-11-16T18:00:00Z",
      "status": "booked",
      "location_type": "Incall",
      "location_address": "456 Rama IV Rd",
      "location_province": "Bangkok",
      "location_district": "Pathum Wan",
      "latitude": 13.7440,
      "longitude": 100.5332,
      "notes": "Auto-created from booking confirmation",
      "is_visible_to_admin": true,
      "created_at": "2025-11-14T11:00:00Z",
      "updated_at": "2025-11-14T11:00:00Z"
    },
    {
      "schedule_id": 47,
      "provider_id": 8,
      "provider_username": "sarah_provider",
      "provider_phone": "0898765432",
      "booking_id": 130,
      "client_id": 15,
      "client_username": "mike_client",
      "start_time": "2025-11-17T10:00:00Z",
      "end_time": "2025-11-17T14:00:00Z",
      "status": "booked",
      "location_type": "Outcall",
      "location_address": "789 Silom Rd",
      "location_province": "Bangkok",
      "location_district": "Bang Rak",
      "latitude": 13.7248,
      "longitude": 100.5334,
      "notes": "Auto-created from booking confirmation",
      "is_visible_to_admin": true,
      "created_at": "2025-11-14T12:00:00Z",
      "updated_at": "2025-11-14T12:00:00Z"
    }
  ]
}
```

**Error Responses:**
- **403 Forbidden**: Not an admin

---

## Auto-Schedule Creation Flow

### When Booking Status Changes to "confirmed"

**Location:** `booking_handlers.go` - `updateBookingStatusHandler()`

**Process:**
1. **Query booking details** - Fetch `start_time`, `end_time`, `location` from `bookings` table
2. **Create schedule entry** - INSERT into `provider_schedules`:
   - `provider_id` - From booking
   - `booking_id` - Link to booking
   - `start_time`, `end_time` - From booking
   - `status` - Set to `'booked'`
   - `location_address` - From booking location
   - `notes` - `'Auto-created from booking confirmation'`
   - `is_visible_to_admin` - `true`
3. **Error handling** - Log warning if schedule creation fails (non-fatal)
4. **ON CONFLICT DO NOTHING** - Prevents duplicate entries if triggered multiple times

**Code Reference:**
```go
case "confirmed":
    // Auto-create schedule entry when booking is confirmed
    var startTime, endTime time.Time
    var location *string
    err = dbPool.QueryRow(ctx, `
        SELECT start_time, end_time, location
        FROM bookings
        WHERE booking_id = $1
    `, bookingID).Scan(&startTime, &endTime, &location)

    if err == nil && !startTime.IsZero() && !endTime.IsZero() {
        _, scheduleErr := dbPool.Exec(ctx, `
            INSERT INTO provider_schedules (
                provider_id, booking_id, start_time, end_time, status,
                location_address, notes, is_visible_to_admin
            ) VALUES ($1, $2, $3, $4, 'booked', $5, 'Auto-created from booking confirmation', true)
            ON CONFLICT DO NOTHING
        `, providerID, bookingID, startTime, endTime, location)

        if scheduleErr != nil {
            log.Printf("Warning: Failed to create schedule entry for booking %d: %v", bookingID, scheduleErr)
        }
    }
```

### When Booking Status Changes to "cancelled"

**Process:**
1. **Delete schedule entry** - Remove from `provider_schedules` WHERE `booking_id = ?`
2. **Cascading**: No error if schedule doesn't exist

**Code Reference:**
```go
case "cancelled":
    // Remove schedule entry if exists
    _, _ = dbPool.Exec(ctx, `DELETE FROM provider_schedules WHERE booking_id = $1`, bookingID)
```

---

## Status Rules

### `available`
- **Created by**: Provider manually creates schedule entry
- **Meaning**: Provider is free and available for booking during this time
- **Can Edit**: Yes, provider can modify all fields
- **Can Delete**: Yes, provider can remove entry

### `booked`
- **Created by**: System automatically when booking confirmed
- **Meaning**: Provider has an active booking during this time
- **Can Edit**: No, system-managed (cannot modify)
- **Can Delete**: No, must cancel booking first
- **Linked to**: `booking_id` references bookings table

### `blocked`
- **Created by**: Provider manually marks time as unavailable
- **Meaning**: Provider is not available (personal time, holiday, maintenance, etc.)
- **Can Edit**: Yes, provider can modify all fields
- **Can Delete**: Yes, provider can remove entry

---

## Security & Access Control

### Provider Access
- **Create**: Own schedules only
- **Read**: Own schedules only (via `/provider/schedule/me`)
- **Update**: Own schedules only (except `status='booked'`)
- **Delete**: Own schedules only (except `status='booked'`)
- **Cannot See**: Other providers' schedules

### Admin/GOD Access
- **Read**: All providers' schedules with full details
- **Endpoints**: 
  - `/admin/schedules/provider/:providerId` - Single provider
  - `/admin/schedules/all` - All providers (GOD dashboard)
- **Details Visible**: Provider username, phone, client username (if booked)
- **Cannot Modify**: Admins view only, cannot create/edit/delete

### Visibility Flag
- `is_visible_to_admin` - Default `true`
- Currently unused (all schedules visible to admin)
- Future use: Allow providers to mark certain schedules as private

---

## Frontend Integration Guide

### Provider Schedule Calendar

**Component: `ProviderScheduleCalendar.tsx`**

**Features:**
- Calendar view showing all schedule entries
- Color coding:
  - 🟢 Green - `available` (free slots)
  - 🔴 Red - `booked` (active bookings)
  - ⚫ Gray - `blocked` (unavailable)
- Add new availability slots (creates `available` entry)
- Edit/delete own schedules (except `booked` status)
- View booking details by clicking on `booked` slots

**TypeScript Interface:**
```typescript
interface ProviderSchedule {
  schedule_id: number;
  provider_id: number;
  booking_id?: number;
  start_time: string; // ISO 8601
  end_time: string; // ISO 8601
  status: 'available' | 'booked' | 'blocked';
  location_type?: 'Incall' | 'Outcall' | 'Both';
  location_address?: string;
  location_province?: string;
  location_district?: string;
  latitude?: number;
  longitude?: number;
  notes?: string;
  is_visible_to_admin: boolean;
  created_at: string;
  updated_at: string;
}

interface CreateScheduleRequest {
  start_time: string; // ISO 8601, required
  end_time: string; // ISO 8601, required
  status?: 'available' | 'blocked'; // Cannot set 'booked'
  location_type?: 'Incall' | 'Outcall' | 'Both';
  location_address?: string;
  location_province?: string;
  location_district?: string;
  latitude?: number;
  longitude?: number;
  notes?: string;
}
```

**Example Usage:**
```typescript
// Fetch provider's schedules
const fetchSchedules = async (startDate?: string, endDate?: string, status?: string) => {
  const params = new URLSearchParams();
  if (startDate) params.append('start_date', startDate);
  if (endDate) params.append('end_date', endDate);
  if (status) params.append('status', status);
  
  const response = await fetch(`/provider/schedule/me?${params}`, {
    headers: { 'Authorization': `Bearer ${token}` }
  });
  return response.json();
};

// Create availability slot
const createSchedule = async (data: CreateScheduleRequest) => {
  const response = await fetch('/provider/schedule', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(data)
  });
  return response.json();
};

// Update schedule
const updateSchedule = async (scheduleId: number, data: Partial<CreateScheduleRequest>) => {
  const response = await fetch(`/provider/schedule/${scheduleId}`, {
    method: 'PATCH',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(data)
  });
  return response.json();
};

// Delete schedule
const deleteSchedule = async (scheduleId: number) => {
  const response = await fetch(`/provider/schedule/${scheduleId}`, {
    method: 'DELETE',
    headers: { 'Authorization': `Bearer ${token}` }
  });
  return response.json();
};
```

---

### Admin Schedule Dashboard

**Component: `AdminScheduleDashboard.tsx`**

**Features:**
- List view of all providers' schedules
- Filter by date range, status, province
- Map view showing provider locations (using lat/long)
- Export schedule reports
- View provider/client details

**TypeScript Interface:**
```typescript
interface ProviderScheduleWithDetails extends ProviderSchedule {
  provider_username: string;
  provider_phone?: string;
  client_id?: number;
  client_username?: string;
}
```

**Example Usage:**
```typescript
// Fetch specific provider's schedule (Admin only)
const fetchProviderSchedule = async (providerId: number, startDate?: string, endDate?: string) => {
  const params = new URLSearchParams();
  if (startDate) params.append('start_date', startDate);
  if (endDate) params.append('end_date', endDate);
  
  const response = await fetch(`/admin/schedules/provider/${providerId}?${params}`, {
    headers: { 'Authorization': `Bearer ${token}` }
  });
  return response.json();
};

// Fetch all providers' schedules (GOD dashboard)
const fetchAllSchedules = async (startDate?: string, status?: string, province?: string) => {
  const params = new URLSearchParams();
  if (startDate) params.append('start_date', startDate);
  if (status) params.append('status', status);
  if (province) params.append('province', province);
  
  const response = await fetch(`/admin/schedules/all?${params}`, {
    headers: { 'Authorization': `Bearer ${token}` }
  });
  return response.json();
};
```

---

### Booking Flow Updates

**Component: `BookingForm.tsx`**

**Check Provider Availability:**
Before creating a booking, fetch provider's schedules to show available time slots:

```typescript
const checkAvailability = async (providerId: number, date: string) => {
  const startDate = `${date}T00:00:00Z`;
  const endDate = `${date}T23:59:59Z`;
  
  const response = await fetch(
    `/admin/schedules/provider/${providerId}?start_date=${startDate}&end_date=${endDate}&status=booked`,
    { headers: { 'Authorization': `Bearer ${token}` } }
  );
  
  const { schedules } = await response.json();
  
  // Show "Provider unavailable" for booked time slots
  return schedules;
};
```

**Auto-Update Calendar:**
- When booking is confirmed (status changed to `confirmed`), schedule entry is automatically created
- WebSocket notification triggers calendar refresh
- Show updated schedule with new `booked` slot

---

## Error Handling

### Overlap Detection

**Scenario:** Provider tries to create overlapping schedule

**Database Constraint:**
```sql
CONSTRAINT no_overlap EXCLUDE USING GIST (
    provider_id WITH =,
    tsrange(start_time, end_time) WITH &&
)
```

**Error Response (409 Conflict):**
```json
{
  "error": "Schedule overlaps with existing entry"
}
```

**Frontend Handling:**
- Show error message: "You already have a schedule during this time"
- Highlight conflicting time slot in calendar
- Suggest available time slots

---

### Booking Schedule Creation Failure

**Scenario:** Auto-schedule creation fails when booking confirmed

**Behavior:**
- Booking confirmation still succeeds (non-fatal)
- Warning logged to server console
- Schedule entry missing in provider calendar

**Frontend Detection:**
- After booking confirmation, fetch provider schedule
- If schedule entry missing, show warning to provider
- Provide manual "Add to Calendar" option

---

## Testing Checklist

### Provider Schedule Management
- [ ] Create schedule with valid data → Success (201)
- [ ] Create schedule with overlapping time → Error (409)
- [ ] Create schedule with `status='booked'` → Error (400, cannot manually create)
- [ ] Create schedule with `end_time <= start_time` → Error (400)
- [ ] Fetch own schedules without filters → Returns all
- [ ] Fetch own schedules with date range filter → Returns filtered
- [ ] Fetch own schedules with status filter → Returns filtered
- [ ] Update own schedule (status='available') → Success
- [ ] Update own schedule (status='booked') → Error (400, cannot modify)
- [ ] Update other provider's schedule → Error (403)
- [ ] Delete own schedule (status='available') → Success
- [ ] Delete own schedule (status='booked') → Error (400)
- [ ] Delete other provider's schedule → Error (403)

### Auto-Schedule Creation
- [ ] Confirm booking → Schedule entry created with `status='booked'`
- [ ] Schedule entry has correct `booking_id` reference
- [ ] Schedule entry visible in provider's calendar
- [ ] Cancel booking → Schedule entry removed
- [ ] Confirm same booking twice → No duplicate schedule (ON CONFLICT DO NOTHING)

### Admin Schedule Viewing
- [ ] Admin fetches single provider's schedule → Success with full details
- [ ] Admin fetches all providers' schedules → Success with all providers
- [ ] Non-admin tries to access admin endpoint → Error (403)
- [ ] Filter by date range → Returns schedules within range
- [ ] Filter by status → Returns matching schedules
- [ ] Filter by province → Returns schedules in province

### Overlap Prevention
- [ ] Create two schedules with overlapping time → Second fails (409)
- [ ] Update schedule to overlap with another → Error (409)
- [ ] GIST index enforces constraint at database level

---

## Migration Notes

### Migration 019: Provider Schedules System

**File:** `migrations.go` (lines 549-629)

**Components:**
1. **btree_gist Extension** - Enables GIST index on integer columns
2. **provider_schedules Table** - Main schedule storage
3. **EXCLUDE Constraint** - Prevents overlapping schedules using `tsrange(start_time, end_time)`
4. **Indexes** - Fast queries on provider_id, time range, status, booking_id
5. **Trigger Function** - Auto-updates `updated_at` timestamp

**Rollback:**
```sql
DROP TRIGGER IF EXISTS update_schedule_timestamp ON provider_schedules;
DROP FUNCTION IF EXISTS update_schedule_timestamp();
DROP TABLE IF EXISTS provider_schedules;
-- Note: btree_gist extension not dropped (may be used by other tables)
```

---

## Performance Considerations

### Index Usage
- **idx_schedules_provider** - Used in provider's own schedule queries
- **idx_schedules_time** - Used in date range filters
- **idx_schedules_status** - Used in status filters
- **idx_schedules_booking** - Used when fetching booking details

### Query Optimization
- **Provider queries**: Index on `(provider_id, start_time)` provides fast lookup
- **Admin queries**: Multiple indexes allow efficient filtering
- **GIST index**: Efficient overlap detection (O(log n) instead of O(n²))

### Recommended Limits
- **Pagination**: Implement for admin endpoint with 50-100 schedules per page
- **Date Range**: Limit to 6 months per query to avoid performance issues
- **Concurrent Bookings**: EXCLUDE constraint handles race conditions at database level

---

## Future Enhancements

### Potential Features
1. **Recurring Schedules** - Create weekly/monthly availability patterns
2. **Time Zone Support** - Handle multiple time zones for international bookings
3. **Private Schedules** - Allow providers to hide certain entries from admin
4. **Schedule Templates** - Save and reuse common availability patterns
5. **Notification Reminders** - Send reminders to providers before scheduled bookings
6. **Schedule Conflicts Dashboard** - Admin view of scheduling issues
7. **Availability Suggestions** - AI-powered suggestions for optimal availability slots
8. **Calendar Export** - Export schedules to Google Calendar, iCal, etc.

---

## Troubleshooting

### GIST Index Error
**Error:** `ERROR: data type integer has no default operator class for access method "gist"`

**Solution:** Add `btree_gist` extension before creating table:
```sql
CREATE EXTENSION IF NOT EXISTS btree_gist;
```

### Schedule Not Created on Booking
**Issue:** Booking confirmed but no schedule entry

**Check:**
1. Server logs for warnings: `"Failed to create schedule entry"`
2. Booking has valid `start_time` and `end_time`
3. No database constraint violations

### Cannot Update/Delete Booked Schedule
**Issue:** Provider tries to modify schedule with `status='booked'`

**Expected Behavior:** Returns 400 error
**Reason:** Booked schedules are system-managed, must cancel booking first

---

## Summary

The Provider Schedule System provides comprehensive calendar management for providers and full visibility for admins/GOD. Key benefits:

✅ **Automatic Integration** - Bookings automatically create schedule entries  
✅ **Overlap Prevention** - Database-level constraint prevents double-booking  
✅ **Admin Visibility** - Track where providers are going and what they're doing  
✅ **Flexible Management** - Providers can create availability and block time  
✅ **Secure Access Control** - Providers see only their schedules, admins see everything  

**Total Endpoints:** 6 (4 provider + 2 admin)  
**Database Tables:** 1 (`provider_schedules`)  
**Migrations:** Migration 019  
**Handler File:** `schedule_handlers.go` (543 lines)
