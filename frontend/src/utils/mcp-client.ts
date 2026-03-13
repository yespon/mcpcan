/**
 * 轻量级 MCP 前端客户端
 * 直接通过浏览器 fetch 实现 MCP JSON-RPC 协议，无需后端中转。
 *
 * 支持协议：
 *   - Streamable HTTP（POST 模式，新标准）
 *   - SSE（GET 建连 + POST 消息，旧标准）
 */

export interface McpTool {
  name: string
  description?: string
  inputSchema?: {
    type: string
    properties?: Record<string, any>
    required?: string[]
    [key: string]: any
  }
}

export interface McpCallResult {
  content: Array<{ type: string; text?: string; data?: string; mimeType?: string }>
  isError?: boolean
}

interface JsonRpcRequest {
  jsonrpc: '2.0'
  id: number
  method: string
  params?: any
}

interface JsonRpcResponse {
  jsonrpc: '2.0'
  id: number
  result?: any
  error?: { code: number; message: string; data?: any }
}

/**
 * 判断是否是 SSE 协议 URL（以 /sse 结尾）
 */
function isSSEUrl(url: string): boolean {
  return /\/sse\s*$/.test(url.split('?')[0])
}

/**
 * 构建请求 Headers
 */
function buildHeaders(token?: string): Record<string, string> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    Accept: 'application/json, text/event-stream',
  }
  if (token) {
    // 支持 "Bearer xxx" 或直接 token 字符串
    headers['Authorization'] = token.startsWith('Bearer ') ? token : `Bearer ${token}`
  }
  return headers
}

/**
 * 解析 SSE 流，返回第一条 JSON 消息
 */
async function readFirstSSEMessage(response: Response): Promise<any> {
  const reader = response.body?.getReader()
  if (!reader) throw new Error('No response body')
  const decoder = new TextDecoder()
  let buffer = ''
  while (true) {
    const { done, value } = await reader.read()
    if (done) break
    buffer += decoder.decode(value, { stream: true })
    // 解析 SSE 事件
    const lines = buffer.split('\n')
    for (let i = 0; i < lines.length; i++) {
      const line = lines[i].trim()
      if (line.startsWith('data:')) {
        const data = line.slice(5).trim()
        if (data && data !== '[DONE]') {
          try {
            reader.cancel()
            return JSON.parse(data)
          } catch {}
        }
      }
    }
    buffer = lines[lines.length - 1]
  }
  throw new Error('SSE stream ended without message')
}

/**
 * Streamable HTTP 模式：发送单条 JSON-RPC 请求并返回结果
 */
async function sendStreamableHttp(
  url: string,
  request: JsonRpcRequest,
  token?: string,
  sessionId?: string,
): Promise<JsonRpcResponse> {
  const headers = buildHeaders(token)
  if (sessionId) headers['Mcp-Session-Id'] = sessionId

  const resp = await fetch(url, {
    method: 'POST',
    headers,
    body: JSON.stringify(request),
  })

  if (!resp.ok) {
    throw new Error(`HTTP ${resp.status}: ${resp.statusText}`)
  }

  const contentType = resp.headers.get('content-type') || ''

  // 处理 SSE 流式响应
  if (contentType.includes('text/event-stream')) {
    return await readFirstSSEMessage(resp)
  }

  // 处理普通 JSON 响应
  return await resp.json()
}

/**
 * SSE 模式客户端（基于 fetch + ReadableStream）
 * 优势：支持自定义 Header（Authorization），EventSource API 不支持。
 * 流程：fetch GET /sse → 读流获取 endpoint 事件 → 持续读流接收响应 → POST endpoint 发消息
 */
class SSEMcpClient {
  private endpoint: string = ''
  private reader: ReadableStreamDefaultReader<Uint8Array> | null = null
  private decoder = new TextDecoder()
  private pendingCallbacks = new Map<number, (msg: JsonRpcResponse) => void>()
  private closed = false

  constructor(
    private sseUrl: string,
    private token?: string,
  ) {}

