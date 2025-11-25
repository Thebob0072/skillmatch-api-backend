# 🎨 Frontend Integration Guide - Service Categories

## 📋 Overview
คู่มือนี้สำหรับ Frontend Developer ในการใช้งาน Service Category API

---

## 🔗 API Endpoints Summary

### Base URL
```
http://localhost:8080
```

### Endpoints
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/service-categories` | ❌ | ดูหมวดหมู่ทั้งหมด |
| GET | `/categories/:id/providers` | ❌ | ค้นหาผู้ให้บริการตามหมวดหมู่ |
| GET | `/providers/:id/categories` | ✅ | ดูหมวดหมู่ของผู้ให้บริการ |
| PUT | `/provider/me/categories` | ✅ | อัพเดทหมวดหมู่ของตัวเอง |

---

## 📦 TypeScript Types

```typescript
// types/serviceCategory.ts

/**
 * Service Category Model
 */
export interface ServiceCategory {
  category_id: number;
  name: string;              // English name (e.g., "massage_therapy")
  name_thai: string;         // Thai name (e.g., "นวดบำบัด")
  description?: string;      // Category description
  icon?: string;            // Emoji icon (e.g., "💆")
  is_adult: boolean;        // Requires 18+ verification
  display_order: number;    // Order for sorting
  is_active: boolean;       // Category active status
}

/**
 * Provider with category info
 */
export interface ProviderWithLocation {
  user_id: number;
  username: string;
  gender_id: number;
  age?: number;
  profile_image_url?: string;
  google_profile_picture?: string;
  province?: string;
  district?: string;
  sub_district?: string;
  average_rating: number;
  review_count: number;
  min_price?: number;
}

/**
 * API Response Types
 */
export interface GetCategoriesResponse {
  categories: ServiceCategory[];
  total: number;
}

export interface GetProviderCategoriesResponse {
  provider_id: number;
  categories: ServiceCategory[];
  total: number;
}

export interface UpdateCategoriesResponse {
  message: string;
  category_ids: number[];
  total: number;
}

