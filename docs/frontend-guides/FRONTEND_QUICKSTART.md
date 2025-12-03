# Frontend Quick Start Guide - SkillMatch API

> **Updated:** December 2, 2025  
> **For:** React, Vue, Angular, or any Frontend Framework  
> **Backend:** http://localhost:8080

---

## 🚀 5-Minute Setup

### Step 1: Test Backend Connection
```bash
curl http://localhost:8080/ping
# Expected: {"message":"pong!","postgres_time":"2025-12-02T..."}
```

### Step 2: Create API Helper

#### React/Next.js
```javascript
// lib/api.js
const API_BASE = 'http://localhost:8080';

export async function apiCall(endpoint, options = {}) {
  const token = localStorage.getItem('token');
  
  const config = {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
      ...(token && { 'Authorization': `Bearer ${token}` })
    }
  };
  
  const response = await fetch(`${API_BASE}${endpoint}`, config);
  const data = await response.json();
  
  if (!response.ok) {
    throw new Error(data.error || 'API Error');
  }
  
  return data;
}

// Usage
import { apiCall } from '@/lib/api';

// Public endpoint
const categories = await apiCall('/service-categories');

// Protected endpoint (requires login)
const profile = await apiCall('/users/me');

// POST request
const login = await apiCall('/login', {
  method: 'POST',
  body: JSON.stringify({ email: 'user@example.com', password: 'pass123' })
});
```

#### Vue/Nuxt
```javascript
// composables/useApi.js
export const useApi = () => {
  const apiBase = 'http://localhost:8080';
  
  const call = async (endpoint, options = {}) => {
    const token = localStorage.getItem('token');
    
    const config = {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
        ...(token && { 'Authorization': `Bearer ${token}` })
      }
    };
    
    const response = await fetch(`${apiBase}${endpoint}`, config);
    const data = await response.json();
    
    if (!response.ok) {
      throw new Error(data.error || 'API Error');
    }
    
    return data;
  };
  
  return { call };
};

// Usage in component
const { call } = useApi();
const categories = await call('/service-categories');
```

#### Angular
```typescript
// services/api.service.ts
import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Observable, throwError } from 'rxjs';
import { catchError } from 'rxjs/operators';

@Injectable({ providedIn: 'root' })
export class ApiService {
  private apiBase = 'http://localhost:8080';

  constructor(private http: HttpClient) {}

  private getHeaders(): HttpHeaders {
    const token = localStorage.getItem('token');
    let headers = new HttpHeaders({ 'Content-Type': 'application/json' });
    
    if (token) {
      headers = headers.set('Authorization', `Bearer ${token}`);
    }
    
    return headers;
  }

  get<T>(endpoint: string): Observable<T> {
    return this.http.get<T>(`${this.apiBase}${endpoint}`, {
      headers: this.getHeaders()
    }).pipe(catchError(this.handleError));
  }

  post<T>(endpoint: string, data: any): Observable<T> {
    return this.http.post<T>(`${this.apiBase}${endpoint}`, data, {
      headers: this.getHeaders()
    }).pipe(catchError(this.handleError));
  }

  private handleError(error: any) {
    return throwError(() => new Error(error.error?.error || 'API Error'));
  }
}
```

---

## 🔐 Authentication Flow

### 1. Login with Email/Password

```javascript
// Step 1: Login
const loginData = await apiCall('/login', {
  method: 'POST',
  body: JSON.stringify({
    email: 'user@example.com',
    password: 'SecurePass123!'
  })
});

// Step 2: Save token
localStorage.setItem('token', loginData.token);

// Step 3: Get user profile
const user = await apiCall('/users/me');
console.log(user); // { user_id, username, profile_picture_url, ... }
```

### 2. Register New User

```javascript
// Step 1: Send OTP to email
await apiCall('/auth/send-verification', {
  method: 'POST',
  body: JSON.stringify({ email: 'newuser@example.com' })
});

// Step 2: Verify OTP (user enters 6-digit code)
const verifyData = await apiCall('/auth/verify-email', {
  method: 'POST',
  body: JSON.stringify({
    email: 'newuser@example.com',
    otp: '123456' // From user input
  })
});

// Step 3: Complete registration
const registerData = await apiCall('/register', {
  method: 'POST',
  body: JSON.stringify({
    email: 'newuser@example.com',
    username: 'johndoe',
    password: 'SecurePass123!',
    first_name: 'John',
    last_name: 'Doe',
    gender_id: 1, // 1=Male, 2=Female, 3=Other, 4=Prefer not to say
    verification_token: verifyData.verification_token
  })
});

// Save token
localStorage.setItem('token', registerData.token);
```

