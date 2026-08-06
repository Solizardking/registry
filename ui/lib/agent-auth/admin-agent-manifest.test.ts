/**
 * Structural test: the shipped admin agent example targets the agents repo
 * and documents the same capability names as the Agent Auth provider.
 */

import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

import { adminCapabilityNames } from './capabilities'

const ROOT = resolve(__dirname, '../../..')
const MANIFEST = resolve(ROOT, 'examples/admin-agent.yaml')

describe('examples/admin-agent.yaml', () => {
  const text = readFileSync(MANIFEST, 'utf8')

  it('is an Agent named registry-admin', () => {
    expect(text).toMatch(/kind:\s*Agent/)
    expect(text).toMatch(/name:\s*registry-admin/)
    expect(text).toMatch(/role:\s*admin/)
  })

  it('points source.repository.url at Solizardking/agents', () => {
    expect(text).toMatch(
      /url:\s*https:\/\/github\.com\/Solizardking\/agents/i,
    )
  })

  it('mentions each admin capability name from the shipped provider', () => {
    for (const name of adminCapabilityNames()) {
      expect(text, `manifest should document capability ${name}`).toContain(name)
    }
  })
})