export interface BrowseByCategoryResponse {
  category_id: number;
  providers: ProviderWithLocation[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

export interface ApiError {
  error: string;
  details?: string;
}
```

---

## 🔧 API Service Class

```typescript
// services/categoryService.ts
import axios, { AxiosInstance } from 'axios';
import {
  ServiceCategory,
  GetCategoriesResponse,
  GetProviderCategoriesResponse,
  UpdateCategoriesResponse,
  BrowseByCategoryResponse,
} from '@/types/serviceCategory';

export class CategoryService {
  private api: AxiosInstance;

  constructor(baseURL: string = 'http://localhost:8080') {
    this.api = axios.create({
      baseURL,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Add auth token to requests
    this.api.interceptors.request.use((config) => {
      const token = localStorage.getItem('auth_token');
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
      return config;
    });
  }

  /**
   * Get all service categories
   * @param includeAdult - Include adult categories (default: true)
   */
  async getCategories(includeAdult = true): Promise<GetCategoriesResponse> {
    const response = await this.api.get<GetCategoriesResponse>(
      '/service-categories',
      {
        params: { include_adult: includeAdult },
      }
    );
    return response.data;
  }

  /**
   * Get provider's selected categories
   * @param providerId - Provider user ID
   */
  async getProviderCategories(
    providerId: number
  ): Promise<GetProviderCategoriesResponse> {
    const response = await this.api.get<GetProviderCategoriesResponse>(
      `/providers/${providerId}/categories`
    );
    return response.data;
  }

  /**
   * Update current provider's categories
   * @param categoryIds - Array of category IDs (max 5)
   */
  async updateMyCategories(
    categoryIds: number[]
  ): Promise<UpdateCategoriesResponse> {
    if (categoryIds.length > 5) {
      throw new Error('Cannot select more than 5 categories');
    }

    const response = await this.api.put<UpdateCategoriesResponse>(
      '/provider/me/categories',
      { category_ids: categoryIds }
    );
    return response.data;
  }

  /**
   * Browse providers by category
   * @param categoryId - Category ID
   * @param page - Page number (default: 1)
   * @param limit - Items per page (default: 20, max: 50)
   */
  async browseByCategory(
    categoryId: number,
    page = 1,
    limit = 20
  ): Promise<BrowseByCategoryResponse> {
    const response = await this.api.get<BrowseByCategoryResponse>(
      `/categories/${categoryId}/providers`,
      {
        params: { page, limit },
      }
    );
    return response.data;
  }
}

// Export singleton instance
export const categoryService = new CategoryService();
```

---

## 🎯 React Hooks

```typescript
// hooks/useCategories.ts
import { useState, useEffect } from 'react';
import { categoryService } from '@/services/categoryService';
import { ServiceCategory } from '@/types/serviceCategory';

/**
 * Hook to fetch all service categories
 */
export const useCategories = (includeAdult = true) => {
  const [categories, setCategories] = useState<ServiceCategory[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchCategories = async () => {
      try {
        setLoading(true);
        const data = await categoryService.getCategories(includeAdult);
        setCategories(data.categories);
        setError(null);
      } catch (err: any) {
        setError(err.response?.data?.error || 'Failed to fetch categories');
      } finally {
        setLoading(false);
      }
    };

    fetchCategories();
  }, [includeAdult]);

  return { categories, loading, error };
};

/**
 * Hook to fetch provider categories
 */
export const useProviderCategories = (providerId: number | null) => {
  const [categories, setCategories] = useState<ServiceCategory[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!providerId) {
      setCategories([]);
      setLoading(false);
      return;
    }

    const fetchProviderCategories = async () => {
      try {
        setLoading(true);
        const data = await categoryService.getProviderCategories(providerId);
        setCategories(data.categories);
        setError(null);
      } catch (err: any) {
        setError(err.response?.data?.error || 'Failed to fetch categories');
      } finally {
        setLoading(false);
      }
    };

    fetchProviderCategories();
  }, [providerId]);

  return { categories, loading, error };
};

/**
 * Hook to manage category updates
 */
export const useCategoryUpdate = () => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const updateCategories = async (categoryIds: number[]) => {
    try {
      setLoading(true);
      setError(null);
      const result = await categoryService.updateMyCategories(categoryIds);
      return result;
    } catch (err: any) {
      const errorMsg = err.response?.data?.error || 'Failed to update categories';
      setError(errorMsg);
      throw new Error(errorMsg);
    } finally {
      setLoading(false);
    }
  };

  return { updateCategories, loading, error };
};
```

---

## 🎨 React Components

### 1. Category Selector (Provider Profile)

```tsx
// components/CategorySelector.tsx
import React, { useState, useEffect } from 'react';
import { useCategories, useCategoryUpdate, useProviderCategories } from '@/hooks/useCategories';
import { ServiceCategory } from '@/types/serviceCategory';

interface CategorySelectorProps {
  providerId?: number; // For viewing existing selections
  onSave?: (categoryIds: number[]) => void;
}

