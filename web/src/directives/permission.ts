import { useSessionStore } from '@/store/session';

type PermissionValue = string | string[] | undefined;
type PermissionBinding = {
  value: PermissionValue;
};

function isAllowed(permission: PermissionValue): boolean {
  if (permission == null || permission === '') {
    return true;
  }
  const sessionStore = useSessionStore();
  return sessionStore.hasPermission(permission);
}

function applyVisibility(el: HTMLElement, permission: PermissionValue) {
  el.style.display = isAllowed(permission) ? '' : 'none';
}

export const permissionDirective = {
  mounted(el: HTMLElement, binding: PermissionBinding) {
    applyVisibility(el, binding.value);
  },
  updated(el: HTMLElement, binding: PermissionBinding) {
    applyVisibility(el, binding.value);
  },
};
