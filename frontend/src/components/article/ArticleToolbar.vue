<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { useSettings } from '@/composables/core/useSettings';
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import {
  PhArrowLeft,
  PhX,
  PhCaretDown,
  PhCheck,
  PhGlobe,
  PhArticle,
  PhEnvelopeOpen,
  PhEnvelope,
  PhStar,
  PhClockCountdown,
  PhArrowSquareOut,
  PhTranslate,
} from '@phosphor-icons/vue';
import type { Article } from '@/types/models';
import type { TranslationDisplayMode } from '@/types/translation';

const { t } = useI18n();
const { settings, fetchSettings } = useSettings();

onMounted(async () => {
  try {
    await fetchSettings();
  } catch (e) {
    console.error('Error loading settings:', e);
  }
});

interface Props {
  article: Article;
  showContent: boolean;
  translationMode?: TranslationDisplayMode;
  isModal?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  translationMode: 'original',
  isModal: false,
});

const emit = defineEmits<{
  close: [];
  toggleContentView: [];
  toggleRead: [];
  toggleFavorite: [];
  toggleReadLater: [];
  openOriginal: [];
  setTranslationMode: [mode: TranslationDisplayMode];
  exportToObsidian: [];
  exportToNotion: [];
  exportToZotero: [];
}>();

const isTranslationMenuOpen = ref(false);
const translationMenuRef = ref<HTMLElement | null>(null);

const translationOptions = computed(() => [
  {
    mode: 'original' as const,
    label: t('article.translation.modeOriginal'),
    description: t('article.translation.modeOriginalDesc'),
  },
  {
    mode: 'bilingual' as const,
    label: t('article.translation.modeBilingual'),
    description: t('article.translation.modeBilingualDesc'),
  },
  {
    mode: 'translated' as const,
    label: t('article.translation.modeTranslated'),
    description: t('article.translation.modeTranslatedDesc'),
  },
]);

const translationTooltip = computed(() => {
  const option = translationOptions.value.find((item) => item.mode === props.translationMode);
  return option
    ? t('article.translation.translateButtonWithMode', { mode: option.label })
    : t('article.translation.translateButton');
});

function toggleTranslationMenu() {
  isTranslationMenuOpen.value = !isTranslationMenuOpen.value;
}

function selectTranslationMode(mode: TranslationDisplayMode) {
  emit('setTranslationMode', mode);
  isTranslationMenuOpen.value = false;
}

function handleDocumentClick(event: MouseEvent) {
  if (!translationMenuRef.value?.contains(event.target as Node)) {
    isTranslationMenuOpen.value = false;
  }
}

function handleTranslationSettingsChanged() {
  fetchSettings().catch((e) => {
    console.error('Error refreshing translation settings:', e);
  });
}

onMounted(() => {
  document.addEventListener('click', handleDocumentClick);
  window.addEventListener('translation-settings-changed', handleTranslationSettingsChanged);
});

onBeforeUnmount(() => {
  document.removeEventListener('click', handleDocumentClick);
  window.removeEventListener('translation-settings-changed', handleTranslationSettingsChanged);
});
</script>

