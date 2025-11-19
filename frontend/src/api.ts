import type { Asset, AuthResponse, FileContent, FileNode, Session, User } from './types';

const API_BASE = import.meta.env.VITE_API_BASE ?? '';

let authToken: string | null = null;

export function setAuthToken(token: string | null) {
  authToken = token;
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const isForm = init?.body instanceof FormData;
  const finalHeaders = new Headers(init?.headers);
  if (!isForm && !finalHeaders.has('Content-Type')) {
    finalHeaders.set('Content-Type', 'application/json');
  }
  if (authToken) {
    finalHeaders.set('Authorization', `Bearer ${authToken}`);
  }
  const response = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers: finalHeaders,
  });
  if (!response.ok) {
    const message = await extractError(response);
    throw new Error(message);
  }
  if (response.status === 204) {
    return null as T;
  }
  return (await response.json()) as T;
}

async function extractError(response: Response) {
  try {
    const data = await response.json();
    if (data?.error) {
      return data.error as string;
    }
  } catch {
    /* ignore */
  }
  return `Request failed with status ${response.status}`;
}

export function listAssets() {
  return request<Asset[]>('/api/assets');
}

export function createAsset(name: string) {
  return request<Asset>('/api/assets', {
    method: 'POST',
    body: JSON.stringify({ name }),
  });
}

export function signup(email: string, password: string, name: string) {
  return request<AuthResponse>('/api/auth/signup', {
    method: 'POST',
    body: JSON.stringify({ email, password, name }),
  });
}

export function login(email: string, password: string) {
  return request<AuthResponse>('/api/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  });
}

export function fetchCurrentUser() {
  return request<User>('/api/auth/me');
}

export function uploadAsset(file: File) {
  const formData = new FormData();
  formData.append('file', file);
  return request<Asset>('/api/assets/upload', {
    method: 'POST',
    body: formData,
    headers: {}, // let browser set boundary
  });
}

export function uploadFileToAsset(assetId: string, file: File) {
  const formData = new FormData();
  formData.append('file', file);
  return request<Asset>(`/api/assets/${assetId}/upload`, {
    method: 'POST',
    body: formData,
    headers: {}, // let browser set boundary
  });
}

export function uploadPastedAsset(filename: string, content: string) {
  return request<Asset>('/api/assets/paste', {
    method: 'POST',
    body: JSON.stringify({ filename, content }),
  });
}

export function uploadPasteToAsset(assetId: string, filename: string, content: string) {
  return request<Asset>(`/api/assets/${assetId}/paste`, {
    method: 'POST',
    body: JSON.stringify({ filename, content }),
  });
}

export function fetchFileTree(assetId: string) {
  return request<FileNode[]>(`/api/assets/${assetId}/tree`);
}

export function fetchFileContent(assetId: string, filePath: string) {
  const encoded = encodeURIComponent(filePath);
  return request<FileContent>(`/api/assets/${assetId}/file?path=${encoded}`);
}

export function deleteAsset(assetId: string) {
  return request<void>(`/api/assets/${assetId}`, { method: 'DELETE' });
}

export function renameAsset(assetId: string, name: string) {
  return request<Asset>(`/api/assets/${assetId}/rename`, {
    method: 'POST',
    body: JSON.stringify({ name }),
  });
}

export function createDirectory(assetId: string, path: string) {
  return request<{ message: string }>(`/api/assets/${assetId}/mkdir`, {
    method: 'POST',
    body: JSON.stringify({ path }),
  });
}

export function moveFile(assetId: string, from: string, to: string) {
  return request<{ message: string }>(`/api/assets/${assetId}/move-file`, {
    method: 'POST',
    body: JSON.stringify({ from, to }),
  });
}

export function renameFile(assetId: string, path: string, newName: string) {
  return request<{ message: string }>(`/api/assets/${assetId}/rename-file`, {
    method: 'POST',
    body: JSON.stringify({ path, newName }),
  });
}

export function deleteFile(assetId: string, path: string) {
  const encoded = encodeURIComponent(path);
  return request<void>(`/api/assets/${assetId}/delete-file?path=${encoded}`, {
    method: 'DELETE'
  });
}

export function createSession(assetId: string, filePath: string) {
  return request<Session>('/api/sessions', {
    method: 'POST',
    body: JSON.stringify({ assetId, path: filePath }),
  });
}

export function querySession(assetId: string, filePath: string) {
  const encoded = encodeURIComponent(filePath);
  return request<Session>(`/api/sessions?assetId=${assetId}&path=${encoded}`);
}

export function fetchSession(sessionId: string) {
  return request<Session>(`/api/sessions/${sessionId}`);
}

export function patchSession(sessionId: string, payload: Partial<Pick<Session, 'cursor' | 'errors' | 'durationSeconds'>>) {
  return request<Session>(`/api/sessions/${sessionId}`, {
    method: 'PATCH',
    body: JSON.stringify(payload),
  });
}
