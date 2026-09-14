@description('Azure region used by the existing Container Apps environment.')
param location string = resourceGroup().location

@minLength(2)
@maxLength(18)
param namePrefix string = 'p7rl'

@description('Existing Azure Container Registry name.')
param registryName string

@description('Existing Container Apps environment name.')
param environmentName string

@description('Existing user-assigned identity with AcrPull on the registry.')
param pullIdentityName string

@description('Immutable image tag, normally the Git commit SHA.')
param imageTag string

@secure()
@description('Pooled Neon connection string used by the Go API.')
param databaseUrl string

@secure()
@description('Direct Neon connection string used only by the migration job.')
param databaseUrlDirect string

@secure()
@minLength(16)
@maxLength(128)
@description('Bearer token required by the control API and WebSocket protocol.')
param operatorToken string

@description('Resource tags applied to the release workloads.')
param tags object = {
  project: 'rl-simulation-control-platform'
  release: 'v1.0.0'
  environment: 'demo'
  costProfile: 'free-grant-scale-to-zero'
  managedBy: 'bicep'
}

var appName = '${namePrefix}-platform'
var migrationJobName = '${namePrefix}-migrate'

resource environment 'Microsoft.App/managedEnvironments@2024-03-01' existing = {
  name: environmentName
}

resource registry 'Microsoft.ContainerRegistry/registries@2023-07-01' existing = {
  name: registryName
}

resource pullIdentity 'Microsoft.ManagedIdentity/userAssignedIdentities@2023-01-31' existing = {
  name: pullIdentityName
}

var registryServer = registry.properties.loginServer
var appFqdn = '${appName}.${environment.properties.defaultDomain}'
var publicBaseUrl = 'https://${appFqdn}'

resource migrationJob 'Microsoft.App/jobs@2024-03-01' = {
  name: migrationJobName
  location: location
  tags: tags
  identity: {
    type: 'UserAssigned'
    userAssignedIdentities: {
      '${pullIdentity.id}': {}
    }
  }
  properties: {
    environmentId: environment.id
    configuration: {
      triggerType: 'Manual'
      replicaTimeout: 900
      replicaRetryLimit: 1
      manualTriggerConfig: {
        parallelism: 1
        replicaCompletionCount: 1
      }
      registries: [
        {
          server: registryServer
          identity: pullIdentity.id
        }
      ]
      secrets: [
        {
          name: 'database-url-direct'
          value: databaseUrlDirect
        }
      ]
    }
    template: {
      containers: [
        {
          name: 'migrate'
          image: '${registryServer}/p7/migrate:${imageTag}'
          env: [
            {
              name: 'DATABASE_URL_DIRECT'
              secretRef: 'database-url-direct'
            }
          ]
          resources: {
            cpu: json('0.25')
            memory: '0.5Gi'
          }
        }
      ]
    }
  }
}

