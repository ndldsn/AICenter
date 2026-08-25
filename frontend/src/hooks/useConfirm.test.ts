import { describe, it, expect, beforeEach } from 'vitest';
import { renderHook } from '@testing-library/react';
import { useConfirm } from './useConfirm';

describe('useConfirm', () => {
    beforeEach(() => {
        // jsdom setup is sufficient — Modal renders into document.body.
    });

    it('returns confirm function and ConfirmModal component', () => {
        const { result } = renderHook(() => useConfirm());
        expect(typeof result.current.confirm).toBe('function');
        expect(typeof result.current.ConfirmModal).toBe('function');
    });

    it('confirm() returns a pending promise', () => {
        const { result } = renderHook(() => useConfirm());
        const p = result.current.confirm({ content: 'Are you sure?' });
        expect(p).toBeInstanceOf(Promise);

        let settled = false;
        p.then(() => { settled = true; });

        expect(settled).toBe(false);

    });
});
