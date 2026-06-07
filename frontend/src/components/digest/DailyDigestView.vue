<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import {
  PhArrowClockwise,
  PhCalendarBlank,
  PhNewspaper,
  PhSparkle,
  PhSpinnerGap,
} from '@phosphor-icons/vue';
import type { DailyDigest } from '@/types/models';

interface Props {
  isSidebarOpen?: boolean;
}

defineProps<Props>();

defineEmits<{
  'toggle-sidebar': [];
}>();

const digests = ref<DailyDigest[]>([]);
const selectedDigest = ref<DailyDigest | null>(null);
const isLoading = ref(false);
const isGenerating = ref(false);
const errorMessage = ref('');

const digestLines = computed(() => {
  if (!selectedDigest.value?.content) return [];
  return selectedDigest.value.content
    .split('\n')
    .map((line) => line.trimEnd())
    .filter((line) => line.length > 0);
});

const selectedDateLabel = computed(() => {
  if (!selectedDigest.value?.digest_date) return '';
  return formatDate(selectedDigest.value.digest_date);
});

onMounted(() => {
  loadDigests();
});

async function loadDigests(): Promise<void> {
  isLoading.value = true;
  errorMessage.value = '';

  try {
    const response = await fetch('/api/agent/digests?limit=60');
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`);
    }

    const data: DailyDigest[] = (await response.json()) || [];
    digests.value = data;

    if (data.length === 0) {
      selectedDigest.value = null;
      return;
    }

    const latestResponse = await fetch('/api/agent/digest/latest');
    if (latestResponse.ok) {
      const latest = await latestResponse.json();
      selectedDigest.value = latest?.id ? latest : data[0];
    } else {
      selectedDigest.value = data[0];
    }
  } catch (error) {
    console.error('Failed to load daily digests:', error);
    errorMessage.value = '加载每日简报失败';
  } finally {
    isLoading.value = false;
  }
}

async function generateTodayDigest(): Promise<void> {
  if (isGenerating.value) return;

  isGenerating.value = true;
  errorMessage.value = '';

  try {
    const response = await fetch('/api/agent/digest/generate', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ force: true }),
    });

    if (!response.ok) {
      const message = await response.text();
      throw new Error(message || `HTTP ${response.status}`);
    }

    const digest: DailyDigest = await response.json();
    selectedDigest.value = digest;
    await loadDigests();
    selectedDigest.value = digest;
    window.showToast?.('今日简报已生成', 'success');
  } catch (error) {
    console.error('Failed to generate daily digest:', error);
    errorMessage.value = '生成每日简报失败，请检查 AI 配置和额度';
    window.showToast?.('生成每日简报失败', 'error');
  } finally {
    isGenerating.value = false;
  }
}

function selectDigest(digest: DailyDigest): void {
  selectedDigest.value = digest;
}

function formatDate(value: string): string {
  if (!value) return '';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    weekday: 'short',
  });
}

function formatTime(value: string): string {
  if (!value) return '';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return '';
  return date.toLocaleString('zh-CN', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}
</script>

<template>
  <div class="digest-view flex h-full flex-1 overflow-hidden bg-bg-primary">
    <section class="digest-list border-r border-border bg-bg-secondary">
      <header class="digest-list-header border-b border-border">
        <div>
          <h2 class="text-xl font-semibold text-text-primary">每日简报</h2>
          <p class="mt-1 text-sm text-text-secondary">历史简报和当天摘要</p>
        </div>
        <button
          class="digest-icon-btn"
          :disabled="isLoading"
          aria-label="刷新简报列表"
          @click="loadDigests"
        >
          <PhArrowClockwise :size="20" :class="{ 'animate-spin': isLoading }" />
        </button>
      </header>

      <div v-if="isLoading && digests.length === 0" class="digest-empty">
        <PhSpinnerGap :size="28" class="animate-spin text-accent" />
        <span>正在加载简报...</span>
      </div>

      <div v-else-if="digests.length === 0" class="digest-empty px-6 text-center">
        <PhNewspaper :size="36" class="text-text-tertiary" />
        <div>
          <p class="font-medium text-text-primary">还没有生成过简报</p>
          <p class="mt-1 text-sm text-text-secondary">生成一次后，这里会保存历史记录。</p>
        </div>
        <button class="digest-primary-btn" :disabled="isGenerating" @click="generateTodayDigest">
          <PhSparkle :size="18" :class="{ 'animate-spin': isGenerating }" />
          {{ isGenerating ? '生成中...' : '生成今日简报' }}
        </button>
      </div>

      <div v-else class="digest-history">
        <button
          v-for="digest in digests"
          :key="digest.id"
          class="digest-history-item"
          :class="{ active: selectedDigest?.id === digest.id }"
          @click="selectDigest(digest)"
        >
          <div class="flex items-start justify-between gap-3">
            <h3 class="line-clamp-2 text-left font-semibold text-text-primary">
              {{ digest.title || '每日简报' }}
            </h3>
            <span class="digest-count">{{ digest.article_count || 0 }}</span>
          </div>
          <div class="mt-3 flex items-center gap-2 text-xs text-text-secondary">
            <PhCalendarBlank :size="14" />
            <span>{{ formatDate(digest.digest_date) }}</span>
          </div>
          <p class="mt-2 line-clamp-2 text-left text-sm text-text-secondary">
            {{ digest.content }}
          </p>
        </button>
      </div>
    </section>

    <main class="digest-detail flex-1 overflow-y-auto">
      <div class="digest-detail-inner">
        <div class="digest-detail-toolbar">
          <div>
            <p class="text-sm text-text-secondary">{{ selectedDateLabel }}</p>
            <h1 class="mt-2 text-3xl font-bold text-text-primary">
              {{ selectedDigest?.title || '每日简报' }}
            </h1>
          </div>
          <button class="digest-primary-btn" :disabled="isGenerating" @click="generateTodayDigest">
            <PhSparkle :size="18" :class="{ 'animate-spin': isGenerating }" />
            {{ isGenerating ? '生成中...' : '生成今日简报' }}
          </button>
        </div>

        <div v-if="errorMessage" class="digest-error">
          {{ errorMessage }}
        </div>

        <div v-if="!selectedDigest && !isLoading" class="digest-detail-empty">
          <PhNewspaper :size="44" class="text-text-tertiary" />
          <p class="text-lg font-semibold text-text-primary">选择一份历史简报</p>
          <p class="text-text-secondary">或者生成今天的个性化 RSS 简报。</p>
        </div>

        <article v-else-if="selectedDigest" class="digest-content">
          <div class="digest-meta">
            <span>{{ selectedDigest.article_count || 0 }} 篇文章</span>
            <span v-if="selectedDigest.model">{{ selectedDigest.model }}</span>
            <span v-if="selectedDigest.generated_at">
              {{ formatTime(selectedDigest.generated_at) }} 生成
            </span>
          </div>

          <div class="digest-prose">
            <p v-for="(line, index) in digestLines" :key="index">
              {{ line }}
            </p>
          </div>

          <section v-if="selectedDigest.articles?.length" class="digest-articles">
            <h2 class="text-xl font-semibold text-text-primary">提到的文章</h2>
            <a
              v-for="article in selectedDigest.articles"
              :key="article.id"
              class="digest-article-card"
              :href="article.url"
              target="_blank"
              rel="noopener noreferrer"
            >
              <div class="flex items-start justify-between gap-4">
                <div>
                  <h3 class="font-semibold text-text-primary">{{ article.title }}</h3>
                  <p class="mt-1 text-sm text-accent">{{ article.feed_title }}</p>
                </div>
                <span class="digest-score">{{ article.relevance_score }}</span>
              </div>
              <p v-if="article.summary" class="mt-3 text-sm text-text-secondary">
                {{ article.summary }}
              </p>
              <p v-if="article.recommendation" class="mt-2 text-sm text-text-primary">
                {{ article.recommendation }}
              </p>
            </a>
          </section>
        </article>
      </div>
    </main>
  </div>
</template>

<style scoped>
.digest-view {
  min-width: 0;
}

.digest-list {
  width: var(--article-list-width, 380px);
  min-width: 280px;
  max-width: 520px;
  display: flex;
  flex-direction: column;
}

.digest-list-header {
  min-height: 80px;
  padding: 18px 18px 14px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.digest-icon-btn {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-secondary);
  transition:
    color 0.15s ease,
    background-color 0.15s ease;
}

.digest-icon-btn:hover:not(:disabled) {
  color: var(--color-accent);
  background: var(--color-bg-tertiary);
}

.digest-history {
  overflow-y: auto;
  padding: 8px;
}

.digest-history-item {
  width: 100%;
  border-radius: 8px;
  padding: 14px;
  margin-bottom: 6px;
  color: inherit;
  background: transparent;
  transition:
    background-color 0.15s ease,
    border-color 0.15s ease;
}

.digest-history-item:hover,
.digest-history-item.active {
  background: var(--color-bg-tertiary);
}

.digest-history-item.active {
  box-shadow: inset 3px 0 0 var(--color-accent);
}

.digest-count,
.digest-score {
  min-width: 24px;
  height: 24px;
  padding: 0 7px;
  border-radius: 999px;
  background: var(--color-bg-tertiary);
  color: var(--color-text-secondary);
  font-size: 12px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.digest-detail {
  min-width: 0;
}

.digest-detail-inner {
  max-width: 980px;
  margin: 0 auto;
  padding: 48px 56px 80px;
}

.digest-detail-toolbar {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 24px;
  margin-bottom: 28px;
}

.digest-primary-btn {
  min-height: 40px;
  padding: 0 14px;
  border-radius: 8px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  background: var(--color-accent);
  color: white;
  font-weight: 600;
  white-space: nowrap;
  transition:
    opacity 0.15s ease,
    transform 0.15s ease;
}

.digest-primary-btn:hover:not(:disabled) {
  transform: translateY(-1px);
}

.digest-primary-btn:disabled {
  opacity: 0.65;
  cursor: not-allowed;
}

.digest-empty,
.digest-detail-empty {
  height: 100%;
  min-height: 280px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 14px;
  color: var(--color-text-secondary);
}

.digest-error {
  margin-bottom: 20px;
  padding: 12px 14px;
  border: 1px solid rgba(239, 68, 68, 0.3);
  border-radius: 8px;
  color: #dc2626;
  background: rgba(239, 68, 68, 0.08);
}

.digest-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-bottom: 24px;
  color: var(--color-text-secondary);
  font-size: 14px;
}

.digest-meta span {
  padding: 5px 9px;
  border-radius: 999px;
  background: var(--color-bg-secondary);
}

.digest-prose {
  font-size: 17px;
  line-height: 1.9;
  color: var(--color-text-primary);
}

.digest-prose p + p {
  margin-top: 14px;
}

.digest-articles {
  margin-top: 40px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.digest-article-card {
  display: block;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 16px;
  background: var(--color-bg-secondary);
  text-decoration: none;
  transition:
    border-color 0.15s ease,
    transform 0.15s ease;
}

.digest-article-card:hover {
  border-color: var(--color-accent);
  transform: translateY(-1px);
}

@media (max-width: 900px) {
  .digest-view {
    flex-direction: column;
  }

  .digest-list {
    width: 100%;
    max-width: none;
    height: 260px;
    border-right: 0;
    border-bottom: 1px solid var(--color-border);
  }

  .digest-detail-inner {
    padding: 32px 24px 72px;
  }

  .digest-detail-toolbar {
    flex-direction: column;
  }
}
</style>
