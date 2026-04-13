import type { WizardState } from '../../schema/config'
import { exportConfigAsJson, exportConfigAsYaml } from './export'

interface PreviewPanelProps {
  state: WizardState
  format: 'json' | 'yaml'
  onFormatChange: (format: 'json' | 'yaml') => void
}

export function PreviewPanel({ state, format, onFormatChange }: PreviewPanelProps) {
  const content = format === 'json' ? exportConfigAsJson(state) : exportConfigAsYaml(state)

  return (
    <section className="legacy-card legacy-preview-card">
      <div className="legacy-preview-header">
        <div>
          <p className="legacy-muted">配置预览</p>
          <h3>生成结果</h3>
        </div>
        <div className="legacy-segmented-control" role="tablist" aria-label="Preview format">
          <button
            type="button"
            className={format === 'json' ? 'is-active' : ''}
            onClick={() => onFormatChange('json')}
          >
            JSON
          </button>
          <button
            type="button"
            className={format === 'yaml' ? 'is-active' : ''}
            onClick={() => onFormatChange('yaml')}
          >
            YAML
          </button>
        </div>
      </div>
      <pre className="legacy-preview-code">{content}</pre>
    </section>
  )
}
