import { useState } from 'react'

import type { WizardState } from '../../schema/config'
import { pollServerResult, submitToServer } from './api'

type SubmitStatus = 'idle' | 'submitting' | 'success' | 'error'

interface UseSubmitOptions {
  pollIntervalMs?: number
}

interface SubmitState {
  status: SubmitStatus
  message: string
}

export function useSubmit({ pollIntervalMs = 1500 }: UseSubmitOptions = {}) {
  const [state, setState] = useState<SubmitState>({
    status: 'idle',
    message: '',
  })

  async function submit(wizardState: WizardState) {
    setState({
      status: 'submitting',
      message: '',
    })

    try {
      await submitToServer(wizardState)
      const result = await pollServerResult({ pollIntervalMs })
      setState({
        status: 'success',
        message: result,
      })
    } catch (error) {
      setState({
        status: 'error',
        message: error instanceof Error ? error.message : 'Initialization failed.',
      })
    }
  }

  function reset() {
    setState({
      status: 'idle',
      message: '',
    })
  }

  return {
    ...state,
    submit,
    reset,
  }
}
