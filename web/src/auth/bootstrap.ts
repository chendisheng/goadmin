import { fetchCurrentUser } from '@/api/auth';
import { useLocaleStore } from '@/store/locale';
import { useSessionStore } from '@/store/session';

export async function restoreAuthenticatedSession() {
  const sessionStore = useSessionStore();
  const localeStore = useLocaleStore();
  if (!sessionStore.isAuthenticated) {
    sessionStore.setCurrentUser(null);
    return null;
  }

  const currentUser = await fetchCurrentUser();
  sessionStore.setCurrentUser(currentUser);
  const resolvedLanguage = currentUser.language?.trim() || sessionStore.language || 'zh-CN';
  sessionStore.setLanguage(resolvedLanguage);
  localeStore.syncFromUser(currentUser);
  return currentUser;
}
