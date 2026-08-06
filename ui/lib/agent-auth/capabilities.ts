/**
 * Registry admin capabilities for the Agent Auth provider.
 *
 * These are the scoped actions an approved admin agent may execute against
 * the AgentRegistry HTTP API (default base: NEXT_PUBLIC_API_URL / REGISTRY_URL).
 */

export type JsonSchema = {
  type: string
  properties?: Record<string, unknown>
  required?: string[]
  additionalProperties?: boolean
  description?: string
}

export type AgentCapability = {
  name: string
  description: string
  input?: JsonSchema
}

/** Admin catalog + lifecycle capabilities for AgentRegistry. */
export const REGISTRY_ADMIN_CAPABILITIES: AgentCapability[] = [
  {
    name: 'list_agents',
    description: 'List agents registered in the AgentRegistry catalog.',
  },
  {
    name: 'get_agent',
    description: 'Get a single agent by name and optional tag.',
    input: {
      type: 'object',
      properties: {
        name: { type: 'string', description: 'Agent name' },
        tag: { type: 'string', description: 'Tag (defaults to latest)' },
      },
      required: ['name'],
    },
  },
  {
    name: 'list_servers',
    description: 'List MCP servers in the registry.',
  },
  {
    name: 'get_server',
    description: 'Get a single MCP server by name and optional tag.',
    input: {
      type: 'object',
      properties: {
        name: { type: 'string' },
        tag: { type: 'string' },
      },
      required: ['name'],
    },
  },
  {
    name: 'list_skills',
    description: 'List skills in the registry.',
  },
  {
    name: 'list_prompts',
    description: 'List prompts in the registry.',
  },
  {
    name: 'list_deployments',
    description: 'List deployments tracked by the registry.',
  },
  {
    name: 'get_deployment',
    description: 'Get a deployment by name and optional namespace.',
    input: {
      type: 'object',
      properties: {
        name: { type: 'string' },
        namespace: { type: 'string' },
      },
      required: ['name'],
    },
  },
  {
    name: 'apply_manifest',
    description:
      'Apply one or more ar.dev/v1alpha1 documents via POST /v0/apply (YAML body). Pass multi-doc YAML in `yaml`, or an array of objects in `documents` (serialized to YAML).',
    input: {
      type: 'object',
      properties: {
        yaml: {
          type: 'string',
          description: 'Multi-doc YAML stream (preferred; Content-Type: application/yaml)',
        },
        documents: {
          type: 'array',
          description:
            'Array of resource objects; each is JSON-stringified under a YAML document separator',
          items: { type: 'object', additionalProperties: true },
        },
        dryRun: {
          type: 'boolean',
          description: 'When true, appends ?dryRun=true',
        },
      },
    },
  },
  {
    name: 'delete_agent',
    description: 'Delete an agent tag from the registry.',
    input: {
      type: 'object',
      properties: {
        name: { type: 'string' },
        tag: { type: 'string' },
      },
      required: ['name'],
    },
  },
  {
    name: 'registry_health',
    description: 'Check AgentRegistry health (GET /v0/health).',
  },
  {
    name: 'registry_version',
    description: 'Return AgentRegistry version (GET /v0/version).',
  },
]

export type AgentSessionLite = {
  user?: { id?: string; email?: string; role?: string | null }
  agent?: {
    id?: string
    name?: string
    capabilityGrants?: Array<{ capability: string; status?: string }>
  }
}

export type ExecuteArgs = {
  capability: string
  arguments?: Record<string, unknown> | null
  agentSession: AgentSessionLite
  registryBaseUrl: string
  fetchImpl?: typeof fetch
}

function joinUrl(base: string, path: string): string {
  const b = base.replace(/\/+$/, '')
  const p = path.startsWith('/') ? path : `/${path}`
  return `${b}${p}`
}

async function registryFetch(
  base: string,
  path: string,
  init: RequestInit,
  fetchImpl: typeof fetch,
): Promise<unknown> {
  const url = joinUrl(base, path)
  const headers = new Headers(init.headers || {})
  if (!headers.has('Accept')) {
    headers.set('Accept', 'application/json')
  }
  if (init.body !== undefined && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }
  const res = await fetchImpl(url, {
    ...init,
    headers,
  })
  const text = await res.text()
  let body: unknown = text
  try {
    body = text ? JSON.parse(text) : null
  } catch {
    // keep raw text
  }
  if (!res.ok) {
    const err = new Error(
      `registry ${init.method || 'GET'} ${path} failed: ${res.status} ${typeof body === 'string' ? body : JSON.stringify(body)}`,
    )
    ;(err as Error & { status?: number }).status = res.status
    throw err
  }
  return body
}

