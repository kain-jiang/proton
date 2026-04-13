import type { SubmitRequest, WizardState } from '../../schema/config'
import { toSubmitConfig } from '../../schema/mappers'

export interface SubmitApiOptions {
  pollIntervalMs?: number
}

function delay(ms: number) {
  return new Promise((resolve) => {
    window.setTimeout(resolve, ms)
  })
}

export async function submitToServer(state: WizardState) {
  const payload: SubmitRequest = {
    cluster_config: toSubmitConfig(state),
  }

  const response = await fetch('/init', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(payload),
  })

  if (!response.ok) {
    throw new Error(await response.text())
  }

  return response.text()
}

export async function pollServerResult({ pollIntervalMs = 1500 }: SubmitApiOptions = {}) {
  while (true) {
    const response = await fetch('/alpha/result', {
      method: 'GET',
    })

    if (response.status === 404) {
      await delay(pollIntervalMs)
      continue
    }

    const text = await response.text()
    if (!response.ok || text.toLowerCase().includes('fail')) {
      throw new Error(text)
    }

    return text
  }
}
