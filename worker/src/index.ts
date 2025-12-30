/**
 * Quantlete Strava Proxy Worker
 *
 * Handles:
 * - OAuth token exchange (browser sends code, worker exchanges for token)
 * - API proxying (browser sends request with token, worker forwards to Strava)
 * - Token refresh
 */

interface Env {
  // Optional - for backwards compatibility. If not set, credentials must come from request body.
  STRAVA_CLIENT_ID?: string
  STRAVA_CLIENT_SECRET?: string
  STRAVA_API_BASE: string
  STRAVA_TOKEN_URL: string
  ALLOWED_ORIGIN?: string
}

interface StravaTokenResponse {
  token_type: string
  expires_at: number
  expires_in: number
  refresh_token: string
  access_token: string
  athlete: {
    id: number
    username: string
    firstname: string
    lastname: string
    profile_medium: string
    profile: string
  }
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url)
    const origin = request.headers.get('Origin') || ''

    // CORS preflight
    if (request.method === 'OPTIONS') {
      return corsResponse(env, origin, new Response(null, { status: 204 }))
    }

    try {
      let response: Response

      // Route handling
      if (url.pathname === '/oauth/exchange') {
        response = await handleOAuthExchange(request, env)
      } else if (url.pathname === '/oauth/refresh') {
        response = await handleTokenRefresh(request, env)
      } else if (url.pathname.startsWith('/api/')) {
        response = await handleApiProxy(request, env, url)
      } else if (url.pathname === '/health') {
        response = new Response(JSON.stringify({ status: 'ok' }), {
          headers: { 'Content-Type': 'application/json' },
        })
      } else {
        response = new Response(JSON.stringify({ error: 'Not found' }), {
          status: 404,
          headers: { 'Content-Type': 'application/json' },
        })
      }

      return corsResponse(env, origin, response)
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Unknown error'
      return corsResponse(
        env,
        origin,
        new Response(JSON.stringify({ error: message }), {
          status: 500,
          headers: { 'Content-Type': 'application/json' },
        })
      )
    }
  },
}

/**
 * Exchange OAuth authorization code for tokens.
 * POST /oauth/exchange
 * Body: { code: string, redirect_uri: string, client_id: string, client_secret: string }
 *
 * Client credentials can be provided in the request body or via env vars (for backwards compat).
 */
async function handleOAuthExchange(request: Request, env: Env): Promise<Response> {
  if (request.method !== 'POST') {
    return new Response(JSON.stringify({ error: 'Method not allowed' }), {
      status: 405,
      headers: { 'Content-Type': 'application/json' },
    })
  }

  const body = (await request.json()) as {
    code: string
    redirect_uri: string
    client_id?: string
    client_secret?: string
  }

  if (!body.code) {
    return new Response(JSON.stringify({ error: 'Missing code' }), {
      status: 400,
      headers: { 'Content-Type': 'application/json' },
    })
  }

  // Use credentials from request body, fall back to env vars for backwards compatibility
  const clientId = body.client_id || env.STRAVA_CLIENT_ID
  const clientSecret = body.client_secret || env.STRAVA_CLIENT_SECRET

  if (!clientId || !clientSecret) {
    return new Response(
      JSON.stringify({ error: 'Missing client credentials. Provide client_id and client_secret in request body.' }),
      {
        status: 400,
        headers: { 'Content-Type': 'application/json' },
      }
    )
  }

  const tokenResponse = await fetch(env.STRAVA_TOKEN_URL, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      client_id: clientId,
      client_secret: clientSecret,
      code: body.code,
      grant_type: 'authorization_code',
    }),
  })

  if (!tokenResponse.ok) {
    const error = await tokenResponse.text()
    return new Response(
      JSON.stringify({ error: 'Token exchange failed', details: error }),
      {
        status: tokenResponse.status,
        headers: { 'Content-Type': 'application/json' },
      }
    )
  }

  const tokenData = (await tokenResponse.json()) as StravaTokenResponse

  return new Response(JSON.stringify(tokenData), {
    headers: { 'Content-Type': 'application/json' },
  })
}

/**
 * Refresh an expired access token.
 * POST /oauth/refresh
 * Body: { refresh_token: string, client_id: string, client_secret: string }
 *
 * Client credentials can be provided in the request body or via env vars (for backwards compat).
 */
