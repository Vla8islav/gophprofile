// Shared session helpers for the web pages. The token is the API's HMAC bearer token
'use strict';

const TOKEN_KEY = 'gophprofile_token';

function getToken() {
    return localStorage.getItem(TOKEN_KEY) || '';
}

function setToken(token) {
    localStorage.setItem(TOKEN_KEY, token);
}

function clearToken() {
    localStorage.removeItem(TOKEN_KEY);
}

// currentUserID returns the numeric user id from the stored token
function currentUserID() {
    const token = getToken();
    if (!token) return null;
    try {
        // base64
        const b64 = token.replace(/-/g, '+').replace(/_/g, '/');
        const padded = b64 + '='.repeat((4 - (b64.length % 4)) % 4);
        const payload = atob(padded); // "userID:expiresAt:signature"
        const [userID, expiresAt] = payload.split(':');
        if (!userID || !/^\d+$/.test(userID)) return null;
        if (Number(expiresAt) * 1000 < Date.now()) {
            clearToken();
            return null;
        }
        return userID;
    } catch {
        return null;
    }
}
