import { CapacitorConfig } from '@capacitor/cli';
import { Capacitor } from '@capacitor/core';

const config: CapacitorConfig = {
  appId: 'com.monetr.app',
  appName: 'monetr',
  webDir: 'pkg/ui/static',
  bundledWebRuntime: false,
  server: {
    url: Capacitor.getPlatform() === 'ios' ? 'http://localhost' : 'http://10.0.2.2',
  }
};

export default config;
