import { useCallback, useEffect, useRef, useState } from 'react';
import type { FormEvent } from 'react';
import { tasksApi } from '../api';
import { useAuth } from '../context/AuthContext';
import type { Task } from '../types';
import styles from './Dashboard.module.css';

type FilterType = 'all' | 'pending' | 'completed';

export default function DashboardPage() {
  const { user, logout } = useAuth();
  const [tasks, setTasks] = useState<Task[]>([]);
  const [newTask, setNewTask] = useState('');
  const [filter, setFilter] = useState<FilterType>('all');
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [adding, setAdding] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editText, setEditText] = useState('');
  const [hasMore, setHasMore] = useState(false);
  const [totalTasks, setTotalTasks] = useState(0);
  const offsetRef = useRef(0);
  const PAGE_SIZE = 50;

  const fetchTasks = useCallback(async () => {
    try {
      const res = await tasksApi.getAll(PAGE_SIZE, 0);
      setTasks(res.data);
      setHasMore(res.has_more);
      setTotalTasks(res.total);
      offsetRef.current = res.data.length;
    } catch {
      // silently fail
    } finally {
      setLoading(false);
    }
  }, []);

  const loadMore = useCallback(async () => {
    if (loadingMore || !hasMore) return;
    setLoadingMore(true);
    try {
      const res = await tasksApi.getAll(PAGE_SIZE, offsetRef.current);
      setTasks((prev) => [...prev, ...res.data]);
      setHasMore(res.has_more);
      setTotalTasks(res.total);
      offsetRef.current += res.data.length;
    } catch {
      // silently fail
    } finally {
      setLoadingMore(false);
    }
  }, [loadingMore, hasMore]);

  useEffect(() => {
    fetchTasks();
  }, [fetchTasks]);

  const handleAdd = async (e: FormEvent) => {
    e.preventDefault();
    const text = newTask.trim();
    if (!text) return;

    setAdding(true);
    try {
      const task = await tasksApi.create({ task: text });
      setTasks((prev) => [task, ...prev]);
      setNewTask('');
      setTotalTasks((prev) => prev + 1);
      offsetRef.current += 1;
    } catch {
      // handle error
    } finally {
      setAdding(false);
    }
  };

  const handleToggle = async (task: Task) => {
    const newStatus = task.status === 'pending' ? 'completed' : 'pending';
    try {
      const updated = await tasksApi.update(task.id, { status: newStatus });
      setTasks((prev) => prev.map((t) => (t.id === task.id ? updated : t)));
    } catch {
      // handle error
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await tasksApi.delete(id);
      setTasks((prev) => prev.filter((t) => t.id !== id));
      setTotalTasks((prev) => Math.max(0, prev - 1));
      offsetRef.current = Math.max(0, offsetRef.current - 1);
    } catch {
      // handle error
    }
  };

  const handleEditStart = (task: Task) => {
    setEditingId(task.id);
    setEditText(task.task);
  };

  const handleEditSave = async (id: string) => {
    const text = editText.trim();
    if (!text) return;

    try {
      const updated = await tasksApi.update(id, { task: text });
      setTasks((prev) => prev.map((t) => (t.id === id ? updated : t)));
      setEditingId(null);
    } catch {
      // handle error
    }
  };

  const handleEditCancel = () => {
    setEditingId(null);
    setEditText('');
  };

  const filteredTasks = tasks.filter((t) => {
    if (filter === 'pending') return t.status === 'pending';
    if (filter === 'completed') return t.status === 'completed';
    return true;
  });

  const stats = {
    total: tasks.length,
    completed: tasks.filter((t) => t.status === 'completed').length,
    pending: tasks.filter((t) => t.status === 'pending').length,
  };

  const completionPercent = stats.total > 0 ? Math.round((stats.completed / stats.total) * 100) : 0;

  return (
    <div className={styles.dashboard}>
      {/* Ambient */}
      <div className={`${styles.dashAmbient} ${styles.dashAmbient1}`} />
      <div className={`${styles.dashAmbient} ${styles.dashAmbient2}`} />
      <div className={`${styles.dashAmbient} ${styles.dashAmbient3}`} />

      {/* Header */}
      <header className={styles.dashHeader}>
        <div className={styles.dashHeaderLeft}>
          <div className={styles.dashLogo}>
            <svg width="28" height="28" viewBox="0 0 32 32" fill="none">
              <rect width="32" height="32" rx="8" fill="url(#dash-logo-g)" />
              <path d="M9 16.5L13.5 21L23 11" stroke="white" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" />
              <defs>
                <linearGradient id="dash-logo-g" x1="0" y1="0" x2="32" y2="32">
                  <stop stopColor="#6366f1" />
                  <stop offset="1" stopColor="#8b5cf6" />
                </linearGradient>
              </defs>
            </svg>
            <span>TaskFlow</span>
          </div>
        </div>
        <div className={styles.dashHeaderRight}>
          <div className={styles.userInfo}>
            <div className={styles.userAvatar}>
              {user?.name?.charAt(0).toUpperCase()}
            </div>
            <span className={styles.userName}>{user?.name}</span>
          </div>
          <button className={styles.btnLogout} onClick={logout} id="logout-btn" aria-label="Sign out">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M9 21H5a2 2 0 01-2-2V5a2 2 0 012-2h4" />
              <polyline points="16 17 21 12 16 7" />
              <line x1="21" y1="12" x2="9" y2="12" />
            </svg>
          </button>
        </div>
      </header>

      {/* Main content */}
      <main className={styles.dashMain}>
        {/* Welcome & Stats */}
        <section className={styles.dashWelcome}>
          <div className={styles.welcomeText}>
            <h1>Good {getGreeting()}, <span className={styles.gradientText}>{user?.name?.split(' ')[0]}</span></h1>
            <p>{stats.pending === 0 ? "You're all caught up! 🎉" : `You have ${stats.pending} task${stats.pending === 1 ? '' : 's'} remaining`}</p>
          </div>
        </section>

        {/* Stats cards */}
        <section className={styles.statsRow}>
          <div className={styles.statCard}>
            <div className={`${styles.statIcon} ${styles.statIconTotal}`}>
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z" />
                <polyline points="14 2 14 8 20 8" />
                <line x1="16" y1="13" x2="8" y2="13" />
                <line x1="16" y1="17" x2="8" y2="17" />
              </svg>
            </div>
            <div className={styles.statContent}>
              <span className={styles.statNumber}>{stats.total}</span>
              <span className={styles.statLabel}>Total</span>
            </div>
          </div>

          <div className={styles.statCard}>
            <div className={`${styles.statIcon} ${styles.statIconPending}`}>
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <circle cx="12" cy="12" r="10" />
                <polyline points="12 6 12 12 16 14" />
              </svg>
            </div>
            <div className={styles.statContent}>
              <span className={styles.statNumber}>{stats.pending}</span>
              <span className={styles.statLabel}>Pending</span>
            </div>
          </div>

          <div className={styles.statCard}>
            <div className={`${styles.statIcon} ${styles.statIconDone}`}>
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M22 11.08V12a10 10 0 11-5.93-9.14" />
                <polyline points="22 4 12 14.01 9 11.01" />
              </svg>
            </div>
            <div className={styles.statContent}>
              <span className={styles.statNumber}>{stats.completed}</span>
              <span className={styles.statLabel}>Done</span>
            </div>
          </div>

          <div className={`${styles.statCard} ${styles.statCardProgress}`}>
            <div className={styles.progressRingContainer}>
              <svg className={styles.progressRing} width="52" height="52" viewBox="0 0 52 52">
                <circle className={styles.progressRingBg} cx="26" cy="26" r="22" />
                <circle
                  className={styles.progressRingFill}
                  cx="26" cy="26" r="22"
                  strokeDasharray={`${2 * Math.PI * 22}`}
                  strokeDashoffset={`${2 * Math.PI * 22 * (1 - completionPercent / 100)}`}
                />
              </svg>
              <span className={styles.progressPercent}>{completionPercent}%</span>
            </div>
            <span className={styles.statLabel}>Progress</span>
          </div>
        </section>

        {/* Add task */}
        <form className={styles.addTaskForm} onSubmit={handleAdd} id="add-task-form">
          <div className={styles.addTaskInputWrapper}>
            <svg className={styles.addIcon} width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <circle cx="12" cy="12" r="10" />
              <line x1="12" y1="8" x2="12" y2="16" />
              <line x1="8" y1="12" x2="16" y2="12" />
            </svg>
            <input
              id="new-task-input"
              type="text"
              placeholder="What needs to be done?"
              value={newTask}
              onChange={(e) => setNewTask(e.target.value)}
              disabled={adding}
            />
          </div>
          <button type="submit" className={styles.btnAdd} disabled={adding || !newTask.trim()} id="add-task-btn">
            {adding ? <span className="spinner" /> : 'Add Task'}
          </button>
        </form>

        {/* Filter tabs */}
        <div className={styles.filterTabs}>
          {(['all', 'pending', 'completed'] as FilterType[]).map((f) => (
            <button
              key={f}
              className={`${styles.filterTab} ${filter === f ? styles.active : ''}`}
              onClick={() => setFilter(f)}
              id={`filter-${f}`}
            >
              {f.charAt(0).toUpperCase() + f.slice(1)}
              <span className={styles.filterCount}>
                {f === 'all' ? stats.total : f === 'pending' ? stats.pending : stats.completed}
              </span>
            </button>
          ))}
        </div>

        {/* Task list */}
        <div className={styles.taskList}>
          {loading ? (
            <div className={styles.taskEmpty}>
              <span className="spinner spinner-lg" />
              <p>Loading tasks...</p>
            </div>
          ) : filteredTasks.length === 0 ? (
            <div className={styles.taskEmpty}>
              <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" opacity="0.4">
                <rect x="3" y="3" width="18" height="18" rx="2" ry="2" />
                <line x1="9" y1="9" x2="15" y2="15" />
                <line x1="15" y1="9" x2="9" y2="15" />
              </svg>
              <p>
                {filter === 'all'
                  ? 'No tasks yet. Add your first one above!'
                  : `No ${filter} tasks`}
              </p>
            </div>
          ) : (
            filteredTasks.map((task, index) => (
              <div
                key={task.id}
                className={`${styles.taskItem} ${task.status === 'completed' ? styles.completed : ''}`}
                style={{ animationDelay: `${index * 0.04}s` }}
              >
                <button
                  className={styles.taskCheck}
                  onClick={() => handleToggle(task)}
                  aria-label={task.status === 'completed' ? 'Mark as pending' : 'Mark as completed'}
                  id={`toggle-${task.id}`}
                >
                  <svg width="20" height="20" viewBox="0 0 20 20" fill="none">
                    <circle cx="10" cy="10" r="9" stroke="currentColor" strokeWidth="1.5" />
                    {task.status === 'completed' && (
                      <path d="M6 10l2.5 3L14 7" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
                    )}
                  </svg>
                </button>

                {editingId === task.id ? (
                  <div className={styles.taskEditRow}>
                    <input
                      className={styles.taskEditInput}
                      type="text"
                      value={editText}
                      onChange={(e) => setEditText(e.target.value)}
                      onKeyDown={(e) => {
                        if (e.key === 'Enter') handleEditSave(task.id);
                        if (e.key === 'Escape') handleEditCancel();
                      }}
                      autoFocus
                    />
                    <button className={`${styles.btnIcon} ${styles.btnSave}`} onClick={() => handleEditSave(task.id)} aria-label="Save">
                      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                        <polyline points="20 6 9 17 4 12" />
                      </svg>
                    </button>
                    <button className={`${styles.btnIcon} ${styles.btnCancel}`} onClick={handleEditCancel} aria-label="Cancel">
                      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                        <line x1="18" y1="6" x2="6" y2="18" />
                        <line x1="6" y1="6" x2="18" y2="18" />
                      </svg>
                    </button>
                  </div>
                ) : (
                  <>
                    <span className={styles.taskText}>{task.task}</span>
                    <div className={styles.taskActions}>
                      <button className={`${styles.btnIcon} ${styles.btnEdit}`} onClick={() => handleEditStart(task)} aria-label="Edit" id={`edit-${task.id}`}>
                        <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                          <path d="M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7" />
                          <path d="M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z" />
                        </svg>
                      </button>
                      <button className={`${styles.btnIcon} ${styles.btnDelete}`} onClick={() => handleDelete(task.id)} aria-label="Delete" id={`delete-${task.id}`}>
                        <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                          <polyline points="3 6 5 6 21 6" />
                          <path d="M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2" />
                        </svg>
                      </button>
                    </div>
                  </>
                )}
              </div>
            ))
          )}
          {hasMore && (
            <button
              className={styles.btnLoadMore}
              onClick={loadMore}
              disabled={loadingMore}
              id="load-more-btn"
            >
              {loadingMore ? (
                <span className="spinner" />
              ) : (
                `Load More (${tasks.length} of ${totalTasks})`
              )}
            </button>
          )}
        </div>
      </main>
    </div>
  );
}

function getGreeting(): string {
  const hour = new Date().getHours();
  if (hour < 12) return 'morning';
  if (hour < 17) return 'afternoon';
  return 'evening';
}
