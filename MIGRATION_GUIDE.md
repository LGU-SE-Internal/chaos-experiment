# Migration Guide: Using the New Architecture

## Overview

The new architecture provides complete decoupling between:
1. **Internal Data Storage** (`internal/resourcetypes`) - Raw data types
2. **Handler Endpoints** (`internal/endpoint`) - Unified conversion layer
3. **System Data Model** (`internal/model`) - Container for all resources
4. **Registry** (`internal/registry`) - Auto-registration pattern

## Current Status

### ✅ Completed
- `internal/resourcetypes/types.go` - Decoupled raw data types (HTTPEndpoint, RPCOperation, DatabaseOperation)
- `internal/endpoint/types.go` - Handler endpoint types (Endpoint, CallPair, HTTPEndpointInfo, DNSEndpointInfo, DatabaseInfo)
- `internal/endpoint/converter.go` - Conversion functions from internal data to endpoints
- `internal/model/systemdata.go` - Unified SystemData container
- `internal/registry/registry.go` - Registry pattern implementation
- `tools/clickhouseanalyzer` - NO LONGER imports internal packages (fully decoupled)

### 🔄 In Progress
- Generated code still uses OLD 3-file structure
- resourcelookup still uses OLD switch-case pattern
- cmd/faultpoints still imports ALL system packages

### ❌ Not Started
- Generate single `data.go` per system with auto-registration
- Update resourcelookup to use registry pattern
- Update cmd/faultpoints to use registry pattern
- Update handlers to use new endpoint types

## How to Use the New Architecture

### For New Code (Recommended Pattern)

#### 1. Accessing System Data via Registry

```go
import (
    "github.com/LGU-SE-Internal/chaos-experiment/internal/registry"
    "github.com/LGU-SE-Internal/chaos-experiment/internal/systemconfig"
)

// Set current system (usually done once at startup)
systemconfig.SetCurrentSystem(systemconfig.SystemTrainTicket)

// Get system data from registry
sysData := registry.MustGetCurrent()

// Access HTTP endpoints
httpEndpoints := sysData.GetHTTPEndpointsByService("ts-order-service")

// Access RPC operations
rpcOps := sysData.GetRPCOperationsByService("frontend")

// Access database operations
dbOps := sysData.GetDatabaseOperationsByService("account service")
```

#### 2. Converting to Handler Endpoints

```go
import (
    "github.com/LGU-SE-Internal/chaos-experiment/internal/endpoint"
    "github.com/LGU-SE-Internal/chaos-experiment/internal/registry"
)

sysData := registry.MustGetCurrent()

// For HTTP Chaos - get HTTP-specific info
httpEndpoints := sysData.GetHTTPEndpointsByService("myservice")
for _, ep := range httpEndpoints {
    httpInfo := endpoint.ToHTTPEndpointInfo(ep)
    // Use httpInfo for HTTP chaos
}

// For Network Chaos - combine all operation types
// (Build CallPair from HTTP + RPC + DB endpoints)
var networkPairs []endpoint.CallPair
// ... collect from HTTP, RPC, and DB ...

// For DNS Chaos - HTTP + DB only (exclude RPC)
// (Build DNSEndpointInfo from HTTP + DB endpoints)

// For MySQL Chaos - filter by DB system
mysqlOps := sysData.GetDatabaseOperationsByDBSystem("mysql")
for _, op := range mysqlOps {
    dbInfo := endpoint.ToDatabaseInfo(op)
    // Use dbInfo for MySQL chaos
}
```

#### 3. Example: Network Chaos Handler

```go
func GetNetworkPairs() []endpoint.CallPair {
    sysData := registry.MustGetCurrent()
    pairMap := make(map[string]*endpoint.CallPair)
    
    // Collect HTTP pairs
    for _, service := range sysData.GetAllServices() {
        for _, ep := range sysData.GetHTTPEndpointsByService(service) {
            if ep.ServerAddress != "" {
                key := ep.ServiceName + "->" + ep.ServerAddress
                // Add to pairMap with operation type "http"
            }
        }
    }
    
    // Collect RPC pairs
    for _, service := range sysData.GetAllRPCServices() {
        for _, op := range sysData.GetRPCOperationsByService(service) {
            // Add to pairMap with operation type "rpc"
        }
    }
    
    // Collect DB pairs
    for _, service := range sysData.GetAllDatabaseServices() {
        for _, op := range sysData.GetDatabaseOperationsByService(service) {
            // Add to pairMap with operation type "db"
        }
    }
    
    return convertMapToSlice(pairMap)
}
```

## Benefits of New Architecture

