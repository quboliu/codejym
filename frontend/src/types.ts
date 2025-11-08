export interface Asset {
  id: string;
  name: string;
  sizeBytes: number;
  fileCount: number;
  createdAt: string;
  updatedAt: string;
  sourceName: string;
}

export interface FileNode {
  name: string;
  path: string;
  isDir: boolean;
  children?: FileNode[];
}

export interface FileContent {
  name: string;
  path: string;
  language: string;
  content: string;
}

export interface Session {
  id: string;
  assetId: string;
  relPath: string;
  cursor: number;
  errors: number;
  durationSeconds: number;
  createdAt: string;
  updatedAt: string;
}
