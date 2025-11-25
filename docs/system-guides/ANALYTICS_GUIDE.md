# 📊 Analytics Dashboard Guide

## Overview

The SkillMatch platform includes a **comprehensive analytics system** for service providers to track their performance, revenue, bookings, and customer feedback. This system provides valuable insights to help providers optimize their services and grow their business.

**Key Features:**
- Real-time dashboard overview
- Booking statistics with date filters
- Revenue breakdown by service package
- Rating distribution analysis
- Monthly performance trends
- Profile view tracking

---

## Database Schema

### `profile_views` Table

```sql
CREATE TABLE profile_views (
    id SERIAL PRIMARY KEY,
    provider_id INTEGER REFERENCES users(user_id),
    viewer_id INTEGER REFERENCES users(user_id), -- NULL for anonymous
    view_count INTEGER DEFAULT 1,
    last_viewed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(provider_id, COALESCE(viewer_id, -1))
);
```

**Key Points:**
- Tracks both authenticated and anonymous views
- Increments view_count for repeat viewers
- One record per provider-viewer pair

---

## API Endpoints

### 1. Dashboard Overview

**GET** `/analytics/provider/dashboard`

Returns comprehensive overview statistics for the authenticated provider.

**Response:**
```json
{
  "profile_views": 1250,
  "total_bookings": 87,
  "completed_bookings": 72,
  "cancelled_bookings": 5,
  "pending_bookings": 10,
  "total_revenue": 215000.00,
  "average_rating": 4.7,
  "total_reviews": 68,
  "favorite_count": 142,
  "response_rate": 95.5,
  "average_response_time": 15
}
```

**Metrics Explained:**
- `profile_views`: Total profile page visits
- `total_bookings`: All bookings (any status)
- `completed_bookings`: Successfully completed bookings
- `cancelled_bookings`: Cancelled bookings
- `pending_bookings`: Awaiting confirmation
- `total_revenue`: Total earnings from completed bookings
- `average_rating`: Average review rating (1-5)
- `total_reviews`: Number of reviews received
- `favorite_count`: Number of users who favorited
- `response_rate`: Percentage of messages responded to
- `average_response_time`: Average response time in minutes

---

### 2. Booking Statistics

**GET** `/analytics/provider/bookings?period=30`

Returns daily booking statistics for the specified period.

**Query Parameters:**
- `period` (optional): Number of days (default: 30)

**Response:**
```json
{
  "stats": [
    {
      "date": "2025-11-13",
      "booking_count": 3,
      "revenue": 9000.00,
      "completed_count": 2,
      "cancelled_count": 0
    },
    {
      "date": "2025-11-12",
      "booking_count": 2,
      "revenue": 6000.00,
      "completed_count": 2,
      "cancelled_count": 0
    }
  ],
  "period": "30"
}
```

**Use Cases:**
- Track booking trends
- Identify peak booking days
- Monitor revenue patterns
- Chart visualization

---

### 3. Revenue Breakdown

**GET** `/analytics/provider/revenue`

Returns revenue breakdown by service package.

**Response:**
```json
{
  "revenue_breakdown": [
    {
      "package_name": "Premium 2 Hours",
      "booking_count": 45,
      "total_revenue": 135000.00,
      "avg_price": 3000.00
    },
    {
      "package_name": "Standard 1 Hour",
      "booking_count": 27,
      "total_revenue": 54000.00,
      "avg_price": 2000.00
    }
  ]
}
```

**Use Cases:**
- Identify most popular packages
- Optimize pricing strategy
- Focus on high-revenue services

---

### 4. Rating Distribution

**GET** `/analytics/provider/ratings`

Returns breakdown of ratings by star count.

**Response:**
```json
{
  "breakdown": {
    "rating_5": 52,
    "rating_4": 12,
    "rating_3": 3,
    "rating_2": 1,
    "rating_1": 0
  },
  "total_reviews": 68
}
```

**Use Cases:**
- Visualize rating distribution
- Track service quality
- Identify improvement areas

---

### 5. Monthly Summary

**GET** `/analytics/provider/monthly`

Returns monthly aggregated statistics for the past 12 months.

**Response:**
```json
{
  "monthly_stats": [
    {
      "month": "2025-11",
      "booking_count": 15,
      "completed_count": 12,
      "revenue": 36000.00,
      "new_reviews": 10,
      "average_rating": 4.8
    },
    {
      "month": "2025-10",
      "booking_count": 18,
      "completed_count": 16,
      "revenue": 48000.00,
      "new_reviews": 14,
      "average_rating": 4.7
    }
  ]
}
```

