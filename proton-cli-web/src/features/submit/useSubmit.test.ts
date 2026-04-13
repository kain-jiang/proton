import { act, renderHook } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { defaultWizardState } from '../../schema/defaults'
import { useSubmit } from './useSubmit'

describe('useSubmit', () => {
  it('posts config to /init and polls /alpha/result', async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        new Response('accepted', {
          status: 200,
        }),
      )
      .mockResolvedValueOnce(
        new Response('Success', {
          status: 200,
        }),
      )

    vi.stubGlobal('fetch', fetchMock)

    const { result } = renderHook(() => useSubmit({ pollIntervalMs: 1 }))

    await act(async () => {
      await result.current.submit(defaultWizardState)
    })

    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      '/init',
      expect.objectContaining({
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
      }),
    )
    expect(JSON.parse(fetchMock.mock.calls[0]?.[1]?.body as string)).toMatchObject({
      cluster_config: {
        apiVersion: 'v1',
      },
    })
    expect(JSON.parse(fetchMock.mock.calls[0]?.[1]?.body as string)).not.toHaveProperty('service_package_dir')
    expect(JSON.parse(fetchMock.mock.calls[0]?.[1]?.body as string).cluster_config).not.toHaveProperty('deploy')
    expect(fetchMock).toHaveBeenNthCalledWith(2, '/alpha/result', expect.objectContaining({ method: 'GET' }))
    expect(result.current.status).toBe('success')

    vi.unstubAllGlobals()
  })
})
