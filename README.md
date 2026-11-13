# auth

### Prerequisites
- **Go** (version 1.22 or later required for standard `net/http` method-based routing):
  - **Linux (Ubuntu/Debian):**
    ```bash
    sudo apt update && sudo apt install -y golang-go
    ```
  - **Windows (PowerShell / Winget):**
    ```powershell
    winget install GoLang.Go
    ```
  - **macOS (Homebrew):**
    ```bash
    brew install go
    ```
- **VS Code** with the official **Go** extension (by *Go Team at Google*)
- **Make** (optional, for running Makefile targets):
  - **Windows (PowerShell / Winget):**
    ```powershell
    winget install GnuWin32.Make
    ```
    *(Alternatively via Chocolatey: `choco install make`)*
  - **macOS:** Check with `make --version`. If missing, run:
    ```bash
    xcode-select --install
    ```
  - **Linux (Ubuntu/Debian):**
    ```bash
    sudo apt install -y make
    ```

---

### Environment Setup

1. At the root of the project, duplicate the exemple file to create your local `.env` file:
   ```bash
   cp .env.exemple .env
   ```

## Docker

A development `Dockerfile` runs the API with hot-reload (via [air](https://github.com/air-verse/air)),
meant to be plugged as-is into the project's docker-compose. Fill in `DATABASE_URL` in `.env` first
(see `.env.exemple`).

Build the image:

```sh
make docker-build
```

Run it locally with hot-reload (mounts the source code and exposes the app on the host port set by
`PORT` in `.env`, `8091` by default):

```sh
make docker-dev
```

Then check it with `curl http://127.0.0.1:8091/.well-known/jwks.json`. Editing any `.go` file rebuilds and restarts
the server automatically inside the container. The container itself always listens on `0.0.0.0:8080`
(overridden from `.env`'s `HTTP_ADDR` by the `docker-dev` target); only the host-side port, read from
`PORT` in `.env`, changes.

## JWT RS256 configuration

Auth signs access tokens using a single RSA private key loaded once at startup,
before connecting to PostgreSQL. The existing `id` (Auth USER.id) and `email`
claims and 24-hour lifetime are preserved. Tokens also include `iss`, `aud`,
`iat`, `exp` and a configurable `kid` header. There is no `sub` substitution.
The register response remains `{ "message": "...", "data": { "user": {...}, "token": "..." } }`.

Required environment variables, in addition to `DATABASE_URL`:

| Variable | Meaning | Local example |
| --- | --- | --- |
| `JWT_PRIVATE_KEY_PATH` | Unencrypted PKCS#1 or PKCS#8 RSA PEM, at least 2048 bits | `secrets/jwt-private.pem` |
| `JWT_KEY_ID` | Unique identifier of the signing key | `odin-dev-1` |
| `JWT_ISSUER` | Exact issuer consumers must trust | `http://localhost:8080` |
| `JWT_AUDIENCE` | One logical API audience consumers must require | `odin-api` |

`odin-api` represents the ODIN APIs collectively for this configuration. It does
not grant business permissions. Consumers must validate the signature with an
explicit RS256 allowlist, issuer, audience and expiration before using `id`.
`JWT_SECRET_KEY` is no longer used; old HS256 tokens will not validate against
the new JWKS. Users will need a newly issued token during this migration.

Generate a development key pair with OpenSSL (run only when provisioning a new
key; do not overwrite an active key):

```powershell
New-Item -ItemType Directory -Force secrets
openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:3072 -out secrets/jwt-private.pem
openssl pkey -in secrets/jwt-private.pem -pubout -out secrets/jwt-public.pem
Copy-Item .env.example .env
```

Fill `.env` using the examples above and your database settings. The public PEM
is optional: Auth derives JWKS from the private key, never reads a separate
public key file, and never sends private RSA parameters through the endpoint.
`secrets/`, `*.pem` and `*.key` are excluded from Git and Docker build context.
Keep private keys at these excluded paths and restrict their filesystem access.
For deployment, mount the key read-only, e.g. at `/run/secrets/jwt-private.pem`,
and configure that path. Never bake it into the image or mount it in consumers.
The existing `make docker-dev` mounts the workspace into `/app`, so the relative
development path also works there. Use HTTPS for deployed APIs.

Public endpoint: `GET /.well-known/jwks.json`. Its response has one `keys` entry
with exactly `kty`, `use`, `alg`, `kid`, `n`, and `e`. The RSA modulus and exponent
use unsigned base64url encoding without padding. The current implementation
publishes the active key only; overlapping key rotation is future work. Replacing
the active key immediately removes the old verification key from this endpoint.

## Register endpoint

`POST /api/v1/auth/register` creates a user and returns a signed JWT.
The full contract is also available in Swagger (`make swagger`, files in `docs/`).

Request body (all fields are required):

```json
{
  "email": "jane.doe@example.com",
  "password": "Str0ngPassw0rd!",
  "username": "janedoe",
  "first_name": "Jane",
  "last_name": "Doe",
  "date_of_birth": "1990-01-31"
}
```

| Field | Rules |
| --- | --- |
| `email` | Valid email address. Unique, case-insensitive. |
| `password` | 8 to 100 characters. Stored as a bcrypt hash. |
| `username` | 3 to 50 characters. Unique, case-insensitive (`Jane` and `jane` are the same username). |
| `first_name` | First name, 3 to 100 characters. |
| `last_name` | Last name, 2 to 100 characters. |
| `date_of_birth` | Format `YYYY-MM-DD`. The user must be at least 18 years old on the registration date. |

Email and username are stored as typed; uniqueness checks compare `LOWER(...)` values.

Responses:

| Status | When |
| --- | --- |
| `201 Created` | User registered. Body: `{ "message": "...", "data": { "user": {...}, "token": "..." } }` |
| `400 Bad Request` | Missing or invalid field, wrong date format, or user under 18. |
| `409 Conflict` | Email or username already used. |
| `500 Internal Server Error` | Unexpected error (database, hashing, signing). The detailed error is logged server-side only. |

Example `201` response body. Keys use snake_case, like the request. The password,
even hashed, is never returned:

```json
{
  "message": "user registered successfully",
  "data": {
    "user": {
      "id": "3f6c1c1e-8a52-4a8e-9a6b-2f1f0f6a9b10",
      "email": "jane.doe@example.com",
      "username": "janedoe",
      "first_name": "Jane",
      "last_name": "Doe",
      "date_of_birth": "1990-01-31",
      "role_id": 1,
      "is_active": true,
      "created_at": "2026-11-06T16:10:00Z",
      "updated_at": "2026-11-06T16:10:00Z"
    },
    "token": "<JWT>"
  }
}
```

Database requirements: the `public.user` table needs the `username`, `first_name`,
`last_name` and `date_of_birth` columns, with unique indexes on `LOWER(username)`
and `LOWER(email)`.

## Tests

```powershell
gofmt -w cmd config db internal migrate
go mod tidy
go test ./...
go vet ./...
```

Tests generate ephemeral keys and use a mock repository; no PostgreSQL or real
private key is required. They cover RS256, claims, 24-hour expiration, `kid`,
wrong-key rejection, algorithm restrictions, key loading and verification using
the public key reconstructed from the HTTP JWKS response.

## Manual registration and JWKS check

Two authentication endpoints are available: `POST /api/v1/auth/register` and
`POST /api/v1/auth/login`. Login accepts `{ "email", "password" }` and returns
`200` with `{ "message", "data": { "user", "token" } }`. The user payload is the
public profile and does not include the password. Unknown email and wrong
password both return `401` with `invalid credentials`. A login password outside
8 to 72 characters returns `400`. A running PostgreSQL database with the existing
schema and role ID 1 is required for registration and login. The check below
uses the register endpoint, which also issues a JWT.

Start `go run ./cmd/api`, then in another PowerShell terminal:

```powershell
$base = 'http://127.0.0.1:8080' # use port 8091 with make docker-dev
$email = 'rsa-' + [guid]::NewGuid().ToString('N') + '@example.com'
$username = 'rsa' + [guid]::NewGuid().ToString('N').Substring(0, 12)
$body = @{
    email         = $email
    password      = 'Local-test-password-123!'
    username      = $username
    first_name    = 'Jane'
    last_name     = 'Doe'
    date_of_birth = '1990-01-31'
} | ConvertTo-Json
$result = Invoke-RestMethod -Method Post -Uri "$base/api/v1/auth/register" -ContentType 'application/json' -Body $body
$jwks = Invoke-RestMethod -Uri "$base/.well-known/jwks.json"
$jwks | ConvertTo-Json -Depth 5

function Read-JWTPart([string]$part) {
    $encoded = $part.Replace('-', '+').Replace('_', '/')
    $encoded = $encoded.PadRight($encoded.Length + ((4 - $encoded.Length % 4) % 4), '=')
    [Text.Encoding]::UTF8.GetString([Convert]::FromBase64String($encoded)) | ConvertFrom-Json
}
$parts = $result.data.token.Split('.')
$header = Read-JWTPart $parts[0]
$claims = Read-JWTPart $parts[1]
$header | ConvertTo-Json
$claims | ConvertTo-Json
if ($header.alg -ne 'RS256' -or $header.kid -ne $jwks.keys[0].kid) { throw 'Algorithm/kid mismatch' }
if ($claims.id -ne $result.data.user.id -or $claims.email -ne $email) { throw 'Identity mismatch' }
```

This last snippet inspects the JWT; decoding alone does not verify its signature.
To exercise actual signature validation using a JWKS response without a database:

```powershell
go test ./cmd/api -run TestJWKSVerifiesIssuedToken -v
go test ./internal/utils -run TestGenerateJWT -v
```
