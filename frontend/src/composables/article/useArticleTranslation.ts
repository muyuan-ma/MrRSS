import { ref } from 'vue';
import type { Article } from '@/types/models';

interface TranslationSettings {
  enabled: boolean;
  targetLang: string;
  translationOnlyMode: boolean;
}

export function useArticleTranslation() {
  const translationSettings = ref<TranslationSettings>({
    enabled: false,
    targetLang: 'zh',
    translationOnlyMode: false,
  });

  // Load translation settings
  async function loadTranslationSettings(): Promise<void> {
    try {
      const res = await fetch('/api/settings');
      const data = await res.json();
      translationSettings.value = {
        enabled: data.translation_enabled === 'true',
        targetLang: data.target_language || 'zh',
        translationOnlyMode: data.translation_only_mode === 'true',
      };
    } catch (e) {
      console.error('Error loading translation settings:', e);
    }
  }

  // Legacy hooks kept for ArticleList. Title translation is now manual-only
  // from the reader toolbar, so these intentionally do not observe anything.
  function setupIntersectionObserver(listRef: HTMLElement | null, articles: Article[]): void {
    void listRef;
    void articles;
  }

  function observeArticle(el: Element | null): void {
    void el;
  }

  // Update translation settings from event
  function handleTranslationSettingsChange(enabled: boolean, targetLang: string): void {
    translationSettings.value = {
      enabled,
      targetLang,
      translationOnlyMode: translationSettings.value.translationOnlyMode,
    };

    void enabled;
  }

  // Cleanup
  function cleanup(): void {}

  return {
    translationSettings,
    loadTranslationSettings,
    setupIntersectionObserver,
    observeArticle,
    handleTranslationSettingsChange,
    cleanup,
  };
}
