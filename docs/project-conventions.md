# Project Conventions

## Version Names

Experiment versions use UTC time plus a short git-like random suffix:

```txt
yyyyMMddTHHmmssZ-6hex
```

Example:

```txt
20260601T000132Z-e32490
```

Run this before marking a new experiment snapshot:

```bash
npm run version:stamp
```

The generated value is stored in `shared/version.ts` and shown in the app UI.

## i18n

This MVP keeps i18n lightweight, but the real experiment should treat multilingual support as required project infrastructure. UI copy should go through `src/i18n` instead of being hard-coded directly inside components.
