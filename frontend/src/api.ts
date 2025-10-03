import type { AuthResponse, User } from './types';

const API_BASE = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080/api';

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const response = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(options.headers || {}),
    },
  });

  if (!response.ok) {
    let message = 'Request failed';
    try {
      const data = await response.json();
      if (data?.error) {
        message = data.error;
      }
    } catch (error) {
      // ignore parse errors
    }
    throw new Error(message);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return (await response.json()) as T;
}

export async function registerUser(payload: {
  name: string;
  email: string;
  password: string;
}): Promise<AuthResponse> {
  return request<AuthResponse>('/register', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export async function loginUser(payload: { email: string; password: string }): Promise<AuthResponse> {
  return request<AuthResponse>('/login', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export async function fetchProfile(token: string): Promise<{ user: User }> {
  return request<{ user: User }>('/profile', {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
}

export async function fetchUsers(token: string): Promise<{ users: User[] }> {
  return request<{ users: User[] }>('/users', {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
}

export async function updateUser(
  token: string,
  id: number,
  payload: Partial<{ name: string; email: string; password: string }>,
): Promise<{ user: User }> {
  return request<{ user: User }>(`/users/${id}`, {
    method: 'PUT',
    headers: {
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify(payload),
  });
}

export async function deleteUser(token: string, id: number): Promise<void> {
  await request<void>(`/users/${id}`, {
    method: 'DELETE',
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
}
