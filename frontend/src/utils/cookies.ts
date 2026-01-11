import { User } from "./api";

/**
 * Get a cookie value by name
 * 
 * NOTE: This function CANNOT read HttpOnly cookies (like auth_token).
 * HttpOnly cookies are only accessible by the server for security.
 * For authentication, use /api/auth/me endpoint instead.
 * This function is only useful for reading non-HttpOnly cookies.
 */
export function getCookie(name: string): string | null {
  const nameEQ = name + "=";
  const cookies = document.cookie.split(";");

  for (let i = 0; i < cookies.length; i++) {
    let cookie = cookies[i];
    while (cookie.charAt(0) === " ") {
      cookie = cookie.substring(1, cookie.length);
    }
    if (cookie.indexOf(nameEQ) === 0) {
      return cookie.substring(nameEQ.length, cookie.length);
    }
  }
  return null;
}

/**
 * Delete a cookie by name (for non-HttpOnly cookies only)
 * 
 * NOTE: This function CANNOT delete HttpOnly cookies.
 * HttpOnly cookies can only be deleted by the server.
 * For logout, call /api/auth/logout endpoint instead.
 */
export function deleteCookie(name: string): void {
  document.cookie = name + "=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;";
}

export function parseJWT(token: string): User | null {
  try {
    const base64Url = token.split(".")[1];
    const base64 = base64Url.replace(/-/g, "+").replace(/_/g, "/");
    const jsonPayload = decodeURIComponent(
      atob(base64)
        .split("")
        .map((c) => "%" + ("00" + c.charCodeAt(0).toString(16)).slice(-2))
        .join("")
    );
    const payload = JSON.parse(jsonPayload);
    return {
      id: payload.user_id,
      username: payload.username,
    };
  } catch (error) {
    console.error("Failed to parse JWT:", error);
    return null;
  }
}

export function isTokenExpired(token: string): boolean {
  try {
    const base64Url = token.split(".")[1];
    const base64 = base64Url.replace(/-/g, "+").replace(/_/g, "/");
    const jsonPayload = decodeURIComponent(
      atob(base64)
        .split("")
        .map((c) => "%" + ("00" + c.charCodeAt(0).toString(16)).slice(-2))
        .join("")
    );
    const payload = JSON.parse(jsonPayload);

    if (!payload.exp) {
      return true;
    }
    const currentTime = Date.now() / 1000;
    return payload.exp < currentTime;
  } catch (error) {
    console.error("Failed to check token expiration:", error);
    return true; // If we can't parse it, consider it expired
  }
}