export const CategorySelector: React.FC<CategorySelectorProps> = ({
  providerId,
  onSave,
}) => {
  const { categories, loading: loadingCategories } = useCategories(false); // No adult
  const { categories: providerCategories, loading: loadingProvider } = 
    useProviderCategories(providerId || null);
  const { updateCategories, loading: updating, error } = useCategoryUpdate();
  
  const [selected, setSelected] = useState<number[]>([]);

  // Load provider's existing categories
  useEffect(() => {
    if (providerCategories.length > 0) {
      setSelected(providerCategories.map(c => c.category_id));
    }
  }, [providerCategories]);

  const handleToggle = (categoryId: number) => {
    if (selected.includes(categoryId)) {
      setSelected(selected.filter(id => id !== categoryId));
    } else if (selected.length < 5) {
      setSelected([...selected, categoryId]);
    } else {
      alert('คุณสามารถเลือกได้สูงสุด 5 หมวดหมู่');
    }
  };

  const handleSave = async () => {
    try {
      await updateCategories(selected);
      onSave?.(selected);
      alert('บันทึกสำเร็จ!');
    } catch (err: any) {
      alert(err.message || 'เกิดข้อผิดพลาด');
    }
  };

  if (loadingCategories || loadingProvider) {
    return <div>Loading...</div>;
  }

  return (
    <div className="category-selector">
      <h3 className="text-xl font-bold mb-4">
        เลือกหมวดหมู่บริการของคุณ ({selected.length}/5)
      </h3>

      {error && (
        <div className="bg-red-100 text-red-700 p-3 rounded mb-4">
          {error}
        </div>
      )}

      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3">
        {categories.map((category) => (
          <button
            key={category.category_id}
            onClick={() => handleToggle(category.category_id)}
            disabled={
              !selected.includes(category.category_id) && selected.length >= 5
            }
            className={`
              p-4 rounded-lg border-2 transition-all
              ${
                selected.includes(category.category_id)
                  ? 'border-blue-500 bg-blue-50 shadow-md'
                  : 'border-gray-300 hover:border-gray-400'
              }
              ${
                !selected.includes(category.category_id) && selected.length >= 5
                  ? 'opacity-50 cursor-not-allowed'
                  : 'cursor-pointer'
              }
            `}
          >
            <div className="text-3xl mb-2">{category.icon}</div>
            <div className="text-sm font-medium">{category.name_thai}</div>
            {category.is_adult && (
              <span className="text-xs text-red-500 mt-1">18+</span>
            )}
          </button>
        ))}
      </div>

      <div className="mt-6 flex justify-end">
        <button
          onClick={handleSave}
          disabled={updating || selected.length === 0}
          className="px-6 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:bg-gray-300"
        >
          {updating ? 'กำลังบันทึก...' : 'บันทึก'}
        </button>
      </div>
    </div>
  );
};
```

### 2. Category Filter (Browse Page)

```tsx
// components/CategoryFilter.tsx
import React, { useState } from 'react';
import { useCategories } from '@/hooks/useCategories';

interface CategoryFilterProps {
  selectedCategory: number | null;
  onCategorySelect: (categoryId: number | null) => void;
  includeAdult?: boolean;
}

export const CategoryFilter: React.FC<CategoryFilterProps> = ({
  selectedCategory,
  onCategorySelect,
  includeAdult = false,
}) => {
  const { categories, loading } = useCategories(includeAdult);

  if (loading) {
    return <div>Loading categories...</div>;
  }

  return (
    <div className="category-filter">
      <h4 className="font-semibold mb-3">กรองตามหมวดหมู่</h4>
      
      <div className="flex flex-wrap gap-2">
        <button
          onClick={() => onCategorySelect(null)}
          className={`
            px-4 py-2 rounded-full border transition-all
            ${
              selectedCategory === null
                ? 'bg-blue-500 text-white border-blue-500'
                : 'bg-white text-gray-700 border-gray-300 hover:border-gray-400'
            }
          `}
        >
          ทั้งหมด
        </button>

        {categories.map((category) => (
          <button
            key={category.category_id}
            onClick={() => onCategorySelect(category.category_id)}
            className={`
              px-4 py-2 rounded-full border transition-all
              ${
                selectedCategory === category.category_id
                  ? 'bg-blue-500 text-white border-blue-500'
                  : 'bg-white text-gray-700 border-gray-300 hover:border-gray-400'
              }
            `}
          >
            <span className="mr-1">{category.icon}</span>
            {category.name_thai}
          </button>
        ))}
      </div>
    </div>
  );
};
```

### 3. Provider Category Badges

```tsx
// components/ProviderCategoryBadges.tsx
import React from 'react';
import { useProviderCategories } from '@/hooks/useCategories';

interface ProviderCategoryBadgesProps {
  providerId: number;
  maxDisplay?: number; // Max badges to show
}

