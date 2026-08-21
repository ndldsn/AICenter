import axios, { AxiosError, AxiosRequestConfig } from 'axios';
import { Message } from '@arco-design/web-react';

const api = axios.create({
    baseURL: '/api/v1',
    timeout: 30_000,
    headers: {
        'Content-Type': 'application/json',
    },
});

// Request interceptor - add auth token
api.interceptors.request.use(
    (config) => {
        const token = localStorage.getItem('access_token');
        if (token) {
            config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
    },
    (error) => Promise.reject(error)
);

// Response interceptor - handle errors
api.interceptors.response.use(
    (response) => response.data,
    (error: AxiosError<ApiError>) => {
        const status = error.response?.status;
        const data = error.response?.data as ApiError | undefined;

        if (status === 401) {
            localStorage.removeItem('access_token');
            localStorage.removeItem('refresh_token');
            if (window.location.pathname !== '/login') {
                window.location.href = '/login';
            }
        }

        const message = data?.message || error.message || 'Request failed';
        Message.error(message);

        return Promise.reject(error);
    }
);

// HTTP methods
export const apiGet = <T = unknown>(url: string, params?: Record<string, any>): Promise<T> =>
    api.get(url, { params });

export const apiPost = <T = unknown>(url: string, data?: any): Promise<T> =>
    api.post(url, data);

export const apiPut = <T = unknown>(url: string, data?: any): Promise<T> =>
    api.put(url, data);

export const apiPatch = <T = unknown>(url: string, data?: any): Promise<T> =>
    api.patch(url, data);

export const apiDelete = <T = unknown>(url: string): Promise<T> =>
    api.delete(url);

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
