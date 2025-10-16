# Quick Start Guide - DDD Architecture

## Running Tests

```bash
# Run all tests
go test ./src/...

# Run with coverage
go test -cover ./src/...

# Run specific domain
go test ./src/domain/analytics/...
go test ./src/domain/tournament/...

# Run specific service
go test ./src/app/analytics/...
go test ./src/app/tournaments/...

# Verbose output
go test -v ./src/...
```

## Creating a New Feature (TDD Approach)

### Step 1: Write Domain Test (RED)

```go
// src/domain/myfeature/myfeature_test.go
package myfeature_test

import (
    "testing"
    "github.com/heroiclabs/nakama/v3/src/domain/myfeature"
)

func TestNewFeature(t *testing.T) {
    feature, err := myfeature.NewFeature("test-id", "Test Name")
    if err != nil {
        t.Fatalf("NewFeature() failed: %v", err)
    }
    if feature.Name != "Test Name" {
        t.Errorf("Expected name 'Test Name', got %v", feature.Name)
    }
}
```

### Step 2: Implement Domain (GREEN)

```go
// src/domain/myfeature/myfeature.go
package myfeature

import (
    "errors"
    "github.com/heroiclabs/nakama/v3/src/domain/shared"
)

type Feature struct {
    ID   shared.FeatureID
    Name string
}

func NewFeature(id shared.FeatureID, name string) (*Feature, error) {
    if err := id.Validate(); err != nil {
        return nil, err
    }
    if name == "" {
        return nil, errors.New("name is required")
    }
    return &Feature{
        ID:   id,
        Name: name,
    }, nil
}
```

### Step 3: Add Repository Interface

```go
// src/domain/myfeature/repository.go
package myfeature

import "context"

type Repository interface {
    Save(ctx context.Context, f *Feature) error
    Get(ctx context.Context, id shared.FeatureID) (*Feature, error)
}
```

### Step 4: Create Application Service

```go
// src/app/myfeature/service.go
package myfeature

import (
    "context"
    "github.com/heroiclabs/nakama/v3/src/domain/myfeature"
    "github.com/heroiclabs/nakama/v3/src/domain/shared"
)

type Service struct {
    Repo myfeature.Repository
}

func NewService(repo myfeature.Repository) *Service {
    return &Service{Repo: repo}
}

type CreateFeatureCommand struct {
    ID   shared.FeatureID
    Name string
}

func (s *Service) CreateFeature(ctx context.Context, cmd CreateFeatureCommand) error {
    feature, err := myfeature.NewFeature(cmd.ID, cmd.Name)
    if err != nil {
        return err
    }
    return s.Repo.Save(ctx, feature)
}
```

### Step 5: Implement Infrastructure

```go
// src/infra/myfeature/memory_repository.go
package myfeature

import (
    "context"
    "sync"
    "github.com/heroiclabs/nakama/v3/src/domain/myfeature"
    "github.com/heroiclabs/nakama/v3/src/domain/shared"
)

type MemoryRepository struct {
    mu       sync.RWMutex
    features map[shared.FeatureID]*myfeature.Feature
}

func NewMemoryRepository() *MemoryRepository {
    return &MemoryRepository{
        features: make(map[shared.FeatureID]*myfeature.Feature),
    }
}

func (r *MemoryRepository) Save(ctx context.Context, f *myfeature.Feature) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.features[f.ID] = f
    return nil
}

func (r *MemoryRepository) Get(ctx context.Context, id shared.FeatureID) (*myfeature.Feature, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    f, exists := r.features[id]
    if !exists {
        return nil, myfeature.ErrFeatureNotFound
    }
    return f, nil
}
```

## Common Patterns

### Domain Aggregate with Validation

```go
type MyAggregate struct {
    ID        shared.MyID
    Name      string
    State     MyState
    CreatedAt time.Time
    UpdatedAt time.Time
}

func NewMyAggregate(id shared.MyID, name string, now time.Time) (*MyAggregate, error) {
    // Validate
    if err := id.Validate(); err != nil {
        return nil, err
    }
    if name == "" {
        return nil, errors.New("name is required")
    }
    
    // Create
    return &MyAggregate{
        ID:        id,
        Name:      name,
        State:     StateActive,
        CreatedAt: now,
        UpdatedAt: now,
    }, nil
}

// Behavior methods
func (a *MyAggregate) DoSomething() error {
    if a.State != StateActive {
        return ErrInvalidState
    }
    // Business logic
    a.UpdatedAt = time.Now()
    return nil
}
```