  async connect(): Promise<void> {
    const headers: Record<string, string> = {
      Accept: 'text/event-stream',
      'Cache-Control': 'no-cache',
    }
    if (this.token) {
      headers['Authorization'] = this.token.startsWith('Bearer ')
        ? this.token
        : `Bearer ${this.token}`
    }

    const resp = await fetch(this.sseUrl, { headers })
    if (!resp.ok || !resp.body) {
      throw new Error(`SSE connect failed: HTTP ${resp.status} ${resp.statusText}`)
    }

    this.reader = resp.body.getReader()

    // 等待直到收到 endpoint 事件
    await this._waitForEndpoint()

    // 后台持续读流，派发响应给 pending callbacks
    this._readLoop()
  }

  /** 解析 SSE 流，找到 endpoint event，返回后端 POST 地址 */
  private async _waitForEndpoint(): Promise<void> {
    let buffer = ''
    const deadline = Date.now() + 8000

    while (Date.now() < deadline) {
      const { done, value } = await this.reader!.read()
      if (done) throw new Error('SSE stream closed before endpoint received')

      buffer += this.decoder.decode(value, { stream: true })
      const events = this._parseSSEBuffer(buffer)
      buffer = events.remainder

      for (const ev of events.parsed) {
        if (ev.event === 'endpoint' && ev.data) {
          const path = ev.data.trim()
          const base = new URL(this.sseUrl)
          this.endpoint = path.startsWith('http') ? path : `${base.origin}${path}`
          return
        }
      }
    }
    throw new Error('SSE endpoint event timeout')
  }

  /** 持续读 SSE 流，将 JSON-RPC 响应派发给等待的 callback */
  private async _readLoop(): Promise<void> {
    let buffer = ''
    while (!this.closed) {
      let done = false
      let value: Uint8Array | undefined
      try {
        ;({ done, value } = await this.reader!.read())
      } catch {
        break
      }
      if (done || !value) break

      buffer += this.decoder.decode(value, { stream: true })
      const events = this._parseSSEBuffer(buffer)
      buffer = events.remainder

      for (const ev of events.parsed) {
        if (!ev.data) continue
        try {
          const msg: JsonRpcResponse = JSON.parse(ev.data)
          const cb = this.pendingCallbacks.get(msg.id)
          if (cb) {
            this.pendingCallbacks.delete(msg.id)
            cb(msg)
          }
        } catch {}
      }
    }
  }

  /** 解析 SSE 文本块，返回完整事件列表和剩余未解析的尾部 */
  private _parseSSEBuffer(raw: string): {
    parsed: Array<{ event: string; data: string }>
    remainder: string
  } {
    const parsed: Array<{ event: string; data: string }> = []
    const blocks = raw.split(/\n\n/)
    // 最后一个可能不完整，保留
    const remainder = blocks.pop() ?? ''

    for (const block of blocks) {
      let event = 'message'
      const dataLines: string[] = []
      for (const line of block.split('\n')) {
        if (line.startsWith('event:')) {
          event = line.slice(6).trim()
        } else if (line.startsWith('data:')) {
          dataLines.push(line.slice(5).trim())
        }
      }
      if (dataLines.length) {
        parsed.push({ event, data: dataLines.join('\n') })
      }
    }
    return { parsed, remainder }
  }

  async send(request: JsonRpcRequest): Promise<JsonRpcResponse> {
    return new Promise(async (resolve, reject) => {
      this.pendingCallbacks.set(request.id, resolve)

      const headers = buildHeaders(this.token)

      try {
        const resp = await fetch(this.endpoint, {
          method: 'POST',
          headers,
          body: JSON.stringify(request),
        })
        if (!resp.ok) {
          this.pendingCallbacks.delete(request.id)
          reject(new Error(`POST to endpoint failed: HTTP ${resp.status}`))
          return
        }
      } catch (e) {
        this.pendingCallbacks.delete(request.id)
        reject(e)
        return
      }

      // 15s 超时
      setTimeout(() => {
        if (this.pendingCallbacks.has(request.id)) {
          this.pendingCallbacks.delete(request.id)
          reject(new Error('MCP request timeout'))
        }
      }, 15000)
    })
  }

  close() {
    this.closed = true
    try {
      this.reader?.cancel()
    } catch {}
    this.reader = null
  }
}


/**
 * 主入口：获取工具列表
 */
export async function mcpListTools(mcpServerUrl: string, token?: string): Promise<McpTool[]> {
  if (isSSEUrl(mcpServerUrl)) {
    return listToolsViaSSE(mcpServerUrl, token)
  }
  return listToolsViaStreamable(mcpServerUrl, token)
}

