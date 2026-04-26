import http from './http';
export function fetchAuthorizationStatus() {
    return http.get('/casbin/status');
}
export function reloadAuthorizationPolicies() {
    return http.post('/casbin/reload');
}
export function seedAuthorizationPolicies() {
    return http.post('/casbin/seed');
}
export const fetchCasbinStatus = fetchAuthorizationStatus;
export const reloadCasbinPolicies = reloadAuthorizationPolicies;
export const seedCasbinPolicies = seedAuthorizationPolicies;