**Use Cases:**
- Track month-over-month growth
- Identify seasonal patterns
- Annual performance review

---

### 6. Track Profile View

**POST** `/analytics/profile-view`

Records a profile view (called by frontend when viewing provider profile).

**Request Body:**
```json
{
  "provider_id": 42
}
```

**Response:**
```json
{
  "message": "Profile view tracked"
}
```

**Note:** Automatically tracks viewer_id if authenticated, NULL for anonymous.

---

## Frontend Implementation

### React Dashboard Component

```typescript
import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { Line, Bar, Doughnut } from 'react-chartjs-2';

interface AnalyticsDashboard {
  profile_views: number;
  total_bookings: number;
  completed_bookings: number;
  cancelled_bookings: number;
  pending_bookings: number;
  total_revenue: number;
  average_rating: number;
  total_reviews: number;
  favorite_count: number;
  response_rate: number;
  average_response_time: number;
}

export const ProviderDashboard: React.FC = () => {
  const [dashboard, setDashboard] = useState<AnalyticsDashboard | null>(null);
  const [bookingStats, setBookingStats] = useState<any[]>([]);
  const [revenueBreakdown, setRevenueBreakdown] = useState<any[]>([]);
  const [ratingBreakdown, setRatingBreakdown] = useState<any>(null);
  const [period, setPeriod] = useState(30);

  useEffect(() => {
    fetchDashboard();
    fetchBookingStats();
    fetchRevenueBreakdown();
    fetchRatingBreakdown();
  }, [period]);

  const fetchDashboard = async () => {
    const response = await axios.get('/analytics/provider/dashboard');
    setDashboard(response.data);
  };

  const fetchBookingStats = async () => {
    const response = await axios.get(`/analytics/provider/bookings?period=${period}`);
    setBookingStats(response.data.stats);
  };

  const fetchRevenueBreakdown = async () => {
    const response = await axios.get('/analytics/provider/revenue');
    setRevenueBreakdown(response.data.revenue_breakdown);
  };

  const fetchRatingBreakdown = async () => {
    const response = await axios.get('/analytics/provider/ratings');
    setRatingBreakdown(response.data.breakdown);
  };

  if (!dashboard) return <div>Loading...</div>;

  return (
    <div className="provider-dashboard">
      <h1>Analytics Dashboard</h1>

      {/* Key Metrics */}
      <div className="metrics-grid">
        <MetricCard
          title="Profile Views"
          value={dashboard.profile_views}
          icon="👁️"
        />
        <MetricCard
          title="Total Bookings"
          value={dashboard.total_bookings}
          icon="📅"
        />
        <MetricCard
          title="Total Revenue"
          value={`฿${dashboard.total_revenue.toLocaleString()}`}
          icon="💰"
        />
        <MetricCard
          title="Average Rating"
          value={dashboard.average_rating.toFixed(1)}
          icon="⭐"
        />
        <MetricCard
          title="Favorites"
          value={dashboard.favorite_count}
          icon="❤️"
        />
        <MetricCard
          title="Response Rate"
          value={`${dashboard.response_rate.toFixed(1)}%`}
          icon="💬"
        />
      </div>

      {/* Booking Status */}
      <div className="booking-status">
        <h2>Booking Status</h2>
        <div className="status-grid">
          <StatusCard
            title="Completed"
            count={dashboard.completed_bookings}
            color="green"
          />
          <StatusCard
            title="Pending"
            count={dashboard.pending_bookings}
            color="yellow"
          />
          <StatusCard
            title="Cancelled"
            count={dashboard.cancelled_bookings}
            color="red"
          />
        </div>
      </div>

      {/* Booking Trend Chart */}
      <div className="chart-section">
        <h2>Booking Trend</h2>
        <select value={period} onChange={(e) => setPeriod(Number(e.target.value))}>
          <option value={7}>Last 7 days</option>
          <option value={30}>Last 30 days</option>
          <option value={90}>Last 90 days</option>
        </select>
        <Line
          data={{
            labels: bookingStats.map(s => s.date),
            datasets: [
              {
                label: 'Bookings',
                data: bookingStats.map(s => s.booking_count),
                borderColor: 'rgb(75, 192, 192)',
                tension: 0.1
              },
              {
                label: 'Revenue (฿)',
                data: bookingStats.map(s => s.revenue),
                borderColor: 'rgb(255, 99, 132)',
                yAxisID: 'revenue'
              }
            ]
          }}
          options={{
            scales: {
              y: { position: 'left', title: { display: true, text: 'Bookings' } },
              revenue: { position: 'right', title: { display: true, text: 'Revenue (฿)' } }
            }
          }}
        />
      </div>

      {/* Revenue by Package */}
      <div className="chart-section">
        <h2>Revenue by Package</h2>
        <Bar
          data={{
            labels: revenueBreakdown.map(r => r.package_name),
            datasets: [{
              label: 'Revenue (฿)',
              data: revenueBreakdown.map(r => r.total_revenue),
              backgroundColor: 'rgba(54, 162, 235, 0.5)'
            }]
          }}
        />
      </div>

      {/* Rating Distribution */}
      {ratingBreakdown && (
        <div className="chart-section">
          <h2>Rating Distribution</h2>
          <Doughnut
            data={{
              labels: ['5 Stars', '4 Stars', '3 Stars', '2 Stars', '1 Star'],
              datasets: [{
                data: [
                  ratingBreakdown.rating_5,
                  ratingBreakdown.rating_4,
                  ratingBreakdown.rating_3,
                  ratingBreakdown.rating_2,
                  ratingBreakdown.rating_1
                ],
                backgroundColor: [
                  '#4CAF50',
                  '#8BC34A',
                  '#FFC107',
                  '#FF9800',
                  '#F44336'
                ]
              }]
            }}
          />
        </div>
      )}
    </div>
  );
};

const MetricCard: React.FC<{ title: string; value: any; icon: string }> = ({ 
  title, value, icon 
}) => (
  <div className="metric-card">
    <div className="icon">{icon}</div>
    <div className="content">
      <h3>{title}</h3>
      <p className="value">{value}</p>
    </div>
  </div>
);

const StatusCard: React.FC<{ title: string; count: number; color: string }> = ({
  title, count, color
}) => (
  <div className={`status-card ${color}`}>
    <h4>{title}</h4>
    <p className="count">{count}</p>
  </div>
);
```

