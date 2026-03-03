#!/usr/bin/env node
/**
 * Comprehensive backend API test runner (single JS file).
 *
 * Runs:
 * - public health checks
 * - auth guard + invalid-token checks
 * - authenticated profile checks
 * - diagnostic flow (+ edge cases)
 * - course/dashboard gated checks
 * - profile platform connection CRUD
 * - platform sync trigger/job/overview
 * - IDE run/submit/status edge checks
 *
 * Usage:
 * API_BASE=http://localhost:8080 TEST_JWT=<supabase_access_token> node backend/test.js
 */

const API_BASE = process.env.API_BASE || 'http://localhost:8080';
const TEST_JWT = process.env.TEST_JWT || 'eyJhbGciOiJFUzI1NiIsImtpZCI6ImQ3ZjMxYjliLTM3ZjEtNDI3Ni05ZmMxLTRhZDE2NTdkNDM2NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJodHRwczovL2plZ3dvem1wYWlqcm55Zm9ucmllLnN1cGFiYXNlLmNvL2F1dGgvdjEiLCJzdWIiOiJjN2Q0ZDNmNC03MmVhLTRkY2UtODExZi1mNTkxOWE2MTg0MDgiLCJhdWQiOiJhdXRoZW50aWNhdGVkIiwiZXhwIjoxNzcyMzYzNzA3LCJpYXQiOjE3NzIzNjAxMDcsImVtYWlsIjoidGVzdEBleGFtcGxlLmNvbSIsInBob25lIjoiIiwiYXBwX21ldGFkYXRhIjp7InByb3ZpZGVyIjoiZW1haWwiLCJwcm92aWRlcnMiOlsiZW1haWwiXX0sInVzZXJfbWV0YWRhdGEiOnsiZW1haWxfdmVyaWZpZWQiOnRydWV9LCJyb2xlIjoiYXV0aGVudGljYXRlZCIsImFhbCI6ImFhbDEiLCJhbXIiOlt7Im1ldGhvZCI6InBhc3N3b3JkIiwidGltZXN0YW1wIjoxNzcyMzYwMTA3fV0sInNlc3Npb25faWQiOiJkNzZhNDNiMC03ZmMwLTQ2OGYtYWI3YS1hOTZiNjIxNWRlMWUiLCJpc19hbm9ueW1vdXMiOmZhbHNlfQ.d5J7YbZ7RDKKTLJc4D0iElNv8ijZ3r_3oahOQ0p6tRwellyq9P3KH6E_Uu-uMaW-lDGqmicFyQUvrVyF6efM2Q';
const TEST_TOPIC_IDS = (process.env.TEST_TOPIC_IDS || '22222222-2222-2222-2222-222222222221,22222222-2222-2222-2222-222222222222')
  .split(',')
  .map((s) => s.trim())
  .filter(Boolean);
const IDE_QUESTION_ID = process.env.IDE_QUESTION_ID || '55555555-5555-5555-5555-555555555556';

console.log('\n==============================');
console.log('ArnoCodes backend test runner');
console.log('API_BASE =', API_BASE);
console.log('JWT present =', !!TEST_JWT);
console.log('JWT length =', TEST_JWT?.length);
console.log('==============================\n');

function decodeJWT(token) {
  try {
    const [h, p] = token.split('.');
    return {
      header: JSON.parse(Buffer.from(h, 'base64url').toString()),
      payload: JSON.parse(Buffer.from(p, 'base64url').toString())
    };
  } catch {
    return null;
  }
}

const decoded = decodeJWT(TEST_JWT);
if (decoded) {
  console.log('--- JWT HEADER ---');
  console.log(decoded.header);
  console.log('--- JWT PAYLOAD ---');
  console.log(decoded.payload);
  console.log('-------------------\n');
}

function section(title) {
  console.log(`\n=== ${title} ===`);
}

function ok(msg) {
  console.log(`✅ ${msg}`);
}

function warn(msg) {
  console.log(`⚠️ ${msg}`);
}

function fail(msg, body) {
  console.error(`❌ ${msg}`);
  if (body !== undefined) {
    try {
      console.error(typeof body === 'string' ? body : JSON.stringify(body, null, 2));
    } catch {
      console.error(body);
    }
  }
  process.exit(1);
}

function sleep(ms) {
  return new Promise((r) => setTimeout(r, ms));
}

