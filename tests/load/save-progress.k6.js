import http from 'k6/http';
import { check } from 'k6';
import { SharedArray } from 'k6/data';
import { Rate } from 'k6/metrics';

const vus = new SharedArray('vus', () => JSON.parse(open('/tmp/vus.json')));
const patchErr = new Rate('patch_errors');

export const options = {
  discardResponseBodies: true,
  scenarios: {
    saturate: {
      executor: 'ramping-vus',
      startVUs: 5,
      stages: [
        { duration: '8s',  target: 20 },
        { duration: '12s', target: 50 },
        { duration: '12s', target: 100 },
        { duration: '12s', target: 200 },
        { duration: '6s',  target: 0 },
      ],
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<200', 'p(99)<500'],
    patch_errors: ['rate<0.01'],
  },
};

export default function () {
  const v = vus[(__VU - 1) % vus.length];
  const res = http.patch(
    `${__ENV.BASE_URL || "http://localhost:8081"}/api/sessions/${v.sessionId}`,
    JSON.stringify({ cursor: __ITER % 5000, errors: 0, durationSeconds: __ITER % 100000 }),
    { headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${v.token}` } }
  );
  patchErr.add(res.status !== 200);
  check(res, { 'status is 200': (r) => r.status === 200 });
}
