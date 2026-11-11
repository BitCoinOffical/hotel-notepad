import type {
  CreateGuestRequest,
  Guest,
  LoginRequest,
  RegisterRequest,
  UpdateGuestRequest,
} from './types';

const AUTH_BASE = '/api/v1/auth';
const GUESTBOOK_BASE = '/api/v1/guestbook';

export class ApiError extends Error {
  constructor(
    message: string,
    public readonly status: number,
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

function readErrorMessage(payload: unknown, fallback: string): string {
  if (typeof payload === 'string' && payload.trim()) {
    return payload;
  }

  if (payload && typeof payload === 'object') {
    const body = payload as Record<string, unknown>;
    for (const key of ['error', 'message', 'msg']) {
      const value = body[key];
      if (typeof value === 'string' && value.trim()) {
        return value;
      }
    }
  }

  return fallback;
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers);
  if (init.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }

  let response: Response;
  try {
    response = await fetch(path, {
      ...init,
      headers,
      credentials: 'include',
    });
  } catch {
    throw new ApiError('Не удалось подключиться к серверу. Проверь, запущен ли backend.', 0);
  }

  const rawBody = response.status === 204 ? '' : await response.text();
  let payload: unknown = undefined;

  if (rawBody) {
    try {
      payload = JSON.parse(rawBody) as unknown;
    } catch {
      payload = rawBody;
    }
  }

  if (!response.ok) {
    throw new ApiError(
      readErrorMessage(payload, `Ошибка запроса: HTTP ${response.status}`),
      response.status,
    );
  }

  return payload as T;
}

export const authApi = {
  login(data: LoginRequest): Promise<void> {
    return request<void>(`${AUTH_BASE}/login`, {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },

  register(data: RegisterRequest): Promise<void> {
    return request<void>(`${AUTH_BASE}/register`, {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },

  logout(): Promise<void> {
    return request<void>(`${AUTH_BASE}/logout`, { method: 'GET' });
  },
};

export const guestbookApi = {
  async getEntries(): Promise<Guest[]> {
    const result = await request<Guest[] | null>(`${GUESTBOOK_BASE}/entries`);
    return result ?? [];
  },

  createEntry(data: CreateGuestRequest): Promise<void> {
    return request<void>(`${GUESTBOOK_BASE}/entries`, {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },

  updateEntry(id: string, data: UpdateGuestRequest): Promise<Guest> {
    return request<Guest>(`${GUESTBOOK_BASE}/entries/${encodeURIComponent(id)}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    });
  },

  deleteEntry(id: string): Promise<void> {
    return request<void>(`${GUESTBOOK_BASE}/entries/${encodeURIComponent(id)}`, {
      method: 'DELETE',
    });
  },
};
