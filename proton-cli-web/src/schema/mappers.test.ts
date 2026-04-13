import { describe, expect, it } from 'vitest'

import { defaultWizardState } from './defaults'
import { toSubmitConfig } from './mappers'

describe('toSubmitConfig', () => {
  it('uses internal mariadb root credentials for internal rds submissions', () => {
    const config = toSubmitConfig({
      ...defaultWizardState,
      resource_connect_info: {
        ...defaultWizardState.resource_connect_info,
        rds: {
          ...defaultWizardState.resource_connect_info.rds,
          source_type: 'internal',
          username: '',
          password: '',
        },
      },
      services: {
        ...defaultWizardState.services,
        mariadb: {
          ...defaultWizardState.services.mariadb,
          local: {
            ...defaultWizardState.services.mariadb.local,
            admin_user: 'root',
            admin_passwd: 'Proton123!',
          },
        },
      },
    })

    expect(config.resource_connect_info).toMatchObject({
      rds: {
        source_type: 'internal',
        username: 'root',
        password: 'Proton123!',
      },
    })
  })

  it('maps local deployment to standard mode with local cs and local cr', () => {
    const config = toSubmitConfig({
      ...defaultWizardState,
      resource_connect_info: {
        ...defaultWizardState.resource_connect_info,
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
          mq_type: 'kafka',
        },
      },
    })

    expect(config).not.toHaveProperty('deploy')
    expect(config.component_management).toEqual({})
    expect(config.cs).toMatchObject({ provisioner: 'local' })
    expect(config.cs).not.toHaveProperty('cs_controller_dir')
    expect(config.cr).toHaveProperty('local')
    expect(config.cr).not.toHaveProperty('external')
    expect(config.proton_mariadb).toMatchObject({ hosts: ['node1'] })
    expect(config).not.toHaveProperty('proton_mongodb')
  })

  it('keeps service package path out of cluster config payload', () => {
    const config = toSubmitConfig({
      ...defaultWizardState,
      service_package_dir: '/data/packages/service-package',
    })

    expect(config).not.toHaveProperty('service_package_dir')
    expect(config.cs).not.toHaveProperty('cs_controller_dir')
  })

  it('maps managed deployment to cloud mode with external cs and external cr', () => {
    const config = toSubmitConfig({
      ...defaultWizardState,
      deploymentKind: 'managed',
      resource_connect_info: {
        ...defaultWizardState.resource_connect_info,
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
          mq_type: 'kafka',
        },
      },
    })

    expect(config).not.toHaveProperty('deploy')
    expect(config.cs).toMatchObject({ provisioner: 'external' })
    expect(config.cr).toHaveProperty('external')
    expect(config.cr).not.toHaveProperty('local')
    expect(config.proton_mariadb).toMatchObject({ replica_count: 1 })
    expect(config.proton_mariadb).not.toHaveProperty('hosts')
  })

  it('omits empty kafka local optional fields from submitted config', () => {
    const config = toSubmitConfig({
      ...defaultWizardState,
      resource_connect_info: {
        ...defaultWizardState.resource_connect_info,
        mq: {
          ...defaultWizardState.resource_connect_info.mq,
          source_type: 'internal',
          mq_type: 'kafka',
        },
      },
    })

    expect(config.kafka).toMatchObject({
      hosts: ['node1'],
      data_path: '/sysvol/components/kafka',
      enabled: true,
    })
    expect(config.kafka).toMatchObject({
      env: {
        KAFKA_HEAP_OPTS: '-Xms1g -Xmx1g',
      },
    })
    expect(config.kafka).toEqual(
      expect.not.objectContaining({
        disable_external_service: expect.anything(),
        external_service_list: expect.anything(),
        resources: expect.anything(),
      }),
    )
    expect(config.kafka).toMatchObject({
      env: {
        KAFKA_HEAP_OPTS: '-Xms1g -Xmx1g',
      },
    })
    expect(config.kafka).not.toHaveProperty('disable_external_service')
    expect(config.kafka).not.toHaveProperty('external_service_list')
    expect(config.kafka).not.toHaveProperty('resources')
    expect(config.kafka).toEqual(
      expect.not.objectContaining({
        env: expect.objectContaining({
          KAFKA_LOG_RETENTION_BYTES: expect.anything(),
          KAFKA_LOG_RETENTION_HOURS: expect.anything(),
          KAFKA_LOG_ROLL_HOURS: expect.anything(),
        }),
      }),
    )
  })

  it('omits empty opensearch local optional fields from submitted config', () => {
    const config = toSubmitConfig({
      ...defaultWizardState,
      resource_connect_info: {
        ...defaultWizardState.resource_connect_info,
        opensearch: {
          ...defaultWizardState.resource_connect_info.opensearch,
          source_type: 'internal',
        },
      },
    })

    expect(config.opensearch).toMatchObject({
      hosts: ['node1'],
      data_path: '/sysvol/components/opensearch',
      mode: 'master',
      enabled: true,
      config: {
        jvmOptions: '-Xmx1g -Xms1g',
      },
    })
    expect(config.opensearch).not.toHaveProperty('storage_capacity')
    expect(config.opensearch).toEqual(
      expect.not.objectContaining({
        config: expect.objectContaining({
          hanlpRemoteextDict: expect.anything(),
          hanlpRemoteextStopwords: expect.anything(),
        }),
      }),
    )
  })

  it('omits opensearch storage capacity from submitted config even when provided', () => {
    const config = toSubmitConfig({
      ...defaultWizardState,
      resource_connect_info: {
        ...defaultWizardState.resource_connect_info,
        opensearch: {
          ...defaultWizardState.resource_connect_info.opensearch,
          source_type: 'internal',
        },
      },
      services: {
        ...defaultWizardState.services,
        opensearch: {
          ...defaultWizardState.services.opensearch,
          local: {
            ...defaultWizardState.services.opensearch.local,
            storage_capacity: '20Gi',
          },
        },
      },
    })

    expect(config.opensearch).toBeDefined()
    expect(config.opensearch).not.toHaveProperty('storage_capacity')
  })

  it('omits local service payloads when the corresponding resources are external', () => {
    const config = toSubmitConfig({
      ...defaultWizardState,
      resource_connect_info: {
        rds: {
          ...defaultWizardState.resource_connect_info.rds,
          source_type: 'external',
          rds_type: 'MySQL',
          hosts: '192.168.10.10',
          port: 3306,
          username: 'root',
          password: 'secret',
        },
        redis: {
          ...defaultWizardState.resource_connect_info.redis,
          source_type: 'external',
          connect_type: 'standalone',
          hosts: '192.168.10.11',
          port: 6379,
          username: 'default',
          password: 'secret',
        },
        opensearch: {
          ...defaultWizardState.resource_connect_info.opensearch,
          source_type: 'external',
          hosts: '192.168.10.12',
          port: 9200,
          username: 'admin',
          password: 'secret',
          version: '7.10.0',
        },
        mq: {
          ...defaultWizardState.resource_connect_info.mq,
          source_type: 'external',
          mq_type: 'kafka',
          mq_hosts: '192.168.10.13',
          mq_port: 9092,
          auth: {
            username: 'user',
            password: 'secret',
            mechanism: 'PLAIN',
          },
        },
      },
    })

    expect(config).not.toHaveProperty('proton_mariadb')
    expect(config).not.toHaveProperty('proton_redis')
    expect(config).not.toHaveProperty('opensearch')
    expect(config).not.toHaveProperty('kafka')
    expect(config).not.toHaveProperty('zookeeper')
    expect(config.resource_connect_info).toMatchObject({
      rds: { source_type: 'external' },
      redis: { source_type: 'external' },
      opensearch: { source_type: 'external' },
      mq: { source_type: 'external' },
    })
  })
})
