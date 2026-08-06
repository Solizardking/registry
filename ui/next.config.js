/** @type {import('next').NextConfig} */
const nextConfig = {
  images: {
    unoptimized: true,
  },
  // better-sqlite3 is a native addon used by Better Auth (Agent Auth provider).
  serverExternalPackages: ['better-sqlite3'],
  async rewrites() {
    return [
      {
        source: '/admin/:path*',
        destination: 'http://localhost:12121/admin/:path*',
      },
      {
        source: '/v0/:path*',
        destination: 'http://localhost:12121/v0/:path*',
      },
      {
        source: '/v0.1/:path*',
        destination: 'http://localhost:12121/v0.1/:path*',
      },
      // Note: App Router filesystem routes (e.g. /api/auth/*) take precedence
      // over rewrites, so Better Auth is not proxied to the Go registry.
      {
        source: '/api/:path*',
        destination: 'http://localhost:12121/api/:path*',
      },
    ]
  },
}

// Only use static export for production builds
if (process.env.NEXT_BUILD_EXPORT === 'true') {
  nextConfig.output = 'export'
  // Disable trailingSlash for static export to avoid redirect loops
  nextConfig.trailingSlash = false
  // Remove rewrites for static export (not supported)
  delete nextConfig.rewrites
}

module.exports = nextConfig

