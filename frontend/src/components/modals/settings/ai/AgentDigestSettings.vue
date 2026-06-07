<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import {
  PhBrain,
  PhClock,
  PhListNumbers,
  PhNote,
  PhProhibit,
  PhRobot,
  PhSparkle,
  PhTarget,
} from '@phosphor-icons/vue';
import {
  ButtonControl,
  NestedSettingsContainer,
  NumberControl,
  SettingGroup,
  SettingWithToggle,
  SubSettingItem,
  TextAreaControl,
} from '@/components/settings';
import type { SettingsData } from '@/types/settings';

interface DailyDigest {
  id: number;
  digest_date: string;
  title: string;
  content: string;
  article_count: number;
  generated_at: string;
  notified_at?: string;
}

interface Props {
  settings: SettingsData;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  'update:settings': [settings: SettingsData];
}>();

const isGenerating = ref(false);
const latestDigest = ref<DailyDigest | null>(null);

const digestSummary = computed(() => {
  if (!latestDigest.value) return '还没有生成过每日简报';
  return `${latestDigest.value.digest_date} · ${latestDigest.value.article_count} 篇文章`;
});

function updateSetting(key: keyof SettingsData, value: any) {
  emit('update:settings', {
    ...props.settings,
    [key]: value,
  });
}

function updateDigestTime(event: Event) {
  updateSetting('agent_digest_time', (event.target as HTMLInputElement).value);
}

async function fetchLatestDigest() {
  try {
    const response = await fetch('/api/agent/digest/latest');
    if (!response.ok) return;
    const data = await response.json();
    latestDigest.value = Object.prototype.hasOwnProperty.call(data, 'digest') ? data.digest : data;
  } catch (error) {
    console.error('Failed to fetch latest daily digest:', error);
  }
}

async function generateDigest() {
  isGenerating.value = true;
  try {
    const response = await fetch('/api/agent/digest/generate', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ force: true }),
    });

    if (!response.ok) {
      const text = await response.text();
      throw new Error(text || 'generate failed');
    }

    latestDigest.value = await response.json();
    window.showToast('每日 AI 简报已生成', 'success');
  } catch (error) {
    console.error('Failed to generate daily digest:', error);
    const message = error instanceof Error ? error.message : String(error);
    window.showToast(`每日 AI 简报生成失败：${message}`, 'error', 6000);
  } finally {
    isGenerating.value = false;
  }
}

onMounted(fetchLatestDigest);
</script>

<template>
  <SettingGroup :icon="PhBrain" title="RSS Agent">
    <SettingWithToggle
      :icon="PhRobot"
      title="每日 AI 简报"
      description="每天到点后总结新增文章，并根据你的记忆给出有用性排序和建议"
      :model-value="props.settings.agent_digest_enabled"
      @update:model-value="updateSetting('agent_digest_enabled', $event)"
    />

    <NestedSettingsContainer v-if="props.settings.agent_digest_enabled">
      <SubSettingItem
        :icon="PhClock"
        title="生成时间"
        description="每天首次到达这个时间后生成今日简报"
      >
        <input
          class="time-input"
          type="time"
          :value="props.settings.agent_digest_time || '08:30'"
          @input="updateDigestTime"
        />
      </SubSettingItem>

      <SubSettingItem
        :icon="PhListNumbers"
        title="最多处理文章数"
        description="控制每日 LLM 消耗；按发布时间从旧到新依次总结"
      >
        <NumberControl
          :model-value="props.settings.agent_digest_max_articles"
          :min="1"
          :max="100"
          :step="1"
          suffix="篇"
          @update:model-value="updateSetting('agent_digest_max_articles', $event)"
        />
      </SubSettingItem>

      <SubSettingItem
        :icon="PhSparkle"
        title="立即生成"
        :description="digestSummary"
      >
        <ButtonControl
          type="primary"
          label="生成"
          :icon="PhSparkle"
          :loading="isGenerating"
          :disabled="isGenerating"
          @click="generateDigest"
        />
      </SubSettingItem>
    </NestedSettingsContainer>

    <NestedSettingsContainer>
      <div class="sub-setting-item-col">
        <div class="flex items-start gap-3">
          <PhTarget :size="22" class="text-text-secondary mt-1 shrink-0" />
          <div class="min-w-0 flex-1">
            <div class="font-medium text-sm">我感兴趣的方向</div>
            <div class="text-xs text-text-secondary mt-1">
              例如：LLM 推理、AI Agent、训练系统、工程实践、投资研究方法
            </div>
          </div>
        </div>
        <TextAreaControl
          :model-value="props.settings.agent_memory_interests"
          :rows="4"
          resize
          placeholder="写下你希望 AI 优先关注的主题、问题、人物、公司、研究方向..."
          @update:model-value="updateSetting('agent_memory_interests', $event)"
        />
      </div>

      <div class="sub-setting-item-col">
        <div class="flex items-start gap-3">
          <PhProhibit :size="22" class="text-text-secondary mt-1 shrink-0" />
          <div class="min-w-0 flex-1">
            <div class="font-medium text-sm">我不感兴趣的内容</div>
            <div class="text-xs text-text-secondary mt-1">
              AI 会在每日简报中降低这些内容的优先级，而不是简单按相关性推荐
            </div>
          </div>
        </div>
        <TextAreaControl
          :model-value="props.settings.agent_memory_dislikes"
          :rows="3"
          resize
          placeholder="写下你不想反复看到的主题、噪音来源、太浅或太营销的内容..."
          @update:model-value="updateSetting('agent_memory_dislikes', $event)"
        />
      </div>

      <div class="sub-setting-item-col">
        <div class="flex items-start gap-3">
          <PhNote :size="22" class="text-text-secondary mt-1 shrink-0" />
          <div class="min-w-0 flex-1">
            <div class="font-medium text-sm">其他记忆</div>
            <div class="text-xs text-text-secondary mt-1">
              可以写当前目标、正在做的项目、偏好的输出风格，或希望 AI 扮演的角色
            </div>
          </div>
        </div>
        <TextAreaControl
          :model-value="props.settings.agent_memory_notes"
          :rows="4"
          resize
          placeholder="例如：我现在在改造 MrRSS，希望每天知道哪些文章值得深入读、哪些只需略过..."
          @update:model-value="updateSetting('agent_memory_notes', $event)"
        />
      </div>
    </NestedSettingsContainer>

    <div v-if="latestDigest?.content" class="latest-digest">
      <div class="latest-digest-title">{{ latestDigest.title }}</div>
      <pre>{{ latestDigest.content }}</pre>
    </div>
  </SettingGroup>
</template>

<style scoped>
.time-input {
  @apply w-32 p-1.5 sm:p-2.5 border border-border rounded-md bg-bg-secondary text-text-primary focus:border-accent focus:outline-none transition-colors text-xs sm:text-sm;
}

.latest-digest {
  @apply rounded-lg border border-border bg-bg-secondary p-3 sm:p-4;
}

.latest-digest-title {
  @apply text-sm font-semibold mb-2 text-text-primary;
}

.latest-digest pre {
  @apply whitespace-pre-wrap text-xs sm:text-sm leading-relaxed text-text-secondary font-sans max-h-80 overflow-auto;
}
</style>