export const ProviderCategoryBadges: React.FC<ProviderCategoryBadgesProps> = ({
  providerId,
  maxDisplay = 3,
}) => {
  const { categories, loading } = useProviderCategories(providerId);

  if (loading) {
    return <div className="text-sm text-gray-400">Loading...</div>;
  }

  if (categories.length === 0) {
    return null;
  }

  const displayedCategories = categories.slice(0, maxDisplay);
  const remaining = categories.length - maxDisplay;

  return (
    <div className="flex flex-wrap gap-2 mt-2">
      {displayedCategories.map((category) => (
        <span
          key={category.category_id}
          className="inline-flex items-center px-2 py-1 rounded-full text-xs bg-gray-100 text-gray-700"
        >
          <span className="mr-1">{category.icon}</span>
          {category.name_thai}
        </span>
      ))}
      
      {remaining > 0 && (
        <span className="inline-flex items-center px-2 py-1 rounded-full text-xs bg-gray-200 text-gray-600">
          +{remaining} อื่นๆ
        </span>
      )}
    </div>
  );
};
```

### 4. Browse by Category Page

```tsx
// pages/BrowseByCategory.tsx
import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { categoryService } from '@/services/categoryService';
import { ProviderWithLocation } from '@/types/serviceCategory';
import { ProviderCard } from '@/components/ProviderCard';

export const BrowseByCategoryPage: React.FC = () => {
  const { categoryId } = useParams<{ categoryId: string }>();
  const [providers, setProviders] = useState<ProviderWithLocation[]>([]);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);

  useEffect(() => {
    const fetchProviders = async () => {
      if (!categoryId) return;

      try {
        setLoading(true);
        const data = await categoryService.browseByCategory(
          parseInt(categoryId),
          page,
          20
        );
        setProviders(data.providers);
        setTotalPages(data.total_pages);
      } catch (err) {
        console.error('Failed to fetch providers:', err);
      } finally {
        setLoading(false);
      }
    };

    fetchProviders();
  }, [categoryId, page]);

  if (loading) {
    return <div>Loading providers...</div>;
  }

  return (
    <div className="browse-by-category">
      <h1 className="text-2xl font-bold mb-6">ผู้ให้บริการในหมวดหมู่นี้</h1>

      {providers.length === 0 ? (
        <div className="text-center text-gray-500 py-12">
          ไม่พบผู้ให้บริการในหมวดหมู่นี้
        </div>
      ) : (
        <>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {providers.map((provider) => (
              <ProviderCard key={provider.user_id} provider={provider} />
            ))}
          </div>

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="flex justify-center gap-2 mt-8">
              <button
                onClick={() => setPage(p => Math.max(1, p - 1))}
                disabled={page === 1}
                className="px-4 py-2 border rounded disabled:opacity-50"
              >
                ก่อนหน้า
              </button>
              
              <span className="px-4 py-2">
                หน้า {page} / {totalPages}
              </span>
              
              <button
                onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                disabled={page === totalPages}
                className="px-4 py-2 border rounded disabled:opacity-50"
              >
                ถัดไป
              </button>
            </div>
          )}
        </>
      )}
    </div>
  );
};
```

---

## 🗂️ Category Data Reference

```typescript
// constants/categories.ts

export const CATEGORY_GROUPS = {
  ADULT: 'Adult Services (18+)',
  HEALTHCARE: 'Healthcare & Wellness',
  ENTERTAINMENT: 'Entertainment & Events',
  SOCIAL: 'Social Companionship',
  PROFESSIONAL: 'Professional Services',
} as const;

