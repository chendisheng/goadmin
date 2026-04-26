import { useSessionStore } from '@/store/session';
function isAllowed(permission) {
    if (permission == null || permission === '') {
        return true;
    }
    const sessionStore = useSessionStore();
    return sessionStore.hasPermission(permission);
}
function applyVisibility(el, permission) {
    el.style.display = isAllowed(permission) ? '' : 'none';
}
export const permissionDirective = {
    mounted(el, binding) {
        applyVisibility(el, binding.value);
    },
    updated(el, binding) {
        applyVisibility(el, binding.value);
    },
};
