import type { Asset, FileContent, FileNode, Session } from './types';

const API_BASE = import.meta.env.VITE_API_BASE ?? '';

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const isForm = init?.body instanceof FormData;
  const finalHeaders = new Headers(init?.headers);
  if (!isForm && !finalHeaders.has('Content-Type')) {
    finalHeaders.set('Content-Type', 'application/json');
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

export function uploadAsset(file: File) {
  const formData = new FormData();
  formData.append('file', file);
  return request<Asset>('/api/assets/upload', {
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

export function createSession(assetId: string, filePath: string) {
  return request<Session>('/api/sessions', {
    method: 'POST',
    body: JSON.stringify({ assetId, path: filePath }),
  });
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