### 3. Google OAuth (Recommended)

```html
<!-- Add Google Sign-In Library -->
<script src="https://accounts.google.com/gsi/client" async defer></script>

<div id="g_id_onload"
     data-client_id="171089417301-each0gvj9d5l38bgkklu0n36p5eo5eau.apps.googleusercontent.com"
     data-callback="handleGoogleSignIn">
</div>
```

```javascript
async function handleGoogleSignIn(response) {
  const code = response.code; // Authorization code from Google
  
  // Send to backend
  const data = await apiCall('/auth/google', {
    method: 'POST',
    body: JSON.stringify({ code })
  });
  
  // Save token
  localStorage.setItem('token', data.token);
  
  // Get user profile (includes profile_picture_url from Google)
  const user = await apiCall('/users/me');
  console.log(user.profile_picture_url); // Google profile picture
}
```

---

## 📱 Common Use Cases

### Browse Providers (Public)

```javascript
// Get all service categories
const { categories } = await apiCall('/service-categories');
// Returns: [{ category_id: 1, name: "Massage", icon: "💆", name_thai: "นวด" }, ...]

// Search providers with filters (NEW!)
const results = await apiCall('/browse/search?location=Bangkok&rating=4&sort=rating&limit=20');
console.log(results.providers); // Array of providers
console.log(results.pagination); // { total, page, limit, total_pages }

// Get provider details (public)
const provider = await apiCall('/provider/456/public');
console.log(provider); // { user_id, username, bio, categories, rating_avg, ... }

// Get provider photos
const { photos } = await apiCall('/provider/456/photos');
console.log(photos); // [{ photo_id, photo_url, caption, sort_order, ... }]

// Get provider packages
const { packages } = await apiCall('/packages/456');
console.log(packages); // [{ package_id, name, price, duration_hours, ... }]

// Get provider reviews
const reviews = await apiCall('/reviews/456?page=1&limit=10');
console.log(reviews.reviews); // [{ review_id, rating, comment, ... }]
```

### Favorites (Protected)

```javascript
// Check if favorited (works even without login - returns false)
const { is_favorite } = await apiCall('/favorites/check/456');

// Add to favorites (requires login)
await apiCall('/favorites', {
  method: 'POST',
  body: JSON.stringify({ provider_id: 456 })
});

// Remove from favorites
await apiCall('/favorites/456', { method: 'DELETE' });

// Get my favorites
const { favorites } = await apiCall('/favorites');
console.log(favorites); // Array of favorite providers
```

### Bookings (Protected)

```javascript
// Create booking with Stripe payment
const booking = await apiCall('/bookings/create-with-payment', {
  method: 'POST',
  body: JSON.stringify({
    provider_id: 456,
    package_id: 1,
    booking_date: '2025-12-10',
    booking_time: '14:00:00',
    notes: 'Please bring massage oil'
  })
});

// Redirect to Stripe Checkout
window.location.href = booking.checkout_url;

// After payment, get bookings
const { bookings } = await apiCall('/bookings/my?status=all');
console.log(bookings); // Array of bookings
```

### Messaging (Protected)

```javascript
// Get conversations
const { conversations } = await apiCall('/conversations');
console.log(conversations); // [{ conversation_id, other_user_id, unread_count, ... }]

// Get messages in conversation
const { messages } = await apiCall('/conversations/123/messages?limit=50');

// Send message
await apiCall('/messages', {
  method: 'POST',
  body: JSON.stringify({
    receiver_id: 456,
    content: 'Hello, I am interested in your service'
  })
});

// Mark as read
await apiCall('/messages/read', {
  method: 'PATCH',
  body: JSON.stringify({ message_ids: [1, 2, 3] })
});
```

### Real-time WebSocket

