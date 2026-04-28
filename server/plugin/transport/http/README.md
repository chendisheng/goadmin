# Plugin CRUD API Examples

## Base Path

`/api/v1/plugins`

## 1. List plugins

### Request

```http
GET /api/v1/plugins
```

### Response

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "total": 1,
    "items": [
      {
        "name": "example",
        "description": "Example builtin plugin",
        "enabled": true,
        "menus": [
          {
            "plugin": "example",
            "id": "example-dashboard",
            "parent_id": "",
            "name": "Example Dashboard",
            "path": "/example/dashboard",
            "component": "view/example/dashboard/index",
            "icon": "dashboard",
            "sort": 10,
            "permission": "example:dashboard:view",
            "type": "menu",
            "visible": true,
            "enabled": true,
            "redirect": "",
            "external_url": "",
            "children": [],
            "created_at": "2026-03-27T00:00:00Z",
            "updated_at": "2026-03-27T00:00:00Z"
          }
        ],
        "permissions": [
          {
            "plugin": "example",
            "object": "example:dashboard",
            "action": "view",
            "description": "View example dashboard"
          }
        ],
        "created_at": "2026-03-27T00:00:00Z",
        "updated_at": "2026-03-27T00:00:00Z"
      }
    ]
  },
  "request_id": ""
}
```

## 2. Get plugin detail

### Request

```http
GET /api/v1/plugins/example
```

### Response

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "name": "example",
    "description": "Example builtin plugin",
    "enabled": true,
    "menus": [],
    "permissions": [],
    "created_at": "2026-03-27T00:00:00Z",
    "updated_at": "2026-03-27T00:00:00Z"
  },
  "request_id": ""
}
```

## 3. Create plugin

### Request

```http
POST /api/v1/plugins
Content-Type: application/json
```

```json
{
  "name": "analytics",
  "description": "Analytics plugin",
  "enabled": true,
  "menus": [
    {
      "id": "analytics-home",
      "parent_id": "",
      "name": "Analytics Home",
      "path": "/analytics",
      "component": "view/analytics/index",
      "icon": "chart",
      "sort": 10,
      "permission": "analytics:home:view",
      "type": "menu",
      "visible": true,
      "enabled": true
    }
  ],
  "permissions": [
    {
      "object": "analytics:home",
      "action": "view",
      "description": "View analytics home"
    }
  ]
}
```

### Response

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "name": "analytics",
    "description": "Analytics plugin",
    "enabled": true,
    "menus": [
      {
        "plugin": "analytics",
        "id": "analytics-home",
        "parent_id": "",
        "name": "Analytics Home",
        "path": "/analytics",
        "component": "view/analytics/index",
        "icon": "chart",
        "sort": 10,
        "permission": "analytics:home:view",
        "type": "menu",
        "visible": true,
        "enabled": true,
        "redirect": "",
        "external_url": "",
        "children": [],
        "created_at": "2026-03-27T00:00:00Z",
        "updated_at": "2026-03-27T00:00:00Z"
      }
    ],
    "permissions": [
      {
        "plugin": "analytics",
        "object": "analytics:home",
        "action": "view",
        "description": "View analytics home"
      }
    ],
    "created_at": "2026-03-27T00:00:00Z",
    "updated_at": "2026-03-27T00:00:00Z"
  },
  "request_id": ""
}
```

## 4. Update plugin

### Request

```http
PUT /api/v1/plugins/analytics
Content-Type: application/json
```

```json
{
  "description": "Analytics plugin v2",
  "enabled": false,
  "menus": [
    {
      "id": "analytics-home",
      "parent_id": "",
      "name": "Analytics Home",
      "path": "/analytics",
      "component": "view/analytics/index",
      "icon": "chart",
      "sort": 10,
      "permission": "analytics:home:view",
      "type": "menu",
      "visible": true,
      "enabled": true
    },
    {
      "id": "analytics-report",
      "parent_id": "analytics-home",
      "name": "Analytics Report",
      "path": "/analytics/report",
      "component": "view/analytics/report/index",
      "icon": "report",
      "sort": 20,
      "permission": "analytics:report:view",
      "type": "menu",
      "visible": true,
      "enabled": true
    }
  ],
  "permissions": [
    {
      "object": "analytics:home",
      "action": "view",
      "description": "View analytics home"
    },
    {
      "object": "analytics:report",
      "action": "view",
      "description": "View analytics report"
    }
  ]
}
```

### Response

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "name": "analytics",
    "description": "Analytics plugin v2",
    "enabled": false,
    "menus": [
      {
        "plugin": "analytics",
        "id": "analytics-home",
        "parent_id": "",
        "name": "Analytics Home",
        "path": "/analytics",
        "component": "view/analytics/index",
        "icon": "chart",
        "sort": 10,
        "permission": "analytics:home:view",
        "type": "menu",
        "visible": true,
        "enabled": true,
        "redirect": "",
        "external_url": "",
        "children": [
          {
            "plugin": "analytics",
            "id": "analytics-report",
            "parent_id": "analytics-home",
            "name": "Analytics Report",
            "path": "/analytics/report",
            "component": "view/analytics/report/index",
            "icon": "report",
            "sort": 20,
            "permission": "analytics:report:view",
            "type": "menu",
            "visible": true,
            "enabled": true,
            "redirect": "",
            "external_url": "",
            "children": [],
            "created_at": "2026-03-27T00:00:00Z",
            "updated_at": "2026-03-27T00:00:00Z"
          }
        ],
        "created_at": "2026-03-27T00:00:00Z",
        "updated_at": "2026-03-27T00:00:00Z"
      }
    ],
    "permissions": [
      {
        "plugin": "analytics",
        "object": "analytics:home",
        "action": "view",
        "description": "View analytics home"
      },
      {
        "plugin": "analytics",
        "object": "analytics:report",
        "action": "view",
        "description": "View analytics report"
      }
    ],
    "created_at": "2026-03-27T00:00:00Z",
    "updated_at": "2026-03-27T00:00:00Z"
  },
  "request_id": ""
}
```

## 5. Delete plugin

### Request

```http
DELETE /api/v1/plugins/analytics
```

### Response

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "deleted": true
  },
  "request_id": ""
}
```

## 6. List plugin menus

### Request

```http
GET /api/v1/plugins/menus
```

### Response

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "plugin": "example",
        "id": "example-dashboard",
        "parent_id": "",
        "name": "Example Dashboard",
        "path": "/example/dashboard",
        "component": "view/example/dashboard/index",
        "icon": "dashboard",
        "sort": 10,
        "permission": "example:dashboard:view",
        "type": "menu",
        "visible": true,
        "enabled": true,
        "redirect": "",
        "external_url": "",
        "children": [],
        "created_at": "2026-03-27T00:00:00Z",
        "updated_at": "2026-03-27T00:00:00Z"
      }
    ]
  },
  "request_id": ""
}
```

## 7. List plugin permissions

### Request

```http
GET /api/v1/plugins/permissions
```

### Response

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "plugin": "example",
        "object": "example:dashboard",
        "action": "view",
        "description": "View example dashboard"
      }
    ]
  },
  "request_id": ""
}
```