/**
 * 主入口：调用工具
 */
export async function mcpCallTool(
  mcpServerUrl: string,
  toolName: string,
  args: Record<string, any>,
  token?: string,
): Promise<McpCallResult> {
  if (isSSEUrl(mcpServerUrl)) {
    return callToolViaSSE(mcpServerUrl, toolName, args, token)
  }
  return callToolViaStreamable(mcpServerUrl, toolName, args, token)
}

// ─── Streamable HTTP ────────────────────────────────────────────────────────

async function listToolsViaStreamable(url: string, token?: string): Promise<McpTool[]> {
  let sessionId = ''

  // Step 1: initialize
  const initResp = await sendStreamableHttp(
    url,
    {
      jsonrpc: '2.0',
      id: 1,
      method: 'initialize',
      params: {
        protocolVersion: '2024-11-05',
        capabilities: { tools: {} },
        clientInfo: { name: 'mcpcan-debug', version: '1.0.0' },
      },
    },
    token,
  )
  if (initResp.error) throw new Error(initResp.error.message)
  sessionId = initResp.result?.sessionId || ''

  // Step 2: tools/list
  const listResp = await sendStreamableHttp(
    url,
    { jsonrpc: '2.0', id: 2, method: 'tools/list', params: {} },
    token,
    sessionId,
  )
  if (listResp.error) throw new Error(listResp.error.message)

  return listResp.result?.tools || []
}

async function callToolViaStreamable(
  url: string,
  toolName: string,
  args: Record<string, any>,
  token?: string,
): Promise<McpCallResult> {
  let sessionId = ''

  // initialize
  const initResp = await sendStreamableHttp(
    url,
    {
      jsonrpc: '2.0',
      id: 1,
      method: 'initialize',
      params: {
        protocolVersion: '2024-11-05',
        capabilities: { tools: {} },
        clientInfo: { name: 'mcpcan-debug', version: '1.0.0' },
      },
    },
    token,
  )
  if (initResp.error) throw new Error(initResp.error.message)
  sessionId = initResp.result?.sessionId || ''

  // tools/call
  const callResp = await sendStreamableHttp(
    url,
    {
      jsonrpc: '2.0',
      id: 2,
      method: 'tools/call',
      params: { name: toolName, arguments: args },
    },
    token,
    sessionId,
  )
  if (callResp.error) throw new Error(callResp.error.message)

  return callResp.result as McpCallResult
}

// ─── SSE ─────────────────────────────────────────────────────────────────────

async function listToolsViaSSE(sseUrl: string, token?: string): Promise<McpTool[]> {
  const client = new SSEMcpClient(sseUrl, token)
  try {
    await client.connect()

    // initialize
    const initResp = await client.send({
      jsonrpc: '2.0',
      id: 1,
      method: 'initialize',
      params: {
        protocolVersion: '2024-11-05',
        capabilities: { tools: {} },
        clientInfo: { name: 'mcpcan-debug', version: '1.0.0' },
      },
    })
    if (initResp.error) throw new Error(initResp.error.message)

    // tools/list
    const listResp = await client.send({
      jsonrpc: '2.0',
      id: 2,
      method: 'tools/list',
      params: {},
    })
    if (listResp.error) throw new Error(listResp.error.message)

    return listResp.result?.tools || []
  } finally {
    client.close()
  }
}

async function callToolViaSSE(
  sseUrl: string,
  toolName: string,
  args: Record<string, any>,
  token?: string,
): Promise<McpCallResult> {
  const client = new SSEMcpClient(sseUrl, token)
  try {
    await client.connect()

    const initResp = await client.send({
      jsonrpc: '2.0',
      id: 1,
      method: 'initialize',
      params: {
        protocolVersion: '2024-11-05',
        capabilities: { tools: {} },
        clientInfo: { name: 'mcpcan-debug', version: '1.0.0' },
      },
    })
    if (initResp.error) throw new Error(initResp.error.message)

    const callResp = await client.send({
      jsonrpc: '2.0',
      id: 2,
      method: 'tools/call',
      params: { name: toolName, arguments: args },
    })
    if (callResp.error) throw new Error(callResp.error.message)

    return callResp.result as McpCallResult
  } finally {
    client.close()
  }
}
