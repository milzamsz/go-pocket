# ADR-0005: Standard Webhooks verification with production fail-closed

- Status: accepted
- Date: 2026-05-29
- Decision owner: platform

## Context

go-pocket ingests provider webhooks for billing (Polar) and email (Resend) at
`POST /webhooks/polar` and `POST /webhooks/resend`. These endpoints mutate
tenant state (subscriptions, invoices, organization billing denorm fields,
email delivery audit records), so accepting a forged payload is a privilege- and
data-integrity risk.

Two problems existed in the initial implementation:

1. The Polar verifier used a bespoke `Polar-Signature: t=<ts>,v1=<hex>` scheme
   (HMAC over `timestamp.payload`, hex-encoded). Polar does not send that. Per
   Polar's documentation, Polar follows the
   [Standard Webhooks](https://www.standardwebhooks.com) specification, the same
   family Resend/Svix uses. The custom scheme would have rejected every genuine
   Polar delivery while appearing to "work" in tests that signed with the same
   wrong scheme.

2. Both verifiers were fail-open: when no webhook secret was configured they
   returned success and accepted the payload. That is convenient for local
   development but dangerous in production, where a missing/typo'd secret would
   silently disable signature checking.

## Decision

1. Verify Polar (and continue verifying Resend) using the Standard Webhooks
   specification:
   - Signed content is `{id}.{timestamp}.{payload}`.
   - The signature is a base64-encoded HMAC-SHA256 digest.
   - The secret is base64, optionally prefixed with `whsec_`; if the remainder
     is not valid base64 (operator set a plain custom secret) the raw bytes are
     used.
   - Both canonical `webhook-id` / `webhook-timestamp` / `webhook-signature`
     headers and the legacy `svix-*` aliases are accepted.
   - A 5-minute timestamp tolerance guards against replay.

2. Fail closed in production: when `APP_ENV` is production-like and the relevant
   `*_WEBHOOK_SECRET` is empty, reject the webhook with an invalid-signature
   error instead of accepting it. In development (default) an empty secret still
   skips verification so local testing and provider CLIs remain frictionless.

3. Surface configuration risk at boot: `config.Validate()` logs warnings in
   production for missing webhook secrets, missing provider credentials, and a
   non-HTTPS base URL.

## Consequences

- Real Polar webhooks now validate; the integration is correct rather than
  merely green in tests.
- Misconfigured production deployments fail safe (reject) rather than fail open
  (accept forged events), at the cost of requiring the secret to be set before
  webhooks function in production — which is the desired behavior.
- The verification logic for both providers now shares the same mental model,
  reducing the chance of future divergence.
