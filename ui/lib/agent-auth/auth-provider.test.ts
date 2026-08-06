/**
 * Smoke test: real betterAuth + agentAuth + admin plugins load and expose
 * discovery for the registry admin capability set.
 */

import { describe, expect, it } from 'vitest'
import { betterAuth } from 'better-auth'
import { admin } from 'better-auth/plugins'
import { agentAuth } from '@better-auth/agent-auth'
import Database from 'better-sqlite3'

import {
  REGISTRY_ADMIN_CAPABILITIES,
  adminCapabilityNames,
} from './capabilities'

describe('Agent Auth provider (live betterAuth + agentAuth plugins)', () => {
  it('builds discovery with AgentRegistry provider and admin capabilities', async () => {
    const auth = betterAuth({
      database: new Database(':memory:'),
      secret: 'test-secret-agentregistry-agent-auth-32b',
      baseURL: 'http://localhost:3000',
      emailAndPassword: { enabled: true },
      plugins: [
        admin({ defaultRole: 'user', adminRoles: ['admin'] }),
        agentAuth({
          providerName: 'AgentRegistry',
          providerDescription: 'AgentRegistry admin provider',
          modes: ['delegated', 'autonomous'],
          capabilities: REGISTRY_ADMIN_CAPABILITIES,
          async onExecute({ capability }) {
            return { capability }
          },
        }),
      ],
    })

    const cfg = await auth.api.getAgentConfiguration()
    expect(cfg).toMatchObject({
      provider_name: 'AgentRegistry',
      modes: expect.arrayContaining(['delegated', 'autonomous']),
    })
    expect(cfg.default_location).toContain('/capability/execute')
    expect(cfg.endpoints?.execute).toContain('/capability/execute')

    const listed = await auth.api.listCapabilities({})
    // listCapabilities shape may wrap { capabilities: [...] }
    const caps = Array.isArray(listed)
      ? listed
      : (listed as { capabilities?: Array<{ name: string }> })?.capabilities ||
        (listed as { data?: Array<{ name: string }> })?.data ||
        []
    const names = (caps as Array<{ name: string }>).map((c) => c.name)
    for (const want of adminCapabilityNames()) {
      expect(names).toContain(want)
    }
  })
})
