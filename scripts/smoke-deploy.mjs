#!/usr/bin/env node

const args = new Map();
for (let i = 2; i < process.argv.length; i += 1) {
  const arg = process.argv[i];
  if (arg.startsWith('--')) {
    args.set(arg.slice(2), process.argv[i + 1]);
    i += 1;
  }
}

const baseUrl = (args.get('base-url') || process.env.BASE_URL || 'http://127.0.0.1:8080').replace(/\/$/, '');
const email = args.get('email') || `smoke-${Date.now()}@codejym.test`;
const password = args.get('password') || 'demo1234';
let token = '';

async function request(path, options = {}) {
  const headers = new Headers(options.headers || {});
  if (!(options.body instanceof FormData) && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }
  if (token) {
    headers.set('Authorization', `Bearer ${token}`);
  }
  const response = await fetch(`${baseUrl}${path}`, { ...options, headers });
  const text = await response.text();
  let data = null;
  try {
    data = text ? JSON.parse(text) : null;
  } catch {
    data = text;
  }
  if (!response.ok) {
    const message = data && data.error ? data.error : text || response.statusText;
    const error = new Error(`${options.method || 'GET'} ${path} -> ${response.status}: ${message}`);
    error.status = response.status;
    error.data = data;
    throw error;
  }
  return data;
}

function assert(condition, message) {
  if (!condition) {
    throw new Error(message);
  }
}

function firstFile(nodes) {
  for (const node of nodes) {
    if (!node.isDir) {
      return node;
    }
    const child = firstFile(node.children || []);
    if (child) {
      return child;
    }
  }
  return null;
}

async function main() {
  const health = await request('/healthz');
  assert(health.status === 'ok', 'health endpoint did not report ok');

  const auth = await request('/api/auth/signup', {
    method: 'POST',
    body: JSON.stringify({ email, password, name: 'Smoke Test' }),
  });
  token = auth.token;
  assert(token, 'signup did not return a token');

  const assets = await request('/api/assets');
  assert(Array.isArray(assets) && assets.length > 0, 'new user has no default asset');

  const defaultAsset = assets[0];
  const tree = await request(`/api/assets/${defaultAsset.id}/tree`);
  const goFile = firstFile(tree);
  assert(goFile && goFile.path, 'default asset has no source file');

  const goPractice = await request('/api/fill-in/enter', {
    method: 'POST',
    body: JSON.stringify({ assetId: defaultAsset.id, path: goFile.path }),
  });
  assert(goPractice.session?.id, 'fill-in session is missing');
  assert(goPractice.blanks.length > 0, 'Go source produced no fill-in blanks');
  assert(goPractice.blanks.every((blank) => !blank.answer), 'unrevealed blanks leaked answers');

  const blank = goPractice.blanks[0];
  const wrong = await request(`/api/fill-in/sessions/${goPractice.session.id}/answers/${blank.id}`, {
    method: 'POST',
    body: JSON.stringify({ input: '__wrong_answer__' }),
  });
  assert(wrong.correct === false && wrong.status === 'incorrect', 'wrong answer was not marked incorrect');

  const revealed = await request(`/api/fill-in/sessions/${goPractice.session.id}/blanks/${blank.id}/reveal`, {
    method: 'POST',
  });
  assert(revealed.answer && revealed.status === 'revealed', 'reveal did not return the answer');

  const reset = await request(`/api/fill-in/sessions/${goPractice.session.id}/reset`, {
    method: 'POST',
  });
  assert(reset.completedBlanks === 0 && reset.status === 'in_progress', 'reset did not clear fill-in progress');

  const correct = await request(`/api/fill-in/sessions/${goPractice.session.id}/answers/${blank.id}`, {
    method: 'POST',
    body: JSON.stringify({ input: revealed.answer }),
  });
  assert(correct.correct === true && correct.status === 'correct', 'correct answer was not accepted');

  const rustContent = [
    'pub fn partition(values: &mut [i32]) -> usize {',
    '    let pivot = values[values.len() - 1];',
    '    let mut store_index = 0;',
    '    for scan_index in 0..values.len() - 1 {',
    '        if values[scan_index] <= pivot {',
    '            values.swap(store_index, scan_index);',
    '            store_index += 1;',
    '        }',
    '    }',
    '    values.swap(store_index, values.len() - 1);',
    '    store_index',
    '}',
    '',
  ].join('\n');
  const rustAsset = await request('/api/assets/paste', {
    method: 'POST',
    body: JSON.stringify({ filename: 'partition.rs', content: rustContent }),
  });
  const rustPractice = await request('/api/fill-in/enter', {
    method: 'POST',
    body: JSON.stringify({ assetId: rustAsset.id, path: 'partition.rs' }),
  });
  assert(rustPractice.source.language === 'rust', `Rust detection failed: ${rustPractice.source.language}`);
  assert(rustPractice.blanks.length > 0, 'Rust source produced no fill-in blanks');

  const modelConfig = await request('/api/model-config');
  assert(modelConfig.provider === 'deepseek', 'default model provider is not deepseek');
  const saved = await request('/api/model-config', {
    method: 'POST',
    body: JSON.stringify({
      provider: 'anthropic',
      model: 'claude-3-5-sonnet-latest',
      baseUrl: '',
      apiKey: '',
      sourceAccessEnabled: false,
    }),
  });
  assert(saved.provider === 'anthropic' && saved.sourceAccessEnabled === false, 'model config save failed');
  const deleted = await request('/api/model-config', { method: 'DELETE' });
  assert(deleted.provider === 'deepseek', 'model config delete did not restore default');

  console.log(JSON.stringify({
    ok: true,
    baseUrl,
    email,
    defaultFile: goFile.path,
    goBlanks: goPractice.blanks.length,
    rustBlanks: rustPractice.blanks.length,
    generationMethod: goPractice.template.generationMethod,
  }, null, 2));
}

main().catch((error) => {
  console.error(error.stack || error);
  process.exit(1);
});
