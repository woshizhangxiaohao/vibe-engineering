import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // 设置 Turbopack 根目录，避免警告
  turbopack: {
    root: process.cwd(),
  },
  // 优化开发体验
  experimental: {
    // 启用更快的编译
    optimizePackageImports: [
      'lucide-react',
      '@radix-ui/react-accordion',
      '@radix-ui/react-alert-dialog',
      '@radix-ui/react-dialog',
      '@radix-ui/react-dropdown-menu',
      '@radix-ui/react-popover',
      '@radix-ui/react-select',
      '@radix-ui/react-tabs',
      '@radix-ui/react-tooltip',
    ],
  },
  // 类型检查配置
  typescript: {
    // 生产构建时进行类型检查，开发时跳过以加快速度
    ignoreBuildErrors: false,
  },
};

export default nextConfig;
