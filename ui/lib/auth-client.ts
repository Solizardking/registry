/**
 * Better Auth browser/client with Agent Auth + Admin plugins.
 */

import { createAuthClient } from 'better-auth/client'
import { agentAuthClient } from '@better-auth/agent-auth/client'
import { adminClient } from 'better-auth/client/plugins'

export const authClient = createAuthClient({
  baseURL:
    typeof window !== 'undefined'
      ? window.location.origin
      : process.env.NEXT_PUBLIC_BETTER_AUTH_URL || 'http://localhost:3000',
  plugins: [agentAuthClient(), adminClient()],
})
