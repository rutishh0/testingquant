/** @type {import('next').NextConfig} */
const nextConfig = {
  /**
   * Output static HTML export compatible with modern Next.js.
   * This replaces the removed `next export` command in Next.js 14+.
   */
  output: 'export',
};

module.exports = nextConfig;
