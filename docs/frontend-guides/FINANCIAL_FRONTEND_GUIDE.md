# 💰 Financial System - Frontend Integration Guide

## 📋 Table of Contents

1. [TypeScript Types & Interfaces](#typescript-types)
2. [API Service Class](#api-service)
3. [React Hooks](#react-hooks)
4. [React Components](#react-components)
5. [State Management](#state-management)
6. [Testing Examples](#testing)

---

## 🎯 TypeScript Types & Interfaces

```typescript
// src/types/financial.ts

export enum TransactionType {
  BOOKING_PAYMENT = 'booking_payment',
  COMMISSION = 'commission',
  PROVIDER_EARNING = 'provider_earning',
  WITHDRAWAL = 'withdrawal',
  WITHDRAWAL_FEE = 'withdrawal_fee',
  SUBSCRIPTION_FEE = 'subscription_fee',
  ADMIN_ADJUSTMENT = 'admin_adjustment',
  BONUS = 'bonus',
  PENALTY = 'penalty',
}

export enum TransactionStatus {
  PENDING = 'pending',
  PROCESSING = 'processing',
  COMPLETED = 'completed',
  FAILED = 'failed',
  CANCELLED = 'cancelled',
  REFUNDED = 'refunded',
}

export enum WithdrawalStatus {
  PENDING = 'pending',
  APPROVED = 'approved',
  PROCESSING = 'processing',
  COMPLETED = 'completed',
  REJECTED = 'rejected',
  FAILED = 'failed',
}

export enum AccountType {
  SAVINGS = 'savings',
  CURRENT = 'current',
}

export interface BankAccount {
  bank_account_id: number;
  user_id: number;
  bank_name: string;
  bank_code: string;
  account_number: string;
  account_name: string;
  account_type: AccountType;
  branch?: string;
  is_verified: boolean;
  is_default: boolean;
  is_active: boolean;
  verified_at?: string;
  verified_by?: number;
  created_at: string;
  updated_at: string;
}

export interface Wallet {
  wallet_id: number;
  user_id: number;
  available_balance: number;
  pending_balance: number;
  total_earned: number;
  total_withdrawn: number;
  total_commission_paid: number;
  currency: string;
  created_at: string;
  updated_at: string;
}

export interface Transaction {
  transaction_id: number;
  transaction_uuid: string;
  user_id: number;
  related_user_id?: number;
  type: TransactionType;
  status: TransactionStatus;
  amount: number;
  commission_amount?: number;
  net_amount?: number;
  currency: string;
  booking_id?: number;
  withdrawal_id?: number;
  payment_method?: string;
  payment_intent_id?: string;
  description?: string;
  metadata?: Record<string, any>;
  processed_at?: string;
  created_at: string;
  updated_at: string;
}

export interface Withdrawal {
  withdrawal_id: number;
  withdrawal_uuid: string;
  user_id: number;
  bank_account_id: number;
  requested_amount: number;
  fee: number;
  net_amount: number;
  status: WithdrawalStatus;
  requested_at: string;
  approved_at?: string;
  approved_by?: number;
  rejected_at?: string;
  rejected_by?: number;
  rejection_reason?: string;
  completed_at?: string;
  transfer_reference?: string;
  transfer_slip_url?: string;
  admin_notes?: string;
  created_at: string;
  updated_at: string;
  // Joined data
  bank_account?: BankAccount;
}

export interface WalletSummary {
  wallet: Wallet;
  recent_transactions: Transaction[];
}

export interface FinancialSummary {
  today_revenue: number;
  today_commission: number;
  month_revenue: number;
  month_commission: number;
  pending_withdrawals_count: number;
  pending_withdrawals_amount: number;
  active_providers: number;
  total_transactions_today: number;
}

// Request DTOs
export interface AddBankAccountRequest {
  bank_name: string;
  bank_code: string;
  account_number: string;
  account_name: string;
  account_type: AccountType;
  branch?: string;
  is_default?: boolean;
}

export interface RequestWithdrawalRequest {
  bank_account_id: number;
  amount: number;
}

export interface ProcessWithdrawalRequest {
  action: 'approve' | 'reject' | 'complete';
  rejection_reason?: string;
  transfer_reference?: string;
  transfer_slip_url?: string;
  admin_notes?: string;
}

export interface AdjustWalletRequest {
  amount: number;
  type: 'bonus' | 'penalty' | 'adjustment';
  description: string;
}

export interface GenerateReportRequest {
  report_type: 'daily' | 'weekly' | 'monthly' | 'yearly';
  period_start: string;
  period_end: string;
}

// Response DTOs
export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  message?: string;
  error?: string;
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}
```

---

## 🔌 API Service Class

```typescript
// src/services/FinancialService.ts

import axios, { AxiosInstance } from 'axios';
import {
  BankAccount,
  Wallet,
  Transaction,
  Withdrawal,
  WalletSummary,
  FinancialSummary,
  AddBankAccountRequest,
  RequestWithdrawalRequest,
  ProcessWithdrawalRequest,
  AdjustWalletRequest,
  GenerateReportRequest,
  ApiResponse,
  PaginatedResponse,
  WithdrawalStatus,
  TransactionType,
} from '../types/financial';

class FinancialService {
  private api: AxiosInstance;

  constructor(baseURL: string = 'http://localhost:8080') {
    this.api = axios.create({
      baseURL,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Add auth token interceptor
    this.api.interceptors.request.use((config) => {
      const token = localStorage.getItem('auth_token');
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
      return config;
    });
  }

  // ==================== User Bank Accounts ====================

  async addBankAccount(data: AddBankAccountRequest): Promise<ApiResponse<BankAccount>> {
    const response = await this.api.post<ApiResponse<BankAccount>>('/bank-accounts', data);
    return response.data;
  }

  async getMyBankAccounts(): Promise<ApiResponse<BankAccount[]>> {
    const response = await this.api.get<ApiResponse<BankAccount[]>>('/bank-accounts');
    return response.data;
  }

  async deleteBankAccount(bankAccountId: number): Promise<ApiResponse<void>> {
    const response = await this.api.delete<ApiResponse<void>>(`/bank-accounts/${bankAccountId}`);
    return response.data;
  }

  // ==================== Wallet ====================

  async getMyWallet(): Promise<ApiResponse<WalletSummary>> {
    const response = await this.api.get<ApiResponse<WalletSummary>>('/wallet');
    return response.data;
  }

  // ==================== Withdrawals ====================

  async requestWithdrawal(data: RequestWithdrawalRequest): Promise<ApiResponse<Withdrawal>> {
    const response = await this.api.post<ApiResponse<Withdrawal>>('/withdrawals', data);
    return response.data;
  }

  async getMyWithdrawals(status?: WithdrawalStatus): Promise<ApiResponse<Withdrawal[]>> {
    const params = status ? { status } : {};
    const response = await this.api.get<ApiResponse<Withdrawal[]>>('/withdrawals', { params });
    return response.data;
  }

  // ==================== Transactions ====================

  async getMyTransactions(
    page: number = 1,
    pageSize: number = 20,
    type?: TransactionType
  ): Promise<ApiResponse<PaginatedResponse<Transaction>>> {
    const params = { page, page_size: pageSize, ...(type && { type }) };
    const response = await this.api.get<ApiResponse<PaginatedResponse<Transaction>>>(
      '/transactions',
      { params }
    );
    return response.data;
  }

  // ==================== Admin Endpoints ====================

  async adminGetPendingWithdrawals(status?: WithdrawalStatus): Promise<ApiResponse<Withdrawal[]>> {
    const params = status ? { status } : {};
    const response = await this.api.get<ApiResponse<Withdrawal[]>>('/admin/withdrawals', {
      params,
    });
    return response.data;
  }

  async adminProcessWithdrawal(
    withdrawalId: number,
    data: ProcessWithdrawalRequest
  ): Promise<ApiResponse<Withdrawal>> {
    const response = await this.api.post<ApiResponse<Withdrawal>>(
      `/admin/withdrawals/${withdrawalId}/process`,
      data
    );
    return response.data;
  }

  async adminVerifyBankAccount(
    bankAccountId: number,
    notes?: string
  ): Promise<ApiResponse<BankAccount>> {
    const response = await this.api.post<ApiResponse<BankAccount>>(
      `/admin/bank-accounts/${bankAccountId}/verify`,
      { notes }
    );
    return response.data;
  }

  async adminGetFinancialSummary(): Promise<ApiResponse<FinancialSummary>> {
    const response = await this.api.get<ApiResponse<FinancialSummary>>(
      '/admin/financial/summary'
    );
    return response.data;
  }

  async adminGenerateReport(data: GenerateReportRequest): Promise<ApiResponse<any>> {
    const response = await this.api.post<ApiResponse<any>>('/admin/financial/reports', data);
    return response.data;
  }

  async adminGetUserWallet(userId: number): Promise<ApiResponse<WalletSummary>> {
    const response = await this.api.get<ApiResponse<WalletSummary>>(`/admin/wallets/${userId}`);
    return response.data;
  }

  async adminAdjustWallet(
    userId: number,
    data: AdjustWalletRequest
  ): Promise<ApiResponse<Transaction>> {
    const response = await this.api.post<ApiResponse<Transaction>>(
      `/admin/wallets/${userId}/adjust`,
      data
    );
    return response.data;
  }
}

export default new FinancialService();
```

---

## 🪝 React Hooks

```typescript
// src/hooks/useWallet.ts

import { useState, useEffect } from 'react';
import FinancialService from '../services/FinancialService';
import { WalletSummary } from '../types/financial';

export const useWallet = () => {
  const [wallet, setWallet] = useState<WalletSummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchWallet = async () => {
    try {
      setLoading(true);
      const response = await FinancialService.getMyWallet();
      if (response.success && response.data) {
        setWallet(response.data);
        setError(null);
      } else {
        setError(response.error || 'Failed to fetch wallet');
      }
    } catch (err: any) {
      setError(err.response?.data?.error || 'An error occurred');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchWallet();
  }, []);

  return { wallet, loading, error, refetch: fetchWallet };
};
```

```typescript
// src/hooks/useBankAccounts.ts

import { useState, useEffect } from 'react';
import FinancialService from '../services/FinancialService';
import { BankAccount, AddBankAccountRequest } from '../types/financial';

export const useBankAccounts = () => {
  const [accounts, setAccounts] = useState<BankAccount[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchAccounts = async () => {
    try {
      setLoading(true);
      const response = await FinancialService.getMyBankAccounts();
      if (response.success && response.data) {
        setAccounts(response.data);
        setError(null);
      } else {
        setError(response.error || 'Failed to fetch bank accounts');
      }
    } catch (err: any) {
      setError(err.response?.data?.error || 'An error occurred');
    } finally {
      setLoading(false);
    }
  };

  const addAccount = async (data: AddBankAccountRequest) => {
    try {
      const response = await FinancialService.addBankAccount(data);
      if (response.success) {
        await fetchAccounts(); // Refresh list
        return { success: true };
      }
      return { success: false, error: response.error };
    } catch (err: any) {
      return { success: false, error: err.response?.data?.error || 'Failed to add account' };
    }
  };

  const deleteAccount = async (bankAccountId: number) => {
    try {
      const response = await FinancialService.deleteBankAccount(bankAccountId);
      if (response.success) {
        await fetchAccounts(); // Refresh list
        return { success: true };
      }
      return { success: false, error: response.error };
    } catch (err: any) {
      return { success: false, error: err.response?.data?.error || 'Failed to delete account' };
    }
  };

  useEffect(() => {
    fetchAccounts();
  }, []);

  return { accounts, loading, error, addAccount, deleteAccount, refetch: fetchAccounts };
};
```

```typescript
// src/hooks/useWithdrawals.ts

import { useState, useEffect } from 'react';
import FinancialService from '../services/FinancialService';
import { Withdrawal, RequestWithdrawalRequest, WithdrawalStatus } from '../types/financial';

export const useWithdrawals = (status?: WithdrawalStatus) => {
  const [withdrawals, setWithdrawals] = useState<Withdrawal[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchWithdrawals = async () => {
    try {
      setLoading(true);
      const response = await FinancialService.getMyWithdrawals(status);
      if (response.success && response.data) {
        setWithdrawals(response.data);
        setError(null);
      } else {
        setError(response.error || 'Failed to fetch withdrawals');
      }
    } catch (err: any) {
      setError(err.response?.data?.error || 'An error occurred');
    } finally {
      setLoading(false);
    }
  };

  const requestWithdrawal = async (data: RequestWithdrawalRequest) => {
    try {
      const response = await FinancialService.requestWithdrawal(data);
      if (response.success) {
        await fetchWithdrawals(); // Refresh list
        return { success: true, data: response.data };
      }
      return { success: false, error: response.error };
    } catch (err: any) {
      return { success: false, error: err.response?.data?.error || 'Failed to request withdrawal' };
    }
  };

  useEffect(() => {
    fetchWithdrawals();
  }, [status]);

  return { withdrawals, loading, error, requestWithdrawal, refetch: fetchWithdrawals };
};
```

---

## 🧩 React Components

### 1. Wallet Dashboard Component

```tsx
// src/components/WalletDashboard.tsx

import React from 'react';
import { useWallet } from '../hooks/useWallet';
import { formatCurrency, formatDate } from '../utils/formatters';
import { TransactionType } from '../types/financial';

const WalletDashboard: React.FC = () => {
  const { wallet, loading, error } = useWallet();

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-lg p-4">
        <p className="text-red-800">{error}</p>
      </div>
    );
  }

  if (!wallet) return null;

  const getTransactionTypeLabel = (type: TransactionType): string => {
    const labels: Record<TransactionType, string> = {
      [TransactionType.BOOKING_PAYMENT]: 'การจ่ายเงิน',
      [TransactionType.COMMISSION]: 'ค่าคอมมิชชั่น',
      [TransactionType.PROVIDER_EARNING]: 'รายได้',
      [TransactionType.WITHDRAWAL]: 'ถอนเงิน',
      [TransactionType.WITHDRAWAL_FEE]: 'ค่าถอนเงิน',
      [TransactionType.SUBSCRIPTION_FEE]: 'ค่าสมาชิก',
      [TransactionType.ADMIN_ADJUSTMENT]: 'ปรับยอด',
      [TransactionType.BONUS]: 'โบนัส',
      [TransactionType.PENALTY]: 'ค่าปรับ',
    };
    return labels[type] || type;
  };

  const getTransactionColor = (type: TransactionType): string => {
    if (type === TransactionType.PROVIDER_EARNING || type === TransactionType.BONUS) {
      return 'text-green-600';
    }
    if (
      type === TransactionType.WITHDRAWAL ||
      type === TransactionType.COMMISSION ||
      type === TransactionType.PENALTY
    ) {
      return 'text-red-600';
    }
    return 'text-gray-600';
  };

  return (
    <div className="space-y-6">
      {/* Balance Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {/* Available Balance */}
        <div className="bg-gradient-to-br from-green-500 to-green-600 rounded-lg shadow-lg p-6 text-white">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-green-100 text-sm font-medium">ยอดพร้อมถอน</p>
              <p className="text-3xl font-bold mt-2">
                {formatCurrency(wallet.wallet.available_balance)}
              </p>
            </div>
            <svg
              className="w-12 h-12 text-green-200"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
          </div>
        </div>

        {/* Pending Balance */}
        <div className="bg-gradient-to-br from-yellow-500 to-yellow-600 rounded-lg shadow-lg p-6 text-white">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-yellow-100 text-sm font-medium">ยอดรอยืนยัน</p>
              <p className="text-3xl font-bold mt-2">
                {formatCurrency(wallet.wallet.pending_balance)}
              </p>
              <p className="text-yellow-100 text-xs mt-1">รอ 7 วัน</p>
            </div>
            <svg
              className="w-12 h-12 text-yellow-200"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
          </div>
        </div>

        {/* Total Earned */}
        <div className="bg-gradient-to-br from-indigo-500 to-indigo-600 rounded-lg shadow-lg p-6 text-white">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-indigo-100 text-sm font-medium">รายได้สะสม</p>
              <p className="text-3xl font-bold mt-2">
                {formatCurrency(wallet.wallet.total_earned)}
              </p>
              <p className="text-indigo-100 text-xs mt-1">
                ถอนไปแล้ว {formatCurrency(wallet.wallet.total_withdrawn)}
              </p>
            </div>
            <svg
              className="w-12 h-12 text-indigo-200"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6"
              />
            </svg>
          </div>
        </div>
      </div>

      {/* Recent Transactions */}
      <div className="bg-white rounded-lg shadow-md">
        <div className="px-6 py-4 border-b border-gray-200">
          <h3 className="text-lg font-semibold text-gray-900">ธุรกรรมล่าสุด</h3>
        </div>
        <div className="divide-y divide-gray-200">
          {wallet.recent_transactions.length === 0 ? (
            <div className="px-6 py-8 text-center text-gray-500">ยังไม่มีธุรกรรม</div>
          ) : (
            wallet.recent_transactions.map((transaction) => (
              <div key={transaction.transaction_id} className="px-6 py-4 hover:bg-gray-50">
                <div className="flex items-center justify-between">
                  <div className="flex-1">
                    <p className="text-sm font-medium text-gray-900">
                      {getTransactionTypeLabel(transaction.type)}
                    </p>
                    {transaction.description && (
                      <p className="text-xs text-gray-500 mt-1">{transaction.description}</p>
                    )}
                    <p className="text-xs text-gray-400 mt-1">
                      {formatDate(transaction.created_at)}
                    </p>
                  </div>
                  <div className="text-right">
                    <p className={`text-lg font-semibold ${getTransactionColor(transaction.type)}`}>
                      {transaction.type === TransactionType.PROVIDER_EARNING ||
                      transaction.type === TransactionType.BONUS
                        ? '+'
                        : '-'}
                      {formatCurrency(transaction.net_amount || transaction.amount)}
                    </p>
                    <span
                      className={`inline-block px-2 py-1 text-xs rounded-full ${
                        transaction.status === 'completed'
                          ? 'bg-green-100 text-green-800'
                          : transaction.status === 'pending'
                          ? 'bg-yellow-100 text-yellow-800'
                          : 'bg-gray-100 text-gray-800'
                      }`}
                    >
                      {transaction.status}
                    </span>
                  </div>
                </div>
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  );
};

export default WalletDashboard;
```

### 2. Bank Account Management Component

```tsx
// src/components/BankAccountManager.tsx

import React, { useState } from 'react';
import { useBankAccounts } from '../hooks/useBankAccounts';
import { AccountType } from '../types/financial';

const BankAccountManager: React.FC = () => {
  const { accounts, loading, error, addAccount, deleteAccount } = useBankAccounts();
  const [showForm, setShowForm] = useState(false);
  const [formData, setFormData] = useState({
    bank_name: '',
    bank_code: '',
    account_number: '',
    account_name: '',
    account_type: AccountType.SAVINGS,
    branch: '',
    is_default: false,
  });
  const [submitting, setSubmitting] = useState(false);

  const thBanks = [
    { code: 'KBANK', name: 'ธนาคารกสิกรไทย' },
    { code: 'SCB', name: 'ธนาคารไทยพาณิชย์' },
    { code: 'BBL', name: 'ธนาคารกรุงเทพ' },
    { code: 'KTB', name: 'ธนาคารกรุงไทย' },
    { code: 'TMB', name: 'ธนาคารทหารไทยธนชาต' },
    { code: 'BAY', name: 'ธนาคารกรุงศรีอยุธยา' },
  ];

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);

    const result = await addAccount(formData);

    if (result.success) {
      alert('เพิ่มบัญชีธนาคารสำเร็จ');
      setShowForm(false);
      setFormData({
        bank_name: '',
        bank_code: '',
        account_number: '',
        account_name: '',
        account_type: AccountType.SAVINGS,
        branch: '',
        is_default: false,
      });
    } else {
      alert(result.error || 'เกิดข้อผิดพลาด');
    }

    setSubmitting(false);
  };

  const handleDelete = async (bankAccountId: number) => {
    if (!confirm('ต้องการลบบัญชีนี้?')) return;

    const result = await deleteAccount(bankAccountId);
    if (result.success) {
      alert('ลบบัญชีสำเร็จ');
    } else {
      alert(result.error || 'เกิดข้อผิดพลาด');
    }
  };

  if (loading) {
    return <div className="text-center py-8">กำลังโหลด...</div>;
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-bold text-gray-900">บัญชีธนาคาร</h2>
        <button
          onClick={() => setShowForm(!showForm)}
          className="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700"
        >
          {showForm ? 'ยกเลิก' : '+ เพิ่มบัญชี'}
        </button>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 text-red-800">{error}</div>
      )}

      {/* Add Account Form */}
      {showForm && (
        <div className="bg-white rounded-lg shadow-md p-6">
          <h3 className="text-lg font-semibold mb-4">เพิ่มบัญชีธนาคาร</h3>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">ธนาคาร</label>
              <select
                value={formData.bank_code}
                onChange={(e) => {
                  const bank = thBanks.find((b) => b.code === e.target.value);
                  setFormData({
                    ...formData,
                    bank_code: e.target.value,
                    bank_name: bank?.name || '',
                  });
                }}
                required
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-indigo-500 focus:border-indigo-500"
              >
                <option value="">เลือกธนาคาร</option>
                {thBanks.map((bank) => (
                  <option key={bank.code} value={bank.code}>
                    {bank.name}
                  </option>
                ))}
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">เลขบัญชี</label>
              <input
                type="text"
                value={formData.account_number}
                onChange={(e) => setFormData({ ...formData, account_number: e.target.value })}
                required
                pattern="[0-9]{10,12}"
                placeholder="1234567890"
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-indigo-500 focus:border-indigo-500"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">ชื่อบัญชี</label>
              <input
                type="text"
                value={formData.account_name}
                onChange={(e) => setFormData({ ...formData, account_name: e.target.value })}
                required
                placeholder="นาย สมชาย ใจดี"
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-indigo-500 focus:border-indigo-500"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">ประเภทบัญชี</label>
              <select
                value={formData.account_type}
                onChange={(e) =>
                  setFormData({ ...formData, account_type: e.target.value as AccountType })
                }
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-indigo-500 focus:border-indigo-500"
              >
                <option value={AccountType.SAVINGS}>ออมทรัพย์</option>
                <option value={AccountType.CURRENT}>กระแสรายวัน</option>
              </select>
            </div>

            <div className="flex items-center">
              <input
                type="checkbox"
                checked={formData.is_default}
                onChange={(e) => setFormData({ ...formData, is_default: e.target.checked })}
                className="h-4 w-4 text-indigo-600 rounded focus:ring-indigo-500"
              />
              <label className="ml-2 text-sm text-gray-700">ตั้งเป็นบัญชีหลัก</label>
            </div>

            <button
              type="submit"
              disabled={submitting}
              className="w-full px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:opacity-50"
            >
              {submitting ? 'กำลังบันทึก...' : 'บันทึก'}
            </button>
          </form>
        </div>
      )}

      {/* Account List */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {accounts.length === 0 ? (
          <div className="col-span-2 text-center py-8 text-gray-500">
            ยังไม่มีบัญชีธนาคาร กรุณาเพิ่มบัญชีเพื่อถอนเงิน
          </div>
        ) : (
          accounts.map((account) => (
            <div
              key={account.bank_account_id}
              className="bg-white rounded-lg shadow-md p-6 relative"
            >
              {account.is_default && (
                <span className="absolute top-3 right-3 px-2 py-1 bg-indigo-100 text-indigo-800 text-xs rounded-full">
                  บัญชีหลัก
                </span>
              )}

              <div className="flex items-start justify-between">
                <div className="flex-1">
                  <h4 className="font-semibold text-gray-900">{account.bank_name}</h4>
                  <p className="text-sm text-gray-600 mt-1">{account.account_number}</p>
                  <p className="text-sm text-gray-600">{account.account_name}</p>

                  <div className="flex items-center gap-2 mt-3">
                    {account.is_verified ? (
                      <span className="px-2 py-1 bg-green-100 text-green-800 text-xs rounded-full flex items-center gap-1">
                        <svg className="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
                          <path
                            fillRule="evenodd"
                            d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                            clipRule="evenodd"
                          />
                        </svg>
                        ยืนยันแล้ว
                      </span>
                    ) : (
                      <span className="px-2 py-1 bg-yellow-100 text-yellow-800 text-xs rounded-full">
                        รอการยืนยัน
                      </span>
                    )}
                    <span className="text-xs text-gray-500">
                      {account.account_type === AccountType.SAVINGS ? 'ออมทรัพย์' : 'กระแสรายวัน'}
                    </span>
                  </div>
                </div>

                <button
                  onClick={() => handleDelete(account.bank_account_id)}
                  className="text-red-600 hover:text-red-800"
                >
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                    />
                  </svg>
                </button>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
};

export default BankAccountManager;
```

### 3. Withdrawal Request Component

```tsx
// src/components/WithdrawalRequest.tsx

import React, { useState } from 'react';
import { useBankAccounts } from '../hooks/useBankAccounts';
import { useWallet } from '../hooks/useWallet';
import { useWithdrawals } from '../hooks/useWithdrawals';
import { formatCurrency } from '../utils/formatters';

const WithdrawalRequest: React.FC = () => {
  const { accounts } = useBankAccounts();
  const { wallet } = useWallet();
  const { requestWithdrawal } = useWithdrawals();

  const [amount, setAmount] = useState('');
  const [selectedBankId, setSelectedBankId] = useState<number | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const fee = 10; // Fixed withdrawal fee
  const netAmount = parseFloat(amount) - fee || 0;
  const availableBalance = wallet?.wallet.available_balance || 0;

  const verifiedAccounts = accounts.filter((acc) => acc.is_verified && acc.is_active);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!selectedBankId) {
      alert('กรุณาเลือกบัญชีธนาคาร');
      return;
    }

    const amountNum = parseFloat(amount);

    if (amountNum < 100) {
      alert('ยอดถอนขั้นต่ำ 100 บาท');
      return;
    }

    if (amountNum > availableBalance) {
      alert('ยอดเงินไม่เพียงพอ');
      return;
    }

    setSubmitting(true);

    const result = await requestWithdrawal({
      bank_account_id: selectedBankId,
      amount: amountNum,
    });

    if (result.success) {
      alert('ขอถอนเงินสำเร็จ\nรอ Admin อนุมัติ');
      setAmount('');
      setSelectedBankId(null);
    } else {
      alert(result.error || 'เกิดข้อผิดพลาด');
    }

    setSubmitting(false);
  };

  if (verifiedAccounts.length === 0) {
    return (
      <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-6">
        <p className="text-yellow-800">
          กรุณาเพิ่มและยืนยันบัญชีธนาคารก่อนถอนเงิน
        </p>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-lg shadow-md p-6">
      <h3 className="text-xl font-semibold mb-6">ถอนเงิน</h3>

      <div className="mb-6 p-4 bg-indigo-50 rounded-lg">
        <div className="flex justify-between items-center">
          <span className="text-gray-700">ยอดพร้อมถอน:</span>
          <span className="text-2xl font-bold text-indigo-600">
            {formatCurrency(availableBalance)}
          </span>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            เลือกบัญชีธนาคาร
          </label>
          <select
            value={selectedBankId || ''}
            onChange={(e) => setSelectedBankId(parseInt(e.target.value))}
            required
            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-indigo-500 focus:border-indigo-500"
          >
            <option value="">เลือกบัญชี</option>
            {verifiedAccounts.map((account) => (
              <option key={account.bank_account_id} value={account.bank_account_id}>
                {account.bank_name} - {account.account_number} (
                {account.is_default ? 'บัญชีหลัก' : account.account_name})
              </option>
            ))}
          </select>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            จำนวนเงินที่ต้องการถอน
          </label>
          <input
            type="number"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            required
            min="100"
            step="1"
            placeholder="ขั้นต่ำ 100 บาท"
            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-indigo-500 focus:border-indigo-500"
          />
        </div>

        {amount && parseFloat(amount) >= 100 && (
          <div className="p-4 bg-gray-50 rounded-lg space-y-2">
            <div className="flex justify-between text-sm">
              <span className="text-gray-600">จำนวนที่ขอถอน:</span>
              <span className="font-medium">{formatCurrency(parseFloat(amount))}</span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-gray-600">ค่าธรรมเนียม:</span>
              <span className="text-red-600">-{formatCurrency(fee)}</span>
            </div>
            <div className="flex justify-between text-lg font-semibold pt-2 border-t">
              <span>ยอดที่จะได้รับ:</span>
              <span className="text-green-600">{formatCurrency(netAmount)}</span>
            </div>
          </div>
        )}

        <button
          type="submit"
          disabled={submitting || !amount || parseFloat(amount) < 100}
          className="w-full px-4 py-3 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed font-medium"
        >
          {submitting ? 'กำลังดำเนินการ...' : 'ขอถอนเงิน'}
        </button>
      </form>

      <div className="mt-6 p-4 bg-gray-50 rounded-lg">
        <h4 className="text-sm font-semibold text-gray-700 mb-2">ข้อมูลสำคัญ:</h4>
        <ul className="text-xs text-gray-600 space-y-1 list-disc list-inside">
          <li>ยอดถอนขั้นต่ำ 100 บาท</li>
          <li>ค่าธรรมเนียมการถอน 10 บาทต่อครั้ง</li>
          <li>สามารถถอนได้สูงสุด 3 ครั้งต่อวัน</li>
          <li>ระยะเวลาโอนเงิน 1-3 วันทำการ หลังจาก Admin อนุมัติ</li>
        </ul>
      </div>
    </div>
  );
};

export default WithdrawalRequest;
```

### 4. Admin Withdrawal Approval Component

```tsx
// src/components/admin/WithdrawalApproval.tsx

import React, { useState } from 'react';
import FinancialService from '../../services/FinancialService';
import { Withdrawal, WithdrawalStatus } from '../../types/financial';
import { formatCurrency, formatDate } from '../../utils/formatters';

const WithdrawalApproval: React.FC = () => {
  const [withdrawals, setWithdrawals] = useState<Withdrawal[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedStatus, setSelectedStatus] = useState<WithdrawalStatus | 'all'>('all');
  const [processingId, setProcessingId] = useState<number | null>(null);

  React.useEffect(() => {
    fetchWithdrawals();
  }, [selectedStatus]);

  const fetchWithdrawals = async () => {
    try {
      setLoading(true);
      const status = selectedStatus === 'all' ? undefined : selectedStatus;
      const response = await FinancialService.adminGetPendingWithdrawals(status);
      if (response.success && response.data) {
        setWithdrawals(response.data);
      }
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleApprove = async (withdrawalId: number) => {
    if (!confirm('ต้องการอนุมัติคำขอนี้?')) return;

    setProcessingId(withdrawalId);
    const result = await FinancialService.adminProcessWithdrawal(withdrawalId, {
      action: 'approve',
    });

    if (result.success) {
      alert('อนุมัติสำเร็จ');
      fetchWithdrawals();
    } else {
      alert(result.error || 'เกิดข้อผิดพลาด');
    }
    setProcessingId(null);
  };

  const handleReject = async (withdrawalId: number) => {
    const reason = prompt('เหตุผลในการปฏิเสธ:');
    if (!reason) return;

    setProcessingId(withdrawalId);
    const result = await FinancialService.adminProcessWithdrawal(withdrawalId, {
      action: 'reject',
      rejection_reason: reason,
    });

    if (result.success) {
      alert('ปฏิเสธสำเร็จ');
      fetchWithdrawals();
    } else {
      alert(result.error || 'เกิดข้อผิดพลาด');
    }
    setProcessingId(null);
  };

  const handleComplete = async (withdrawalId: number) => {
    const reference = prompt('เลขที่อ้างอิงการโอน:');
    if (!reference) return;

    setProcessingId(withdrawalId);
    const result = await FinancialService.adminProcessWithdrawal(withdrawalId, {
      action: 'complete',
      transfer_reference: reference,
    });

    if (result.success) {
      alert('โอนเงินสำเร็จ');
      fetchWithdrawals();
    } else {
      alert(result.error || 'เกิดข้อผิดพลาด');
    }
    setProcessingId(null);
  };

  const getStatusBadge = (status: WithdrawalStatus) => {
    const styles: Record<WithdrawalStatus, string> = {
      [WithdrawalStatus.PENDING]: 'bg-yellow-100 text-yellow-800',
      [WithdrawalStatus.APPROVED]: 'bg-blue-100 text-blue-800',
      [WithdrawalStatus.PROCESSING]: 'bg-indigo-100 text-indigo-800',
      [WithdrawalStatus.COMPLETED]: 'bg-green-100 text-green-800',
      [WithdrawalStatus.REJECTED]: 'bg-red-100 text-red-800',
      [WithdrawalStatus.FAILED]: 'bg-gray-100 text-gray-800',
    };
    return styles[status] || 'bg-gray-100 text-gray-800';
  };

  if (loading) {
    return <div className="text-center py-8">กำลังโหลด...</div>;
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-bold">คำขอถอนเงิน</h2>
        <select
          value={selectedStatus}
          onChange={(e) => setSelectedStatus(e.target.value as any)}
          className="px-3 py-2 border border-gray-300 rounded-lg"
        >
          <option value="all">ทั้งหมด</option>
          <option value={WithdrawalStatus.PENDING}>รอดำเนินการ</option>
          <option value={WithdrawalStatus.APPROVED}>อนุมัติแล้ว</option>
          <option value={WithdrawalStatus.COMPLETED}>เสร็จสิ้น</option>
          <option value={WithdrawalStatus.REJECTED}>ปฏิเสธ</option>
        </select>
      </div>

      <div className="grid gap-4">
        {withdrawals.length === 0 ? (
          <div className="text-center py-8 text-gray-500">ไม่มีคำขอ</div>
        ) : (
          withdrawals.map((withdrawal) => (
            <div key={withdrawal.withdrawal_id} className="bg-white rounded-lg shadow-md p-6">
              <div className="flex items-start justify-between mb-4">
                <div>
                  <h3 className="text-lg font-semibold">#{withdrawal.withdrawal_uuid}</h3>
                  <p className="text-sm text-gray-600">User ID: {withdrawal.user_id}</p>
                </div>
                <span className={`px-3 py-1 rounded-full text-sm ${getStatusBadge(withdrawal.status)}`}>
                  {withdrawal.status}
                </span>
              </div>

              <div className="grid grid-cols-2 gap-4 mb-4">
                <div>
                  <p className="text-sm text-gray-600">จำนวนที่ขอถอน</p>
                  <p className="text-xl font-bold">{formatCurrency(withdrawal.requested_amount)}</p>
                </div>
                <div>
                  <p className="text-sm text-gray-600">ยอดที่โอน (หักค่าธรรมเนียม)</p>
                  <p className="text-xl font-bold text-green-600">
                    {formatCurrency(withdrawal.net_amount)}
                  </p>
                </div>
              </div>

              {withdrawal.bank_account && (
                <div className="bg-gray-50 rounded p-3 mb-4">
                  <p className="text-sm font-semibold">{withdrawal.bank_account.bank_name}</p>
                  <p className="text-sm">{withdrawal.bank_account.account_number}</p>
                  <p className="text-sm">{withdrawal.bank_account.account_name}</p>
                </div>
              )}

              <p className="text-xs text-gray-500 mb-4">
                ขอเมื่อ: {formatDate(withdrawal.requested_at)}
              </p>

              {withdrawal.status === WithdrawalStatus.PENDING && (
                <div className="flex gap-2">
                  <button
                    onClick={() => handleApprove(withdrawal.withdrawal_id)}
                    disabled={processingId === withdrawal.withdrawal_id}
                    className="flex-1 px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 disabled:opacity-50"
                  >
                    อนุมัติ
                  </button>
                  <button
                    onClick={() => handleReject(withdrawal.withdrawal_id)}
                    disabled={processingId === withdrawal.withdrawal_id}
                    className="flex-1 px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 disabled:opacity-50"
                  >
                    ปฏิเสธ
                  </button>
                </div>
              )}

              {withdrawal.status === WithdrawalStatus.APPROVED && (
                <button
                  onClick={() => handleComplete(withdrawal.withdrawal_id)}
                  disabled={processingId === withdrawal.withdrawal_id}
                  className="w-full px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:opacity-50"
                >
                  โอนเงินเรียบร้อย
                </button>
              )}

              {withdrawal.rejection_reason && (
                <div className="mt-4 p-3 bg-red-50 rounded">
                  <p className="text-sm text-red-800">
                    <strong>เหตุผล:</strong> {withdrawal.rejection_reason}
                  </p>
                </div>
              )}
            </div>
          ))
        )}
      </div>
    </div>
  );
};

export default WithdrawalApproval;
```

---

## 📊 Utils & Formatters

```typescript
// src/utils/formatters.ts

export const formatCurrency = (amount: number, currency: string = 'THB'): string => {
  return new Intl.NumberFormat('th-TH', {
    style: 'currency',
    currency: currency,
    minimumFractionDigits: 0,
    maximumFractionDigits: 2,
  }).format(amount);
};

export const formatDate = (dateString: string): string => {
  const date = new Date(dateString);
  return new Intl.DateTimeFormat('th-TH', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date);
};

export const formatNumber = (num: number): string => {
  return new Intl.NumberFormat('th-TH').format(num);
};
```

---

## 🧪 Testing Examples

```typescript
// Test API calls
import FinancialService from './services/FinancialService';

// Get wallet
const { data } = await FinancialService.getMyWallet();
console.log('Wallet:', data);

// Add bank account
const newAccount = await FinancialService.addBankAccount({
  bank_name: 'ธนาคารกสิกรไทย',
  bank_code: 'KBANK',
  account_number: '1234567890',
  account_name: 'นาย สมชาย ใจดี',
  account_type: AccountType.SAVINGS,
  is_default: true,
});
console.log('New account:', newAccount);

// Request withdrawal
const withdrawal = await FinancialService.requestWithdrawal({
  bank_account_id: 1,
  amount: 1000,
});
console.log('Withdrawal:', withdrawal);
```

---

## ✅ Quick Start Checklist

1. **Install dependencies:**
   ```bash
   npm install axios
   npm install @types/node --save-dev  # If using TypeScript
   ```

2. **Copy types:**
   - Copy `src/types/financial.ts` to your project

3. **Add API service:**
   - Copy `src/services/FinancialService.ts`
   - Update `baseURL` to match your backend

4. **Add components:**
   - Choose components you need (Wallet, Bank Accounts, Withdrawals)
   - Customize styling to match your design system

5. **Setup routing:**
   ```typescript
   import WalletDashboard from './components/WalletDashboard';
   import BankAccountManager from './components/BankAccountManager';
   import WithdrawalRequest from './components/WithdrawalRequest';

   // In your router
   <Route path="/wallet" element={<WalletDashboard />} />
   <Route path="/bank-accounts" element={<BankAccountManager />} />
   <Route path="/withdrawals" element={<WithdrawalRequest />} />
   ```

6. **Test the flow:**
   - Provider adds bank account → shows unverified
   - Admin verifies bank account
   - Provider sees available balance
   - Provider requests withdrawal
   - Admin approves withdrawal
   - Funds transferred

---

## 🔐 Security Notes

- **Always use HTTPS** in production
- Store auth tokens securely (httpOnly cookies preferred)
- Validate all inputs on frontend AND backend
- Never store sensitive data in localStorage
- Use environment variables for API URLs
- Implement rate limiting for withdrawal requests
- Add 2FA for large withdrawals (optional)

---

## 🎨 Customization Tips

1. **Colors:** Replace Tailwind classes with your brand colors
2. **Icons:** Use your preferred icon library (Heroicons, Font Awesome, etc.)
3. **Animations:** Add transitions with `react-spring` or `framer-motion`
4. **Forms:** Use `react-hook-form` for complex form validation
5. **State:** Integrate with Redux/Zustand if needed
6. **Notifications:** Add toast notifications with `react-hot-toast`

---

**Last Updated:** November 14, 2025  
**Framework:** React + TypeScript + Tailwind CSS  
**Status:** ✅ Ready for Integration
