import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const DEFAULT_LOCALES_DIR = path.resolve(__dirname, '../src/i18n/locales');

const REQUIRED_LANGUAGES = ['zh-CN', 'en-US'];
const NAMESPACE_FILE_PATTERN = /^[a-z0-9][a-z0-9_-]*\.json$/;

function readJsonFile(filePath) {
  const raw = fs.readFileSync(filePath, 'utf8');
  return JSON.parse(raw);
}

function isNonEmptyObject(value) {
  return !!value && typeof value === 'object' && !Array.isArray(value) && Object.keys(value).length > 0;
}

export function validateLocales(localesDir = DEFAULT_LOCALES_DIR) {
  const errors = [];

  if (!fs.existsSync(localesDir)) {
    return {
      ok: false,
      errors: [`locales directory not found: ${localesDir}`],
    };
  }

  const languageNamespaces = new Map();

  for (const language of REQUIRED_LANGUAGES) {
    const langDir = path.join(localesDir, language);
    if (!fs.existsSync(langDir)) {
      errors.push(`missing language directory: ${language}`);
      languageNamespaces.set(language, new Set());
      continue;
    }

    const namespaceFiles = fs.readdirSync(langDir).filter((name) => name.endsWith('.json'));
    const namespaces = new Set();

    for (const fileName of namespaceFiles) {
      if (!NAMESPACE_FILE_PATTERN.test(fileName)) {
        errors.push(`invalid namespace filename in ${language}: ${fileName}`);
        continue;
      }

      const namespace = fileName.replace(/\.json$/, '');
      const fullPath = path.join(langDir, fileName);

      let payload;
      try {
        payload = readJsonFile(fullPath);
      } catch (error) {
        errors.push(`invalid json file: ${fullPath} (${error instanceof Error ? error.message : String(error)})`);
        continue;
      }

      if (!isNonEmptyObject(payload)) {
        errors.push(`empty namespace payload: ${language}/${fileName}`);
        continue;
      }

      namespaces.add(namespace);
    }

    languageNamespaces.set(language, namespaces);
  }

  const union = new Set();
  for (const language of REQUIRED_LANGUAGES) {
    for (const namespace of languageNamespaces.get(language) ?? []) {
      union.add(namespace);
    }
  }

  for (const language of REQUIRED_LANGUAGES) {
    const namespaces = languageNamespaces.get(language) ?? new Set();
    for (const namespace of union) {
      if (!namespaces.has(namespace)) {
        errors.push(`missing namespace file: ${language}/${namespace}.json`);
      }
    }
  }

  return {
    ok: errors.length === 0,
    errors,
  };
}

export function runLocaleValidation() {
  const result = validateLocales();
  if (result.ok) {
    console.log('i18n locale validation passed');
    return;
  }

  console.error('i18n locale validation failed:');
  for (const error of result.errors) {
    console.error(`- ${error}`);
  }
  process.exitCode = 1;
}

const invokedAsScript = process.argv[1] && path.resolve(process.argv[1]) === __filename;
if (invokedAsScript) {
  runLocaleValidation();
}
