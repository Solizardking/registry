/**
 * Device authorization approval page for Agent Auth.
 *
 * Plugin default deviceAuthorizationPage is "/device/capabilities".
 * Query params: agent_id, code (and optional user_code).
 *
 * Minimal operator surface — wire approveCapability / device flows for production.
 */

type SearchParams = Record<string, string | string[] | undefined>

function first(value: string | string[] | undefined): string {
  if (Array.isArray(value)) return value[0] || ''
  return value || ''
}

export default async function DeviceCapabilitiesPage({
  searchParams,
}: {
  searchParams?: Promise<SearchParams> | SearchParams
}) {
  const params = (await Promise.resolve(searchParams)) || {}
  const agentId = first(params.agent_id ?? params.agentId)
  const code = first(params.code ?? params.user_code)

  return (
    <main className="mx-auto max-w-lg p-8 font-sans">
      <h1 className="text-2xl font-semibold tracking-tight">
        Agent capability approval
      </h1>
      <p className="mt-2 text-sm text-muted-foreground">
        An AI agent is requesting capability grants. Sign in as an admin
        operator, then approve or deny the request.
      </p>
      <dl className="mt-6 space-y-2 rounded-md border p-4 text-sm">
        <div>
          <dt className="font-medium">Agent ID</dt>
          <dd className="break-all font-mono">{agentId || '—'}</dd>
        </div>
        <div>
          <dt className="font-medium">User code</dt>
          <dd className="break-all font-mono">{code || '—'}</dd>
        </div>
      </dl>
      <p className="mt-4 text-xs text-muted-foreground">
        See <code>docs/agent-auth.md</code> for capability list and setup.
      </p>
    </main>
  )
}
