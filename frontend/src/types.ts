export interface Asset {
  id: string;
  userId: string;
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
  userId: string;
  assetId: string;
  relPath: string;
  cursor: number;
  errors: number;
  durationSeconds: number;
  createdAt: string;
  updatedAt: string;
}

export interface FillInTemplate {
  id: string;
  difficulty: 'easy' | 'medium' | 'hard';
  intent: string;
  generationMethod: 'model' | 'fallback';
  provider: string;
  model: string;
  status: string;
}

export type FillInBlankStatus = 'empty' | 'incorrect' | 'correct' | 'revealed';

export interface FillInBlank {
  id: string;
  startOffset: number;
  endOffset: number;
  lineStart: number;
  lineEnd: number;
  kind: string;
  hint?: string;
  status: FillInBlankStatus;
  currentInput: string;
  errorCount: number;
  revealed: boolean;
  answer?: string;
}

export interface FillInSession {
  id: string;
  status: 'in_progress' | 'completed';
  completionOutcome: '' | 'independent_completion' | 'assisted_completion';
  completedBlanks: number;
  totalBlanks: number;
}

export interface FillInPractice {
  template: FillInTemplate;
  source: FileContent;
  blanks: FillInBlank[];
  session: FillInSession;
}

export interface FillInAnswerResult {
  blankId: string;
  correct: boolean;
  status: FillInBlankStatus;
  errorCount: number;
  sessionStatus: FillInSession['status'];
  outcome: FillInSession['completionOutcome'];
}

export interface FillInRevealResult {
  blankId: string;
  answer: string;
  status: FillInBlankStatus;
  sessionStatus: FillInSession['status'];
  outcome: FillInSession['completionOutcome'];
}

export interface ModelConfig {
  provider: string;
  model: string;
  baseUrl: string;
  keyHint: string;
  hasKey: boolean;
  sourceAccessEnabled: boolean;
  usingDevelopmentKey: boolean;
}

export interface User {
  id: string;
  email: string;
  name: string;
  createdAt: string;
  updatedAt: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}