export const CATEGORY_LIST = [
  // Adult (18+)
  { id: 1, name: 'adult_entertainment', nameThai: 'บริการผู้ใหญ่', icon: '🔞', group: 'ADULT' },
  { id: 2, name: 'escort', nameThai: 'แอสคอร์ท', icon: '💋', group: 'ADULT' },
  
  // Healthcare
  { id: 3, name: 'massage_therapy', nameThai: 'นวดบำบัด', icon: '💆', group: 'HEALTHCARE' },
  { id: 4, name: 'spa_wellness', nameThai: 'สปาและเวลเนส', icon: '🧖', group: 'HEALTHCARE' },
  { id: 5, name: 'personal_care', nameThai: 'ดูแลส่วนตัว', icon: '🤲', group: 'HEALTHCARE' },
  { id: 6, name: 'healthcare_companion', nameThai: 'เพื่อนดูแลสุขภาพ', icon: '🏥', group: 'HEALTHCARE' },
  
  // Entertainment
  { id: 7, name: 'bartender', nameThai: 'บาร์เทนเดอร์', icon: '🍷', group: 'ENTERTAINMENT' },
  { id: 8, name: 'party_host', nameThai: 'พิธีกรงานปาร์ตี้', icon: '🎉', group: 'ENTERTAINMENT' },
  { id: 9, name: 'karaoke_companion', nameThai: 'เพื่อนร้องเพลง', icon: '🎤', group: 'ENTERTAINMENT' },
  
  // Social
  { id: 10, name: 'dining_companion', nameThai: 'เพื่อนทานอาหาร', icon: '🍽️', group: 'SOCIAL' },
  { id: 11, name: 'movie_companion', nameThai: 'เพื่อนดูหนัง', icon: '🎬', group: 'SOCIAL' },
  { id: 12, name: 'shopping_companion', nameThai: 'เพื่อนช็อปปิ้ง', icon: '🛍️', group: 'SOCIAL' },
  { id: 13, name: 'travel_companion', nameThai: 'เพื่อนเดินทาง', icon: '✈️', group: 'SOCIAL' },
  
  // Professional
  { id: 14, name: 'personal_assistant', nameThai: 'ผู้ช่วยส่วนตัว', icon: '👔', group: 'PROFESSIONAL' },
  { id: 15, name: 'event_companion', nameThai: 'เพื่อนร่วมงานอีเว้นท์', icon: '🎪', group: 'PROFESSIONAL' },
  { id: 16, name: 'language_practice', nameThai: 'ฝึกภาษา', icon: '📚', group: 'PROFESSIONAL' },
  { id: 17, name: 'fitness_trainer', nameThai: 'เทรนเนอร์ส่วนตัว', icon: '💪', group: 'PROFESSIONAL' },
  { id: 18, name: 'hobby_companion', nameThai: 'เพื่อนทำกิจกรรม', icon: '🎨', group: 'PROFESSIONAL' },
  { id: 19, name: 'photo_model', nameThai: 'นางแบบถ่ายภาพ', icon: '📷', group: 'PROFESSIONAL' },
  { id: 20, name: 'music_companion', nameThai: 'เพื่อนฟังเพลง', icon: '🎵', group: 'PROFESSIONAL' },
] as const;
```

---

## 🧪 Testing Examples

```typescript
// tests/categoryService.test.ts
import { categoryService } from '@/services/categoryService';

describe('CategoryService', () => {
  test('should fetch all categories', async () => {
    const result = await categoryService.getCategories(true);
    expect(result.total).toBe(20);
    expect(result.categories).toHaveLength(20);
  });

  test('should filter adult categories', async () => {
    const result = await categoryService.getCategories(false);
    expect(result.total).toBe(18);
    expect(result.categories.every(c => !c.is_adult)).toBe(true);
  });

  test('should update provider categories', async () => {
    const categoryIds = [3, 7, 10];
    const result = await categoryService.updateMyCategories(categoryIds);
    expect(result.total).toBe(3);
    expect(result.category_ids).toEqual(categoryIds);
  });

  test('should throw error for more than 5 categories', async () => {
    const categoryIds = [1, 2, 3, 4, 5, 6];
    await expect(
      categoryService.updateMyCategories(categoryIds)
    ).rejects.toThrow('Cannot select more than 5 categories');
  });
});
```

---

## 🎯 Quick Start Checklist

- [ ] Copy TypeScript types to `types/serviceCategory.ts`
- [ ] Copy API service to `services/categoryService.ts`
- [ ] Copy hooks to `hooks/useCategories.ts`
- [ ] Copy components to `components/`
- [ ] Update your API base URL in service constructor
- [ ] Add category routes to your router
- [ ] Test with Postman/Thunder Client
- [ ] Style components with your design system

---

## 📞 Support

**Backend API:** http://localhost:8080  
**Documentation:** SERVICE_CATEGORY_API.md  
**Migration:** migrations/012_add_service_categories.sql

---

**Last Updated:** November 14, 2025  
**Frontend Framework:** React + TypeScript  
**State Management:** React Hooks  
**HTTP Client:** Axios
