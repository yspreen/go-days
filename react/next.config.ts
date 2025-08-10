import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "export",
  distDir: "dist",
  redirects: async () => [
    {
      source: "/room/:roomId",
      destination: "/?roomId=:roomId",
      permanent: false,
    },
  ],
};

export default nextConfig;
