import type { CodegenConfig } from '@graphql-codegen/cli';

const config: CodegenConfig = {
  schema: '../internal/graphql/schema/*.graphqls',
  documents: ['src/api/**/*.graphql', 'src/features/**/*.graphql'],
  generates: {
    'src/api/graphql/generated/': {
      preset: 'client',
      presetConfig: { fragmentMasking: false },
    },
  },
};

export default config;
