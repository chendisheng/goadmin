import { fetchCurrentUser } from '@/api/auth';
import { useSessionStore } from '@/store/session';

export async function restoreAuthenticatedSession() {
  const sessionStore = useSessionStore();
  if (!sessionStore.isAuthenticated) {
    sessionStore.setCurrentUser(null);
    return null;
  }

  const currentUser = await fetchCurrentUser();
  sessionStore.setCurrentUser(currentUser);
  return currentUser;
}
