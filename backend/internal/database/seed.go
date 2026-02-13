package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func RunSeed(ctx context.Context, db *DB) error {
	if db == nil || db.pool == nil {
		return fmt.Errorf("database is not initialized")
	}

	tx, err := db.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin seed tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `INSERT INTO tests (id, type, duration_minutes, total_marks)
		VALUES ('11111111-1111-1111-1111-111111111111','diagnostic',20,10)
		ON CONFLICT (id) DO NOTHING`); err != nil {
		return fmt.Errorf("seed tests: %w", err)
	}

	if _, err := tx.Exec(ctx, `INSERT INTO topics (id, name) VALUES
		('22222222-2222-2222-2222-222222222221','Arrays'),
		('22222222-2222-2222-2222-222222222222','Strings'),
		('22222222-2222-2222-2222-222222222223','Trees')
		ON CONFLICT (id) DO NOTHING`); err != nil {
		return fmt.Errorf("seed topics: %w", err)
	}

	if _, err := tx.Exec(ctx, `INSERT INTO subtopics (id, topic_id, title, order_index) VALUES
		('33333333-3333-3333-3333-333333333331','22222222-2222-2222-2222-222222222221','Array basics',1),
		('33333333-3333-3333-3333-333333333332','22222222-2222-2222-2222-222222222222','String basics',1),
		('33333333-3333-3333-3333-333333333333','22222222-2222-2222-2222-222222222223','Tree basics',1)
		ON CONFLICT (id) DO NOTHING`); err != nil {
		return fmt.Errorf("seed subtopics: %w", err)
	}

	if _, err := tx.Exec(ctx, `INSERT INTO learning_questions (id, topic_id, source, difficulty, link) VALUES
		('44444444-4444-4444-4444-444444444441','22222222-2222-2222-2222-222222222221','leetcode','easy','https://leetcode.com/problems/two-sum'),
		('44444444-4444-4444-4444-444444444442','22222222-2222-2222-2222-222222222223','gfg','easy','https://www.geeksforgeeks.org/tree-traversals-inorder-preorder-and-postorder/')
		ON CONFLICT (id) DO NOTHING`); err != nil {
		return fmt.Errorf("seed learning_questions: %w", err)
	}

	if _, err := tx.Exec(ctx, `INSERT INTO questions (id, test_id, question_type, content, options, correct_option, marks, order_index) VALUES
		('55555555-5555-5555-5555-555555555551','11111111-1111-1111-1111-111111111111','slide','Arrays: contiguous memory and index access.',NULL,NULL,0,1),
		('55555555-5555-5555-5555-555555555552','11111111-1111-1111-1111-111111111111','slide','Strings: common matching patterns and complexity.',NULL,NULL,0,2),
		('55555555-5555-5555-5555-555555555553','11111111-1111-1111-1111-111111111111','slide','Trees: recursion + traversal fundamentals.',NULL,NULL,0,3),
		('55555555-5555-5555-5555-555555555554','11111111-1111-1111-1111-111111111111','mcq','Best complexity for array index lookup?', '["O(n)","O(1)","O(log n)","O(n log n)"]'::jsonb,2,2,10),
		('55555555-5555-5555-5555-555555555555','11111111-1111-1111-1111-111111111111','mcq','Queue-based traversal of a tree is?', '["Inorder","Preorder","Level-order","Postorder"]'::jsonb,3,2,11),
		('55555555-5555-5555-5555-555555555556','11111111-1111-1111-1111-111111111111','coding','Write a function to reverse an array.', NULL,NULL,6,12)
		ON CONFLICT (id) DO NOTHING`); err != nil {
		return fmt.Errorf("seed questions: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit seed: %w", err)
	}

	return nil
}
