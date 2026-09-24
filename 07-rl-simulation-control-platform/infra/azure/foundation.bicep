@description('Azure region for the zero-fixed-cost demo foundation.')
param location string = resourceGroup().location

@minLength(2)
@maxLength(18)
@description('Short lowercase prefix used for globally unique resource names.')
param namePrefix string = 'p7rl'

@description('Resource tags applied to the Azure release foundation.')
param tags object = {
  project: 'rl-simulation-control-platform'
  release: 'v1.0.1-ghcr'
  environment: 'demo'
  costProfile: 'consumption-scale-to-zero-ghcr'
  managedBy: 'bicep'
}

var environmentName = '${namePrefix}-env'

resource environment 'Microsoft.App/managedEnvironments@2024-03-01' = {
  name: environmentName
  location: location
  tags: tags
  properties: {
    appLogsConfiguration: {
      // Azure Monitor without diagnostic settings stores no application logs and
      // avoids the unsupported literal `none` in older regional control planes.
      destination: 'azure-monitor'
    }
  }
}

output environmentName string = environment.name
output environmentDefaultDomain string = environment.properties.defaultDomain