<template>
  <div
    class="p-2 sm:p-4 border-b border-border flex justify-between items-center bg-bg-primary shrink-0"
  >
    <!-- Modal mode: X button always visible -->
    <button
      v-if="isModal"
      class="flex items-center gap-1.5 sm:gap-2 text-text-secondary hover:text-text-primary text-sm sm:text-base"
      :aria-label="t('common.close')"
      @click="$emit('close')"
    >
      <PhX :size="20" class="sm:w-5 sm:h-5" />
    </button>
    <!-- Normal mode: Back button on mobile -->
    <button
      v-else
      class="md:hidden flex items-center gap-1.5 sm:gap-2 text-text-secondary hover:text-text-primary text-sm sm:text-base"
      @click="$emit('close')"
    >
      <PhArrowLeft :size="18" class="sm:w-5 sm:h-5" />
      <span class="hidden xs:inline">{{ t('common.back') }}</span>
    </button>
    <div class="flex gap-1 sm:gap-2 ml-auto">
      <button
        class="action-btn"
        :aria-label="
          showContent ? t('article.action.viewOriginal') : t('article.action.viewContent')
        "
        :data-tooltip="
          showContent ? t('article.action.viewOriginal') : t('article.action.viewContent')
        "
        @click="$emit('toggleContentView')"
      >
        <PhGlobe v-if="showContent" :size="18" class="sm:w-5 sm:h-5" />
        <PhArticle v-else :size="18" class="sm:w-5 sm:h-5" />
      </button>
      <div
        v-if="showContent && settings.translation_enabled"
        ref="translationMenuRef"
        class="relative"
      >
        <button
          :class="[
            'action-btn',
            translationMode !== 'original' ? 'text-accent hover:text-accent' : '',
          ]"
          :aria-label="translationTooltip"
          :aria-expanded="isTranslationMenuOpen"
          :data-tooltip="translationTooltip"
          @click.stop="toggleTranslationMenu"
        >
          <PhTranslate
            :size="18"
            class="sm:w-5 sm:h-5"
            :weight="translationMode !== 'original' ? 'fill' : 'regular'"
          />
          <PhCaretDown :size="10" class="-ml-1 hidden sm:block" />
        </button>

        <Transition name="translation-menu">
          <div
            v-if="isTranslationMenuOpen"
            class="translation-menu absolute right-0 top-full z-50 mt-2 w-56 overflow-hidden rounded-md border border-border bg-bg-primary shadow-lg"
            @click.stop
          >
            <button
              v-for="option in translationOptions"
              :key="option.mode"
              class="translation-menu-item"
              :class="{ active: option.mode === translationMode }"
              @click="selectTranslationMode(option.mode)"
            >
              <span class="flex h-4 w-4 items-center justify-center">
                <PhCheck v-if="option.mode === translationMode" :size="14" />
              </span>
              <span class="min-w-0 flex-1 text-left">
                <span class="block text-sm font-medium">{{ option.label }}</span>
                <span class="block truncate text-xs text-text-secondary">{{
                  option.description
                }}</span>
              </span>
            </button>
          </div>
        </Transition>
      </div>
      <button
        class="action-btn"
        :aria-label="
          article.is_read ? t('article.action.markAsUnread') : t('article.action.markAsRead')
        "
        :data-tooltip="
          article.is_read ? t('article.action.markAsUnread') : t('article.action.markAsRead')
        "
        @click="$emit('toggleRead')"
      >
        <PhEnvelopeOpen v-if="article.is_read" :size="18" class="sm:w-5 sm:h-5" />
        <PhEnvelope v-else :size="18" class="sm:w-5 sm:h-5" />
      </button>
      <button
        :class="[
          'action-btn',
          article.is_favorite ? 'text-yellow-500 hover:text-yellow-600' : 'hover:text-yellow-500',
        ]"
        :aria-label="
          article.is_favorite
            ? t('article.action.removeFromFavorite')
            : t('article.toolbar.addToFavorite')
        "
        :data-tooltip="
          article.is_favorite
            ? t('article.action.removeFromFavorite')
            : t('article.toolbar.addToFavorite')
        "
        @click="$emit('toggleFavorite')"
      >
        <PhStar
          :size="18"
          class="sm:w-5 sm:h-5"
          :weight="article.is_favorite ? 'fill' : 'regular'"
        />
      </button>
      <button
        :class="[
          'action-btn',
          article.is_read_later ? 'text-blue-500 hover:text-blue-600' : 'hover:text-blue-500',
        ]"
        :aria-label="
          article.is_read_later
            ? t('article.action.removeFromReadLater')
            : t('article.toolbar.addToReadLater')
        "
        :data-tooltip="
          article.is_read_later
            ? t('article.action.removeFromReadLater')
            : t('article.toolbar.addToReadLater')
        "
        @click="$emit('toggleReadLater')"
      >
        <PhClockCountdown
          :size="18"
          class="sm:w-5 sm:h-5"
          :weight="article.is_read_later ? 'fill' : 'regular'"
        />
      </button>
      <button
        class="action-btn"
        :aria-label="t('article.action.openInBrowser')"
        :data-tooltip="t('article.action.openInBrowser')"
        @click="$emit('openOriginal')"
      >
        <PhArrowSquareOut :size="18" class="sm:w-5 sm:h-5" />
      </button>
      <button
        v-if="settings.obsidian_enabled"
        class="action-btn"
        :aria-label="t('setting.plugins.obsidian.exportTo')"
        :data-tooltip="t('setting.plugins.obsidian.exportTo')"
        @click="$emit('exportToObsidian')"
      >
        <img
          src="/assets/plugin_icons/obsidian.svg"
          class="w-[18px] h-[18px] sm:w-5 sm:h-5"
          alt="Obsidian"
        />
      </button>
      <button
        v-if="settings.notion_enabled"
        class="action-btn"
        :aria-label="t('setting.plugins.notion.exportTo')"
        :data-tooltip="t('setting.plugins.notion.exportTo')"
        @click="$emit('exportToNotion')"
      >
        <img
          src="/assets/plugin_icons/notion.svg"
          class="w-[18px] h-[18px] sm:w-5 sm:h-5"
          alt="Notion"
        />
      </button>
      <button
        v-if="settings.zotero_enabled"
        class="action-btn"
        :aria-label="t('setting.plugins.zotero.exportTo')"
        :data-tooltip="t('setting.plugins.zotero.exportTo')"
        @click="$emit('exportToZotero')"
      >
        <img
          src="/assets/plugin_icons/zotero.png"
          class="w-[18px] h-[18px] sm:w-5 sm:h-5"
          alt="Zotero"
        />
      </button>
    </div>
  </div>
</template>

<style scoped>
.action-btn {
  @apply relative flex items-center justify-center text-lg sm:text-xl cursor-pointer text-text-secondary p-1 sm:p-1.5 rounded-md transition-colors hover:bg-bg-tertiary hover:text-text-primary;
}

.action-btn[data-tooltip]::after {
  content: attr(data-tooltip);
  position: absolute;
  top: calc(100% + 6px);
  right: 50%;
  transform: translateX(50%) translateY(-2px);
  z-index: 60;
  max-width: 220px;
  padding: 0.35rem 0.5rem;
  border-radius: 0.375rem;
  background: var(--text-primary);
  color: var(--bg-primary);
  font-size: 0.75rem;
  line-height: 1.2;
  white-space: nowrap;
  opacity: 0;
  pointer-events: none;
  transition:
    opacity 80ms ease,
    transform 80ms ease;
}

.action-btn[data-tooltip]:hover::after,
.action-btn[data-tooltip]:focus-visible::after {
  opacity: 1;
  transform: translateX(50%) translateY(0);
}

.translation-menu-item {
  @apply flex w-full items-start gap-2 px-3 py-2 text-left text-text-primary transition-colors hover:bg-bg-tertiary;
}

.translation-menu-item.active {
  @apply bg-bg-secondary text-accent;
}

.translation-menu-enter-active,
.translation-menu-leave-active {
  transition:
    opacity 100ms ease,
    transform 100ms ease;
}

.translation-menu-enter-from,
.translation-menu-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
