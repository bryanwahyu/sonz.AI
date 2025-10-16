# Testing Guide

This document describes the testing strategy and best practices for the DDD-structured codebase.

## Test Structure

Tests are organized by layer and follow the same directory structure as the source code:

```
src/
├── domain/
│   ├── analytics/
│   │   ├── event.go
│   │   ├── event_test.go          # Domain tests
│   │   ├── session.go
│   │   └── session_test.go
│   └── tournament/
│       ├── tournament.go
│       ├── tournament_test.go
│       ├── participant.go
│       └── participant_test.go
├── app/
│   ├── analytics/
│   │   ├── service.go
│   │   └── service_test.go        # Application tests
│   └── tournaments/
│       ├── service.go
│       └── service_test.go
└── infra/
    ├── analytics/
    │   ├── segment_dispatcher.go
    │   └── segment_dispatcher_test.go  # Infrastructure tests (optional)
    └── tournament/
        ├── memory_repository.go
        └── memory_repository_test.go
```

## Test-Driven Development (TDD) Workflow

### 1. Red Phase - Write Failing Test

Start by writing a test that describes the desired behavior:

```go
func TestTournament_End(t *testing.T) {
    // Arrange
    now := time.Now()
    startTime := now.Add(1 * time.Hour)
    tour, _ := tournament.NewTournament(
        "tournament-123",
        "Test Tournament",
        "Description",
        1,
        tournament.SortOrderDescending,
        tournament.OperatorBest,
        "",
        true,
        false,
        100,
        10,
        startTime,
        24*time.Hour,
        now,
    )
    
    // Act
    endTime := startTime.Add(2 * time.Hour)
    err := tour.End(endTime)
    
    // Assert
    if err != nil {
        t.Errorf("End() failed: %v", err)
    }
    if tour.State != tournament.StateEnded {
        t.Errorf("Expected state %v, got %v", tournament.StateEnded, tour.State)
    }
}
```

### 2. Green Phase - Make It Pass

Implement the minimum code to make the test pass:

```go
func (t *Tournament) End(endTime time.Time) error {
    if t.State == StateEnded {
        return ErrTournamentAlreadyEnded
    }
    if endTime.Before(t.StartTime) {
        return errors.New("end time cannot be before start time")
    }
    t.State = StateEnded
    t.EndTime = &endTime
    t.UpdatedAt = endTime
    return nil
}
```

### 3. Refactor Phase - Improve Code

Clean up the code while keeping tests green:
- Extract common logic
- Improve naming
- Remove duplication
- Optimize performance

## Testing Patterns by Layer

### Domain Layer Tests

Domain tests are **pure unit tests** with no external dependencies.

**Characteristics:**
- No mocks needed
- Fast execution
- Test business rules
- Test invariants
- Test edge cases

**Example:**

```go
func TestNewSession(t *testing.T) {
    tests := []struct {
        name      string
        userID    shared.PlayerID
        version   string
        startedAt time.Time
        wantErr   bool
    }{
        {
            name:      "valid session",
            userID:    "player-123",
            version:   "1.0.0",
            startedAt: time.Now(),
            wantErr:   false,
        },
        {
            name:      "empty user id",
            userID:    "",
            version:   "1.0.0",
            startedAt: time.Now(),
            wantErr:   true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            session, err := analytics.NewSession(tt.userID, tt.version, "prod", tt.startedAt)
            if (err != nil) != tt.wantErr {
                t.Errorf("NewSession() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Application Layer Tests

Application tests use **mocks** for dependencies.

**Characteristics:**
- Mock repositories and providers
- Test orchestration logic
- Test command/query handling
- Test error handling

**Example:**

```go
type mockDispatcher struct {
    dispatchFunc func(ctx context.Context, events []*analytics.Event) error
}

func (m *mockDispatcher) Dispatch(ctx context.Context, events []*analytics.Event) error {
    if m.dispatchFunc != nil {
        return m.dispatchFunc(ctx, events)
    }
    return nil
}

