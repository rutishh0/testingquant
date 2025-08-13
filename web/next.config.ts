import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: 'export',
  // Inject backend URL at build-time so the static export knows where to call
  env: {
    NEXT_PUBLIC_BACKEND_URL: process.env.NEXT_PUBLIC_BACKEND_URL,
    NEXT_PUBLIC_API_KEY: process.env.NEXT_PUBLIC_API_KEY,
  },
};

export default nextConfig;
