import { defineStore } from 'pinia';
const SIDEBAR_COLLAPSED_KEY = 'goadmin.sidebar.collapsed';
function canUseStorage() {
    return typeof window !== 'undefined' && typeof window.localStorage !== 'undefined';
}
function readCollapsedState() {
    if (!canUseStorage()) {
        return false;
    }
    return window.localStorage.getItem(SIDEBAR_COLLAPSED_KEY) === '1';
}
function persistCollapsedState(collapsed) {
    if (!canUseStorage()) {
        return;
    }
    window.localStorage.setItem(SIDEBAR_COLLAPSED_KEY, collapsed ? '1' : '0');
}
export const useAppStore = defineStore('app', {
    state: () => ({
        sidebarCollapsed: readCollapsedState(),
    }),
    actions: {
        hydrate() {
            this.sidebarCollapsed = readCollapsedState();
        },
        toggleSidebar() {
            this.sidebarCollapsed = !this.sidebarCollapsed;
            persistCollapsedState(this.sidebarCollapsed);
        },
        setSidebarCollapsed(collapsed) {
            this.sidebarCollapsed = collapsed;
            persistCollapsedState(collapsed);
        },
    },
});