---

## Tracking Profile Views

### Implementation in Provider Profile Page

```typescript
// When a user visits a provider profile
useEffect(() => {
  const trackView = async () => {
    try {
      await axios.post('/analytics/profile-view', {
        provider_id: providerId
      });
    } catch (error) {
      console.error('Failed to track profile view:', error);
    }
  };

  trackView();
}, [providerId]);
```

**Note:** Profile views are tracked:
- Once per visit per user
- Increments count for repeat visitors
- Supports anonymous users (viewer_id = NULL)

---

## Best Practices

### 1. Data Refresh
- Dashboard should auto-refresh every 5-10 minutes
- Manual refresh button for immediate updates
- Cache data to reduce API calls

### 2. Date Filters
- Provide multiple period options (7, 30, 90 days)
- Allow custom date range selection
- Display period in chart titles

### 3. Visualization
- Use appropriate chart types (Line for trends, Bar for comparisons, Doughnut for distributions)
- Color-code metrics (green = positive, red = negative, yellow = neutral)
- Show tooltips with detailed information

### 4. Performance
- Lazy load charts (only fetch when visible)
- Paginate large datasets
- Use CDN for chart libraries

---

## SQL Queries

### Custom Analytics Queries

**Get top performing days:**
```sql
SELECT 
    booking_date::date,
    COUNT(*) as bookings,
    SUM(total_price) as revenue
FROM bookings
WHERE provider_id = $1 
AND status = 'completed'
AND booking_date >= CURRENT_DATE - INTERVAL '90 days'
GROUP BY booking_date::date
ORDER BY revenue DESC
LIMIT 10;
```

**Get repeat customers:**
```sql
SELECT 
    client_id,
    u.username,
    COUNT(*) as booking_count,
    SUM(total_price) as total_spent
FROM bookings b
JOIN users u ON b.client_id = u.user_id
WHERE provider_id = $1 
AND status = 'completed'
GROUP BY client_id, u.username
HAVING COUNT(*) > 1
ORDER BY booking_count DESC;
```

**Get peak hours:**
```sql
SELECT 
    EXTRACT(HOUR FROM start_time) as hour,
    COUNT(*) as booking_count
FROM bookings
WHERE provider_id = $1
AND status = 'completed'
GROUP BY EXTRACT(HOUR FROM start_time)
ORDER BY booking_count DESC;
```

---

## Future Enhancements

- [ ] Export data to CSV/Excel
- [ ] Email weekly/monthly reports
- [ ] Comparison with platform average
- [ ] Predictive analytics (forecast revenue)
- [ ] Customer retention rate
- [ ] Conversion funnel (views → favorites → bookings)
- [ ] Geographic distribution of clients
- [ ] Peak booking time recommendations
- [ ] A/B testing for packages
- [ ] Goal setting and tracking

---

**Last Updated:** November 13, 2025
