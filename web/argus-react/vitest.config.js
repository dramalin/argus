"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
/// <reference types="vitest" />
var vite_1 = require("vite");
var plugin_react_1 = require("@vitejs/plugin-react");
var path_1 = require("path");
// https://vitejs.dev/config/
exports.default = (0, vite_1.defineConfig)({
    plugins: [(0, plugin_react_1.default)()],
    resolve: {
        alias: {
            '@': (0, path_1.resolve)(__dirname, './src'),
            '@components': (0, path_1.resolve)(__dirname, './src/components'),
            '@hooks': (0, path_1.resolve)(__dirname, './src/hooks'),
            '@utils': (0, path_1.resolve)(__dirname, './src/utils'),
            '@contexts': (0, path_1.resolve)(__dirname, './src/contexts'),
            '@types': (0, path_1.resolve)(__dirname, './src/types'),
        },
    },
    test: {
        globals: true,
        environment: 'jsdom',
        setupFiles: ['./src/tests/setup.ts'],
        css: false,
        reporters: ['verbose'],
        coverage: {
            provider: 'v8',
            reporter: ['text', 'json', 'html'],
            exclude: [
                'node_modules/',
                'src/tests/',
                '**/*.d.ts',
                '**/*.test.{ts,tsx}',
                'src/main.tsx',
            ],
        },
    },
});