async function handleTokenRefresh(request: Request, env: Env): Promise<Response> {
  if (request.method !== 'POST') {
    return new Response(JSON.stringify({ error: 'Method not allowed' }), {
      status: 405,
      headers: { 'Content-Type': 'application/json' },
    })
  }

  const body = (await request.json()) as {
    refresh_token: string
    client_id?: string
    client_secret?: string
  }

  if (!body.refresh_token) {
    return new Response(JSON.stringify({ error: 'Missing refresh_token' }), {
      status: 400,
      headers: { 'Content-Type': 'application/json' },
    })
  }

  // Use credentials from request body, fall back to env vars for backwards compatibility
  const clientId = body.client_id || env.STRAVA_CLIENT_ID
  const clientSecret = body.client_secret || env.STRAVA_CLIENT_SECRET

  if (!clientId || !clientSecret) {
    return new Response(
      JSON.stringify({ error: 'Missing client credentials. Provide client_id and client_secret in request body.' }),
      {
        status: 400,
        headers: { 'Content-Type': 'application/json' },
      }
    )
  }

  const tokenResponse = await fetch(env.STRAVA_TOKEN_URL, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      client_id: clientId,
      client_secret: clientSecret,
      refresh_token: body.refresh_token,
      grant_type: 'refresh_token',
    }),
  })

  if (!tokenResponse.ok) {
    const error = await tokenResponse.text()
    return new Response(
      JSON.stringify({ error: 'Token refresh failed', details: error }),
      {
        status: tokenResponse.status,
        headers: { 'Content-Type': 'application/json' },
      }
    )
  }

  const tokenData = await tokenResponse.json()

  return new Response(JSON.stringify(tokenData), {
    headers: { 'Content-Type': 'application/json' },
  })
}

/**
 * Proxy API requests to Strava.
 * GET/POST /api/* → Strava API
 * Header: Authorization: Bearer <token>
 */
async function handleApiProxy(
  request: Request,
  env: Env,
  url: URL
): Promise<Response> {
  const authHeader = request.headers.get('Authorization')

  if (!authHeader?.startsWith('Bearer ')) {
    return new Response(JSON.stringify({ error: 'Missing authorization' }), {
      status: 401,
      headers: { 'Content-Type': 'application/json' },
    })
  }

  // Strip /api prefix and forward to Strava
  const stravaPath = url.pathname.replace(/^\/api/, '')
  const stravaUrl = `${env.STRAVA_API_BASE}${stravaPath}${url.search}`

  const stravaResponse = await fetch(stravaUrl, {
    method: request.method,
    headers: {
      Authorization: authHeader,
      'Content-Type': 'application/json',
    },
    body: request.method !== 'GET' ? await request.text() : undefined,
  })

  // Forward rate limit headers
  const responseHeaders = new Headers({
    'Content-Type': 'application/json',
  })

  const rateLimitLimit = stravaResponse.headers.get('X-RateLimit-Limit')
  const rateLimitUsage = stravaResponse.headers.get('X-RateLimit-Usage')

  if (rateLimitLimit) {
    responseHeaders.set('X-RateLimit-Limit', rateLimitLimit)
  }
  if (rateLimitUsage) {
    responseHeaders.set('X-RateLimit-Usage', rateLimitUsage)
  }

  return new Response(await stravaResponse.text(), {
    status: stravaResponse.status,
    headers: responseHeaders,
  })
}

/**
 * Add CORS headers to response.
 */
function corsResponse(env: Env, origin: string, response: Response): Response {
  const headers = new Headers(response.headers)

  // Allow the frontend origin or any origin in development
  const allowedOrigin = env.ALLOWED_ORIGIN || '*'
  headers.set('Access-Control-Allow-Origin', allowedOrigin === '*' ? origin || '*' : allowedOrigin)
  headers.set('Access-Control-Allow-Methods', 'GET, POST, OPTIONS')
  headers.set('Access-Control-Allow-Headers', 'Content-Type, Authorization')
  headers.set('Access-Control-Max-Age', '86400')

  return new Response(response.body, {
    status: response.status,
    statusText: response.statusText,
    headers,
  })
}
