import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const config: Config = {
  title: 'MobiCode',
  tagline: 'Mobile first coding agent for on the go development',
  favicon: 'img/favicon.svg',
  url: 'https://soumajitgh.github.io',
  baseUrl: '/mobicode/',
  organizationName: 'soumajitgh',
  projectName: 'mobicode',
  trailingSlash: false,
  onBrokenLinks: 'throw',
  i18n: {defaultLocale: 'en', locales: ['en']},
  presets: [
    ['classic', {
      docs: {
        sidebarPath: './sidebars.ts',
        editUrl: 'https://github.com/soumajitgh/mobicode/edit/master/website/',
      },
      blog: false,
      theme: {customCss: './src/css/custom.css'},
    } satisfies Preset.Options],
  ],
  themeConfig: {
    colorMode: {respectPrefersColorScheme: true},
    navbar: {
      title: 'MobiCode',
      items: [
        {type: 'docSidebar', sidebarId: 'docs', position: 'left', label: 'Docs'},
        {href: 'https://github.com/soumajitgh/mobicode', label: 'GitHub', position: 'right'},
      ],
    },
    footer: {
      style: 'dark',
      links: [{title: 'Project', items: [
        {label: 'Documentation', to: '/docs/intro'},
        {label: 'GitHub', href: 'https://github.com/soumajitgh/mobicode'},
      ]}],
      copyright: `Copyright © ${new Date().getFullYear()} MobiCode contributors. Apache-2.0.`,
    },
    prism: {theme: prismThemes.github, darkTheme: prismThemes.dracula},
  } satisfies Preset.ThemeConfig,
};

export default config;