/** Serialize document objects into a multi-doc YAML-ish stream for /v0/apply. */
export function documentsToYamlStream(documents: unknown[]): string {
  return documents
    .map((doc) => {
      if (typeof doc === 'string') return doc.trim()
      // Prefer real YAML when available; JSON is valid YAML subset for objects.
      return JSON.stringify(doc, null, 2)
    })
    .filter(Boolean)
    .join('\n---\n')
}

/**
 * Execute a registry admin capability against the live AgentRegistry HTTP API.
 * Pure enough to unit-test with a mock fetch.
 */
export async function executeRegistryCapability({
  capability,
  arguments: args,
  agentSession,
  registryBaseUrl,
  fetchImpl = fetch,
}: ExecuteArgs): Promise<unknown> {
  const a = (args || {}) as Record<string, unknown>
  const actor = {
    userId: agentSession.user?.id,
    email: agentSession.user?.email,
    agentId: agentSession.agent?.id,
    agentName: agentSession.agent?.name,
  }

  switch (capability) {
    case 'list_agents':
      return registryFetch(registryBaseUrl, '/v0/agents', { method: 'GET' }, fetchImpl)

    case 'get_agent': {
      const name = String(a.name || '')
      if (!name) throw new Error('name is required')
      const tag = a.tag ? String(a.tag) : undefined
      const path = tag
        ? `/v0/agents/${encodeURIComponent(name)}/${encodeURIComponent(tag)}`
        : `/v0/agents/${encodeURIComponent(name)}`
      return registryFetch(registryBaseUrl, path, { method: 'GET' }, fetchImpl)
    }

    case 'list_servers':
      return registryFetch(registryBaseUrl, '/v0/servers', { method: 'GET' }, fetchImpl)

    case 'get_server': {
      const name = String(a.name || '')
      if (!name) throw new Error('name is required')
      const tag = a.tag ? String(a.tag) : undefined
      const path = tag
        ? `/v0/servers/${encodeURIComponent(name)}/${encodeURIComponent(tag)}`
        : `/v0/servers/${encodeURIComponent(name)}`
      return registryFetch(registryBaseUrl, path, { method: 'GET' }, fetchImpl)
    }

    case 'list_skills':
      return registryFetch(registryBaseUrl, '/v0/skills', { method: 'GET' }, fetchImpl)

    case 'list_prompts':
      return registryFetch(registryBaseUrl, '/v0/prompts', { method: 'GET' }, fetchImpl)

    case 'list_deployments':
      return registryFetch(registryBaseUrl, '/v0/deployments', { method: 'GET' }, fetchImpl)

    case 'get_deployment': {
      const name = String(a.name || '')
      if (!name) throw new Error('name is required')
      const qs = a.namespace
        ? `?namespace=${encodeURIComponent(String(a.namespace))}`
        : ''
      return registryFetch(
        registryBaseUrl,
        `/v0/deployments/${encodeURIComponent(name)}${qs}`,
        { method: 'GET' },
        fetchImpl,
      )
    }

    case 'apply_manifest': {
      // Registry expects application/yaml multi-doc stream (see openapi apply-batch).
      let yamlBody: string
      if (typeof a.yaml === 'string' && a.yaml.trim()) {
        yamlBody = a.yaml
      } else if (Array.isArray(a.documents)) {
        yamlBody = documentsToYamlStream(a.documents as unknown[])
      } else {
        throw new Error('apply_manifest requires `yaml` string or `documents` array')
      }
      const qs = a.dryRun === true ? '?dryRun=true' : ''
      return registryFetch(
        registryBaseUrl,
        `/v0/apply${qs}`,
        {
          method: 'POST',
          body: yamlBody,
          headers: {
            'Content-Type': 'application/yaml',
            Accept: 'application/json',
            'X-Agent-Auth-Actor': JSON.stringify(actor),
          },
        },
        fetchImpl,
      )
    }

    case 'delete_agent': {
      const name = String(a.name || '')
      if (!name) throw new Error('name is required')
      const tag = a.tag ? String(a.tag) : undefined
      const path = tag
        ? `/v0/agents/${encodeURIComponent(name)}/${encodeURIComponent(tag)}`
        : `/v0/agents/${encodeURIComponent(name)}`
      return registryFetch(registryBaseUrl, path, { method: 'DELETE' }, fetchImpl)
    }

    case 'registry_health':
      return registryFetch(registryBaseUrl, '/v0/health', { method: 'GET' }, fetchImpl)

    case 'registry_version':
      return registryFetch(registryBaseUrl, '/v0/version', { method: 'GET' }, fetchImpl)

    default:
      throw new Error(`Unsupported capability: ${capability}`)
  }
}

/** Capability names only — useful for grants / admin agent metadata. */
export function adminCapabilityNames(): string[] {
  return REGISTRY_ADMIN_CAPABILITIES.map((c) => c.name)
}
