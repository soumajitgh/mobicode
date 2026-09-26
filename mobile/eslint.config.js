const { defineConfig, globalIgnores } = require('eslint/config');
const expoConfig = require('eslint-config-expo/flat');

module.exports = defineConfig([
  globalIgnores(['dist/*', 'src/api/graphql/generated/*', 'src/components/ui/**']),
  expoConfig,
]);