async function api(method, path, { token, body, headers } = {}) {
  const h = { ...(headers || {}) };

  if (body !== undefined) h['Content-Type'] = 'application/json';
  if (token) h.Authorization = `Bearer ${token}`;

  console.log('\n================ REQUEST ================');
  console.log('URL:', `${API_BASE}${path}`);
  console.log('Method:', method);
  console.log('Headers:', h);
  if (body !== undefined) console.log('Body:', body);

  const res = await fetch(`${API_BASE}${path}`, {
    method,
    headers: h,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });

  const text = await res.text();

  console.log('================ RESPONSE ================');
  console.log('Status:', res.status);
  console.log('Response Headers:', Object.fromEntries(res.headers.entries()));
  console.log('Body:', text);

  if (text.includes('No API key found in request')) {
    console.log('\n🔥 RESPONSE IS FROM SUPABASE REST (not your Go backend)');
  }

  let json = null;
  try {
    json = text ? JSON.parse(text) : null;
  } catch {}

  return { status: res.status, text, json };
}

function expectStatus(resp, allowed, label) {
  if (!allowed.includes(resp.status)) {
    fail(`${label}: expected [${allowed.join(', ')}], got ${resp.status}`, resp.text || resp.json);
  }
  ok(`${label}: HTTP ${resp.status}`);
}

async function unauthenticatedChecks() {
  section('Public + auth guard checks');

  const h1 = await api('GET', '/health');
  expectStatus(h1, [200], 'GET /health');

  const h2 = await api('GET', '/api/v1/health');
  expectStatus(h2, [200], 'GET /api/v1/health');

  const p1 = await api('GET', '/api/v1/profiles/me/status');
  expectStatus(p1, [401], 'GET /api/v1/profiles/me/status without token');

  const p2 = await api('GET', '/api/v1/profiles/me/status', { token: 'invalid.token.value' });
  expectStatus(p2, [401], 'GET /api/v1/profiles/me/status with invalid token');

  const d1 = await api('POST', '/api/v1/diagnostic/start', {
    body: { selected_topics: TEST_TOPIC_IDS },
  });
  expectStatus(d1, [401], 'POST /api/v1/diagnostic/start without token');
}

