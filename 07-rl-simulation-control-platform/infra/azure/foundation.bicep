@description('Azure region for the zero-cost demo resources.')
param location string = resourceGroup().location

@minLength(2)
@maxLength(18)
@description('Short lowercase prefix used for globally unique resource names.')
param namePrefix string = 'p7rl'

@description('Resource tags applied to the Azure release foundation.')
param tags object = {
  project: 'rl-simulation-control-platform'
  release: 'v1.0.0'
  environment: 'demo'
  costProfile: 'free-grant-scale-to-zero'
  managedBy: 'bicep'
}

var suffix = uniqueString(subscription().id, resourceGroup().id)
var registryName = toLower('${replace(namePrefix, '-', '')}${suffix}')
var environmentName = '${namePrefix}-env'
var identityName = '${namePrefix}-pull'
var acrPullRoleDefinitionId = subscriptionResourceId(
  'Microsoft.Authorization/roleDefinitions',
  '7f951dda-4ed3-4680-a7ca-43fe172d538d'
)

resource environment 'Microsoft.App/managedEnvironments@2024-03-01' = {
  name: environmentName
  location: location
  tags: tags
  properties: {
    appLogsConfiguration: {
      destination: 'none'
    }
  }
}

// New Azure accounts currently include one Standard registry for 12 months.
// Standard is intentional: Basic is not part of that documented free grant.
resource registry 'Microsoft.ContainerRegistry/registries@2023-07-01' = {
  name: registryName
  location: location
  tags: tags
  sku: {
    name: 'Standard'
  }
  properties: {
    adminUserEnabled: false
    publicNetworkAccess: 'Enabled'
  }
}

resource pullIdentity 'Microsoft.ManagedIdentity/userAssignedIdentities@2023-01-31' = {
  name: identityName
  location: location
  tags: tags
}

resource acrPull 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  name: guid(registry.id, pullIdentity.id, acrPullRoleDefinitionId)
  scope: registry
  properties: {
    principalId: pullIdentity.properties.principalId
    principalType: 'ServicePrincipal'
    roleDefinitionId: acrPullRoleDefinitionId
  }
}

output registryName string = registry.name
output registryLoginServer string = registry.properties.loginServer
output environmentName string = environment.name
output environmentDefaultDomain string = environment.properties.defaultDomain
output pullIdentityName string = pullIdentity.name
