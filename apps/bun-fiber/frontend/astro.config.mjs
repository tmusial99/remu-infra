// @ts-check
import { defineConfig } from 'astro/config';
import path from "node:path";

const SITE = process.env.SITE ?? (() => { throw new Error("SITE env not provided."); })();
const SITE_PATH = path.resolve(`src/sites/${SITE}`);

// https://astro.build/config
export default defineConfig({
    root: '.',
    srcDir: SITE_PATH,
    outDir: `dist/${SITE}`,
    vite: {
        resolve: {
            alias: {
                '@shared': path.resolve('./src/shared'),
            }
        }
    },
    server: {
        allowedHosts: ["remu"],
    }
});