func TestService_StartSession(t *testing.T) {
    ctx := context.Background()
    
    var capturedEvents []*analytics.Event
    dispatcher := &mockDispatcher{
        dispatchFunc: func(ctx context.Context, events []*analytics.Event) error {
            capturedEvents = events
            return nil
        },
    }
    
    sessionRepo := &mockSessionRepo{}
    service := analytics.NewService(dispatcher, sessionRepo)
    
    cmd := analytics.StartSessionCommand{
        UserID:  "player-123",
        Version: "1.0.0",
        Variant: "production",
    }
    
    err := service.StartSession(ctx, cmd)
    if err != nil {
        t.Fatalf("StartSession() failed: %v", err)
    }
    
    if len(capturedEvents) != 2 {
        t.Errorf("Expected 2 events, got %d", len(capturedEvents))
    }
}
```

### Infrastructure Layer Tests

Infrastructure tests can be **unit tests** or **integration tests**.

**Unit Tests:**
- Test in-memory implementations
- Test data transformations
- Fast and isolated

**Integration Tests (Optional):**
- Test real external systems
- Use test databases
- Slower but more realistic

**Example (Unit Test):**

```go
func TestMemoryRepository_Save(t *testing.T) {
    ctx := context.Background()
    repo := infraTournament.NewMemoryRepository()
    
    tour, _ := tournament.NewTournament(...)
    
    err := repo.Save(ctx, tour)
    if err != nil {
        t.Fatalf("Save() failed: %v", err)
    }
    
    retrieved, err := repo.Get(ctx, tour.ID)
    if err != nil {
        t.Fatalf("Get() failed: %v", err)
    }
    
    if retrieved.ID != tour.ID {
        t.Errorf("Expected ID %v, got %v", tour.ID, retrieved.ID)
    }
}
```

## Test Organization

### Table-Driven Tests

Use table-driven tests for multiple scenarios:

```go
func TestParticipant_AddAttempts(t *testing.T) {
    tests := []struct {
        name    string
        count   int
        wantErr bool
    }{
        {
            name:    "add positive attempts",
            count:   5,
            wantErr: false,
        },
        {
            name:    "add zero attempts",
            count:   0,
            wantErr: true,
        },
        {
            name:    "add negative attempts",
            count:   -1,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            now := time.Now()
            p, _ := tournament.NewParticipant("tournament-123", "player-456", now)
            
            err := p.AddAttempts(tt.count, now)
            if (err != nil) != tt.wantErr {
                t.Errorf("AddAttempts() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Test Helpers

Create helpers for common test setup:

```go
func createTestTournament(t *testing.T) *tournament.Tournament {
    t.Helper()
    
    tour, err := tournament.NewTournament(
        "tournament-123",
        "Test Tournament",
        "Description",
        1,
        tournament.SortOrderDescending,
        tournament.OperatorBest,
        "",
        true,
        false,
        100,
        10,
        time.Now().Add(1*time.Hour),
        24*time.Hour,
        time.Now(),
    )
    
    if err != nil {
        t.Fatalf("Failed to create test tournament: %v", err)
    }
    
    return tour
}
```

## Running Tests

### Run All Tests

```bash
go test ./src/...
```

### Run Specific Package

```bash
go test ./src/domain/tournament/
go test ./src/app/analytics/
```

### Run with Coverage

```bash
go test -cover ./src/...
go test -coverprofile=coverage.out ./src/...
go tool cover -html=coverage.out
```

### Run with Verbose Output

```bash
go test -v ./src/...
```

### Run Specific Test

```bash
go test -run TestTournament_End ./src/domain/tournament/
```

### Run Tests in Parallel

```bash
go test -parallel 4 ./src/...
```

## Coverage Goals

- **Domain Layer**: 90%+ coverage (critical business logic)
- **Application Layer**: 80%+ coverage (orchestration)
- **Infrastructure Layer**: 70%+ coverage (implementation details)

## Best Practices

### 1. Test Behavior, Not Implementation

❌ Bad:
```go
func TestService_CreateTournament(t *testing.T) {
    // Testing internal implementation details
    if service.repo != nil {
        t.Error("repo should be set")
    }
}
```

✅ Good:
```go
func TestService_CreateTournament(t *testing.T) {
    // Testing observable behavior
    result, err := service.CreateTournament(ctx, cmd)
    if err != nil {
        t.Fatalf("CreateTournament() failed: %v", err)
    }
    if result.TournamentID == "" {
        t.Error("Expected tournament ID to be set")
    }
}
```

### 2. Use Descriptive Test Names

❌ Bad:
```go
func TestTournament1(t *testing.T) { ... }
func TestTournament2(t *testing.T) { ... }
```

✅ Good:
```go
func TestTournament_End_WithValidEndTime_ShouldSucceed(t *testing.T) { ... }
func TestTournament_End_WithEndTimeBeforeStart_ShouldFail(t *testing.T) { ... }
```

### 3. Arrange-Act-Assert Pattern

```go
func TestExample(t *testing.T) {
    // Arrange - Set up test data
    tournament := createTestTournament(t)
    endTime := time.Now().Add(2 * time.Hour)
    
    // Act - Execute the behavior
    err := tournament.End(endTime)
    
    // Assert - Verify the results
    if err != nil {
        t.Errorf("End() failed: %v", err)
    }
    if tournament.State != tournament.StateEnded {
        t.Errorf("Expected state %v, got %v", tournament.StateEnded, tournament.State)
    }
}
```

### 4. Test Edge Cases

Always test:
- Empty/nil inputs
- Boundary values
- Error conditions
- Concurrent access (if applicable)

### 5. Keep Tests Independent

Each test should be able to run in isolation:

```go
func TestA(t *testing.T) {
    // Don't depend on TestB
}

func TestB(t *testing.T) {
    // Don't depend on TestA
}
```

### 6. Use Test Fixtures Sparingly

Prefer explicit setup in each test over shared fixtures:

❌ Bad:
```go
var globalTournament *tournament.Tournament

func init() {
    globalTournament = createTournament()
}
```

✅ Good:
```go
func TestTournament_End(t *testing.T) {
    tour := createTestTournament(t)
    // Test with fresh instance
}
```

## Continuous Integration

Tests should run automatically on:
- Every commit
- Every pull request
- Before deployment

Example CI configuration:

```yaml
# .github/workflows/test.yml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: '1.21'
      - run: go test -v -cover ./src/...
```

## Debugging Tests

### Print Debug Information

```go
func TestDebug(t *testing.T) {
    result := someFunction()
    t.Logf("Result: %+v", result)  // Only shown if test fails or -v flag
}
```

### Use Test Helpers

```go
func assertEqual(t *testing.T, expected, actual interface{}) {
    t.Helper()
    if expected != actual {
        t.Errorf("Expected %v, got %v", expected, actual)
    }
}
```

## Summary

- **Write tests first** (TDD)
- **Test behavior**, not implementation
- **Use mocks** for external dependencies
- **Keep tests simple** and focused
- **Aim for high coverage** in domain layer
- **Run tests frequently** during development
- **Maintain tests** as code evolves
