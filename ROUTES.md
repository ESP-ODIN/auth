# API Routes

All routes are prefixed with `/api/v1/auth` and protected by a global rate limiter (1 req/s, burst of 5).

Swagger UI is available at `/swagger/index.html` when the server is running.

---

## POST /api/v1/auth/register

Create a new user account and return a signed JWT.

**Request body:**

| Field | Type | Rules |
| --- | --- | --- |
| `email` | string | Required. Valid email, unique (case-insensitive). |
| `password` | string | Required. 8–72 characters. Stored as bcrypt hash. |
| `username` | string | Required. 3–50 characters. Unique (case-insensitive). |
| `first_name` | string | Required. 3–100 characters. |
| `last_name` | string | Required. 2–100 characters. |
| `date_of_birth` | string | Required. Format `YYYY-MM-DD`. Must be at least 18 years old. |

**Responses:**

| Status | Description |
| --- | --- |
| `201 Created` | `{ "message": "...", "data": { "user": {...}, "token": "..." } }` |
| `400 Bad Request` | Invalid field, wrong date format, or user under 18. |
| `409 Conflict` | Email or username already exists. |
| `500 Internal Server Error` | Unexpected error (logged server-side). |

---

## POST /api/v1/auth/login

Authenticate a user and return a signed JWT.

**Request body:**

| Field | Type | Rules |
| --- | --- | --- |
| `email` | string | Required. Valid email. |
| `password` | string | Required. 8–72 characters. |

**Responses:**

| Status | Description |
| --- | --- |
| `200 OK` | `{ "message": "...", "data": { "user": {...}, "token": "..." } }` |
| `400 Bad Request` | Missing or invalid field. |
| `401 Unauthorized` | Invalid credentials (same message for unknown email or wrong password). |
| `500 Internal Server Error` | Unexpected error (logged server-side). |

---

## POST /api/v1/auth/forgot-password

Request a password reset link. An email with a reset token is sent if the account exists and is active.

**Request body:**

| Field | Type | Rules |
| --- | --- | --- |
| `email` | string | Required. Valid email. |

**Responses:**

| Status | Description |
| --- | --- |
| `202 Accepted` | `{ "message": "if this email is registered, a reset link has been sent" }` — always returned regardless of whether the email exists, to prevent user enumeration. |
| `400 Bad Request` | Missing or invalid email field. |

**Notes:**
- The reset token expires after 1 hour.
- Response time is kept constant to prevent timing-based enumeration.
- Requires SMTP configuration (see [README.md](README.md#smtp-configuration)).

---

## POST /api/v1/auth/reset-password

Validate a reset token and set a new password.

**Request body:**

| Field | Type | Rules |
| --- | --- | --- |
| `token` | string | Required. The token received in the reset email. |
| `new_password` | string | Required. 8–72 characters. |

**Responses:**

| Status | Description |
| --- | --- |
| `200 OK` | `{ "message": "password reset successfully" }` |
| `400 Bad Request` | Token is invalid, expired, or already used (generic message: `"invalid or expired token"`). |
| `500 Internal Server Error` | Unexpected error (logged server-side). |

**Notes:**
- The token is single-use. Once consumed, all other pending reset tokens for the same user are also invalidated.
- Tokens are stored as SHA-256 hashes in the database (never in plaintext).

---

## GET /.well-known/jwks.json

Returns the public RSA key used to verify JWTs issued by this service.

**Responses:**

| Status | Description |
| --- | --- |
| `200 OK` | `{ "keys": [{ "kty", "use", "alg", "kid", "n", "e" }] }` |

No authentication or rate limiting.