### 1. Complete Decoupling
- **Tool Layer** (clickhouseanalyzer): Defines its OWN types, doesn't import internal packages
- **Storage Layer** (resourcetypes): Raw data storage, type-specific fields only
- **Model Layer** (model): Container for all resource types
- **Conversion Layer** (endpoint): Transforms storage to handler needs
- **Registry Layer** (registry): Auto-registration, no switch-case

### 2. No More Switch-Case Statements
OLD:
```go
switch systemconfig.GetCurrentSystem() {
case systemconfig.SystemTrainTicket:
    return tsendpoints.GetEndpoints()
case systemconfig.SystemOtelDemo:
    return oteldemoendpoints.GetEndpoints()
// ... 6 systems ...
}
```

NEW:
```go
sysData := registry.MustGetCurrent()
return sysData.GetHTTPEndpointsByService(service)
```

### 3. No System-Specific Imports
OLD (cmd/faultpoints):
```go
import (
    hsendpoints "github.com/LGU-SE-Internal/chaos-experiment/internal/hs/serviceendpoints"
    mediaendpoints "github.com/LGU-SE-Internal/chaos-experiment/internal/media/serviceendpoints"
    // ... 18+ imports ...
)
```

NEW:
```go
import (
    "github.com/LGU-SE-Internal/chaos-experiment/internal/registry"
    "github.com/LGU-SE-Internal/chaos-experiment/internal/endpoint"
)
```

### 4. Type-Specific Data Storage
Each resource type only stores relevant fields:
- **HTTPEndpoint**: Method, Route, Status (HTTP-specific)
- **RPCOperation**: RPCSystem, RPCService, RPCMethod (RPC-specific)
- **DatabaseOperation**: DBName, DBTable, Operation, DBSystem (DB-specific)

### 5. Fault-Type-Specific Endpoints
Different handlers get different endpoint types based on their needs:
- **Network Chaos**: `CallPair` (all operation types)
- **HTTP Chaos**: `HTTPEndpointInfo` (HTTP only)
- **DNS Chaos**: `DNSEndpointInfo` (HTTP + DB, excludes RPC)
- **MySQL Chaos**: `DatabaseInfo` (MySQL DB only)

## Next Steps

### Phase 1: Update Code Generator (HIGH PRIORITY)
Modify `tools/clickhouseanalyzer/datagenerator.go` to:
1. Generate single `data.go` per system instead of 3 files
2. Add `init()` function that registers data with registry
3. Keep types as aliases to `resourcetypes`

Example generated file structure:
```go
// internal/ts/data/data.go
package data

import (
    "github.com/LGU-SE-Internal/chaos-experiment/internal/model"
    "github.com/LGU-SE-Internal/chaos-experiment/internal/registry"
    "github.com/LGU-SE-Internal/chaos-experiment/internal/resourcetypes"
    "github.com/LGU-SE-Internal/chaos-experiment/internal/systemconfig"
)

func init() {
    // Auto-register on import
    registry.Register(systemconfig.SystemTrainTicket, &model.SystemData{
        SystemName: "ts",
        HTTPEndpoints: map[string][]resourcetypes.HTTPEndpoint{ /* ... */ },
        RPCOperations: map[string][]resourcetypes.RPCOperation{ /* ... */ },
        DatabaseOperations: map[string][]resourcetypes.DatabaseOperation{ /* ... */ },
        AllServices: []string{ /* ... */ },
    })
}
```

### Phase 2: Update resourcelookup (HIGH PRIORITY)
Replace switch-case logic with registry-based access:
- Use `registry.MustGetCurrent()` instead of system-specific imports
- Implement conversion to endpoint types
- Remove OLD functions gradually

### Phase 3: Update cmd/faultpoints (MEDIUM PRIORITY)
- Remove all system-specific imports
- Import only `internal/registry` and `internal/endpoint`
- Use registry to access data

### Phase 4: Update Handlers (LOW PRIORITY)
- Migrate to use new endpoint types from `internal/endpoint`
- Remove conversions from old types

## Testing the New Code

Currently, the new architecture can be tested but requires:
1. Manually populating registry with test data
2. Using the conversion functions from `internal/endpoint`

Example test:
```go
func TestRegistryPattern(t *testing.T) {
    // Setup
    testData := &model.SystemData{
        HTTPEndpoints: map[string][]resourcetypes.HTTPEndpoint{
            "service1": {{ServiceName: "service1", Route: "/api/test"}},
        },
    }
    registry.Register(systemconfig.SystemTrainTicket, testData)
    systemconfig.SetCurrentSystem(systemconfig.SystemTrainTicket)
    
    // Use
    sysData := registry.MustGetCurrent()
    eps := sysData.GetHTTPEndpointsByService("service1")
    assert.Equal(t, 1, len(eps))
}
```