async function authenticatedFlow() {
  if (!TEST_JWT) {
    warn('TEST_JWT not provided. Skipping authenticated flow.');
    return;
  }

  section('Authenticated profile checks + edge cases');
  const me = await api('GET', '/api/v1/profiles/me/status', { token: TEST_JWT });
  expectStatus(me, [200], 'GET /api/v1/profiles/me/status');

  const badMethod = await api('PUT', '/api/v1/profiles/me/status', { token: TEST_JWT });
  expectStatus(badMethod, [405], 'PUT /api/v1/profiles/me/status method guard');

  section('Diagnostic edge cases + flow');

  const badStartBody = await api('POST', '/api/v1/diagnostic/start', {
    token: TEST_JWT,
    body: { selected_topics: [] },
  });
  expectStatus(badStartBody, [422], 'POST /api/v1/diagnostic/start invalid topic payload');

  const start = await api('POST', '/api/v1/diagnostic/start', {
    token: TEST_JWT,
    body: { selected_topics: TEST_TOPIC_IDS },
  });

  if (start.status === 403) {
    warn('Diagnostic start returned 403 (possibly DIAGNOSTIC_BLOCKED for this user). Continuing gated/read checks.');
  } else {
    expectStatus(start, [201], 'POST /api/v1/diagnostic/start');
  }

  let attemptId = start.json?.data?.attempt_id;

  if (attemptId) {
    ok(`Diagnostic attempt_id: ${attemptId}`);

    const next = await api('GET', `/api/v1/diagnostic/${attemptId}/next`, { token: TEST_JWT });
    expectStatus(next, [200, 404], 'GET /api/v1/diagnostic/{attempt}/next');

    const badAnswer = await api('POST', `/api/v1/diagnostic/${attemptId}/answer`, {
      token: TEST_JWT,
      body: { question_id: '', question_type: 'mcq', selected_option: 1 },
    });
    expectStatus(badAnswer, [422], 'POST /api/v1/diagnostic/{attempt}/answer invalid payload');

    const status = await api('GET', `/api/v1/diagnostic/${attemptId}/status`, { token: TEST_JWT });
    expectStatus(status, [200], 'GET /api/v1/diagnostic/{attempt}/status');

    const submit = await api('POST', `/api/v1/diagnostic/${attemptId}/submit`, { token: TEST_JWT });
    expectStatus(submit, [202, 403], 'POST /api/v1/diagnostic/{attempt}/submit');
  }

  section('Course + dashboard guarded reads');

  const course = await api('GET', '/api/v1/course', { token: TEST_JWT });
  expectStatus(course, [200, 403], 'GET /api/v1/course');

  const dashboard = await api('GET', '/api/v1/dashboard/summary', { token: TEST_JWT });
  expectStatus(dashboard, [200, 403], 'GET /api/v1/dashboard/summary');

  section('Profile platform connection CRUD + sync');

  const listBefore = await api('GET', '/api/v1/profiles/me/platform-connections', { token: TEST_JWT });
  expectStatus(listBefore, [200], 'GET /api/v1/profiles/me/platform-connections');

  const invalidPlatform = await api('POST', '/api/v1/profiles/me/platform-connections', {
    token: TEST_JWT,
    body: { platform: 'unknown', handle: 'abc' },
  });
  expectStatus(invalidPlatform, [422], 'POST /api/v1/profiles/me/platform-connections invalid platform');

  const connect = await api('POST', '/api/v1/profiles/me/platform-connections', {
    token: TEST_JWT,
    body: { platform: 'leetcode', handle: `sample_${Date.now()}` },
  });
  expectStatus(connect, [202], 'POST /api/v1/profiles/me/platform-connections leetcode');

  const trigger = await api('POST', '/api/v1/platform-sync/trigger', { token: TEST_JWT });
  expectStatus(trigger, [202, 503], 'POST /api/v1/platform-sync/trigger');

  const overview = await api('GET', '/api/v1/platform-sync/overview', { token: TEST_JWT });
  expectStatus(overview, [200], 'GET /api/v1/platform-sync/overview');

  const jobId = trigger.json?.data?.job_id;
  if (jobId) {
    const job = await api('GET', `/api/v1/platform-sync/jobs/${jobId}`, { token: TEST_JWT });
    expectStatus(job, [200], 'GET /api/v1/platform-sync/jobs/{jobId}');
  } else {
    warn('No job_id returned from trigger; skipped job details check.');
  }

  const disconnect = await api('DELETE', '/api/v1/profiles/me/platform-connections/leetcode', { token: TEST_JWT });
  expectStatus(disconnect, [202], 'DELETE /api/v1/profiles/me/platform-connections/leetcode');

  section('IDE endpoints + edge cases');

  const ideMissing = await api('POST', '/api/v1/ide/submit', {
    token: TEST_JWT,
    body: { question_id: IDE_QUESTION_ID, code: '', language: 'python' },
  });
  expectStatus(ideMissing, [422], 'POST /api/v1/ide/submit missing code');

  const run = await api('POST', '/api/v1/ide/run', {
    token: TEST_JWT,
    body: { question_id: IDE_QUESTION_ID, code: 'print(1)', language: 'python' },
  });
  expectStatus(run, [200, 404], 'POST /api/v1/ide/run');

  const submit = await api('POST', '/api/v1/ide/submit', {
    token: TEST_JWT,
    body: { question_id: IDE_QUESTION_ID, code: 'print(1)', language: 'python' },
  });
  expectStatus(submit, [202, 404], 'POST /api/v1/ide/submit');

  const sid = submit.json?.data?.submission_id;
  if (sid) {
    let terminal = false;
    for (let i = 0; i < 15; i++) {
      const st = await api('GET', `/api/v1/ide/status?id=${encodeURIComponent(sid)}`, { token: TEST_JWT });
      expectStatus(st, [200], `GET /api/v1/ide/status poll #${i + 1}`);
      const evalStatus = st.json?.data?.evaluation_status;
      if (['completed', 'failed'].includes(evalStatus)) {
        ok(`IDE submission terminal status: ${evalStatus}`);
        terminal = true;
        break;
      }
      await sleep(1500);
    }
    if (!terminal) {
      warn('IDE submission did not reach terminal status within poll window.');
    }
  } else {
    warn('No submission_id returned from IDE submit; polling skipped.');
  }
}

async function main() {
  await unauthenticatedChecks();
  await authenticatedFlow();

  console.log('\n🎉 All selected checks completed.');
}

main().catch((err) => fail(err?.message || String(err)));