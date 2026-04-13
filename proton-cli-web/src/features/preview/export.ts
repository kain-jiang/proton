import YAML from 'yaml'

import type { WizardState } from '../../schema/config'
import { toSubmitConfig } from '../../schema/mappers'

export function exportConfigAsJson(state: WizardState) {
  return JSON.stringify(toSubmitConfig(state), null, 2)
}

export function exportConfigAsYaml(state: WizardState) {
  return YAML.stringify(toSubmitConfig(state), {
    aliasDuplicateObjects: false,
    lineWidth: 0,
  })
}