```javascript
// Connect
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onopen = () => {
  // Authenticate
  ws.send(JSON.stringify({
    type: 'auth',
    payload: { token: localStorage.getItem('token') }
  }));
};

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  
  switch(data.type) {
    case 'new_message':
      // Show notification
      showNotification('New message from ' + data.payload.sender);
      // Update chat UI
      addMessageToChat(data.payload);
      break;
      
    case 'notification':
      // Update notification badge
      updateNotificationCount(data.payload);
      break;
      
    case 'booking_update':
      // Refresh bookings list
      refreshBookings();
      break;
  }
};

ws.onerror = (error) => console.error('WebSocket error:', error);
ws.onclose = () => console.log('WebSocket closed - reconnect logic here');
```

---

## 🎨 React Example Components

### Provider Search with Filters

```jsx
import { useState, useEffect } from 'react';
import { apiCall } from '@/lib/api';

function ProviderSearch() {
  const [providers, setProviders] = useState([]);
  const [loading, setLoading] = useState(false);
  const [filters, setFilters] = useState({
    location: '',
    rating: '',
    tier: '',
    category: '',
    service_type: '',
    sort: 'rating',
    page: 1,
    limit: 20
  });
  const [pagination, setPagination] = useState(null);

  useEffect(() => {
    const fetchProviders = async () => {
      setLoading(true);
      try {
        // Build query string (skip empty values)
        const params = new URLSearchParams();
        Object.entries(filters).forEach(([key, value]) => {
          if (value) params.append(key, value);
        });
        
        const data = await apiCall(`/browse/search?${params}`);
        setProviders(data.providers);
        setPagination(data.pagination);
      } catch (error) {
        console.error('Search failed:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchProviders();
  }, [filters]);

  return (
    <div className="provider-search">
      {/* Filters */}
      <div className="filters">
        <input
          type="text"
          placeholder="Location (e.g., Bangkok)"
          value={filters.location}
          onChange={(e) => setFilters({...filters, location: e.target.value})}
        />
        
        <select
          value={filters.rating}
          onChange={(e) => setFilters({...filters, rating: e.target.value})}
        >
          <option value="">All Ratings</option>
          <option value="4">4+ stars</option>
          <option value="4.5">4.5+ stars</option>
        </select>
        
        <select
          value={filters.service_type}
          onChange={(e) => setFilters({...filters, service_type: e.target.value})}
        >
          <option value="">All Types</option>
          <option value="Incall">Incall</option>
          <option value="Outcall">Outcall</option>
          <option value="Both">Both</option>
        </select>
        
        <select
          value={filters.sort}
          onChange={(e) => setFilters({...filters, sort: e.target.value})}
        >
          <option value="rating">Best Rating</option>
          <option value="reviews">Most Reviews</option>
          <option value="price">Lowest Price</option>
        </select>
      </div>

      {/* Results */}
      {loading ? (
        <div>Loading...</div>
      ) : (
        <>
          <div className="providers-grid">
            {providers.map(provider => (
              <ProviderCard key={provider.user_id} provider={provider} />
            ))}
          </div>
          
          {pagination && (
            <Pagination
              currentPage={pagination.page}
              totalPages={pagination.total_pages}
              onPageChange={(page) => setFilters({...filters, page})}
            />
          )}
        </>
      )}
    </div>
  );
}

function ProviderCard({ provider }) {
  return (
    <div className="provider-card">
      <img 
        src={provider.profile_picture_url || '/default-avatar.png'} 
        alt={provider.username} 
      />
      <h3>{provider.username}</h3>
      <p>{provider.bio}</p>
      <div className="rating">
        ⭐ {provider.rating_avg.toFixed(1)} ({provider.review_count} reviews)
      </div>
      <div className="price">
        From ฿{provider.min_price}
      </div>
      <span className="tier">{provider.provider_level_name}</span>
    </div>
  );
}
```

### Authentication Component

```jsx
import { useState } from 'react';
import { apiCall } from '@/lib/api';

function LoginForm() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');

  const handleLogin = async (e) => {
    e.preventDefault();
    setError('');
    
    try {
      const data = await apiCall('/login', {
        method: 'POST',
        body: JSON.stringify({ email, password })
      });
      
      localStorage.setItem('token', data.token);
      window.location.href = '/dashboard'; // Redirect after login
    } catch (err) {
      setError(err.message);
    }
  };

  return (
    <form onSubmit={handleLogin}>
      {error && <div className="error">{error}</div>}
      
      <input
        type="email"
        placeholder="Email"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        required
      />
      
      <input
        type="password"
        placeholder="Password"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
        required
      />
      
      <button type="submit">Login</button>
      
      {/* Google OAuth Button */}
      <div id="g_id_onload"
           data-client_id="171089417301-each0gvj9d5l38bgkklu0n36p5eo5eau.apps.googleusercontent.com"
           data-callback="handleGoogleSignIn">
      </div>
    </form>
  );
}
```

