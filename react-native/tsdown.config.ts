import { defineConfig } from 'tsdown';

export default defineConfig({
  entry: ['src/index.ts'],
  format: ['cjs'],
  dts: true,
  sourcemap: true,
  clean: true,
  // `react-native-webview`/`expo-asset` are peer deps; `*.html` must be left
  // as a literal `require()` call untouched - Metro (React Native's bundler)
  // is the one that turns it into an asset reference at app build time.
  deps: {
    neverBundle: ['react', 'react-native', 'react-native-webview', 'expo-asset', /\.html$/],
  },
});
