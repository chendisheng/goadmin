export class ApiError extends Error {
    code;
    payload;
    requestId;
    constructor(message, code, payload, requestId) {
        super(message);
        this.name = 'ApiError';
        this.code = code;
        this.payload = payload;
        this.requestId = requestId;
    }
}
