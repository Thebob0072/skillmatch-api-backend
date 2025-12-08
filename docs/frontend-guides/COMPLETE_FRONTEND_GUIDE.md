# 📚 SkillMatch API - Complete Frontend Integration Guide

> **Version:** 2.0 (December 2, 2025)  
> **Status:** Production Ready  
> **Backend:** http://localhost:8080  
> **For:** React, Vue, Angular, or any Frontend Framework

---

## 📖 Table of Contents

1. [🚀 Quick Start (5 Minutes)](#-quick-start-5-minutes)
2. [🌐 Translation & Localization (Thai/English)](#-translation--localization-thaienglish)
3. [⚠️ Breaking Changes (December 2025)](#️-breaking-changes-december-2025)
4. [🔐 Authentication & Authorization](#-authentication--authorization)
5. [🔍 Browse & Search System](#-browse--search-system)
6. [👤 Profile Management](#-profile-management)
7. [💬 Messaging System](#-messaging-system)
8. [📦 Booking & Payment System](#-booking--payment-system)
9. [💰 Financial System (Provider Wallet)](#-financial-system-provider-wallet)
10. [🎨 Service Categories](#-service-categories)
11. [📡 Complete API Reference](#-complete-api-reference)
12. [🧩 React Components Examples](#-react-components-examples)
13. [🎯 Vue/Angular Examples](#-vueangular-examples)
14. [📱 Real-time WebSocket](#-real-time-websocket)
15. [🧪 Testing Guide](#-testing-guide)
16. [🔧 Troubleshooting](#-troubleshooting)

---

## 🚀 Quick Start (5 Minutes)

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

// Usage Examples:
// Public: const categories = await apiCall('/service-categories');
// Protected: const profile = await apiCall('/users/me');
// POST: const login = await apiCall('/login', { method: 'POST', body: JSON.stringify({...}) });
```

### Step 3: Test Authentication
```javascript
// Login
const loginData = await apiCall('/login', {
  method: 'POST',
  body: JSON.stringify({
    email: 'god@skillmatch.com',
    password: 'godpass123'
  })
});

// Save token
localStorage.setItem('token', loginData.token);

// Get user profile
const user = await apiCall('/users/me');
console.log(user); // { user_id, username, profile_picture_url, ... }
```

---

## 🌐 Translation & Localization (Thai/English)

### Overview

SkillMatch API supports **Thai (ไทย)** and **English** languages. Some data comes from the backend in Thai, while other fields require frontend translation.

### ✅ Backend Provides Thai Data (No Translation Needed)

#### 1. Service Categories
```javascript
const { categories } = await apiCall('/service-categories');
// Response includes BOTH languages:
{
  "category_id": 1,
  "name": "Massage",           // English
  "name_thai": "นวดแผนไทย",    // Thai ⭐
  "icon": "💆"
}

// Display based on user's language preference:
const displayName = language === 'th' ? category.name_thai : category.name;
```

#### 2. Location Fields (Already Thai)
```javascript
// These fields are stored in Thai - display as-is:
{
  "province": "กรุงเทพมหานคร",
  "district": "บางรัก", 
  "sub_district": "สีลม",
  "address_line1": "123 ถนนสีลม"
}
```

### ❌ Frontend Must Translate (Backend Sends English)

#### 1. Provider Tier Names
```javascript
// Backend returns: "General", "Silver", "Diamond", "Premium"
const tierTranslations = {
  th: {
    "General": "ทั่วไป",
    "Silver": "เงิน", 
    "Diamond": "เพชร",
    "Premium": "พรีเมียม"
  },
  en: {
    "General": "General",
    "Silver": "Silver",
    "Diamond": "Diamond", 
    "Premium": "Premium"
  }
};

// Usage:
const displayTier = tierTranslations[language][provider.provider_level_name];
```

#### 2. Service Types
```javascript
// Backend returns: "Incall", "Outcall", "Both"
const serviceTypeTranslations = {
  th: {
    "Incall": "บริการที่ร้าน",
    "Outcall": "บริการนอกสถานที่",
    "Both": "ทั้งสองแบบ"
  },
  en: {
    "Incall": "Incall",
    "Outcall": "Outcall",
    "Both": "Both"
  }
};
```

#### 3. Booking Status
```javascript
// Backend returns: "pending", "paid", "confirmed", "completed", "cancelled"
const bookingStatusTranslations = {
  th: {
    "pending": "รอดำเนินการ",
    "paid": "ชำระแล้ว",
    "confirmed": "ยืนยันแล้ว",
    "completed": "เสร็จสิ้น",
    "cancelled": "ยกเลิก"
  }
};
```

#### 4. Transaction Types
```javascript
// Backend returns: "earning", "withdrawal", "refund"
const transactionTypeTranslations = {
  th: {
    "earning": "รายได้",
    "withdrawal": "ถอนเงิน",
    "refund": "คืนเงิน"
  }
};
```

#### 5. Gender
```javascript
// Backend returns: gender_id (1, 2, 3, 4)
const genderTranslations = {
  th: {
    1: "ชาย",
    2: "หญิง",
    3: "อื่นๆ",
    4: "ไม่ระบุ"
  },
  en: {
    1: "Male",
    2: "Female",
    3: "Other",
    4: "Prefer not to say"
  }
};
```

### 🎯 Implementation Guide

#### React Translation Hook
```javascript
// hooks/useTranslation.js
import { createContext, useContext, useState } from 'react';

const translations = {
  th: {
    common: {
      save: "บันทึก",
      cancel: "ยกเลิก",
      search: "ค้นหา",
      loading: "กำลังโหลด..."
    },
    provider: {
      tiers: {
        "General": "ทั่วไป",
        "Silver": "เงิน",
        "Diamond": "เพชร", 
        "Premium": "พรีเมียม"
      },
      serviceTypes: {
        "Incall": "บริการที่ร้าน",
        "Outcall": "บริการนอกสถานที่",
        "Both": "ทั้งสองแบบ"
      }
    },
    booking: {
      status: {
        "pending": "รอดำเนินการ",
        "paid": "ชำระแล้ว",
        "confirmed": "ยืนยันแล้ว",
        "completed": "เสร็จสิ้น",
        "cancelled": "ยกเลิก"
      }
    }
  },
  en: {
    common: {
      save: "Save",
      cancel: "Cancel",
      search: "Search",
      loading: "Loading..."
    },
    provider: {
      tiers: {
        "General": "General",
        "Silver": "Silver",
        "Diamond": "Diamond",
        "Premium": "Premium"
      },
      serviceTypes: {
        "Incall": "Incall",
        "Outcall": "Outcall",
        "Both": "Both"
      }
    },
    booking: {
      status: {
        "pending": "Pending",
        "paid": "Paid",
        "confirmed": "Confirmed",
        "completed": "Completed",
        "cancelled": "Cancelled"
      }
    }
  }
};

const TranslationContext = createContext();

export function TranslationProvider({ children }) {
  const [language, setLanguage] = useState(
    localStorage.getItem('language') || 'th'
  );

  const t = (key) => {
    const keys = key.split('.');
    let value = translations[language];
    for (const k of keys) {
      value = value?.[k];
      if (!value) return key;
    }
    return value;
  };

  const changeLanguage = (lang) => {
    setLanguage(lang);
    localStorage.setItem('language', lang);
  };

  return (
    <TranslationContext.Provider value={{ t, language, changeLanguage }}>
      {children}
    </TranslationContext.Provider>
  );
}

export function useTranslation() {
  return useContext(TranslationContext);
}
```

#### Usage in Components
```jsx
import { useTranslation } from '@/hooks/useTranslation';

function ProviderCard({ provider }) {
  const { t, language } = useTranslation();

  return (
    <div className="provider-card">
      {/* Category name (from backend) */}
      <p className="category">
        {language === 'th' ? provider.category_name_thai : provider.category_name}
      </p>
      
      {/* Provider tier (translate) */}
      <span className="tier">
        {t(`provider.tiers.${provider.provider_level_name}`)}
      </span>
      
      {/* Service type (translate) */}
      <p className="service-type">
        {t(`provider.serviceTypes.${provider.service_type}`)}
      </p>
      
      {/* Location (Thai from backend - no translation) */}
      <p className="location">{provider.province}, {provider.district}</p>
      
      <button>{t('common.search')}</button>
    </div>
  );
}

function LanguageSwitcher() {
  const { language, changeLanguage } = useTranslation();

  return (
    <div className="language-switcher">
      <button
        onClick={() => changeLanguage('th')}
        className={language === 'th' ? 'active' : ''}
      >
        ไทย
      </button>
      <button
        onClick={() => changeLanguage('en')}
        className={language === 'en' ? 'active' : ''}
      >
        EN
      </button>
    </div>
  );
}
```

### 📋 Translation Checklist

**✅ Already Thai (display as-is):**
- ☑ Category `name_thai` field
- ☑ Province, district, sub_district
- ☑ Address fields (user input)
- ☑ User bio, package names, review comments

**❌ Requires Frontend Translation:**
- ☐ Provider tier names (General → ทั่วไป)
- ☐ Service types (Incall → บริการที่ร้าน)
- ☐ Booking status (confirmed → ยืนยันแล้ว)
- ☐ Transaction types (earning → รายได้)
- ☐ Gender labels (1 → ชาย/Male)
- ☐ UI labels (buttons, messages, errors)
- ☐ Form labels and validation messages

### 🌏 Best Practices

1. **Store language preference** in `localStorage`
2. **Provide language switcher** in navbar
3. **Use consistent translation keys** (e.g., `provider.tiers.Premium`)
4. **Don't translate location data** - already in Thai
5. **Test both languages** before deployment
6. **Default to Thai** for Thai audience

### 🗣️ Provider Languages Filter

Backend stores provider's spoken languages in the `languages` field as an array of language codes.

#### Supported Language Codes:
```javascript
const supportedLanguages = {
  th: 'Thai (ไทย)',
  en: 'English',
  zh: 'Chinese (中文)',
  ja: 'Japanese (日本語)',
  ko: 'Korean (한국어)',
  fr: 'French (Français)',
  de: 'German (Deutsch)',
  es: 'Spanish (Español)',
  ru: 'Russian (Русский)',
  ar: 'Arabic (العربية)'
};
```

#### Filter by Languages:
```javascript
// Single language
const results = await apiCall('/browse/search?languages=th');

// Multiple languages (providers who speak ANY of these)
const results = await apiCall('/browse/search?languages=th,en,zh');

// Combined with other filters
const results = await apiCall('/browse/search?location=Bangkok&languages=en,zh&rating=4');
```

#### Display Languages in Profile:
```jsx
function ProviderProfile({ provider }) {
  const languageNames = {
    th: { th: 'ไทย', en: 'Thai' },
    en: { th: 'อังกฤษ', en: 'English' },
    zh: { th: 'จีน', en: 'Chinese' },
    ja: { th: 'ญี่ปุ่น', en: 'Japanese' },
    ko: { th: 'เกาหลี', en: 'Korean' }
  };
  
  const { language } = useTranslation();

  return (
    <div>
      <h3>Languages Spoken (ภาษาที่สื่อสารได้)</h3>
      <div className="languages">
        {provider.languages.map(langCode => (
          <span key={langCode} className="language-badge">
            {languageNames[langCode]?.[language] || langCode}
          </span>
        ))}
      </div>
    </div>
  );
}
```

#### Advanced Language Filter Component:
```jsx
function LanguageFilter({ selectedLanguages, onChange }) {
  const { t } = useTranslation();
  
  const languages = [
    { code: 'th', label: 'ไทย', flag: '🇹🇭' },
    { code: 'en', label: 'English', flag: '🇬🇧' },
    { code: 'zh', label: '中文', flag: '🇨🇳' },
    { code: 'ja', label: '日本語', flag: '🇯🇵' },
    { code: 'ko', label: '한국어', flag: '🇰🇷' }
  ];

  const toggleLanguage = (code) => {
    const current = selectedLanguages.split(',').filter(Boolean);
    const newSelection = current.includes(code)
      ? current.filter(l => l !== code)
      : [...current, code];
    onChange(newSelection.join(','));
  };

  return (
    <div className="language-filter">
      <label>{t('search.languages')}</label>
      <div className="language-checkboxes">
        {languages.map(({ code, label, flag }) => (
          <label key={code} className="language-option">
            <input
              type="checkbox"
              checked={selectedLanguages.includes(code)}
              onChange={() => toggleLanguage(code)}
            />
            <span>{flag} {label}</span>
          </label>
        ))}
      </div>
    </div>
  );
}

// Usage:
<LanguageFilter
  selectedLanguages={filters.languages}
  onChange={(langs) => setFilters({...filters, languages: langs, page: 1})}
/>
```

---

## ⚠️ Breaking Changes (December 2025)

### 🔴 CRITICAL: Profile Picture Field Renamed

**What Changed:**
- ❌ Removed: `profile_image_url` (from user_profiles table)
- ❌ Removed: `google_profile_picture` (from users table)  
- ✅ New: `profile_picture_url` (unified field in users table)

**Migration Required:**

```javascript
// ❌ OLD CODE (WILL NOT WORK)
const ProfileCard = ({ user }) => {
  return <img src={user.profile_image_url} alt="Profile" />;
};

// ✅ NEW CODE (CORRECT)
const ProfileCard = ({ user }) => {
  return <img src={user.profile_picture_url} alt="Profile" />;
};
```

**Affected Endpoints:**
- `GET /users/me`
- `GET /profile/me`
- `GET /provider/:userId`
- `GET /provider/:userId/public`
- `GET /browse/search` ⭐ (NEW)
- `GET /favorites`

**Backward Compatibility (Temporary):**
```javascript
// If you need to support both old and new backends
const getProfilePicture = (user) => {
  return user.profile_picture_url || user.profile_image_url || user.google_profile_picture || '/default-avatar.png';
};
```

### ✨ New Features Added

#### 1. Advanced Browse/Search Endpoint
```javascript
// OLD: Limited to category-based browse
const providers = await apiCall('/categories/1/providers');

// NEW: Powerful multi-filter search
const results = await apiCall('/browse/search?location=Bangkok&rating=4&tier=3&sort=rating&limit=20');
```

**Query Parameters:**
- `location` - Search location text
- `province` - Exact province
- `district` - Exact district
- `rating` - Minimum rating (1-5)
- `tier` - Provider level (1-4)
- `category` - Category ID
- `service_type` - "Incall", "Outcall", "Both"
- `languages` - Comma-separated language codes (e.g., "th,en,zh")
- `sort` - "rating", "reviews", "price"
- `page`, `limit` - Pagination

**Response:**
```json
{
  "providers": [...],
  "pagination": {
    "total": 50,
    "page": 1,
    "limit": 20,
    "total_pages": 3
  },
  "filters_applied": {...}
}
```

#### 2. Performance Improvements
- Database optimized: 50-80% faster queries
- Added 9 new performance indexes
- Removed 2 duplicate indexes
- VACUUM ANALYZE completed

---

## 🔐 Authentication & Authorization

### Public Endpoints (No Token Required)
```javascript
// Service categories
const categories = await apiCall('/service-categories');

// Provider public profile
const provider = await apiCall('/provider/456/public');

// Provider photos
const photos = await apiCall('/provider/456/photos');

// Provider packages
const packages = await apiCall('/packages/456');

// Provider reviews
const reviews = await apiCall('/reviews/456');

// Review statistics
const stats = await apiCall('/reviews/stats/456');

// Favorites check (returns false if no token)
const { is_favorite } = await apiCall('/favorites/check/456');
```

### Protected Endpoints (Token Required)
```javascript
// Full provider profile (with sensitive data)
const provider = await apiCall('/provider/456'); // age, height, service_type included

// Browse/Search (multi-filter)
const results = await apiCall('/browse/search?location=Bangkok');

// User profile
const profile = await apiCall('/users/me');

// Favorites
await apiCall('/favorites', { method: 'POST', body: JSON.stringify({ provider_id: 456 }) });
const favorites = await apiCall('/favorites');

// Bookings
const bookings = await apiCall('/bookings/my');

// Messages
const conversations = await apiCall('/conversations');
```

### Authentication Flow

#### 1. Email/Password Login
```javascript
const loginData = await apiCall('/login', {
  method: 'POST',
  body: JSON.stringify({
    email: 'user@example.com',
    password: 'SecurePass123!'
  })
});

localStorage.setItem('token', loginData.token);
// Token valid for 7 days
```

#### 2. Google OAuth (Recommended)
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
  const code = response.code;
  
  const data = await apiCall('/auth/google', {
    method: 'POST',
    body: JSON.stringify({ code })
  });
  
  localStorage.setItem('token', data.token);
  
  // User profile includes profile_picture_url from Google
  const user = await apiCall('/users/me');
  console.log(user.profile_picture_url); // Google profile picture URL
}
```

#### 3. Register New User
```javascript
// Step 1: Send OTP
await apiCall('/auth/send-verification', {
  method: 'POST',
  body: JSON.stringify({ email: 'newuser@example.com' })
});

// Step 2: Verify OTP
const verifyData = await apiCall('/auth/verify-email', {
  method: 'POST',
  body: JSON.stringify({
    email: 'newuser@example.com',
    otp: '123456'
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
    gender_id: 1,
    verification_token: verifyData.verification_token
  })
});

localStorage.setItem('token', registerData.token);
```

#### 4. Provider Registration (Enhanced)
```javascript
const providerData = await apiCall('/register/provider', {
  method: 'POST',
  body: JSON.stringify({
    username: 'massage_pro',
    email: 'provider@example.com',
    password: 'SecurePass123!',
    gender_id: 2,
    phone: '0812345678',
    otp: '123456',
    category_ids: [1, 2, 5], // 1-5 categories
    service_type: 'Both', // "Incall", "Outcall", "Both"
    bio: 'Professional massage therapist...',
    province: 'Bangkok',
    district: 'Sukhumvit'
  })
});

localStorage.setItem('token', providerData.token);

// Next step: Upload documents (National ID, Health Certificate)
```

---

## 🔍 Browse & Search System

### New Advanced Search Endpoint

```javascript
// Multi-filter search
async function searchProviders(filters = {}) {
  const params = new URLSearchParams();
  
  // Location filters
  if (filters.location) params.append('location', filters.location);
  if (filters.province) params.append('province', filters.province);
  if (filters.district) params.append('district', filters.district);
  
  // Quality filters
  if (filters.rating) params.append('rating', filters.rating);
  if (filters.tier) params.append('tier', filters.tier);
  
  // Category & Service type
  if (filters.category) params.append('category', filters.category);
  if (filters.service_type) params.append('service_type', filters.service_type);
  
  // Languages filter
  if (filters.languages) params.append('languages', filters.languages);
  
  // Sorting & Pagination
  if (filters.sort) params.append('sort', filters.sort);
  if (filters.page) params.append('page', filters.page);
  if (filters.limit) params.append('limit', filters.limit);
  
  const data = await apiCall(`/browse/search?${params}`);
  return data;
}

// Usage examples:
// 1. Basic search
const allProviders = await searchProviders();

// 2. Location-based
const bangkokProviders = await searchProviders({ location: 'Bangkok' });

// 3. High-rated providers
const topRated = await searchProviders({ rating: 4.5, sort: 'rating' });

// 4. Advanced filters with languages
const results = await searchProviders({
  location: 'Bangkok',
  rating: 4,
  tier: 3,
  category: 1,
  service_type: 'Both',
  languages: 'th,en,zh', // Filter by spoken languages
  sort: 'reviews',
  page: 1,
  limit: 20
});
```

### Response Structure
```typescript
interface SearchResponse {
  providers: Array<{
    user_id: number;
    username: string;
    profile_picture_url: string;
    bio: string;
    provider_level_id: number;
    provider_level_name: string;
    rating_avg: number;
    review_count: number;
    service_type: string;
    languages: string[]; // Array of language codes: ["th", "en", "zh"]
    location: string;
    min_price: number;
  }>;
  pagination: {
    total: number;
    page: number;
    limit: number;
    total_pages: number;
  };
  filters_applied: Record<string, string>;
}
```

### React Search Component Example

```jsx
import { useState, useEffect } from 'react';
import { apiCall } from '@/lib/api';

function ProviderSearch() {
  const [providers, setProviders] = useState([]);
  const [filters, setFilters] = useState({
    location: '',
    rating: '',
    tier: '',
    category: '',
    service_type: '',
    languages: '', // e.g., "th,en" or "th,en,zh"
    sort: 'rating',
    page: 1,
    limit: 20
  });imit: 20
  });
  const [pagination, setPagination] = useState(null);

  useEffect(() => {
    const fetchProviders = async () => {
      setLoading(true);
      try {
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
          placeholder="Location"
          value={filters.location}
          onChange={(e) => setFilters({...filters, location: e.target.value, page: 1})}
        />
        
        <select
          value={filters.rating}
          onChange={(e) => setFilters({...filters, rating: e.target.value, page: 1})}
        >
          <option value="">All Ratings</option>
          <option value="4">4+ stars</option>
          <option value="4.5">4.5+ stars</option>
        </select>
        <select
          value={filters.service_type}
          onChange={(e) => setFilters({...filters, service_type: e.target.value, page: 1})}
        >
          <option value="">All Types</option>
          <option value="Incall">Incall</option>
          <option value="Outcall">Outcall</option>
          <option value="Both">Both</option>
        </select>
        
        {/* Language Filter */}
        <select
          multiple
          value={filters.languages.split(',').filter(Boolean)}
          onChange={(e) => {
            const selected = Array.from(e.target.selectedOptions, option => option.value);
            setFilters({...filters, languages: selected.join(','), page: 1});
          }}
        >
          <option value="th">Thai (ไทย)</option>
          <option value="en">English</option>
          <option value="zh">Chinese (中文)</option>
          <option value="ja">Japanese (日本語)</option>
          <option value="ko">Korean (한국어)</option>
        </select>
        
        <select
          value={filters.sort}
          onChange={(e) => setFilters({...filters, sort: e.target.value, page: 1})}
        >
          <option value="rating">Best Rating</option>
          <option value="reviews">Most Reviews</option>
          <option value="price">Lowest Price</option>
        </select> value="price">Lowest Price</option>
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
function ProviderCard({ provider }) {
  // Language name mapping
  const languageNames = {
    th: 'ไทย',
    en: 'English',
    zh: '中文',
    ja: '日本語',
    ko: '한국어'
  };

  return (
    <div className="provider-card">
      <img 
        src={provider.profile_picture_url || '/default-avatar.png'} 
        alt={provider.username} 
      />
      <h3>{provider.username}</h3>
      <p>{provider.bio}</p>
      
      {/* Languages spoken */}
      {provider.languages && provider.languages.length > 0 && (
        <div className="languages">
          🗣️ {provider.languages.map(lang => languageNames[lang] || lang).join(', ')}
        </div>
      )}
      
      <div className="rating">
        ⭐ {provider.rating_avg.toFixed(1)} ({provider.review_count} reviews)
      </div>
      <div className="price">From ฿{provider.min_price}</div>
      <span className="tier">{provider.provider_level_name}</span>
    </div>
  );
}unction ProviderCard({ provider }) {
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
      <div className="price">From ฿{provider.min_price}</div>
      <span className="tier">{provider.provider_level_name}</span>
    </div>
  );
}
```

---

## 👤 Profile Management

### Public vs Protected Profile

#### Public Profile (No Login Required)
```javascript
// Limited information (no age, height, service_type)
const provider = await apiCall('/provider/456/public');

// Returns:
// {
//   user_id, username, bio, location, province, district,
//   tier_name, average_rating, review_count, profile_picture_url,
//   is_available
// }
// ❌ Missing: age, height, weight, service_type, working_hours
```

#### Full Profile (Login Required)
```javascript
// Complete information
const provider = await apiCall('/provider/456');

// Returns ALL fields including:
// age, height, weight, service_type, working_hours, languages, etc.
```

### Update Profile
```javascript
await apiCall('/profile/me', {
  method: 'PUT',
  body: JSON.stringify({
    bio: 'Professional service provider',
    province: 'กรุงเทพมหานคร',
    district: 'บางรัก',
    sub_district: 'สีลม',
    postal_code: '10500',
    address_line1: '123 ถนนสีลม',
    latitude: 13.7278,
    longitude: 100.5318
  })
});
```

### Provider Photos Management

#### Get Photos (Public)
```javascript
const photos = await apiCall('/provider/456/photos');
// Returns array sorted by sort_order ASC
// [
//   {
//     photo_id: 1,
//     photo_url: 'photos/user456_photo1.jpg',
//     caption: 'Professional workspace',
//     sort_order: 1,
//     uploaded_at: '2025-11-10T10:00:00Z'
//   },
//   ...
// ]
```

#### Photo Gallery Display
```jsx
function ProviderPhotoGallery({ userId }) {
  const [photos, setPhotos] = useState([]);
  const [selectedIndex, setSelectedIndex] = useState(null);

  useEffect(() => {
    apiCall(`/provider/${userId}/photos`)
      .then(data => setPhotos(data))
      .catch(console.error);
  }, [userId]);

  return (
    <>
      <div className="photo-gallery">
        {photos.map((photo, index) => (
          <img
            key={photo.photo_id}
            src={`https://storage.googleapis.com/sex-worker-bucket/${photo.photo_url}`}
            alt={photo.caption || 'Provider photo'}
            onClick={() => setSelectedIndex(index)}
            className="cursor-pointer hover:opacity-80"
          />
        ))}
      </div>

      {/* Full-screen modal */}
      {selectedIndex !== null && (
        <PhotoModal
          photos={photos}
          currentIndex={selectedIndex}
          onClose={() => setSelectedIndex(null)}
          onPrev={() => setSelectedIndex(i => Math.max(0, i - 1))}
          onNext={() => setSelectedIndex(i => Math.min(photos.length - 1, i + 1))}
        />
      )}
    </>
  );
}
```

---

## 💬 Messaging System

⚠️ **Important Policy:** Users can only send automated/templated messages for booking-related communication. Direct contact exchange (phone, Line ID, email, social media) is **strictly prohibited** and actively monitored.

### Send Message
```javascript
await apiCall('/messages', {
  method: 'POST',
  body: JSON.stringify({
    receiver_id: 456,
    content: 'Hello! I am interested in your Thai massage service.'
    // ❌ Cannot include: phone numbers, email, Line ID, social media handles
  })
});
```

### Get Conversations
```javascript
const { conversations } = await apiCall('/conversations');
// [
//   {
//     conversation_id: 123,
//     other_user_id: 456,
//     other_username: 'massage_pro',
//     last_message: 'Hello! I am interested...',
//     last_message_at: '2025-11-14T10:30:00Z',
//     unread_count: 2
//   }
// ]
```

### Get Messages
```javascript
const { messages } = await apiCall('/conversations/123/messages?limit=50&offset=0');
```

### Mark as Read
```javascript
await apiCall('/messages/read', {
  method: 'PATCH',
  body: JSON.stringify({ message_ids: [1, 2, 3] })
});
```

---

## 📦 Booking & Payment System

### Booking Flow

#### 1. Create Booking with Payment
```javascript
const booking = await apiCall('/bookings/create-with-payment', {
  method: 'POST',
  body: JSON.stringify({
    provider_id: 456,
    package_id: 1,
    booking_date: '2025-12-10T14:00:00Z',
    notes: 'First time booking',
    success_url: `${window.location.origin}/booking/success?session_id={CHECKOUT_SESSION_ID}`,
    cancel_url: `${window.location.origin}/booking/cancel`
  })
});

// Response:
// {
//   checkout_url: 'https://checkout.stripe.com/pay/cs_test_...',
//   session_id: 'cs_test_...',
//   booking_id: 789,
//   total_amount: 1000.00
// }

// Redirect to Stripe Checkout
window.location.href = booking.checkout_url;
```

#### 2. Payment Success Page
```jsx
// Route: /booking/success?session_id=cs_test_...
function BookingSuccessPage() {
  const searchParams = new URLSearchParams(window.location.search);
  const sessionId = searchParams.get('session_id');
  
  useEffect(() => {
    // Booking status automatically updated by webhook
    // You can refresh booking list here
  }, []);
  
  return (
    <div>
      <h1>Payment Successful! 🎉</h1>
      <p>Your booking has been confirmed.</p>
      <p>Session ID: {sessionId}</p>
      <Link to="/bookings/my">View My Bookings</Link>
    </div>
  );
}
```

#### 3. Get My Bookings
```javascript
const { bookings } = await apiCall('/bookings/my?status=all');
// status: 'all', 'pending', 'paid', 'confirmed', 'completed', 'cancelled'
```

#### 4. Provider: Get Incoming Bookings
```javascript
const { bookings } = await apiCall('/bookings/provider');
```

#### 5. Update Booking Status
```javascript
// Provider confirms booking
await apiCall('/bookings/789/status', {
  method: 'PATCH',
  body: JSON.stringify({ status: 'confirmed' })
});

// Provider completes booking (releases payment)
await apiCall('/bookings/789/status', {
  method: 'PATCH',
  body: JSON.stringify({ status: 'completed' })
});
```

### Payment Flow Explanation

**Client pays ฿1,000 (full price) → Stripe Checkout → Webhook processes:**
1. Stripe deducts 2.75% (฿27.50) - payment processing
2. Platform retains 10% (฿100) - commission  
3. Provider receives 87.25% (฿872.50) in `pending_balance`
4. After 7 days → moves to `available_balance` (withdrawable)
5. Provider can withdraw after booking status = `completed`

**Fee Display:**
- ❌ **Clients:** Pay full price, NO fee breakdown shown
- ✅ **Providers:** See detailed breakdown (12.75% deduction)

---

## 💰 Financial System (Provider Wallet)

### Get Wallet Balance
```javascript
const wallet = await apiCall('/wallet/balance');
// {
//   pending_balance: 2500.00,      // 7-day hold
//   available_balance: 8725.00,    // Ready to withdraw
//   total_earned: 15437.50,        // Lifetime earnings (87.25%)
//   total_withdrawn: 4212.50
// }
```

### Transaction History
```javascript
const { transactions } = await apiCall('/wallet/transactions?limit=20&offset=0&type=earning');
// [
//   {
//     transaction_id: 456,
//     transaction_type: 'earning',
//     amount: 872.50,
//     fee_breakdown: {
//       original_amount: 1000.00,
//       stripe_fee: 27.50,
//       platform_commission: 100.00,
//       total_fee_percentage: 12.75,
//       net_amount: 872.50
//     },
//     status: 'completed',
//     created_at: '2025-11-14T15:00:00Z'
//   }
// ]
```

### Request Withdrawal
```javascript
await apiCall('/wallet/withdraw', {
  method: 'POST',
  body: JSON.stringify({
    amount: 5000.00,
    bank_name: 'Kasikorn Bank',
    bank_account_number: '1234567890',
    account_holder_name: 'Sarah Johnson'
  })
});

// Provider receives:
// 1. Admin reviews and approves
// 2. GOD transfers via platform bank account
// 3. Masked transfer slip sent via WebSocket + Email
```

### Withdrawal History
```javascript
const { withdrawals } = await apiCall('/wallet/withdrawals?status=completed');
```

---

## 🎨 Service Categories

### Get All Categories
```javascript
const { categories } = await apiCall('/service-categories?include_adult=false');
// [
//   {
//     category_id: 1,
//     name: 'Massage',
//     name_thai: 'นวดแผนไทย',
//     icon: '💆',
//     description: 'Traditional massage services',
//     is_adult: false
//   },
//   ...
// ]
```

### Browse by Category
```javascript
const data = await apiCall('/categories/1/providers?page=1&limit=20');
// {
//   category_id: 1,
//   providers: [...],
//   pagination: { total, page, limit, total_pages }
// }
```

### Provider: Update Categories
```javascript
await apiCall('/provider/me/categories', {
  method: 'PUT',
  body: JSON.stringify({
    category_ids: [1, 2, 5] // Max 5 categories
  })
});
```

---

## 📡 Complete API Reference

### Public Endpoints (No Token)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/ping` | Health check |
| POST | `/register` | User registration |
| POST | `/register/provider` | Provider registration |
| POST | `/login` | Email/password login |
| POST | `/auth/google` | Google OAuth login |
| POST | `/auth/send-verification` | Send OTP email |
| GET | `/service-categories` | Get all categories |
| GET | `/provider/:id/public` | Public profile (limited) |
| GET | `/provider/:id/photos` | Provider photos |
| GET | `/packages/:providerId` | Service packages |
| GET | `/reviews/:providerId` | Provider reviews |
| GET | `/reviews/stats/:providerId` | Review statistics |
| GET | `/categories/:id/providers` | Browse by category |

### Protected Endpoints (Token Required)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/users/me` | Current user profile |
| GET | `/profile/me` | Current user full profile |
| PUT | `/profile/me` | Update profile |
| GET | `/provider/:id` | Full provider profile |
| **GET** | **`/browse/search`** | **⭐ Advanced multi-filter search** |
| GET | `/favorites` | My favorites |
| POST | `/favorites` | Add favorite |
| DELETE | `/favorites/:id` | Remove favorite |
| GET | `/favorites/check/:id` | Check if favorited |
| GET | `/bookings/my` | My bookings |
| POST | `/bookings/create-with-payment` | Create booking + Stripe |
| PATCH | `/bookings/:id/status` | Update booking status |
| GET | `/conversations` | My conversations |
| GET | `/conversations/:id/messages` | Conversation messages |
| POST | `/messages` | Send message |
| PATCH | `/messages/read` | Mark as read |
| GET | `/wallet/balance` | Wallet balance |
| GET | `/wallet/transactions` | Transaction history |
| POST | `/wallet/withdraw` | Request withdrawal |
| GET | `/wallet/withdrawals` | Withdrawal history |

### Admin Endpoints (Admin Only)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/admin/providers/pending` | Pending providers |
| PATCH | `/admin/verify-document/:id` | Verify document |
| PATCH | `/admin/approve-provider/:id` | Approve provider |
| POST | `/admin/recalculate-provider-tiers` | Recalculate tiers |
| GET | `/admin/withdrawals` | Pending withdrawals |
| POST | `/admin/withdrawals/:id/process` | Process withdrawal |
| GET | `/admin/financial/summary` | Financial dashboard |

---

## 🧩 React Components Examples

### Authentication Hook
```jsx
// hooks/useAuth.js
import { useState, useEffect } from 'react';
import { apiCall } from '@/lib/api';

export function useAuth() {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const token = localStorage.getItem('token');
    if (token) {
      apiCall('/users/me')
        .then(setUser)
        .catch(() => {
          localStorage.removeItem('token');
          setUser(null);
        })
        .finally(() => setLoading(false));
    } else {
      setLoading(false);
    }
  }, []);

  const login = async (email, password) => {
    const data = await apiCall('/login', {
      method: 'POST',
      body: JSON.stringify({ email, password })
    });
    localStorage.setItem('token', data.token);
    setUser(data.user);
  };

  const logout = () => {
    localStorage.removeItem('token');
    setUser(null);
  };

  return { user, loading, login, logout };
}
```

### Provider Profile Page
```jsx
// pages/ProviderProfile.jsx
import { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { useAuth } from '@/hooks/useAuth';
import { apiCall } from '@/lib/api';

function ProviderProfile() {
  const { userId } = useParams();
  const { user: currentUser } = useAuth();
  const [provider, setProvider] = useState(null);
  const [photos, setPhotos] = useState([]);
  const [packages, setPackages] = useState([]);
  const [reviews, setReviews] = useState([]);
  const [isFavorite, setIsFavorite] = useState(false);

  useEffect(() => {
    const fetchData = async () => {
      try {
        // Get profile (public or full based on auth)
        const profileEndpoint = currentUser
          ? `/provider/${userId}`
          : `/provider/${userId}/public`;
        const profileData = await apiCall(profileEndpoint);
        setProvider(profileData);

        // Get photos (public)
        const photosData = await apiCall(`/provider/${userId}/photos`);
        setPhotos(photosData);

        // Get packages (public)
        const packagesData = await apiCall(`/packages/${userId}`);
        setPackages(packagesData);

        // Get reviews (public)
        const reviewsData = await apiCall(`/reviews/${userId}`);
        setReviews(reviewsData.reviews);

        // Check if favorited (requires auth)
        if (currentUser) {
          const favData = await apiCall(`/favorites/check/${userId}`);
          setIsFavorite(favData.is_favorite);
        }
      } catch (error) {
        console.error('Failed to load provider:', error);
      }
    };

    fetchData();
  }, [userId, currentUser]);

  const toggleFavorite = async () => {
    try {
      if (isFavorite) {
        await apiCall(`/favorites/${userId}`, { method: 'DELETE' });
        setIsFavorite(false);
      } else {
        await apiCall('/favorites', {
          method: 'POST',
          body: JSON.stringify({ provider_id: parseInt(userId) })
        });
        setIsFavorite(true);
      }
    } catch (error) {
      alert('Please login to add favorites');
    }
  };

  if (!provider) return <div>Loading...</div>;

  return (
    <div className="provider-profile">
      {/* Header */}
      <div className="profile-header">
        <img
          src={provider.profile_picture_url || '/default-avatar.png'}
          alt={provider.username}
          className="profile-avatar"
        />
        <div className="profile-info">
          <h1>{provider.username}</h1>
          <span className="tier-badge">{provider.tier_name}</span>
          <div className="rating">
            ⭐ {provider.average_rating.toFixed(1)} ({provider.review_count} reviews)
          </div>
          <button onClick={toggleFavorite} className="favorite-btn">
            {isFavorite ? '❤️ Favorited' : '🤍 Add to Favorites'}
          </button>
        </div>
      </div>

      {/* Bio */}
      <div className="bio">
        <h2>About</h2>
        <p>{provider.bio}</p>
      </div>

      {/* Sensitive info (only if logged in) */}
      {currentUser && provider.age && (
        <div className="detailed-info">
          <h3>Details (Members Only)</h3>
          <p>Age: {provider.age}</p>
          <p>Height: {provider.height} cm</p>
          <p>Service Type: {provider.service_type}</p>
          <p>Working Hours: {provider.working_hours}</p>
        </div>
      )}

      {/* Photo Gallery */}
      <div className="photo-gallery">
        <h2>Photos ({photos.length})</h2>
        <div className="photos-grid">
          {photos.map(photo => (
            <img
              key={photo.photo_id}
              src={`https://storage.googleapis.com/sex-worker-bucket/${photo.photo_url}`}
              alt={photo.caption}
            />
          ))}
        </div>
      </div>

      {/* Packages */}
      <div className="packages">
        <h2>Service Packages</h2>
        {packages.map(pkg => (
          <div key={pkg.package_id} className="package-card">
            <h3>{pkg.name}</h3>
            <p>{pkg.description}</p>
            <p className="price">฿{pkg.price}</p>
            <p className="duration">{pkg.duration_hours} hours</p>
            <button>Book Now</button>
          </div>
        ))}
      </div>

      {/* Reviews */}
      <div className="reviews">
        <h2>Reviews</h2>
        {reviews.map(review => (
          <div key={review.review_id} className="review-card">
            <div className="review-header">
              <span className="reviewer">{review.client_username}</span>
              <span className="rating">⭐ {review.rating}</span>
            </div>
            <p>{review.comment}</p>
            <span className="date">{new Date(review.created_at).toLocaleDateString()}</span>
          </div>
        ))}
      </div>
    </div>
  );
}

export default ProviderProfile;
```

---

## 🎯 Vue/Angular Examples

### Vue Composable
```javascript
// composables/useProvider.js
import { ref, onMounted } from 'vue';

export function useProvider(userId, isAuthenticated) {
  const provider = ref(null);
  const photos = ref([]);
  const loading = ref(true);
  const error = ref(null);

  onMounted(async () => {
    try {
      loading.value = true;
      
      const endpoint = isAuthenticated
        ? `/provider/${userId}`
        : `/provider/${userId}/public`;
      
      const response = await fetch(`http://localhost:8080${endpoint}`, {
        headers: isAuthenticated ? {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        } : {}
      });

      if (!response.ok) throw new Error('Provider not found');

      provider.value = await response.json();

      const photosRes = await fetch(`http://localhost:8080/provider/${userId}/photos`);
      photos.value = await photosRes.json();
    } catch (err) {
      error.value = err.message;
    } finally {
      loading.value = false;
    }
  });

  return { provider, photos, loading, error };
}
```

### Angular Service
```typescript
// services/provider.service.ts
import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Observable } from 'rxjs';

@Injectable({ providedIn: 'root' })
export class ProviderService {
  private apiBase = 'http://localhost:8080';

  constructor(private http: HttpClient) {}

  private getHeaders(): HttpHeaders {
    const token = localStorage.getItem('token');
    return new HttpHeaders({
      'Content-Type': 'application/json',
      ...(token && { 'Authorization': `Bearer ${token}` })
    });
  }

  getProvider(userId: number, isAuthenticated: boolean): Observable<any> {
    const endpoint = isAuthenticated
      ? `/provider/${userId}`
      : `/provider/${userId}/public`;
    
    return this.http.get(`${this.apiBase}${endpoint}`, {
      headers: this.getHeaders()
    });
  }

  searchProviders(filters: any): Observable<any> {
    const params = new URLSearchParams(filters).toString();
    return this.http.get(`${this.apiBase}/browse/search?${params}`, {
      headers: this.getHeaders()
    });
  }
}
```

---

## 📱 Real-time WebSocket

### Connection Setup
```javascript
// lib/websocket.js
class WebSocketManager {
  constructor() {
    this.ws = null;
    this.handlers = new Map();
  }

  connect(token) {
    this.ws = new WebSocket('ws://localhost:8080/ws');

    this.ws.onopen = () => {
      console.log('WebSocket connected');
      
      // Authenticate
      this.ws.send(JSON.stringify({
        type: 'auth',
        payload: { token }
      }));
    };

    this.ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      const handler = this.handlers.get(data.type);
      if (handler) handler(data.payload);
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    this.ws.onclose = () => {
      console.log('WebSocket closed - reconnecting in 5s');
      setTimeout(() => this.connect(token), 5000);
    };
  }

  on(eventType, handler) {
    this.handlers.set(eventType, handler);
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }
}

export const wsManager = new WebSocketManager();
```

### Usage in React
```jsx
import { useEffect } from 'react';
import { wsManager } from '@/lib/websocket';

function App() {
  useEffect(() => {
    const token = localStorage.getItem('token');
    if (token) {
      wsManager.connect(token);

      // Handle new messages
      wsManager.on('new_message', (payload) => {
        showNotification(`New message from ${payload.sender}`);
        // Update UI
      });

      // Handle booking updates
      wsManager.on('booking_update', (payload) => {
        showNotification(`Booking #${payload.booking_id} ${payload.status}`);
        // Refresh bookings
      });

      // Handle notifications
      wsManager.on('notification', (payload) => {
        updateNotificationBadge(payload);
      });

      return () => wsManager.disconnect();
    }
  }, []);

  return <div>{/* App content */}</div>;
}
```

---

## 🧪 Testing Guide

### API Connection Test
```javascript
describe('API Connection', () => {
  it('should ping backend', async () => {
    const response = await fetch('http://localhost:8080/ping');
    const data = await response.json();
    expect(data.message).toBe('pong!');
  });
});
```

### Authentication Test
```javascript
describe('Authentication', () => {
  it('should login successfully', async () => {
    const data = await apiCall('/login', {
      method: 'POST',
      body: JSON.stringify({
        email: 'god@skillmatch.com',
        password: 'godpass123'
      })
    });
    
    expect(data).toHaveProperty('token');
    expect(data).toHaveProperty('message', 'Login successful');
  });

  it('should return user profile', async () => {
    localStorage.setItem('token', 'test-token');
      "service_type": "Both",
      "languages": ["th", "en", "zh"],
      "location": "Bangkok, Sukhumvit",
      "min_price": 1500.00perty('profile_picture_url'); // NEW field name
  });
});
```

### Search Test
```javascript
describe('Provider Search', () => {
  it('should search providers', async () => {
    const data = await apiCall('/browse/search?location=Bangkok&rating=4');
    
    expect(data).toHaveProperty('providers');
    expect(data).toHaveProperty('pagination');
    expect(Array.isArray(data.providers)).toBe(true);
    expect(data.pagination).toHaveProperty('total_pages');
  });
});
```

---

## 🔧 Troubleshooting

### Issue 1: Profile Pictures Not Showing
**Problem:** Broken images after migration

**Solution:**
```javascript
// Check field name
console.log(user); // Should have profile_picture_url, not profile_image_url

// Add fallback
const profilePic = user.profile_picture_url || '/default-avatar.png';
```

### Issue 2: CORS Errors
**Problem:** `Access-Control-Allow-Origin` error

**Solution:**
- Backend allows: `localhost:3000`, `localhost:5173`, `localhost:8080`
- Check your dev server port matches allowed origins
- For production, update backend CORS config

### Issue 3: 401 Unauthorized
**Problem:** Token expired or invalid

**Solution:**
```javascript
// Check token expiry (JWT tokens valid for 7 days)
if (response.status === 401) {
  localStorage.removeItem('token');
  window.location.href = '/login';
}
```

### Issue 4: Browse Search Returns Empty
**Problem:** No results from `/browse/search`

**Solution:**
```javascript
// Check filters are valid
const validFilters = {
  rating: 4,      // 1-5 only
  tier: 3,        // 1-4 only (1=General, 2=Silver, 3=Diamond, 4=Premium)
  service_type: 'Incall' // Must be: "Incall", "Outcall", or "Both"
};

// Only "verified" or "approved" providers appear in search
```

### Issue 5: WebSocket Connection Failed
**Problem:** WebSocket disconnects immediately

**Solution:**
```javascript
// Check authentication
ws.send(JSON.stringify({
  type: 'auth',
  payload: { token: localStorage.getItem('token') }
}));

// Add reconnection logic
ws.onclose = () => {
  console.log('Reconnecting in 5s...');
  setTimeout(() => connect(), 5000);
};
```

---

## 📊 API Response Examples

### Browse Search Response
```json
{
  "providers": [
    {
      "user_id": 456,
      "username": "massage_pro",
      "profile_picture_url": "https://...",
      "bio": "Professional massage therapist...",
      "provider_level_id": 3,
      "provider_level_name": "Diamond",
      "rating_avg": 4.8,
      "review_count": 120,
      "service_type": "Both",
      "location": "Bangkok, Sukhumvit",
      "min_price": 1500.00
    }
  ],
  "pagination": {
    "total": 50,
    "page": 1,
    "limit": 20,
    "total_pages": 3
  },
  "filters_applied": {
    "location": "Bangkok",
    "rating": "4",
    "sort": "rating"
  }
}
```

### User Profile Response
```json
{
  "user_id": 123,
  "username": "john_doe",
  "email": "john@example.com",
  "profile_picture_url": "https://...",
  "tier_id": 1,
  "tier_name": "General",
  "verification_status": "verified",
  "created_at": "2025-01-15T10:00:00Z"
}
```

---

## 🎉 Quick Reference Card

### Essential Endpoints
```javascript
// Auth
POST /login
POST /auth/google
GET /users/me

// Browse
GET /browse/search?location=Bangkok&rating=4&sort=rating
GET /provider/:id/public
GET /provider/:id/photos

// Booking
POST /bookings/create-with-payment
GET /bookings/my

// Favorites
POST /favorites
GET /favorites
DELETE /favorites/:id

// Messaging
POST /messages
GET /conversations
```

### Common Status Codes
- **200** - Success
- **401** - Unauthorized (token expired/invalid)
- **403** - Forbidden (insufficient permissions)
- **404** - Not Found
- **409** - Conflict (duplicate data)
- **500** - Server Error

---

## 📞 Support Resources

### Quick Diagnostics
```bash
# Test backend
curl http://localhost:8080/ping

# Test auth (use GOD token from test accounts)
curl -H "Authorization: Bearer eyJhbGc..." http://localhost:8080/users/me

# Test search
curl "http://localhost:8080/browse/search?location=Bangkok"
```

### Test Accounts
```
GOD Admin:
Email: god@skillmatch.com
Password: godpass123
Token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

Verified Provider:
Email: provider@test.com
Password: test123
User ID: 456
```

---

**Documentation Version:** 2.0 (December 2, 2025)  
**Backend Server:** http://localhost:8080  
**Total Endpoints:** 119  
**Database Status:** Optimized (50-80% faster)  
**Ready for Production:** ✅

**Happy Coding! 🚀**