---

## 🧪 Testing

```javascript
// Test API connection
describe('API Connection', () => {
  it('should ping backend', async () => {
    const response = await fetch('http://localhost:8080/ping');
    const data = await response.json();
    expect(data.message).toBe('pong!');
  });
});

// Test authentication
describe('Authentication', () => {
  it('should login successfully', async () => {
    const data = await apiCall('/login', {
      method: 'POST',
      body: JSON.stringify({
        email: 'test@example.com',
        password: 'password123'
      })
    });
    
    expect(data).toHaveProperty('token');
    expect(data).toHaveProperty('message', 'Login successful');
  });
});

// Test browse/search
describe('Provider Search', () => {
  it('should search providers', async () => {
    const data = await apiCall('/browse/search?location=Bangkok');
    
    expect(data).toHaveProperty('providers');
    expect(data).toHaveProperty('pagination');
    expect(Array.isArray(data.providers)).toBe(true);
  });
});
```

---

## 📚 TypeScript Types

```typescript
// types/api.ts

export interface User {
  user_id: number;
  username: string;
  email: string;
  first_name?: string;
  last_name?: string;
  tier_id: number;
  tier_name: string;
  profile_picture_url?: string;
  verification_status: string;
  created_at: string;
}

export interface Provider {
  user_id: number;
  username: string;
  profile_picture_url?: string;
  bio?: string;
  provider_level_id: number;
  provider_level_name: string;
  rating_avg: number;
  review_count: number;
  service_type: 'Incall' | 'Outcall' | 'Both';
  location?: string;
  min_price?: number;
}

export interface SearchFilters {
  location?: string;
  province?: string;
  district?: string;
  rating?: number;
  tier?: number;
  category?: number;
  service_type?: 'Incall' | 'Outcall' | 'Both';
  sort?: 'rating' | 'reviews' | 'price';
  page?: number;
  limit?: number;
}

export interface SearchResponse {
  providers: Provider[];
  pagination: {
    total: number;
    page: number;
    limit: number;
    total_pages: number;
  };
  filters_applied: Record<string, string>;
}

export interface ServiceCategory {
  category_id: number;
  name: string;
  name_thai: string;
  icon: string;
  description?: string;
}

export interface Booking {
  booking_id: number;
  provider_id: number;
  provider_username: string;
  package_name: string;
  booking_date: string;
  booking_time: string;
  status: 'pending' | 'paid' | 'confirmed' | 'completed' | 'cancelled';
  total_price: number;
  created_at: string;
}

export interface Message {
  message_id: number;
  sender_id: number;
  receiver_id: number;
  content: string;
  is_read: boolean;
  created_at: string;
}
```

---

## 🎯 Next Steps

1. ✅ Setup API helper (5 min)
2. ✅ Test backend connection (2 min)
3. ✅ Implement authentication (15 min)
4. ✅ Build provider search (30 min)
5. ✅ Add booking flow (30 min)
6. ✅ Integrate WebSocket (15 min)
7. ✅ Add tests (30 min)

**Total Time:** ~2 hours for complete integration

---

## 📞 Need Help?

### Quick Diagnostics
```bash
# Test backend
curl http://localhost:8080/ping

# Test auth (use GOD token from FRONTEND_SETUP.md)
curl -H "Authorization: Bearer eyJhbGc..." http://localhost:8080/users/me

# Test search
curl "http://localhost:8080/browse/search?location=Bangkok"
```

### Common Issues

**Issue:** CORS errors  
**Solution:** Backend allows `localhost:3000`, `localhost:5173`, `localhost:8080`

**Issue:** 401 Unauthorized  
**Solution:** Check if token is expired (7 days) or invalid

**Issue:** Profile pictures not showing  
**Solution:** Use `profile_picture_url` (not `profile_image_url`)

---

**Ready to build? Start coding! 🚀**
