import http from './http';
export function fetchMenuRoutes() {
    return http.get('/menus/routes');
}
