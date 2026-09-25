import type {ReactNode} from 'react';
import Link from '@docusaurus/Link';
import Layout from '@theme/Layout';
import Heading from '@theme/Heading';

import styles from './index.module.css';

export default function Home(): ReactNode {
  return (
    <Layout title="Home" description="Mobile first coding agent for on the go development">
      <header className={styles.hero}>
        <div className="container">
          <Heading as="h1">Code from anywhere.</Heading>
          <p>Mobile first coding agent for on the go development.</p>
          <div className={styles.actions}>
            <Link className="button button--primary button--lg" to="/docs/intro">Get started</Link>
            <Link className="button button--secondary button--lg" href="https://github.com/soumajitgh/mobicode">View on GitHub</Link>
          </div>
        </div>
      </header>
      <main className="container">
        <div className={styles.columns}>
          <section><Heading as="h2">Go server</Heading><p>A small HTTP foundation with a health endpoint.</p></section>
          <section><Heading as="h2">Mobile app</Heading><p>Expo and React Native with gluestack UI ready for product work.</p></section>
          <section><Heading as="h2">Documentation</Heading><p>Project guides published from this repository.</p></section>
        </div>
      </main>
    </Layout>
  );
}
