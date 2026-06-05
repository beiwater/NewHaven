# Target Architecture Prototype

Last updated: 2026-06-06

This is the first high-level pseudocode sketch for `backend-next`.

## Runtime Shape

```txt
cmd/simapi-next
  load config
  load static catalog data
  create clock, id generator, logger
  create storage adapter
  create domain services
  create application use cases
  create HTTP router
  start server
  wait for shutdown
  stop schedulers
  flush state if enabled
```

## Request Flow

```txt
HTTP request
  -> middleware
       request id
       auth token extraction
       panic recovery
       logging
  -> route handler
       parse request
       call use case
       map domain error to API error
       write response DTO
```

## Use Case Flow

```txt
UseCase(input, actor)
  validate actor permission
  validate input shape
  load current domain state
  call domain logic
  commit mutation if command
  emit event or ledger entry if needed
  return read model / command result
```

## Dependency Direction

```txt
httpapi -> service -> domain -> model
service -> storage interfaces
service -> platform clock/id/logger interfaces
storage adapters -> database/files/memory
formula -> no side effects
```

## Source Of Truth During Migration

```txt
old mode:
  A owns state and responses

shadow mode:
  A owns state and responses
  B may read copied state or dry-run commands

next mode:
  B owns response for selected route
  state ownership must be explicit before command routes move
```

## First Interfaces To Sketch

```txt
Clock
  Now() time

IDGenerator
  New(prefix) string

Catalog
  Resource(id) Resource
  Building(id) Building
  Recipe(resourceID) Recipe

CompanyReader
  Company(id) CompanyReadModel
  Inventory(companyID) InventoryReadModel

ProductionService
  Jobs(companyID) []ProductionJob
  Start(companyID, command) Result
  Claim(companyID, jobID) Result

MarketService
  Ticker(resourceID) Ticker
  Depth(resourceID, quality) Depth
  CreateOrder(companyID, command) Result
  TakeOrder(companyID, command) Result
```

