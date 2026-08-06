/**
 * Agent Auth Protocol discovery document.
 *
 * Must live at /.well-known/agent-configuration (app root), even though
 * Better Auth itself is mounted under /api/auth.
 */

import { auth } from '@/lib/auth'
import { NextResponse } from 'next/server'

export async function GET(request: Request) {
  try {
    const configuration = await auth.api.getAgentConfiguration({
      headers: request.headers,
    })

    return NextResponse.json(configuration, {
      headers: {
        'Cache-Control': 'public, max-age=60',
      },
    })
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err)
    return NextResponse.json(
      { error: 'failed to load agent configuration', detail: message },
      { status: 500 },
    )
  }
}
