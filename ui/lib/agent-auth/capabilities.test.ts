import { describe, expect, it, vi } from 'vitest'

import {
  REGISTRY_ADMIN_CAPABILITIES,
  adminCapabilityNames,
  executeRegistryCapability,
} from './capabilities'

describe('REGISTRY_ADMIN_CAPABILITIES', () => {
  it('includes core admin capabilities for registry ops', () => {
    const names = adminCapabilityNames()
    expect(names).toContain('list_agents')
    expect(names).toContain('get_agent')
    expect(names).toContain('apply_manifest')
    expect(names).toContain('delete_agent')
    expect(names).toContain('registry_health')
    expect(names).toContain('list_deployments')
    // No duplicates
    expect(new Set(names).size).toBe(names.length)
  })

  it('defines input schemas for mutating ops', () => {
    const apply = REGISTRY_ADMIN_CAPABILITIES.find((c) => c.name === 'apply_manifest')
    const del = REGISTRY_ADMIN_CAPABILITIES.find((c) => c.name === 'delete_agent')
    expect(apply?.input?.type).toBe('object')
    expect(del?.input?.required).toContain('name')
  })
})

describe('executeRegistryCapability', () => {
  const agentSession = {
    user: { id: 'user_1', email: 'admin@example.com', role: 'admin' },
    agent: { id: 'agent_1', name: 'registry-admin' },
  }

  it('GETs list_agents from the real registry path', async () => {
    const fetchImpl = vi.fn(async (url: string | URL | Request) => {
      expect(String(url)).toBe('http://registry.test/v0/agents')
      return new Response(JSON.stringify([{ name: 'registry-admin' }]), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      })
    }) as unknown as typeof fetch

    const out = await executeRegistryCapability({
      capability: 'list_agents',
      agentSession,
      registryBaseUrl: 'http://registry.test',
      fetchImpl,
    })
    expect(out).toEqual([{ name: 'registry-admin' }])
    expect(fetchImpl).toHaveBeenCalledOnce()
  })

  it('GETs get_agent with name and tag', async () => {
    const fetchImpl = vi.fn(async (url: string | URL | Request) => {
      expect(String(url)).toBe(
        'http://registry.test/v0/agents/registry-admin/stable',
      )
      return new Response(JSON.stringify({ name: 'registry-admin', tag: 'stable' }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      })
    }) as unknown as typeof fetch

    const out = await executeRegistryCapability({
      capability: 'get_agent',
      arguments: { name: 'registry-admin', tag: 'stable' },
      agentSession,
      registryBaseUrl: 'http://registry.test/',
      fetchImpl,
    })
    expect(out).toMatchObject({ name: 'registry-admin' })
  })

  it('POSTs apply_manifest YAML to /v0/apply with actor header', async () => {
    const fetchImpl = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      expect(String(url)).toBe('http://registry.test/v0/apply')
      expect(init?.method).toBe('POST')
      const headers = new Headers(init?.headers)
      expect(headers.get('Content-Type')).toBe('application/yaml')
      expect(headers.get('X-Agent-Auth-Actor')).toContain('user_1')
      expect(String(init?.body)).toContain('"kind": "Agent"')
      expect(String(init?.body)).toContain('registry-admin')
      return new Response(JSON.stringify({ applied: 1 }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      })
    }) as unknown as typeof fetch

    const out = await executeRegistryCapability({
      capability: 'apply_manifest',
      arguments: {
        documents: [
          {
            apiVersion: 'ar.dev/v1alpha1',
            kind: 'Agent',
            metadata: { name: 'registry-admin' },
          },
        ],
      },
      agentSession,
      registryBaseUrl: 'http://registry.test',
      fetchImpl,
    })
    expect(out).toEqual({ applied: 1 })
  })

  it('POSTs raw yaml for apply_manifest', async () => {
    const fetchImpl = vi.fn(async (_url: string | URL | Request, init?: RequestInit) => {
      expect(String(init?.body)).toBe('apiVersion: ar.dev/v1alpha1\nkind: Agent\n')
      return new Response(JSON.stringify({ ok: true }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      })
    }) as unknown as typeof fetch

    await executeRegistryCapability({
      capability: 'apply_manifest',
      arguments: { yaml: 'apiVersion: ar.dev/v1alpha1\nkind: Agent\n' },
      agentSession,
      registryBaseUrl: 'http://registry.test',
      fetchImpl,
    })
  })

  it('DELETEs delete_agent by name', async () => {
    const fetchImpl = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      expect(String(url)).toBe('http://registry.test/v0/agents/old-agent')
      expect(init?.method).toBe('DELETE')
      return new Response(JSON.stringify({ deleted: true }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      })
    }) as unknown as typeof fetch

    await executeRegistryCapability({
      capability: 'delete_agent',
      arguments: { name: 'old-agent' },
      agentSession,
      registryBaseUrl: 'http://registry.test',
      fetchImpl,
    })
  })

  it('rejects unknown capabilities', async () => {
    await expect(
      executeRegistryCapability({
        capability: 'drop_database',
        agentSession,
        registryBaseUrl: 'http://registry.test',
        fetchImpl: vi.fn() as unknown as typeof fetch,
      }),
    ).rejects.toThrow(/Unsupported capability/)
  })

  it('surfaces non-OK registry responses', async () => {
    const fetchImpl = vi.fn(async () => {
      return new Response('nope', { status: 403 })
    }) as unknown as typeof fetch

    await expect(
      executeRegistryCapability({
        capability: 'list_agents',
        agentSession,
        registryBaseUrl: 'http://registry.test',
        fetchImpl,
      }),
    ).rejects.toThrow(/403/)
  })
})
