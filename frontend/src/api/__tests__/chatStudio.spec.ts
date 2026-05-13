import { beforeEach, describe, expect, it, vi } from 'vitest'
import {
  ChatStudioError,
  createChatCompletionStream,
  extractChatStudioDelta,
  isAbortError,
} from '@/api/chatStudio'

function streamFromChunks(chunks: string[]): ReadableStream<Uint8Array> {
  const encoder = new TextEncoder()
  return new ReadableStream<Uint8Array>({
    start(controller) {
      for (const chunk of chunks) {
        controller.enqueue(encoder.encode(chunk))
      }
      controller.close()
    },
  })
}

function okStream(chunks: string[]): Response {
  return new Response(streamFromChunks(chunks), { status: 200 })
}

describe('chatStudio api', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
  })

  it('extracts deltas from OpenAI-compatible chat chunks', () => {
    expect(extractChatStudioDelta({
      choices: [{ delta: { content: '你好' } }],
    })).toBe('你好')
    expect(extractChatStudioDelta({
      choices: [{ message: { content: '完成' } }],
    })).toBe('完成')
  })

  it('streams multiple SSE chunks until DONE', async () => {
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue(okStream([
      'data: {"choices":[{"delta":{"content":"你"}}]}\n\n',
      'data: {"choices":[{"delta":{"content":"好"}}]}\n\n',
      'data: [DONE]\n\n',
    ]))
    const onDelta = vi.fn()

    const result = await createChatCompletionStream({
      apiKey: 'sk-test',
      model: 'gpt-5.4-mini',
      messages: [{ role: 'user', content: 'hi' }],
      onDelta,
    })

    expect(result.content).toBe('你好')
    expect(onDelta).toHaveBeenNthCalledWith(1, '你')
    expect(onDelta).toHaveBeenNthCalledWith(2, '好')
    expect(fetchMock).toHaveBeenCalledWith('/v1/chat/completions', expect.objectContaining({
      method: 'POST',
      headers: expect.objectContaining({
        Authorization: 'Bearer sk-test',
      }),
      body: JSON.stringify({
        model: 'gpt-5.4-mini',
        messages: [{ role: 'user', content: 'hi' }],
        stream: true,
      }),
    }))
  })

  it('handles SSE events split across transport chunks', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(okStream([
      'data: {"choices":[{"delta"',
      ':{"content":"拆"}}]}\n\n',
      'data: {"choices":[{"delta":{"content":"分"}}]}\n\n',
      'data: [DONE]\n\n',
    ]))

    const result = await createChatCompletionStream({
      apiKey: 'sk-test',
      model: 'gpt-5.4-mini',
      messages: [{ role: 'user', content: 'hi' }],
    })

    expect(result.content).toBe('拆分')
  })

  it('throws gateway error messages for non-2xx responses', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response(
      JSON.stringify({ error: { message: 'rate limited', code: 'rate_limit' } }),
      { status: 429, statusText: 'Too Many Requests' },
    ))

    await expect(createChatCompletionStream({
      apiKey: 'sk-test',
      model: 'gpt-5.4-mini',
      messages: [{ role: 'user', content: 'hi' }],
    })).rejects.toMatchObject<Partial<ChatStudioError>>({
      name: 'ChatStudioError',
      message: 'rate limited',
      status: 429,
      code: 'rate_limit',
    })
  })

  it('throws ChatStudioError when a streaming event carries an error payload', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(okStream([
      'data: {"error":{"message":"upstream overloaded","code":"overloaded"}}\n\n',
      'data: [DONE]\n\n',
    ]))

    await expect(createChatCompletionStream({
      apiKey: 'sk-test',
      model: 'gpt-5.4-mini',
      messages: [{ role: 'user', content: 'hi' }],
    })).rejects.toMatchObject<Partial<ChatStudioError>>({
      name: 'ChatStudioError',
      message: 'upstream overloaded',
      code: 'overloaded',
    })
  })

  it('preserves AbortError detection', async () => {
    const abortError = new DOMException('aborted', 'AbortError')
    vi.spyOn(globalThis, 'fetch').mockRejectedValue(abortError)

    await expect(createChatCompletionStream({
      apiKey: 'sk-test',
      model: 'gpt-5.4-mini',
      messages: [{ role: 'user', content: 'hi' }],
    })).rejects.toBe(abortError)
    expect(isAbortError(abortError)).toBe(true)
  })
})
