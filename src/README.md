# Domain-Driven Design Architecture

This directory contains the DDD-structured codebase for Nakama, organized into three main layers following clean architecture principles.

## Directory Structure

```
src/
├── domain/          # Domain Layer (Business Logic)
│   ├── analytics/   # Analytics tracking domain
│   ├── battle/      # Battle management domain
│   ├── bot/         # Bot command domain
│   ├── group/       # Group management domain
│   ├── leaderboard/ # Leaderboard domain
│   ├── player/      # Player/account domain
│   ├── tournament/  # Tournament domain
│   └── shared/      # Shared domain types
├── app/             # Application Layer (Use Cases)
│   ├── analytics/   # Analytics service
│   ├── auth/        # Authentication service
│   ├── battles/     # Battle service
│   ├── bot/         # Bot service
│   ├── groups/      # Group service
│   ├── leaderboard/ # Leaderboard service
│   └── tournaments/ # Tournament service
└── infra/           # Infrastructure Layer (External Integrations)
    ├── analytics/   # Analytics infrastructure (Segment, etc.)
    ├── runtime/     # Nakama runtime integrations
    └── tournament/  # Tournament infrastructure
```

## Layer Responsibilities

### Domain Layer (`domain/`)

The domain layer contains:
- **Aggregates**: Core business entities with behavior (e.g., `Tournament`, `Session`, `Battle`)
- **Value Objects**: Immutable objects representing concepts (e.g., `EventType`, `SortOrder`)
- **Domain Events**: Events that represent state changes
- **Repository Interfaces**: Contracts for persistence
- **Domain Errors**: Business rule violations

**Rules:**
- No dependencies on other layers
- Pure business logic only
- Framework-agnostic
- Highly testable

**Example:**
```go
// domain/tournament/tournament.go
type Tournament struct {
    ID            shared.TournamentID
    Title         string
    State         TournamentState
    // ... other fields
}

func (t *Tournament) End(endTime time.Time) error {
    if t.State == StateEnded {
        return ErrTournamentAlreadyEnded
    }
    t.State = StateEnded
    t.EndTime = &endTime
    return nil
}
```

### Application Layer (`app/`)

The application layer contains:
- **Services**: Orchestrate domain objects and use cases
- **Commands**: Input DTOs for write operations
- **Queries**: Input DTOs for read operations
- **Results**: Output DTOs
- **Provider Interfaces**: Abstractions for external systems

**Rules:**
- Depends only on domain layer
- Orchestrates domain logic
- Handles transactions
- Coordinates infrastructure

**Example:**
```go
// app/tournaments/service.go
type Service struct {
    Repo     tournament.Repository
    Provider NakamaProvider
}

func (s *Service) CreateTournament(ctx context.Context, cmd CreateTournamentCommand) (CreateTournamentResult, error) {
    // 1. Create domain aggregate
    t, err := tournament.NewTournament(...)
    
    // 2. Save to repository
    if err := s.Repo.Save(ctx, t); err != nil {
        return CreateTournamentResult{}, err
    }
    
    // 3. Sync with external system
    if err := s.Provider.CreateTournament(ctx, params); err != nil {
        return CreateTournamentResult{}, err
    }
    
    return CreateTournamentResult{TournamentID: t.ID}, nil
}
```

### Infrastructure Layer (`infra/`)

The infrastructure layer contains:
- **Repository Implementations**: Concrete persistence (memory, DB, etc.)
- **External Service Clients**: API clients (Segment, Nakama, etc.)
- **Adapters**: Convert between domain and external formats

**Rules:**
- Implements interfaces from domain/app layers
- Handles external dependencies
- Framework-specific code
- Database access

**Example:**
```go
// infra/analytics/segment_dispatcher.go
type SegmentDispatcher struct {
    APIKey     string
    HTTPClient *http.Client
}

func (d *SegmentDispatcher) Dispatch(ctx context.Context, events []*analytics.Event) error {
    // Convert domain events to Segment format
    // Make HTTP request
    // Handle response
}
```

## Test-Driven Development (TDD)

Each layer has comprehensive test coverage:

### Domain Tests
- Unit tests for aggregates and value objects
- Test business rules and invariants
- No mocks needed (pure logic)

