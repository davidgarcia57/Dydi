import { supabase } from './supabase';

const BASE = process.env.EXPO_PUBLIC_API_URL || 'https://dydi-25hj.onrender.com';

const delay = (ms: number) => new Promise((res) => setTimeout(res, ms));

const MAX_RETRIES = 3;
const PER_ATTEMPT_TIMEOUT = 30_000; // ms

export async function api(path: string, options: RequestInit = {}, retries = MAX_RETRIES): Promise<any> {
  let lastErr: any = null;

  for (let i = 0; i < retries; i++) {
    const controller = new AbortController();
    const timer = setTimeout(() => controller.abort(), PER_ATTEMPT_TIMEOUT);
    
    try {
      // Fetch session on every request to guarantee the freshest token
      const { data: { session } } = await supabase.auth.getSession();
      const token = session?.access_token ?? '';

      const headers: Record<string, string> = {
        'Content-Type': 'application/json',
        ...(options.headers as Record<string, string> | undefined),
      };

      if (token) {
        headers['Authorization'] = `Bearer ${token}`;
      }

      const res = await fetch(`${BASE}${path}`, {
        ...options,
        signal: controller.signal,
        headers,
      });

      const text = await res.text();
      let body: any = null;
      try {
        body = text ? JSON.parse(text) : null;
      } catch {
        body = { message: text }; // Tolerate non-JSON response bodies
      }

      if (!res.ok) {
        const err = { status: res.status, ...(body || {}) };
        
        // Render free-tier cold-start triggers 502/503/504 which are worth retrying.
        if (res.status >= 502 && res.status <= 504 && i < retries - 1) {
          lastErr = err;
          await delay(Math.min(1000 * Math.pow(2, i), 8000));
          continue;
        }
        throw err;
      }

      return body;
    } catch (err: any) {
      // Retry transient transport / abort failures
      const isTransient = err instanceof TypeError || err?.name === 'AbortError';
      if (!isTransient || i === retries - 1) throw err;
      lastErr = err;
      await delay(Math.min(1000 * Math.pow(2, i), 8000));
    } finally {
      clearTimeout(timer);
    }
  }
  
  throw lastErr;
}
