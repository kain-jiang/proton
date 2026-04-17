import { describe, expect, it } from 'vitest'

import { defaultWizardState } from '../../schema/defaults'
import { exportConfigAsJson, exportConfigAsYaml } from './export'

describe('preview export', () => {
  it('omits deploy from json output', () => {
    const output = exportConfigAsJson(defaultWizardState)

    expect(output).not.toContain('"deploy"')
    expect(output).toContain('"component_management": {}')
  })

  it('writes expected top-level keys in yaml output', () => {
    const output = exportConfigAsYaml({
      ...defaultWizardState,
      resource_connect_info: {
        rds: {
          ...defaultWizardState.resource_connect_info.rds,
          source_type: 'internal',
        },
        redis: {
          ...defaultWizardState.resource_connect_info.redis,
          source_type: 'internal',
        },
        opensearch: {
          ...defaultWizardState.resource_connect_info.opensearch,
          source_type: 'internal',
        },
        mq: {
          ...defaultWizardState.resource_connect_info.mq,
          source_type: 'internal',
        },
      },
    })

    expect(output).toContain('apiVersion: v1')
    expect(output).toContain('component_management: {}')
    expect(output).toContain('proton_mariadb:')
    expect(output).not.toContain('deploy:')
  })
})
