import { describe, it, expect, vi, beforeEach } from 'vitest';

// Module-scope state, hoisted so the vi.mock factory can read/write it
// var is required: let/const would hit a TDZ because the factory body runs
// before the lexical binding is initialised, even though vi.mock hoists the call.
var _mockState: {
    configs: any[];
    _response: any;
    _reset(): void;
    setResponse(v: any): void;
};

// Replace the entire ./api module
vi.mock('./api', () => {
    const configs: any[] = [];

    const requestHandler = (config: any) => {
        const token = (() => {
            try { return localStorage.getItem('access_token'); } catch { return null; }
        })();
        if (token) config.headers.Authorization = `Bearer ${token}`;
        configs.push({ ...config });
        return config;
    };

    const responseHandler = (response: any) => response.data;

    const state = {
        configs,
        _response: { data: {} },
        _reset() { configs.length = 0; },
        setResponse(v: any) { state._response = v; },
    };
    _mockState = state;

    // window.location / localStorage refs captured once (closure-safe)
    const getPath = () => {
        try { return window.location.pathname; } catch { return '/'; }
    };
    const setHref = (v: string) => {
        try { window.location.href = v; } catch { /* noop in fake env */ }
    };

    const makeSpy = (method: string) => {
        const fn = vi.fn(function (url: string, _data?: any, _config?: any) {
            const cfg: any = { url, method, headers: {}, ..._config };
            return Promise.resolve(cfg)
                .then((c) => requestHandler(c))
                .then(() => state._response)
                .then(
                    (res: any) => responseHandler(res),  // success: unwrap response.data
                    (err: any) => {                      // rejection side-effects
                        const status = err?.response?.status;
                        if (status === 401 && getPath() !== '/login') {
                            try { localStorage.removeItem('access_token'); } catch {}
                            try { localStorage.removeItem('refresh_token'); } catch {}
                            setHref('/login');
                        }
                        return Promise.reject(err);
                    }
                );
        });
        return fn;
    };

    return {
        __esModule: true,
        apiGet:    makeSpy('get'),
        apiPost:   makeSpy('post'),
        apiPut:    makeSpy('put'),
        apiPatch:  makeSpy('patch'),
        apiDelete: makeSpy('delete'),
    } as any;
});

import { apiGet, apiPost, apiPut, apiPatch, apiDelete } from './api';

function stubResponse(value: any) { _mockState.setResponse(value); }

const fakeLocation = {
    pathname: '/',
    assign:    vi.fn(),
    replace:   vi.fn(),
    reload:    vi.fn(),
    _hrefValue: 'http://localhost/' as string,
    get href() { return this._hrefValue; },
    set href(v: string) { this._hrefValue = v; },
};

Object.defineProperty(window, 'location', {
    value: fakeLocation,
    writable: true,
    configurable: true,
});

beforeEach(() => {
    vi.restoreAllMocks();
    localStorage.clear();
    fakeLocation.pathname = '/';
    fakeLocation._hrefValue = 'http://localhost/';
    fakeLocation.assign = vi.fn();
    _mockState._reset();
});

describe('request interceptor', () => {
    it('adds Authorization header when access_token exists', async () => {
        localStorage.setItem('access_token', 'tok-123');
        stubResponse({ data: { id: 1 } });
        await apiGet('/users');
        expect(_mockState.configs.length).toBe(1);
        expect(_mockState.configs[0].headers.Authorization).toBe('Bearer tok-123');
    });

    it('does not add Authorization header when no token', async () => {
        stubResponse({ data: {} });
        await apiGet('/users');
        expect(_mockState.configs.length).toBe(1);
        expect(_mockState.configs[0].headers.Authorization).toBeUndefined();
    });
});

describe('response interceptor', () => {
    it('returns response.data on success', async () => {
        stubResponse({ data: { x: 1 } });
        const result = await apiGet('/ok');
        expect(result).toEqual({ x: 1 });
    });

    it('clears tokens and redirects to /login on 401', async () => {
        localStorage.setItem('access_token', 'tok');
        localStorage.setItem('refresh_token', 'ref');
        stubResponse(Promise.reject((() => {
            const err: any = new Error('Unauthorized');
            err.config = { url: '/secure', method: 'get' };
            err.response = { status: 401 };
            return err;
        })()));
        await expect(apiGet('/secure')).rejects.toThrow();
        expect(localStorage.getItem('access_token')).toBeNull();
        expect(localStorage.getItem('refresh_token')).toBeNull();
        expect(fakeLocation._hrefValue).toBe('/login');
    });

    it('does not redirect on non-401 error', async () => {
        stubResponse(Promise.reject((() => {
            const err: any = new Error('Server error');
            err.config = { url: '/fail', method: 'get' };
            err.response = { status: 500 };
            return err;
        })()));
        await expect(apiGet('/fail')).rejects.toThrow();
        expect(fakeLocation._hrefValue).toBe('http://localhost/');
        expect(fakeLocation.assign).not.toHaveBeenCalled();
    });

    it('does not redirect when already on /login', async () => {
        fakeLocation.pathname = '/login';
        stubResponse(Promise.reject((() => {
            const err: any = new Error('Unauthorized');
            err.config = { url: '/login', method: 'get' };
            err.response = { status: 401 };
            return err;
        })()));
        await expect(apiGet('/login')).rejects.toThrow();
        expect(fakeLocation._hrefValue).toBe('http://localhost/');
    });
});

describe('HTTP helpers', () => {
    it('apiGet calls instance.get and unwraps response.data', async () => {
        stubResponse({ data: [1, 2, 3] });
        const result = await apiGet<number[]>('/items');
        expect(result).toEqual([1, 2, 3]);
        expect(apiGet).toHaveBeenCalledWith('/items');
    });

    it('apiPost calls instance.post', async () => {
        stubResponse({ data: { id: 99 } });
        const result = await apiPost('/items', { name: 'new' });
        expect(result).toEqual({ id: 99 });
        expect(apiPost).toHaveBeenCalledWith('/items', { name: 'new' });
    });

    it('apiPut calls instance.put', async () => {
        stubResponse({ data: { id: 1 } });
        const result = await apiPut('/items/1', { name: 'updated' });
        expect(result).toEqual({ id: 1 });
        expect(apiPut).toHaveBeenCalledWith('/items/1', { name: 'updated' });
    });

    it('apiPatch calls instance.patch', async () => {
        stubResponse({ data: { id: 1 } });
        const result = await apiPatch('/items/1', { name: 'patched' });
        expect(result).toEqual({ id: 1 });
        expect(apiPatch).toHaveBeenCalledWith('/items/1', { name: 'patched' });
    });

    it('apiDelete calls instance.delete', async () => {
        stubResponse({ data: null });
        const result = await apiDelete('/items/1');
        expect(result).toBeNull();
        expect(apiDelete).toHaveBeenCalledWith('/items/1');
    });
});
