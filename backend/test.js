/**
 * Diagnostic API end-to-end smoke script for local development.
 *
 * Run:
 *   1) Start backend first:
 *      cd backend && go run cmd/api/main.go
 *   2) In another terminal run:
 *      cd backend && node test.js
 *
 * Optional env vars:
 *   API_BASE=http://localhost:8080
 *   USER_ID=aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa
 *
 * IMPORTANT:
 * - USER_ID must exist in auth.users + profiles.
 * - For Supabase SQL editor, execute:
 *     INSERT INTO auth.users(id) VALUES ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa') ON CONFLICT DO NOTHING;
 *     INSERT INTO profiles(id, full_name) VALUES ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Local Test User') ON CONFLICT DO NOTHING;
 */

const API_BASE = process.env.API_BASE || 'http://localhost:8080';
const USER_ID = process.env.USER_ID || 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa';

const TOPIC_IDS = [
  '22222222-2222-2222-2222-222222222221', // Arrays (seed)
  '22222222-2222-2222-2222-222222222222', // Strings (seed)
];

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

async function call(method, path, body) {
  const res = await fetch(`${API_BASE}${path}`, {
    method,
    headers: {
      'Content-Type': 'application/json',
      'X-User-ID': USER_ID,
    },
    body: body ? JSON.stringify(body) : undefined,
  });

  let json = null;
  try {
    json = await res.json();
  } catch {
    json = null;
  }

  return { status: res.status, json };
}

function expect(condition, message) {
  if (!condition) {
    throw new Error(`❌ ${message}`);
  }
  console.log(`✅ ${message}`);
}

function logStep(title) {
  console.log(`\n=== ${title} ===`);
}

async function answerQuestion(attemptID, question) {
  if (question.question_type === 'mcq') {
    const res = await call('POST', `/api/v1/diagnostic/${attemptID}/answer`, {
      question_id: question.question_id,
      question_type: 'mcq',
      selected_option: 1,
    });
    expect(res.status === 202, `MCQ answer accepted for question ${question.question_id}`);
    return;
  }

  if (question.question_type === 'coding') {
    const res = await call('POST', `/api/v1/diagnostic/${attemptID}/coding`, {
      question_id: question.question_id,
      question_type: 'coding',
      language: 'javascript',
      code: 'function solve(arr){ return [...arr].reverse(); }',
    });
    expect(res.status === 202, `Coding submission accepted for question ${question.question_id}`);
    return;
  }

  throw new Error(`❌ Unsupported question type in script: ${question.question_type}`);
}

async function main() {
  logStep('1) Health check');
  const health = await call('GET', '/api/v1/health');
  expect(health.status === 200, 'Health endpoint is up');

  logStep('2) Profile status and guard before diagnostic submit');
  const statusBefore = await call('GET', '/api/v1/profiles/me/status');
  expect(statusBefore.status === 200, 'Profile status fetched');

  const dashboardBefore = await call('GET', '/api/v1/dashboard/summary');
  const allowedBefore = dashboardBefore.status === 200;
  if (allowedBefore) {
    console.log('ℹ️ Dashboard already unlocked for this user (a previous submitted diagnostic exists).');
  } else {
    expect(dashboardBefore.status === 403, 'Dashboard is locked before diagnostic completion');
  }

  logStep('3) Start diagnostic');
  const start = await call('POST', '/api/v1/diagnostic/start', { selected_topics: TOPIC_IDS });
  console.log('Start response:', start);

  if (start.status === 403 && start.json?.error === 'DIAGNOSTIC_BLOCKED') {
    console.log('⚠️ DIAGNOSTIC_BLOCKED: This user has already reached retake limit (2 attempts in 48 hours).');
    console.log('Use a different USER_ID or wait for retake window, then run again.');
    return;
  }

  expect(start.status === 201, 'Diagnostic attempt created');
  const attemptID = start.json?.data?.attempt_id;
  expect(Boolean(attemptID), 'Attempt ID received');
  console.log(`Attempt ID: ${attemptID}`);

  logStep('4) Fetch and answer questions in sequence');
  for (let i = 0; i < 20; i++) {
    const next = await call('GET', `/api/v1/diagnostic/${attemptID}/next`);
    if (next.status === 404 && next.json?.error === 'NOT_FOUND') {
      console.log('ℹ️ No more unanswered questions.');
      break;
    }

    expect(next.status === 200, 'Fetched next question');
    const q = next.json?.data;
    expect(Boolean(q?.question_id), 'Question payload has question_id');
    expect(q.correct_option === undefined, 'Safe question payload (no correct_option exposed)');

    await answerQuestion(attemptID, q);
  }

  logStep('5) Check status before submit');
  const statusMid = await call('GET', `/api/v1/diagnostic/${attemptID}/status`);
  expect(statusMid.status === 200, 'Attempt status fetched');
  console.log('Status payload:', statusMid.json?.data);

  logStep('6) Wait for worker to process coding submissions (optional but useful)');
  console.log('Waiting 12s for polling worker cycle...');
  await sleep(12_000);

  logStep('7) Submit diagnostic');
  const submit = await call('POST', `/api/v1/diagnostic/${attemptID}/submit`);
  expect(submit.status === 202, 'Diagnostic submitted');

  logStep('8) Verify dashboard unlock after submit');
  const dashboardAfter = await call('GET', '/api/v1/dashboard/summary');
  expect(dashboardAfter.status === 200, 'Dashboard unlocked after diagnostic submission');

  const statusAfter = await call('GET', '/api/v1/profiles/me/status');
  expect(statusAfter.status === 200, 'Profile status fetched after submission');
  console.log('Final profile status:', statusAfter.json?.data);

  console.log('\n🎉 Diagnostic API smoke test completed successfully.');
}

main().catch((err) => {
  console.error(err.message || err);
  process.exit(1);
});
