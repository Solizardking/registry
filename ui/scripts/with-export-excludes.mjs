#!/usr/bin/env node
/**
 * Static export (NEXT_BUILD_EXPORT) cannot include App Router route handlers.
 * Temporarily move server-only Agent Auth routes out of app/ for the export
 * build used by Fly (embedded UI), then restore them.
 *
 * Usage: node scripts/with-export-excludes.mjs <command...>
 */
import { spawnSync } from 'node:child_process'
import { existsSync, mkdirSync, renameSync, rmSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

const uiRoot = join(dirname(fileURLToPath(import.meta.url)), '..')
const stashRoot = join(uiRoot, '.export-stash')

const PATHS = [
  'app/api',
  'app/.well-known',
  'app/device',
  // Auth modules import better-sqlite3 (native) — not needed for static catalog UI
]

function stash() {
  mkdirSync(stashRoot, { recursive: true })
  for (const rel of PATHS) {
    const from = join(uiRoot, rel)
    const to = join(stashRoot, rel)
    if (!existsSync(from)) continue
    mkdirSync(dirname(to), { recursive: true })
    if (existsSync(to)) rmSync(to, { recursive: true, force: true })
    renameSync(from, to)
    console.log(`[export-excludes] stashed ${rel}`)
  }
}

function restore() {
  for (const rel of PATHS) {
    const from = join(stashRoot, rel)
    const to = join(uiRoot, rel)
    if (!existsSync(from)) continue
    mkdirSync(dirname(to), { recursive: true })
    if (existsSync(to)) rmSync(to, { recursive: true, force: true })
    renameSync(from, to)
    console.log(`[export-excludes] restored ${rel}`)
  }
  rmSync(stashRoot, { recursive: true, force: true })
}

const cmd = process.argv.slice(2)
if (cmd.length === 0) {
  console.error('usage: node scripts/with-export-excludes.mjs <command...>')
  process.exit(2)
}

stash()
let code = 1
try {
  const result = spawnSync(cmd[0], cmd.slice(1), {
    cwd: uiRoot,
    stdio: 'inherit',
    env: process.env,
    shell: process.platform === 'win32',
  })
  code = result.status ?? 1
} finally {
  restore()
}
process.exit(code)