// The complete demo shares one HTTP-scaled replica. Redis is an internal transient
// transport, while all product state remains durable in Neon. No replica stays warm.
resource platform 'Microsoft.App/containerApps@2024-03-01' = {
  name: appName
  location: location
  tags: tags
  identity: {
    type: 'UserAssigned'
    userAssignedIdentities: {
      '${pullIdentity.id}': {}
    }
  }
  properties: {
    environmentId: environment.id
    configuration: {
      activeRevisionsMode: 'Single'
      ingress: {
        external: true
        allowInsecure: false
        targetPort: 8088
        transport: 'auto'
      }
      registries: [
        {
          server: registryServer
          identity: pullIdentity.id
        }
      ]
      secrets: [
        {
          name: 'database-url'
          value: databaseUrl
        }
        {
          name: 'operator-token'
          value: operatorToken
        }
      ]
    }
    template: {
      containers: [
        {
          name: 'gateway'
          image: '${registryServer}/p7/gateway:${imageTag}'
          resources: {
            cpu: json('0.25')
            memory: '0.5Gi'
          }
          probes: [
            {
              type: 'Liveness'
              httpGet: {
                path: '/health/live'
                port: 8088
                scheme: 'HTTP'
              }
              initialDelaySeconds: 20
              periodSeconds: 15
              timeoutSeconds: 5
              failureThreshold: 4
            }
          ]
        }
        {
          name: 'redis'
          image: 'redis:8.2.1-alpine'
          command: [
            'redis-server'
          ]
          args: [
            '--save'
            ''
            '--appendonly'
            'no'
            '--maxmemory'
            '128mb'
            '--maxmemory-policy'
            'noeviction'
          ]
          resources: {
            cpu: json('0.25')
            memory: '0.5Gi'
          }
        }
        {
          name: 'control-api'
          image: '${registryServer}/p7/control-api:${imageTag}'
          env: [
            {
              name: 'HTTP_ADDRESS'
              value: '0.0.0.0:8080'
            }
            {
              name: 'DATABASE_URL'
              secretRef: 'database-url'
            }
            {
              name: 'REDIS_URL'
              value: 'redis://127.0.0.1:6379/0'
            }
            {
              name: 'OPERATOR_TOKEN'
              secretRef: 'operator-token'
            }
            {
              name: 'COMMAND_STREAM'
              value: 'rl.commands.v1'
            }
            {
              name: 'EVENT_STREAM'
              value: 'rl.events.v1'
            }
            {
              name: 'DEAD_LETTER_STREAM'
              value: 'rl.dead-letter.v1'
            }
            {
              name: 'PROJECTOR_CONSUMER_GROUP'
              value: 'control-api-projector-v1'
            }
            {
              name: 'ALLOWED_ORIGINS'
              value: publicBaseUrl
            }
            {
              name: 'OPENAPI_PATH'
              value: '/app/contracts/http/openapi.json'
            }
            {
              name: 'LOG_LEVEL'
              value: 'INFO'
            }
          ]
          resources: {
            cpu: json('0.25')
            memory: '0.5Gi'
          }
        }
        {
          name: 'rl-runner'
          image: '${registryServer}/p7/rl-runner:${imageTag}'
          env: [
            {
              name: 'REDIS_URL'
              value: 'redis://127.0.0.1:6379/0'
            }
            {
              name: 'COMMAND_STREAM'
              value: 'rl.commands.v1'
            }
            {
              name: 'EVENT_STREAM'
              value: 'rl.events.v1'
            }
            {
              name: 'DEAD_LETTER_STREAM'
              value: 'rl.dead-letter.v1'
            }
            {
              name: 'RUNNER_CONSUMER_GROUP'
              value: 'rl-runner-v1'
            }
            {
              name: 'POLICY_ARTIFACT_ROOT'
              value: '/workspace/artifacts/policies'
            }
            {
              name: 'RUNNER_STEP_DELAY_MS'
              value: '120'
            }
            {
              name: 'DQN_STEP_DELAY_MS'
              value: '10'
            }
            {
              name: 'LOG_LEVEL'
              value: 'INFO'
            }
          ]
          resources: {
            cpu: json('0.25')
            memory: '0.5Gi'
          }
        }
        {
          name: 'web'
          image: '${registryServer}/p7/web:${imageTag}'
          env: [
            {
              name: 'NUXT_PUBLIC_API_BASE'
              value: publicBaseUrl
            }
            {
              name: 'NUXT_PUBLIC_WS_BASE'
              value: 'wss://${appFqdn}'
            }
          ]
          resources: {
            cpu: json('0.25')
            memory: '0.5Gi'
          }
        }
      ]
      scale: {
        minReplicas: 0
        maxReplicas: 1
        rules: [
          {
            name: 'http-demo-activation'
            http: {
              metadata: {
                concurrentRequests: '10'
              }
            }
          }
        ]
      }
    }
  }
}

output apiUrl string = publicBaseUrl
output webUrl string = publicBaseUrl
output migrationJobName string = migrationJob.name
output appName string = platform.name
