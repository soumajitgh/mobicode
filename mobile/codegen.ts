import type { CodegenConfig } from '@graphql-codegen/cli';

const config: CodegenConfig = {
  schema: '../internal/graphql/schema/*.graphqls',
  documents: 'src/graphql/*.graphql',
  generates: {
    'src/graphql/generated/': {
      preset: 'client',
      presetConfig: { fragmentMasking: false },
    },
  },
};

export default config;
