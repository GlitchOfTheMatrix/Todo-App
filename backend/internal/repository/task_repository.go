package repository

import (
	"context"
	"fmt"

	"github.com/GlitchOfTheMatrix/Todo-App/backend/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskRepository struct {
	db *pgxpool.Pool
}

func NewTaskRepository(db *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(ctx context.Context, task *models.Task) error {
	query := `
		INSERT INTO todos (user_id, task)
		VALUES ($1, $2)
		RETURNING id, status, created_at, updated_at`

	err := r.db.QueryRow(ctx, query, task.UserId, task.Task).
		Scan(&task.Id, &task.Status, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	return nil
}

func (r *TaskRepository) GetAllByUserId(ctx context.Context, userId uuid.UUID) ([]models.Task, error) {
	query := `
		SELECT id, user_id, task, status, created_at, updated_at
		FROM todos
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		err := rows.Scan(&task.Id, &task.UserId, &task.Task, &task.Status, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	if tasks == nil {
		tasks = []models.Task{}
	}

	return tasks, nil
}

func (r *TaskRepository) GetById(ctx context.Context, id uuid.UUID, userId uuid.UUID) (*models.Task, error) {
	query := `
		SELECT id, user_id, task, status, created_at, updated_at
		FROM todos
		WHERE id = $1 AND user_id = $2`

	task := &models.Task{}
	err := r.db.QueryRow(ctx, query, id, userId).
		Scan(&task.Id, &task.UserId, &task.Task, &task.Status, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return task, nil
}

func (r *TaskRepository) Update(ctx context.Context, id uuid.UUID, userId uuid.UUID, req *models.UpdateTaskRequest) (*models.Task, error) {
	query := `
		UPDATE todos
		SET task = COALESCE($1, task),
		    status = COALESCE($2, status),
		    updated_at = NOW()
		WHERE id = $3 AND user_id = $4
		RETURNING id, user_id, task, status, created_at, updated_at`

	task := &models.Task{}
	err := r.db.QueryRow(ctx, query, req.Task, req.Status, id, userId).
		Scan(&task.Id, &task.UserId, &task.Task, &task.Status, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	return task, nil
}

func (r *TaskRepository) Delete(ctx context.Context, id uuid.UUID, userId uuid.UUID) error {
	query := `DELETE FROM todos WHERE id = $1 AND user_id = $2`

	result, err := r.db.Exec(ctx, query, id, userId)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}
