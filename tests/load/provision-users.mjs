const B = process.env.BASE_URL || 'http://localhost:8081';
const N = parseInt(process.argv[2] || '200', 10);
const CONC = 30;
const vus = [];
let next = 0, done = 0;

async function provisionOne(i) {
  const email = `load_${i}_${Date.now()}@test.com`;
  const su = await fetch(`${B}/api/auth/signup`, { method:'POST', headers:{'Content-Type':'application/json'},
    body: JSON.stringify({ email, name:`u${i}`, password:'pw123456' }) });
  if (!su.ok) throw new Error('signup '+su.status);
  const token = (await su.json()).token;
  const auth = { 'Content-Type':'application/json', Authorization:`Bearer ${token}` };
  const pa = await fetch(`${B}/api/assets/paste`, { method:'POST', headers:auth,
    body: JSON.stringify({ filename:'practice.go', content:'package main\nfunc main(){ println("hi") }\n' }) });
  if (!pa.ok) throw new Error('paste '+pa.status);
  const assetId = (await pa.json()).id;
  const se = await fetch(`${B}/api/sessions`, { method:'POST', headers:auth,
    body: JSON.stringify({ assetId, path:'practice.go' }) });
  if (!se.ok) throw new Error('session '+se.status);
  const sessionId = (await se.json()).id;
  vus.push({ token, sessionId });
}

async function worker() {
  while (next < N) {
    const i = next++;
    try { await provisionOne(i); } catch (e) { /* count */ }
    done++;
  }
}
const t0 = Date.now();
await Promise.all(Array.from({length: CONC}, worker));
const fs = await import('fs');
fs.writeFileSync('/tmp/vus.json', JSON.stringify(vus));
console.log(`provisioned ${vus.length}/${N} users in ${((Date.now()-t0)/1000).toFixed(1)}s`);
