/**
 * Better Auth catch-all route (sessions, agent registration, capability execute, admin).
 *
 * Note: next.config.js rewrites /api/* to the Go registry, but App Router
 * filesystem routes take precedence, so this handler is not proxied away.
 */

import { auth } from '@/lib/auth'
import { toNextJsHandler } from 'better-auth/next-js'

export const { GET, POST } = toNextJsHandler(auth)