```go
// domain/tournament/tournament_test.go
func TestTournament_End(t *testing.T) {
    tour, _ := tournament.NewTournament(...)
    err := tour.End(endTime)
    // Assert state changes
}
```

### Application Tests
- Service-level tests with mocks
- Test use case orchestration
- Verify command/query handling

```go
// app/tournaments/service_test.go
func TestService_CreateTournament(t *testing.T) {
    mockRepo := &mockTournamentRepo{}
    mockProvider := &mockNakamaProvider{}
    service := tournaments.NewService(mockRepo, mockProvider)
    // Test service behavior
}
```

### Infrastructure Tests
- Integration tests (optional)
- Test external system interactions
- Can use test doubles

## Adding New Features

### 1. Start with Domain (TDD)

```go
// 1. Write failing test
func TestNewFeature(t *testing.T) {
    feature, err := NewFeature(...)
    // Assert expected behavior
}

// 2. Implement domain logic
type Feature struct {
    ID shared.FeatureID
    // fields
}

func NewFeature(...) (*Feature, error) {
    // Validation and creation
}

// 3. Add repository interface
type Repository interface {
    Save(ctx context.Context, f *Feature) error
    Get(ctx context.Context, id shared.FeatureID) (*Feature, error)
}
```

### 2. Create Application Service

```go
// app/features/service.go
type Service struct {
    Repo feature.Repository
}

type CreateFeatureCommand struct {
    // Input fields
}

func (s *Service) CreateFeature(ctx context.Context, cmd CreateFeatureCommand) error {
    // Orchestrate domain objects
}
```

### 3. Implement Infrastructure

```go
// infra/feature/memory_repository.go
type MemoryRepository struct {
    mu       sync.RWMutex
    features map[shared.FeatureID]*feature.Feature
}

func (r *MemoryRepository) Save(ctx context.Context, f *feature.Feature) error {
    // Implementation
}
```

### 4. Create Adapter (if needed)

```go
// Adapt to existing APIs
type FeatureAdapter struct {
    service *features.Service
}

func (a *FeatureAdapter) LegacyMethod(...) error {
    // Convert to command and call service
}
```

## Benefits of This Architecture

### 1. **Testability**
- Domain logic is pure and easy to test
- Services use dependency injection
- Mocks are simple to create

### 2. **Maintainability**
- Clear separation of concerns
- Business logic isolated from infrastructure
- Easy to understand and modify

### 3. **Flexibility**
- Swap implementations easily (memory → DB)
- Change external services without touching domain
- Add new features without breaking existing code

### 4. **Domain Focus**
- Business rules are explicit and visible
- Ubiquitous language in code
- Domain experts can read the code

## Migration Strategy

The codebase supports both legacy and DDD approaches:

1. **Legacy code** (`se/se.go`, `sample_go_module/tournament.go`) remains functional
2. **Adapters** (`se/adapter.go`, `tournament_adapter.go`) bridge old and new
3. **New features** should use DDD structure
4. **Gradual migration** of existing features as needed

## Running Tests

```bash
# Run all tests
go test ./src/...

# Run domain tests only
go test ./src/domain/...

# Run with coverage
go test -cover ./src/...

# Run specific package
go test ./src/app/tournaments/...
```

## Examples

### Analytics Tracking

```go
// Using the service directly
dispatcher := infraAnalytics.NewSegmentDispatcher(apiKey, baseURL)
sessionRepo := infraAnalytics.NewMemorySessionRepository()
service := analytics.NewService(dispatcher, sessionRepo)

cmd := analytics.StartSessionCommand{
    UserID:  shared.PlayerID("player-123"),
    Version: "1.0.0",
    Variant: "production",
}
err := service.StartSession(ctx, cmd)
```

### Tournament Management

```go
// Using the service directly
repo := infraTournament.NewMemoryRepository()
participantRepo := infraTournament.NewMemoryParticipantRepository()
provider := infraTournament.NewNakamaProvider(nk)
service := tournaments.NewService(repo, participantRepo, provider)

cmd := tournaments.CreateTournamentCommand{
    ID:       shared.TournamentID("tournament-123"),
    Title:    "Championship",
    Category: 1,
    // ... other fields
}
result, err := service.CreateTournament(ctx, cmd)
```

## References

- [Domain-Driven Design by Eric Evans](https://www.domainlanguage.com/ddd/)
- [Clean Architecture by Robert C. Martin](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/)
