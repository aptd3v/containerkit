import { defineConfig } from "eslint/config";
import tseslint from "typescript-eslint";
import { recommended } from "./dist/eslint.js";

export default defineConfig(
	{ ignores: ["dist/**", "dist-test/**", "dist-examples/**", "node_modules/**", "src/**"] },
	{ files: ["examples/**/*.{ts,tsx}"], languageOptions: { parser: tseslint.parser } },
	...recommended,
);
