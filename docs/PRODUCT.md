# Product Overview

## Vision
ArnoCodes is a unified DSA learning and practice platform that combines a prerequisite-based learning DAG, cross-platform analytics, streak tracking, and AI-assisted learning tools (chat bot, code helper, and visualizer) into a single, college-friendly experience.

## Core User Flows

### New User (Onboarding)
1. Select known topics from the topic DAG (prerequisite-aware).
2. Receive a diagnostic test generated from the selected topics.
3. Complete the test to determine eligibility for the course structure.
4. Unlock dashboard and learning features based on diagnostic results.

### Returning User (Dashboard + Learning)
- **Dashboard**
  - Student profile
  - Analytics with cross-platform activity
  - Heatmap and streaks
  - Mini DAG visualization showing completion
  - AI-generated pain points
- **Course**
  - Full DAG view
  - Topic-level navigation
  - Subtopic learning flow (theory → practice → guided help → auto-completion)
- **AI Tools**
  - Chat bot (topic-restricted; SLM unlimited + LLM limited)
  - Code helper (structured, interview-style, step-by-step)
  - Code visualizer (interactive animations)

## Key Differentiators
- Prerequisite-aware DAG-driven progression
- Cross-platform analytics (LeetCode, Codeforces, GFG, etc.)
- Streaks and heatmap derived from both internal and external activity
- AI assistance constrained to learning context and topic relevance

## Tech Stack
- **Frontend:** React, Next.js, TypeScript
- **Backend:** Go, PostgreSQL (Supabase)
- **AI:** Python + vector DB (RAG)

