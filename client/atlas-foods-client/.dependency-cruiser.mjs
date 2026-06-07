export default {
  "forbidden": [
    {
      "name": "no-circular",
      "severity": "error",
      "comment": "禁止循环依赖",
      "from": {},
      "to": {
        "circular": true
      }
    },
    {
      "name": "no-feature-cross-import",
      "comment": "features 模块不能互相引用",
      "severity": "warn",
      "from": {
        "path": "^src/features"
      },
      "to": {
        "path": "^src/features",
        "pathNot": "^src/features/[^/]+/shared"
      }
    }
  ],
  "options": {
    "doNotFollow": {
      "path": "node_modules"
    },
    "tsPreCompilationDeps": true,
    "tsConfig": {
      "fileName": "tsconfig.app.json"
    },
    "enhancedResolveOptions": {
      "exportsFields": [],
      "conditionNames": [
        "import",
        "require",
        "default"
      ]
    }
  }
}