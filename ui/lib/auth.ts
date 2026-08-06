/**
 * Better Auth server config for AgentRegistry UI.
 *
 * Plugins:
 * - agentAuth: Agent Auth Protocol provider (discovery + capabilities + JWT execute)
 * - admin: user role / ban / impersonation for human operators who approve agents
 *
 * Database: SQLite file (BETTER_AUTH_DB_PATH) for local agent sessions/grants.
 * Capability execution proxies to the Go registry (REGISTRY_URL / NEXT_PUBLIC_API_URL).
 */

import { betterAuth } from 'better-auth'
import { admin } from 'better-auth/plugins'
import { agentAuth } from '@better-auth/agent-auth'
import Database from 'better-sqlite3'
import { mkdirSync } from 'node:fs'
import { dirname, resolve } from 'node:path'

import {
  REGISTRY_ADMIN_CAPABILITIES,
  executeRegistryCapability,
} from '@/lib/agent-auth/capabilities'

function resolveRegistryBaseUrl(): string {
  return (
    process.env.REGISTRY_URL ||
    process.env.AGENT_REGISTRY_URL ||
    process.env.NEXT_PUBLIC_API_URL ||
    'http://localhost:12121'
  )
}

function openAuthDatabase() {
  // Use in-memory when BETTER_AUTH_DB_PATH=:memory: (tests / ephemeral).
  if (process.env.BETTER_AUTH_DB_PATH === ':memory:') {
    return new Database(':memory:')
  }
  const raw =
    process.env.BETTER_AUTH_DB_PATH ||
    resolve(process.cwd(), '.data', 'better-auth.sqlite')
  const path = resolve(raw)
  mkdirSync(dirname(path), { recursive: true })
  return new Database(path)
}

const secret =
  process.env.BETTER_AUTH_SECRET ||
  process.env.AGENT_REGISTRY_JWT_PRIVATE_KEY ||
  // Dev-only fallback so local `next dev` works without extra env.
  // Production MUST set BETTER_AUTH_SECRET.
  'dev-only-agentregistry-better-auth-secret-change-me'

const baseURL =
  process.env.BETTER_AUTH_URL ||
  process.env.NEXT_PUBLIC_BETTER_AUTH_URL ||
  'http://localhost:3000'

export const auth = betterAuth({
  database: openAuthDatabase(),
  secret,
  baseURL,
  emailAndPassword: {
    enabled: true,
  },
  plugins: [
    admin({
      defaultRole: 'user',
      adminRoles: ['admin'],
    }),
    agentAuth({
      providerName: 'AgentRegistry',
      providerDescription:
        'AgentRegistry admin provider — discover, register, and manage agents, MCP servers, skills, prompts, and deployments via capability-scoped JWTs.',
      modes: ['delegated', 'autonomous'],
      capabilities: REGISTRY_ADMIN_CAPABILITIES,
      async onExecute({ capability, arguments: args, agentSession }) {
        return executeRegistryCapability({
          capability,
          arguments: (args as Record<string, unknown>) || null,
          agentSession,
          registryBaseUrl: resolveRegistryBaseUrl(),
        })
      },
    }),
  ],
})

export type Auth = typeof auth
