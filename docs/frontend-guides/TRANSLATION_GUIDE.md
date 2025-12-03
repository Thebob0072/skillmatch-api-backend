# 🌐 SkillMatch Frontend - Translation Guide (Thai/English)

> **คู่มือการแปลภาษาสำหรับ Frontend**  
> **Backend API:** http://localhost:8080  
> **Languages Supported:** Thai (ไทย), English

---

## 📋 Table of Contents

1. [Fields with Thai Data from Backend](#1-fields-with-thai-data-from-backend)
2. [Fields Frontend Must Translate](#2-fields-frontend-must-translate)
3. [Translation Implementation](#3-translation-implementation)
4. [Complete Translation Dictionary](#4-complete-translation-dictionary)

---

## 1. Fields with Thai Data from Backend

### ✅ Category Names (API: `/service-categories`)
```json
{
  "category_id": 1,
  "name": "Massage",
  "name_thai": "นวดแผนไทย",  // ⭐ Backend provides Thai
  "icon": "💆"
}
```

**Frontend Usage:**
```javascript
const displayName = lang === 'th' ? category.name_thai : category.name;
```

### ✅ Location Fields
```json
{
  "province": "กรุงเทพมหานคร",      // ⭐ Thai from backend
  "district": "บางรัก",              // ⭐ Thai from backend
  "sub_district": "สีลม",            // ⭐ Thai from backend
  "address_line1": "123 ถนนสีลม"    // ⭐ User input (Thai)
}
```

**No translation needed** - display as-is.

---

## 2. Fields Frontend Must Translate

### ❌ Provider Tier Names
**Backend sends:** "General", "Silver", "Diamond", "Premium"  
**Frontend must translate to Thai**

### ❌ Service Types
**Backend sends:** "Incall", "Outcall", "Both"  
**Frontend must translate to Thai**

### ❌ Booking Status
**Backend sends:** "pending", "paid", "confirmed", "completed", "cancelled"  
**Frontend must translate to Thai**

### ❌ Transaction Types
**Backend sends:** "earning", "withdrawal", "refund"  
**Frontend must translate to Thai**

### ❌ Gender
**Backend sends:** `gender_id` (1, 2, 3, 4)  
**Frontend must map to text**

### ❌ UI Labels
**All buttons, messages, errors** - Frontend must provide translations

---

## 3. Translation Implementation

### React Implementation

#### Step 1: Create Translation Files

**translations/th.json:**
```json
{
  "common": {
    "save": "บันทึก",
    "cancel": "ยกเลิก",
    "confirm": "ยืนยัน",
    "delete": "ลบ",
    "edit": "แก้ไข",
    "search": "ค้นหา",
    "loading": "กำลังโหลด...",
    "error": "เกิดข้อผิดพลาด",
    "success": "สำเร็จ"
  },
  "auth": {
    "login": "เข้าสู่ระบบ",
    "logout": "ออกจากระบบ",
    "register": "สมัครสมาชิก",
    "email": "อีเมล",
    "password": "รหัสผ่าน",
    "forgotPassword": "ลืมรหัสผ่าน?",
    "loginSuccess": "เข้าสู่ระบบสำเร็จ",
    "loginFailed": "เข้าสู่ระบบไม่สำเร็จ"
  },
  "provider": {
    "tiers": {
      "General": "ทั่วไป",
      "Silver": "เงิน",
      "Diamond": "เพชร",
      "Premium": "พรีเมียม"
    },
    "serviceTypes": {
      "Incall": "บริการที่ร้าน",
      "Outcall": "บริการนอกสถานที่",
      "Both": "ทั้งสองแบบ"
    },
    "profile": "โปรไฟล์",
    "reviews": "รีวิว",
    "packages": "แพ็คเกจ",
    "photos": "รูปภาพ",
    "bio": "เกี่ยวกับ",
    "rating": "คะแนน",
    "reviewCount": "รีวิว"
  },
  "booking": {
    "status": {
      "pending": "รอดำเนินการ",
      "paid": "ชำระแล้ว",
      "confirmed": "ยืนยันแล้ว",
      "completed": "เสร็จสิ้น",
      "cancelled": "ยกเลิก"
    },
    "createBooking": "สร้างการจอง",
    "myBookings": "การจองของฉัน",
    "bookingDate": "วันที่จอง",
    "totalPrice": "ราคารวม",
    "payNow": "ชำระเงินตอนนี้"
  },
  "financial": {
    "transactionTypes": {
      "earning": "รายได้",
      "withdrawal": "ถอนเงิน",
      "refund": "คืนเงิน"
    },
    "wallet": "กระเป๋าเงิน",
    "pendingBalance": "ยอดค้างรับ",
    "availableBalance": "ยอดพร้อมถอน",
    "withdraw": "ถอนเงิน",
    "transactions": "ธุรกรรม"
  },
  "messages": {
    "conversations": "บทสนทนา",
    "sendMessage": "ส่งข้อความ",
    "typeMessage": "พิมพ์ข้อความ...",
    "noMessages": "ไม่มีข้อความ",
    "unreadCount": "ยังไม่อ่าน"
  },
  "gender": {
    "1": "ชาย",
    "2": "หญิง",
    "3": "อื่นๆ",
    "4": "ไม่ระบุ"
  },
  "search": {
    "filters": "ตัวกรอง",
    "location": "สถานที่",
    "category": "หมวดหมู่",
    "rating": "คะแนน",
    "priceRange": "ช่วงราคา",
    "serviceType": "ประเภทบริการ",
    "sortBy": "เรียงตาม",
    "sortOptions": {
      "rating": "คะแนนสูงสุด",
      "reviews": "รีวิวมากสุด",
      "price": "ราคาต่ำสุด"
    }
  }
}
```

**translations/en.json:**
```json
{
  "common": {
    "save": "Save",
    "cancel": "Cancel",
    "confirm": "Confirm",
    "delete": "Delete",
    "edit": "Edit",
    "search": "Search",
    "loading": "Loading...",
    "error": "Error",
    "success": "Success"
  },
  "auth": {
    "login": "Login",
    "logout": "Logout",
    "register": "Register",
    "email": "Email",
    "password": "Password",
    "forgotPassword": "Forgot Password?",
    "loginSuccess": "Login successful",
    "loginFailed": "Login failed"
  },
  "provider": {
    "tiers": {
      "General": "General",
      "Silver": "Silver",
      "Diamond": "Diamond",
      "Premium": "Premium"
    },
    "serviceTypes": {
      "Incall": "Incall",
      "Outcall": "Outcall",
      "Both": "Both"
    },
    "profile": "Profile",
    "reviews": "Reviews",
    "packages": "Packages",
    "photos": "Photos",
    "bio": "About",
    "rating": "Rating",
    "reviewCount": "reviews"
  },
  "booking": {
    "status": {
      "pending": "Pending",
      "paid": "Paid",
      "confirmed": "Confirmed",
      "completed": "Completed",
      "cancelled": "Cancelled"
    },
    "createBooking": "Create Booking",
    "myBookings": "My Bookings",
    "bookingDate": "Booking Date",
    "totalPrice": "Total Price",
    "payNow": "Pay Now"
  },
  "financial": {
    "transactionTypes": {
      "earning": "Earning",
      "withdrawal": "Withdrawal",
      "refund": "Refund"
    },
    "wallet": "Wallet",
    "pendingBalance": "Pending Balance",
    "availableBalance": "Available Balance",
    "withdraw": "Withdraw",
    "transactions": "Transactions"
  },
  "messages": {
    "conversations": "Conversations",
    "sendMessage": "Send Message",
    "typeMessage": "Type a message...",
    "noMessages": "No messages",
    "unreadCount": "unread"
  },
  "gender": {
    "1": "Male",
    "2": "Female",
    "3": "Other",
    "4": "Prefer not to say"
  },
  "search": {
    "filters": "Filters",
    "location": "Location",
    "category": "Category",
    "rating": "Rating",
    "priceRange": "Price Range",
    "serviceType": "Service Type",
    "sortBy": "Sort By",
    "sortOptions": {
      "rating": "Best Rating",
      "reviews": "Most Reviews",
      "price": "Lowest Price"
    }
  }
}
```

#### Step 2: Create Translation Hook

**hooks/useTranslation.js:**
```javascript
import { createContext, useContext, useState } from 'react';
import th from '@/translations/th.json';
import en from '@/translations/en.json';

const TranslationContext = createContext();

export function TranslationProvider({ children }) {
  const [language, setLanguage] = useState(
    localStorage.getItem('language') || 'th'
  );

  const translations = {
    th,
    en
  };

  const t = (key) => {
    const keys = key.split('.');
    let value = translations[language];
    
    for (const k of keys) {
      value = value?.[k];
      if (!value) return key; // Fallback to key if not found
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

#### Step 3: Usage in Components

```jsx
import { useTranslation } from '@/hooks/useTranslation';

function ProviderCard({ provider }) {
  const { t, language } = useTranslation();

  return (
    <div className="provider-card">
      <img src={provider.profile_picture_url} alt={provider.username} />
      <h3>{provider.username}</h3>
      
      {/* Category name (from backend) */}
      <p className="category">
        {language === 'th' ? provider.category_name_thai : provider.category_name}
      </p>
      
      {/* Provider tier (translate on frontend) */}
      <span className="tier">
        {t(`provider.tiers.${provider.provider_level_name}`)}
      </span>
      
      {/* Service type (translate on frontend) */}
      <p className="service-type">
        {t(`provider.serviceTypes.${provider.service_type}`)}
      </p>
      
      {/* Rating label */}
      <div className="rating">
        ⭐ {provider.rating_avg.toFixed(1)} 
        ({provider.review_count} {t('provider.reviewCount')})
      </div>
      
      {/* Location (Thai from backend) */}
      <p className="location">{provider.location}</p>
      
      <button>{t('provider.viewProfile')}</button>
    </div>
  );
}

function BookingList({ bookings }) {
  const { t } = useTranslation();

  return (
    <div className="bookings">
      <h2>{t('booking.myBookings')}</h2>
      {bookings.map(booking => (
        <div key={booking.booking_id} className="booking-card">
          <p>{t('booking.bookingDate')}: {booking.booking_date}</p>
          <p>{t('booking.totalPrice')}: ฿{booking.total_price}</p>
          
          {/* Status translation */}
          <span className={`status ${booking.status}`}>
            {t(`booking.status.${booking.status}`)}
          </span>
        </div>
      ))}
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

---

## 4. Complete Translation Dictionary

### API Response Field Mapping

| Backend Field | Type | Translation Needed? | Example |
|--------------|------|---------------------|---------|
| `name_thai` | string | ❌ No (display as-is) | "นวดแผนไทย" |
| `provider_level_name` | string | ✅ Yes | "Premium" → "พรีเมียม" |
| `service_type` | string | ✅ Yes | "Incall" → "บริการที่ร้าน" |
| `status` (booking) | string | ✅ Yes | "confirmed" → "ยืนยันแล้ว" |
| `transaction_type` | string | ✅ Yes | "earning" → "รายได้" |
| `gender_id` | number | ✅ Yes | 1 → "ชาย" / "Male" |
| `province` | string | ❌ No | "กรุงเทพมหานคร" |
| `district` | string | ❌ No | "บางรัก" |

### Complete Enum Translations

```javascript
// helpers/translations.js
export const TRANSLATIONS = {
  th: {
    // Provider Tiers
    tiers: {
      'General': 'ทั่วไป',
      'Silver': 'เงิน',
      'Diamond': 'เพชร',
      'Premium': 'พรีเมียม'
    },
    
    // Service Types
    serviceTypes: {
      'Incall': 'บริการที่ร้าน',
      'Outcall': 'บริการนอกสถานที่',
      'Both': 'ทั้งสองแบบ'
    },
    
    // Booking Status
    bookingStatus: {
      'pending': 'รอดำเนินการ',
      'paid': 'ชำระแล้ว',
      'confirmed': 'ยืนยันแล้ว',
      'completed': 'เสร็จสิ้น',
      'cancelled': 'ยกเลิก'
    },
    
    // Transaction Types
    transactionTypes: {
      'earning': 'รายได้',
      'withdrawal': 'ถอนเงิน',
      'refund': 'คืนเงิน'
    },
    
    // Verification Status
    verificationStatus: {
      'unverified': 'ยังไม่ยืนยัน',
      'pending': 'รอการตรวจสอบ',
      'documents_submitted': 'ส่งเอกสารแล้ว',
      'approved': 'อนุมัติแล้ว',
      'verified': 'ยืนยันแล้ว',
      'rejected': 'ปฏิเสธ'
    },
    
    // Gender
    gender: {
      1: 'ชาย',
      2: 'หญิง',
      3: 'อื่นๆ',
      4: 'ไม่ระบุ'
    },
    
    // Subscription Tiers (Client)
    subscriptionTiers: {
      'General': 'ทั่วไป (ฟรี)',
      'Silver': 'เงิน',
      'Gold': 'ทอง',
      'Platinum': 'แพลทินัม'
    }
  },
  
  en: {
    // Keep original values or provide English
    tiers: {
      'General': 'General',
      'Silver': 'Silver',
      'Diamond': 'Diamond',
      'Premium': 'Premium'
    },
    // ... same structure
  }
};

// Helper function
export function translate(category, value, language = 'th') {
  return TRANSLATIONS[language]?.[category]?.[value] || value;
}
```

**Usage:**
```javascript
import { translate } from '@/helpers/translations';

// Provider tier
const tierName = translate('tiers', provider.provider_level_name, 'th');
// "Premium" → "พรีเมียม"

// Service type
const serviceType = translate('serviceTypes', provider.service_type, 'th');
// "Both" → "ทั้งสองแบบ"

// Booking status
const status = translate('bookingStatus', booking.status, 'th');
// "confirmed" → "ยืนยันแล้ว"
```

---

## 5. Vue/Angular Translation

### Vue 3 with Composition API

**composables/useI18n.js:**
```javascript
import { ref, computed } from 'vue';
import th from '@/translations/th.json';
import en from '@/translations/en.json';

const language = ref(localStorage.getItem('language') || 'th');

const translations = { th, en };

export function useI18n() {
  const t = (key) => {
    const keys = key.split('.');
    let value = translations[language.value];
    
    for (const k of keys) {
      value = value?.[k];
      if (!value) return key;
    }
    
    return value;
  };

  const setLanguage = (lang) => {
    language.value = lang;
    localStorage.setItem('language', lang);
  };

  return {
    t,
    language: computed(() => language.value),
    setLanguage
  };
}
```

### Angular Service

**services/translation.service.ts:**
```typescript
import { Injectable } from '@angular/core';
import { BehaviorSubject } from 'rxjs';
import th from '@/assets/translations/th.json';
import en from '@/assets/translations/en.json';

@Injectable({ providedIn: 'root' })
export class TranslationService {
  private languageSubject = new BehaviorSubject<string>(
    localStorage.getItem('language') || 'th'
  );
  
  language$ = this.languageSubject.asObservable();
  
  private translations = { th, en };

  translate(key: string): string {
    const lang = this.languageSubject.value;
    const keys = key.split('.');
    let value: any = this.translations[lang];
    
    for (const k of keys) {
      value = value?.[k];
      if (!value) return key;
    }
    
    return value;
  }

  setLanguage(lang: string): void {
    this.languageSubject.next(lang);
    localStorage.setItem('language', lang);
  }
}
```

---

## 6. Best Practices

### ✅ DO

1. **Store language preference** in localStorage
2. **Provide language switcher** in navbar/settings
3. **Use translation keys** consistently (`provider.tiers.Premium`)
4. **Fallback to English** if translation missing
5. **Display Thai location data** as-is (no translation)
6. **Test both languages** before deployment

### ❌ DON'T

1. ❌ Hardcode Thai/English text in components
2. ❌ Translate location fields (province, district) - already Thai
3. ❌ Mix translation approaches (use one system)
4. ❌ Forget to translate error messages
5. ❌ Assume all users speak Thai (provide EN option)

---

## 7. API Endpoints Language Behavior

### Endpoints with Thai Data
- ✅ `GET /service-categories` - Returns `name_thai` field
- ✅ All location fields - Thai text

### Endpoints with English Enums
- ❌ `GET /provider/:id` - `service_type` (Incall/Outcall/Both)
- ❌ `GET /bookings/my` - `status` (pending/confirmed/etc)
- ❌ `GET /wallet/transactions` - `transaction_type` (earning/withdrawal)

**Frontend must translate these enums** - Backend does not provide Thai versions.

---

## 8. Quick Reference

### Translation Checklist

```plaintext
☐ Provider tier names (General → ทั่วไป)
☐ Service types (Incall → บริการที่ร้าน)
☐ Booking status (confirmed → ยืนยันแล้ว)
☐ Transaction types (earning → รายได้)
☐ Gender (1 → ชาย)
☐ UI labels (Save → บันทึก)
☐ Error messages
☐ Success messages
☐ Button labels
☐ Form labels
☐ Validation messages
```

### No Translation Needed

```plaintext
✓ Category names (use name_thai field)
✓ Province (already Thai)
✓ District (already Thai)
✓ Sub-district (already Thai)
✓ Address lines (user input)
✓ User bio (user input)
✓ Package names (user input)
✓ Review comments (user input)
```

---

**Translation Guide Version:** 1.0 (December 2, 2025)  
**Supported Languages:** Thai (ไทย), English  
**Default Language:** Thai  

**Happy Translating! 🌏**