### Service with Command Pattern

```go
type Service struct {
    Repo     domain.Repository
    Provider ExternalProvider
    Clock    func() time.Time
}

func NewService(repo domain.Repository, provider ExternalProvider) *Service {
    return &Service{
        Repo:     repo,
        Provider: provider,
        Clock:    func() time.Time { return time.Now().UTC() },
    }
}

type MyCommand struct {
    ID   shared.MyID
    Name string
}

type MyResult struct {
    ID shared.MyID
}

func (s *Service) ExecuteCommand(ctx context.Context, cmd MyCommand) (MyResult, error) {
    // 1. Create domain aggregate
    now := s.Clock()
    aggregate, err := domain.NewAggregate(cmd.ID, cmd.Name, now)
    if err != nil {
        return MyResult{}, err
    }
    
    // 2. Save to repository
    if err := s.Repo.Save(ctx, aggregate); err != nil {
        return MyResult{}, err
    }
    
    // 3. Call external provider if needed
    if err := s.Provider.DoSomething(ctx, aggregate); err != nil {
        return MyResult{}, err
    }
    
    return MyResult{ID: aggregate.ID}, nil
}
```

### Mock for Testing

```go
type mockRepository struct {
    saveFunc func(ctx context.Context, a *domain.Aggregate) error
    getFunc  func(ctx context.Context, id shared.MyID) (*domain.Aggregate, error)
}

func (m *mockRepository) Save(ctx context.Context, a *domain.Aggregate) error {
    if m.saveFunc != nil {
        return m.saveFunc(ctx, a)
    }
    return nil
}

func (m *mockRepository) Get(ctx context.Context, id shared.MyID) (*domain.Aggregate, error) {
    if m.getFunc != nil {
        return m.getFunc(ctx, id)
    }
    return nil, domain.ErrNotFound
}
```

## Directory Structure Template

```
src/
├── domain/
│   └── myfeature/
│       ├── myfeature.go        # Aggregate
│       ├── myfeature_test.go   # Domain tests
│       ├── repository.go       # Repository interface
│       └── errors.go           # Domain errors
├── app/
│   └── myfeature/
│       ├── service.go          # Application service
│       └── service_test.go     # Service tests
└── infra/
    └── myfeature/
        ├── memory_repository.go      # In-memory implementation
        └── memory_repository_test.go # Infrastructure tests
```

## Checklist for New Features

- [ ] Write domain tests first (TDD)
- [ ] Implement domain aggregate with validation
- [ ] Define repository interface in domain
- [ ] Create application service with commands
- [ ] Write service tests with mocks
- [ ] Implement infrastructure (memory/database)
- [ ] Add error types in domain
- [ ] Document public APIs
- [ ] Run tests with coverage
- [ ] Update README if needed

## Common Commands

```bash
# Format code
go fmt ./src/...

# Lint code (if golangci-lint installed)
golangci-lint run ./src/...

# Build
go build ./...

# Test specific function
go test -run TestMyFunction ./src/domain/myfeature/

# Generate coverage report
go test -coverprofile=coverage.out ./src/...
go tool cover -html=coverage.out

# Benchmark (if benchmarks exist)
go test -bench=. ./src/...
```

## Tips

1. **Start with domain tests** - They're easiest and most valuable
2. **Keep domain pure** - No external dependencies
3. **Use table-driven tests** - Cover multiple scenarios easily
4. **Mock at service boundaries** - Not inside domain
5. **Name tests descriptively** - `TestAggregate_Method_Scenario_ExpectedResult`
6. **Test edge cases** - Empty values, nil, boundaries
7. **One assertion per test** - Or use subtests
8. **Use test helpers** - Reduce duplication in setup

## Resources

- **Architecture**: `src/README.md`
- **Testing Guide**: `src/TESTING.md`
- **Migration Summary**: `DDD_MIGRATION_SUMMARY.md`
- **Examples**: See `src/domain/analytics/` and `src/domain/tournament/`
