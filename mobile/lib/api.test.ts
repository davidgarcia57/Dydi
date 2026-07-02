/// <reference types="jest" />
import { api } from './api';
import { supabase } from './supabase';

jest.mock('./supabase', () => ({
  supabase: {
    auth: {
      getSession: jest.fn(),
    },
  },
}));

const mockFetch = jest.fn();
global.fetch = mockFetch as any;

describe('api retry policy', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    (supabase.auth.getSession as jest.Mock).mockResolvedValue({
      data: { session: { access_token: 'test_token' } },
    });
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  it('should retry a safe request (GET) automatically on 502', async () => {
    mockFetch
      .mockResolvedValueOnce({ ok: false, status: 502, text: async () => 'Bad Gateway' })
      .mockResolvedValueOnce({ ok: true, status: 200, text: async () => JSON.stringify({ success: true }) });

    const result = await api('/test-path', { method: 'GET' });
    expect(result).toEqual({ success: true });
    expect(mockFetch).toHaveBeenCalledTimes(2);
  });

  it('should not retry a mutating request (POST) automatically on 502', async () => {
    mockFetch.mockResolvedValueOnce({ ok: false, status: 502, text: async () => 'Bad Gateway' });

    await expect(api('/test-path', { method: 'POST' })).rejects.toMatchObject({ status: 502 });
    expect(mockFetch).toHaveBeenCalledTimes(1); // Should only be called once, no retry
  });

  it('should not retry a mutating request on network abort/transient error', async () => {
    const abortErr = new Error('AbortError');
    abortErr.name = 'AbortError';
    mockFetch.mockRejectedValueOnce(abortErr);

    await expect(api('/test-path', { method: 'POST' })).rejects.toThrow('AbortError');
    expect(mockFetch).toHaveBeenCalledTimes(1);
  });

  it('should clear timeout correctly', async () => {
    jest.useFakeTimers();
    const clearTimeoutSpy = jest.spyOn(global, 'clearTimeout');
    
    mockFetch.mockResolvedValueOnce({ ok: true, status: 200, text: async () => JSON.stringify({ success: true }) });

    await api('/test-path', { method: 'GET' });
    
    expect(clearTimeoutSpy).toHaveBeenCalled();
  });
});
