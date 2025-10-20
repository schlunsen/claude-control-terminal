<template>
  <div v-if="show" class="todo-write-box">
    <div class="todo-box-header">
      <div class="todo-box-icon">📋</div>
      <div class="todo-box-title">Tasks</div>
    </div>
    <div class="todo-list">
      <div
        v-for="(todo, index) in todos"
        :key="index"
        class="todo-item"
        :class="todo.status"
      >
        <div class="todo-status-icon">
          <span v-if="todo.status === 'completed'">✅</span>
          <span v-else-if="todo.status === 'in_progress'" class="in-progress-icon">🔄</span>
          <span v-else>📝</span>
        </div>
        <div class="todo-content">
          <div class="todo-text">{{ todo.content }}</div>
          <div v-if="todo.activeForm && todo.status === 'in_progress'" class="todo-active-form">
            {{ todo.activeForm }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Todo {
  content: string
  status: 'pending' | 'in_progress' | 'completed'
  activeForm?: string
}

interface Props {
  show: boolean
  todos: Todo[]
}

defineProps<Props>()
</script>

<style scoped>
.todo-write-box {
  position: absolute;
  top: 16px;
  right: 16px;
  width: 320px;
  max-height: 400px;
  background: var(--card-bg);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.15);
  z-index: 100;
  animation: fadeIn 0.3s ease-out;
  overflow: hidden;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.todo-box-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: linear-gradient(135deg, var(--accent-purple), var(--accent-purple-hover));
  color: white;
  border-bottom: 1px solid var(--border-color);
}

.todo-box-icon {
  font-size: 1.2rem;
}

.todo-box-title {
  font-size: 0.9rem;
  font-weight: 600;
}

.todo-list {
  max-height: 320px;
  overflow-y: auto;
  padding: 8px;
}

.todo-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 8px 12px;
  margin-bottom: 6px;
  background: var(--bg-secondary);
  border-radius: 8px;
  transition: all 0.2s ease;
  animation: slideInRight 0.3s ease-out;
}

@keyframes slideInRight {
  from {
    opacity: 0;
    transform: translateX(20px);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

.todo-item:hover {
  background: var(--bg-tertiary);
  transform: translateX(-2px);
}

.todo-item.completed {
  opacity: 0.7;
}

.todo-item.completed .todo-text {
  text-decoration: line-through;
  color: var(--text-secondary);
}

.todo-status-icon {
  flex-shrink: 0;
  margin-top: 2px;
}

.in-progress-icon {
  animation: spin 2s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.todo-content {
  flex: 1;
  min-width: 0;
}

.todo-text {
  font-size: 0.9rem;
  color: var(--text-primary);
  line-height: 1.4;
  word-wrap: break-word;
}

.todo-active-form {
  font-size: 0.85rem;
  color: var(--accent-purple);
  font-style: italic;
  margin-top: 2px;
}

@media (max-width: 768px) {
  .todo-write-box {
    width: calc(100% - 32px);
    left: 16px;
    right: 16px;
  }
}
</style>
