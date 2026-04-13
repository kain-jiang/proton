import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { defaultWizardState } from '../../schema/defaults'
import { Wizard } from './Wizard'

afterEach(() => {
  vi.unstubAllGlobals()
})

async function fillRequiredServicePasswords(user: ReturnType<typeof userEvent.setup>) {
  const passwordInputs = screen
    .getAllByDisplayValue('')
    .filter((element): element is HTMLInputElement => element instanceof HTMLInputElement && element.type === 'password')

  await user.type(passwordInputs[0], 'mariadb-pass')
  await user.type(passwordInputs[1], 'redis-pass')
}

describe('Wizard', () => {
  it('starts on a template chooser before any step forms render', () => {
    render(<Wizard />)

    expect(screen.getByRole('heading', { name: '初始化模板' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: '标准模式部署' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: '托管Kubernetes部署' })).toBeDisabled()
    expect(screen.getByText('功能持续完善中')).toBeInTheDocument()
    expect(screen.queryByText('节点配置')).not.toBeInTheDocument()
  })

  it('enters the old local step flow after choosing standard mode', async () => {
    const user = userEvent.setup()
    render(<Wizard />)

    await user.click(screen.getByRole('button', { name: '标准模式部署' }))

    expect(screen.getByText('节点配置')).toBeInTheDocument()
    expect(screen.getByText('kubernetes配置')).toBeInTheDocument()
    expect(screen.getByText('仓库配置')).toBeInTheDocument()
    expect(screen.getByText('基础服务配置')).toBeInTheDocument()
    expect(screen.getByText('连接配置')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: '下一步' })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: '上一步' })).not.toBeInTheDocument()
  })

  it('does not enter the managed step flow while the template is disabled', async () => {
    const user = userEvent.setup()
    render(<Wizard />)

    await user.click(screen.getByRole('button', { name: '托管Kubernetes部署' }))

    expect(screen.queryByText('节点配置')).not.toBeInTheDocument()
    expect(screen.queryByText('kubernetes配置')).not.toBeInTheDocument()
    expect(screen.queryByText('仓库配置')).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: '下一步' })).not.toBeInTheDocument()
    expect(screen.getByRole('heading', { name: '初始化模板' })).toBeInTheDocument()
  })

  it('defaults kubernetes addons to ingress-nginx only', () => {
    expect(defaultWizardState.cs.local.addons).toEqual(['ingress-nginx'])
    expect(defaultWizardState.cs.managed.addons).toEqual(['ingress-nginx'])
  })

  it('renders global node settings above a multi-node table and supports add/remove', async () => {
    const user = userEvent.setup()
    render(<Wizard />)

    await user.click(screen.getByRole('button', { name: '标准模式部署' }))

    expect(screen.getByText('时间同步模式')).toBeInTheDocument()
    expect(screen.getByText('防火墙模式')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: '新增节点' })).toBeInTheDocument()
    expect(screen.getByRole('columnheader', { name: '节点名称' })).toBeInTheDocument()
    expect(screen.getByRole('columnheader', { name: 'IPv4地址' })).toBeInTheDocument()
    expect(screen.getByRole('columnheader', { name: 'IPv6地址' })).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: '新增节点' }))
    expect(screen.getAllByLabelText('节点名称').length).toBe(1)
    expect(screen.getAllByLabelText('Node IPv4').length).toBe(1)

    await user.click(screen.getByRole('button', { name: '新增节点' }))
    expect(screen.getAllByLabelText('节点名称').length).toBe(2)
    expect(screen.getAllByLabelText('Node IPv4').length).toBe(2)

    await user.click(screen.getAllByRole('button', { name: '删除' })[1])
    expect(screen.getAllByLabelText('节点名称').length).toBe(1)
  })

  it('keeps chrony and firewall controls in one compact node settings section', async () => {
    const user = userEvent.setup()
    render(<Wizard />)

    await user.click(screen.getByRole('button', { name: '标准模式部署' }))

    expect(screen.getByRole('heading', { name: '节点通用配置' })).toBeInTheDocument()
    expect(screen.queryByRole('heading', { name: '时间同步配置' })).not.toBeInTheDocument()
    expect(screen.queryByRole('heading', { name: '防火墙配置' })).not.toBeInTheDocument()
    expect(screen.getByText('时间同步模式')).toBeInTheDocument()
    expect(screen.getByText('防火墙模式')).toBeInTheDocument()
  })

  it('does not show the service package path in the standard flow', async () => {
    const user = userEvent.setup()
    render(<Wizard />)

    await user.click(screen.getByRole('button', { name: '标准模式部署' }))

    expect(screen.queryByLabelText('Service Package 路径')).not.toBeInTheDocument()
    expect(screen.queryByText('cs_controller_dir')).not.toBeInTheDocument()
  })

  it('does not show the service package path in the managed flow', async () => {
    const user = userEvent.setup()
    render(<Wizard />)

    await user.click(screen.getByRole('button', { name: '托管Kubernetes部署' }))

    expect(screen.queryByLabelText('Service Package 路径')).not.toBeInTheDocument()
    expect(screen.getByRole('heading', { name: '初始化模板' })).toBeInTheDocument()
  })

  it('shows a submitting state after clicking 完成 and before the server finishes', async () => {
    let resolveResultPoll!: (value: Response) => void
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        new Response('accepted', {
          status: 200,
        }),
      )
      .mockImplementationOnce(
        () =>
          new Promise<Response>((resolve) => {
            resolveResultPoll = resolve
          }),
      )

    vi.stubGlobal('fetch', fetchMock)

    const user = userEvent.setup()
    render(<Wizard />)

    await user.click(screen.getByRole('button', { name: '标准模式部署' }))
    await user.click(screen.getByRole('button', { name: '新增节点' }))
    await user.type(screen.getByLabelText('Node IPv4'), '192.168.40.11')
    await user.click(screen.getByRole('button', { name: '下一步' }))
    await user.click(screen.getByRole('button', { name: '下一步' }))
    await user.click(screen.getByRole('button', { name: '下一步' }))
    await fillRequiredServicePasswords(user)
    await user.click(screen.getByRole('button', { name: '下一步' }))
    await user.click(screen.getByRole('button', { name: '完成' }))

    expect(screen.getByRole('heading', { name: '正在提交初始化请求' })).toBeInTheDocument()
    expect(screen.getByText('正在向 proton-cli server 提交配置，请稍候。')).toBeInTheDocument()

    resolveResultPoll(
      new Response('Success', {
        status: 200,
      }),
    )

    expect(await screen.findByRole('heading', { name: '集群初始化成功' })).toBeInTheDocument()
  })

  it('keeps footer actions and shows an error dialog after initialization fails so retry can continue', async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        new Response('accepted', {
          status: 200,
        }),
      )
      .mockResolvedValueOnce(
        new Response('root cause failed', {
          status: 200,
        }),
      )
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

    const user = userEvent.setup()
    render(<Wizard />)

    await user.click(screen.getByRole('button', { name: '标准模式部署' }))
    await user.click(screen.getByRole('button', { name: '新增节点' }))
    await user.type(screen.getByLabelText('Node IPv4'), '192.168.40.11')
    await user.click(screen.getByRole('button', { name: '下一步' }))
    await user.click(screen.getByRole('button', { name: '下一步' }))
    await user.click(screen.getByRole('button', { name: '下一步' }))
    await fillRequiredServicePasswords(user)
    await user.click(screen.getByRole('button', { name: '下一步' }))
    await user.click(screen.getByRole('button', { name: '完成' }))

    expect(await screen.findByRole('dialog', { name: '初始化失败' })).toBeInTheDocument()
    expect(screen.getByText('root cause failed')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: '上一步' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: '完成' })).toBeInTheDocument()
    expect(screen.getByText('连接配置')).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: '完成' }))

    expect(await screen.findByRole('heading', { name: '集群初始化成功' })).toBeInTheDocument()
    expect(fetchMock).toHaveBeenCalledTimes(4)
  })
})
