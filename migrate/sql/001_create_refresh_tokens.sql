-- ODIN-221 : refresh tokens stockés en base.
-- Seul le hash SHA-256 du token est conservé, jamais le token en clair.
CREATE TABLE IF NOT EXISTS public.refresh_tokens (
    id          uuid PRIMARY KEY,
    user_id     uuid NOT NULL REFERENCES public."user"(id) ON DELETE CASCADE,
    token_hash  text NOT NULL UNIQUE,
    expires_at  timestamptz NOT NULL,
    revoked_at  timestamptz,
    created_at  timestamptz NOT NULL DEFAULT now()
);

-- Révocation de tous les tokens d'un utilisateur (détection de vol, futur logout).
CREATE INDEX IF NOT EXISTS refresh_tokens_user_id_idx ON public.refresh_tokens (user_id);
