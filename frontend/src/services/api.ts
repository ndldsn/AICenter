import axios from 'axios';
import type { ApiError } from './api';

const api = axios.create({
    baseURL: 'http://127.0.0.1:8081/api/v1',
    timeout: 30000,
    headers: {
        'Content-Type': 'application/json',
    },
});

// Request interceptor - add auth token
api.interceptors.request.use((config) => {
    const token = localStorage.getItem('access_token');
    if (token) {
        config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
}, (error) => Promise.reject(error));

// Response interceptor - handle errors
api.interceptors.response.use((response) => response.data, (error) => {
    const status = error.response?.status;
    const data = error.response?.data as ApiError | undefined;
    if (status === 401) {
        localStorage.removeItem('access_token');
        localStorage.removeItem('refresh_token');
        if (window.location.pathname !== '/login') {
            window.location.href = '/login';
        }
    }
    return Promise.reject(error);
});

// HTTP methods
export const apiGet = <T = unknown>(url: string, params?: Record<string, any>): Promise<T> =>
    api.get(url, { params }) as unknown as Promise<T>;

export const apiPost = <T = unknown>(url: string, data?: any): Promise<T> =>
    api.post(url, data) as unknown as Promise<T>;

export const apiPut = <T = unknown>(url: string, data?: any): Promise<T> =>
    api.put(url, data) as unknown as Promise<T>;

export const apiPatch = <T = unknown>(url: string, data?: any): Promise<T> =>
    api.patch(url, data) as unknown as Promise<T>;

export const apiDelete = <T = unknown>(url: string): Promise<T> =>
    api.delete(url) as unknown as Promise<T>;

export interface ApiError {
    code: number;
    message: string;
    errors?: Array<{ field: string; message: string }>;
}

export interface ApiResponse<T = unknown> {
    code: number;
    message: string;
    data: T;
}

export interface PaginatedResponse<T> {
    items: T[];
    total: number;
    page: number;
    limit: number;
}

export default api;
