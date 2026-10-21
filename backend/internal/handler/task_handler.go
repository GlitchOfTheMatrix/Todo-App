package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/GlitchOfTheMatrix/Todo-App/backend/internal/middleware"
	"github.com/GlitchOfTheMatrix/Todo-App/backend/internal/models"
	"github.com/GlitchOfTheMatrix/Todo-App/backend/internal/repository"
	"github.com/google/uuid"
)

type TaskHandler struct {
	taskRepo *repository.TaskRepository
}

func NewTaskHandler(taskRepo *repository.TaskRepository) *TaskHandler {
	return &TaskHandler{taskRepo: taskRepo}
}

// GetAll returns all tasks for the authenticated user.
// GET /api/tasks
func (h *TaskHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	userId := getUserIDFromContext(r)

	tasks, err := h.taskRepo.GetAllByUserId(r.Context(), userId)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch tasks")
		return
	}

	writeJSON(w, http.StatusOK, tasks)
}

// GetOne returns a single task by ID.
// GET /api/tasks/{id}
func (h *TaskHandler) GetOne(w http.ResponseWriter, r *http.Request) {
	userId := getUserIDFromContext(r)
	taskId, err := parseTaskID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	task, err := h.taskRepo.GetById(r.Context(), taskId, userId)
	if err != nil {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}

	writeJSON(w, http.StatusOK, task)
}

// Create creates a new task for the authenticated user.
// POST /api/tasks
func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	userId := getUserIDFromContext(r)

	var req models.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Task = strings.TrimSpace(req.Task)
	if req.Task == "" {
		writeError(w, http.StatusBadRequest, "task is required")
		return
	}

	task := &models.Task{
		UserId: userId,
		Task:   req.Task,
	}

	if err := h.taskRepo.Create(r.Context(), task); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create task")
		return
	}

	writeJSON(w, http.StatusCreated, task)
}

// Update modifies an existing task.
// PATCH /api/tasks/{id}
func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	userId := getUserIDFromContext(r)
	taskId, err := parseTaskID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	var req models.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate status if provided
	if req.Status != nil {
		status := strings.TrimSpace(*req.Status)
		if status != "pending" && status != "completed" {
			writeError(w, http.StatusBadRequest, "status must be 'pending' or 'completed'")
			return
		}
		req.Status = &status
	}

	if req.Task != nil {
		task := strings.TrimSpace(*req.Task)
		if task == "" {
			writeError(w, http.StatusBadRequest, "task cannot be empty")
			return
		}
		req.Task = &task
	}

	task, err := h.taskRepo.Update(r.Context(), taskId, userId, &req)
	if err != nil {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}

	writeJSON(w, http.StatusOK, task)
}

// Delete removes a task.
// DELETE /api/tasks/{id}
func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userId := getUserIDFromContext(r)
	taskId, err := parseTaskID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	if err := h.taskRepo.Delete(r.Context(), taskId, userId); err != nil {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "task deleted"})
}

// parseTaskID extracts the task UUID from the URL path.
func parseTaskID(r *http.Request) (uuid.UUID, error) {
	idStr := r.PathValue("id")
	return uuid.Parse(idStr)
}

// getUserIDFromContext extracts the authenticated user's UUID from context.
func getUserIDFromContext(r *http.Request) uuid.UUID {
	return middleware.GetUserID(r.Context())
}
