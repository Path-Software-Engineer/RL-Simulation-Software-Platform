using './workloads.bicep'

param location = readEnvironmentVariable('AZURE_RELEASE_LOCATION')
param namePrefix = readEnvironmentVariable('AZURE_RELEASE_NAME_PREFIX')
param registryName = readEnvironmentVariable('AZURE_RELEASE_REGISTRY_NAME')
param environmentName = readEnvironmentVariable('AZURE_RELEASE_ENVIRONMENT_NAME')
param pullIdentityName = readEnvironmentVariable('AZURE_RELEASE_PULL_IDENTITY_NAME')
param imageTag = readEnvironmentVariable('AZURE_RELEASE_IMAGE_TAG')
param databaseUrl = readEnvironmentVariable('NEON_DATABASE_URL')
param databaseUrlDirect = readEnvironmentVariable('NEON_DATABASE_URL_DIRECT')
param operatorToken = readEnvironmentVariable('OPERATOR_TOKEN')
