package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/adapter/postgres"
	"github.com/joshu-sajeev/paisa/internal/domain/account"
	"github.com/joshu-sajeev/paisa/internal/domain/jar"
)

func newTestJar(
	name string,
	allocationType jar.AllocationType,
	allocationValue int64,
) *jar.Jar {
	now := time.Now().UTC().Truncate(time.Microsecond)

	return &jar.Jar{
		ID:              uuid.New(),
		Name:            name + " " + uuid.NewString(),
		AllocationType:  allocationType,
		AllocationValue: allocationValue,
		IsArchived:      false,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func assertJar(t *testing.T, got, want *jar.Jar) {
	t.Helper()

	if got.ID != want.ID {
		t.Errorf("ID = %v, want %v", got.ID, want.ID)
	}

	if got.Name != want.Name {
		t.Errorf("Name = %q, want %q", got.Name, want.Name)
	}

	if got.AllocationType != want.AllocationType {
		t.Errorf(
			"AllocationType = %q, want %q",
			got.AllocationType,
			want.AllocationType,
		)
	}

	if got.AllocationValue != want.AllocationValue {
		t.Errorf(
			"AllocationValue = %d, want %d",
			got.AllocationValue,
			want.AllocationValue,
		)
	}

	if got.IsArchived != want.IsArchived {
		t.Errorf(
			"IsArchived = %v, want %v",
			got.IsArchived,
			want.IsArchived,
		)
	}

	if got.CreatedAt.IsZero() {
		t.Errorf("CreatedAt = %v, want non-zero time", got.CreatedAt)
	}

	if got.UpdatedAt.IsZero() {
		t.Errorf("UpdatedAt = %v, want non-zero time", got.UpdatedAt)
	}
}

// queryJar reads a jar directly from PostgreSQL.
// It is intentionally a test helper, not part of the repository port.
func queryJar(t *testing.T, id uuid.UUID) *jar.Jar {
	t.Helper()

	var j jar.Jar

	err := db.QueryRow(
		ctx,
		`
		SELECT
			id,
			name,
			allocation_type,
			allocation_value,
			is_archived,
			created_at,
			updated_at
		FROM jars
		WHERE id = $1
		`,
		id,
	).Scan(
		&j.ID,
		&j.Name,
		&j.AllocationType,
		&j.AllocationValue,
		&j.IsArchived,
		&j.CreatedAt,
		&j.UpdatedAt,
	)
	if err != nil {
		t.Fatalf("query jar: %v", err)
	}

	return &j
}

func TestJarCreate(t *testing.T) {
	t.Cleanup(func() {
		truncateTables(t, ctx, db)
	})

	tests := []struct {
		name string
		jar  *jar.Jar
	}{
		{
			name: "percentage allocation",
			jar:  newTestJar("Needs", jar.AllocationTypePercentage, 50),
		},
		{
			name: "fixed allocation",
			jar:  newTestJar("Insurance", jar.AllocationTypeFixed, 500000),
		},
		{
			name: "remainder allocation",
			jar:  newTestJar("Remainder", jar.AllocationTypeRemainder, 0),
		},
		{
			name: "one percent allocation",
			jar:  newTestJar("One Percent", jar.AllocationTypePercentage, 1),
		},
		{
			name: "one hundred percent allocation",
			jar:  newTestJar("Full Allocation", jar.AllocationTypePercentage, 100),
		},
		{
			name: "special characters",
			jar:  newTestJar(`Jar's "Savings"`, jar.AllocationTypePercentage, 25),
		},
		{
			name: "unicode",
			jar:  newTestJar("生活費", jar.AllocationTypePercentage, 30),
		},
		{
			name: "long name",
			jar: newTestJar(
				"VeryLongJarNameWithManyCharactersForTesting",
				jar.AllocationTypePercentage,
				20,
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := jarRepo.Create(ctx, tt.jar); err != nil {
				t.Fatalf("Create() error = %v", err)
			}

			got := queryJar(t, tt.jar.ID)

			assertJar(t, got, tt.jar)
		})
	}
}

func TestJarCreate_DuplicateName(t *testing.T) {
	t.Cleanup(func() {
		truncateTables(t, ctx, db)
	})

	first := newTestJar(
		"Savings",
		jar.AllocationTypePercentage,
		50,
	)

	if err := jarRepo.Create(ctx, first); err != nil {
		t.Fatalf("first Create() error = %v", err)
	}

	second := newTestJar(
		"Other",
		jar.AllocationTypePercentage,
		20,
	)
	second.Name = first.Name

	err := jarRepo.Create(ctx, second)

	if !errors.Is(err, jar.ErrJarNameExists) {
		t.Errorf(
			"Create() error = %v, want jar.ErrJarNameExists",
			err,
		)
	}
}

func TestJarCreate_ConcurrentDuplicateName(t *testing.T) {
	t.Cleanup(func() {
		truncateTables(t, ctx, db)
	})

	const numOps = 10
	const name = "Concurrent Savings"

	errChan := make(chan error, numOps)

	for range numOps {
		go func() {
			j := newTestJar(
				"Other",
				jar.AllocationTypePercentage,
				20,
			)
			j.Name = name

			errChan <- jarRepo.Create(ctx, j)
		}()
	}

	var successCount int
	var duplicateCount int

	for range numOps {
		err := <-errChan

		switch {
		case err == nil:
			successCount++

		case errors.Is(err, jar.ErrJarNameExists):
			duplicateCount++

		default:
			t.Errorf("unexpected Create() error = %v", err)
		}
	}

	if successCount != 1 {
		t.Errorf("successful creates = %d, want 1", successCount)
	}

	if duplicateCount != numOps-1 {
		t.Errorf(
			"duplicate errors = %d, want %d",
			duplicateCount,
			numOps-1,
		)
	}
}

func TestJarCreate_ConcurrentRemainder(t *testing.T) {
	t.Cleanup(func() {
		truncateTables(t, ctx, db)
	})

	const numOps = 10

	errChan := make(chan error, numOps)

	for range numOps {
		go func() {
			j := newTestJar(
				"Remainder",
				jar.AllocationTypeRemainder,
				0,
			)

			errChan <- jarRepo.Create(ctx, j)
		}()
	}

	var successCount int
	var duplicateCount int

	for range numOps {
		err := <-errChan

		switch {
		case err == nil:
			successCount++

		case errors.Is(err, jar.ErrJarNameExists):
			duplicateCount++

		default:
			t.Errorf("unexpected Create() error = %v", err)
		}
	}

	if successCount != 1 {
		t.Errorf("successful creates = %d, want 1", successCount)
	}

	if duplicateCount != numOps-1 {
		t.Errorf(
			"duplicate errors = %d, want %d",
			duplicateCount,
			numOps-1,
		)
	}
}

func TestJarList(t *testing.T) {
	t.Cleanup(func() {
		truncateTables(t, ctx, db)
	})

	jar1 := newTestJar(
		"Needs",
		jar.AllocationTypePercentage,
		50,
	)
	jar1.CreatedAt = time.Date(
		2026, 1, 1, 0, 0, 0, 0, time.UTC,
	)
	jar1.UpdatedAt = jar1.CreatedAt

	jar2 := newTestJar(
		"Leisure",
		jar.AllocationTypePercentage,
		30,
	)
	jar2.CreatedAt = time.Date(
		2026, 1, 2, 0, 0, 0, 0, time.UTC,
	)
	jar2.UpdatedAt = jar2.CreatedAt

	archived := newTestJar(
		"Archived",
		jar.AllocationTypePercentage,
		20,
	)
	archived.IsArchived = true
	archived.CreatedAt = time.Date(
		2026, 1, 3, 0, 0, 0, 0, time.UTC,
	)
	archived.UpdatedAt = archived.CreatedAt

	for _, j := range []*jar.Jar{
		jar1,
		jar2,
		archived,
	} {
		if err := jarRepo.Create(ctx, j); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	got, err := jarRepo.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(got) != 3 {
		t.Fatalf("List() returned %d jars, want 3", len(got))
	}

	assertJar(t, got[0], archived)
	assertJar(t, got[1], jar2)
	assertJar(t, got[2], jar1)
}

func TestJarList_Empty(t *testing.T) {
	t.Cleanup(func() {
		truncateTables(t, ctx, db)
	})

	got, err := jarRepo.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if got == nil {
		t.Fatal("List() returned nil, want empty slice")
	}

	if len(got) != 0 {
		t.Fatalf("List() returned %d jars, want 0", len(got))
	}
}

func TestJarFindByID(t *testing.T) {
	t.Cleanup(func() {
		truncateTables(t, ctx, db)
	})

	j := newTestJar(
		"Needs",
		jar.AllocationTypePercentage,
		50,
	)

	if err := jarRepo.Create(ctx, j); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	tests := []struct {
		name    string
		id      uuid.UUID
		wantErr error
	}{
		{
			name: "found",
			id:   j.ID,
		},
		{
			name:    "not found",
			id:      uuid.New(),
			wantErr: jar.ErrJarNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jarRepo.FindByID(ctx, tt.id)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf(
						"FindByID() error = %v, want %v",
						err,
						tt.wantErr,
					)
				}
				return
			}

			if err != nil {
				t.Fatalf("FindByID() error = %v", err)
			}

			assertJar(t, got, j)
		})
	}
}

func TestJarSave(t *testing.T) {
	t.Cleanup(func() {
		truncateTables(t, ctx, db)
	})

	tests := []struct {
		name    string
		setup   func(t *testing.T) *jar.Jar
		wantErr error
	}{
		{
			name: "update name",
			setup: func(t *testing.T) *jar.Jar {
				j := newTestJar(
					"Original",
					jar.AllocationTypePercentage,
					50,
				)

				if err := jarRepo.Create(ctx, j); err != nil {
					t.Fatalf("Create() error = %v", err)
				}

				j.Name = "Updated Name"
				j.UpdatedAt = j.UpdatedAt.Add(time.Second)

				return j
			},
		},
		{
			name: "archive",
			setup: func(t *testing.T) *jar.Jar {
				j := newTestJar(
					"Original",
					jar.AllocationTypePercentage,
					50,
				)

				if err := jarRepo.Create(ctx, j); err != nil {
					t.Fatalf("Create() error = %v", err)
				}

				j.IsArchived = true
				j.UpdatedAt = j.UpdatedAt.Add(time.Second)

				return j
			},
		},
		{
			name: "update allocation",
			setup: func(t *testing.T) *jar.Jar {
				j := newTestJar(
					"Original",
					jar.AllocationTypePercentage,
					50,
				)

				if err := jarRepo.Create(ctx, j); err != nil {
					t.Fatalf("Create() error = %v", err)
				}

				j.AllocationType = jar.AllocationTypeFixed
				j.AllocationValue = 500000
				j.UpdatedAt = j.UpdatedAt.Add(time.Second)

				return j
			},
		},
		{
			name: "update name and archive",
			setup: func(t *testing.T) *jar.Jar {
				j := newTestJar(
					"Original",
					jar.AllocationTypePercentage,
					50,
				)

				if err := jarRepo.Create(ctx, j); err != nil {
					t.Fatalf("Create() error = %v", err)
				}

				j.Name = "Archived Jar"
				j.IsArchived = true
				j.UpdatedAt = j.UpdatedAt.Add(time.Second)

				return j
			},
		},
		{
			name: "unicode name",
			setup: func(t *testing.T) *jar.Jar {
				j := newTestJar(
					"Original",
					jar.AllocationTypePercentage,
					50,
				)

				if err := jarRepo.Create(ctx, j); err != nil {
					t.Fatalf("Create() error = %v", err)
				}

				j.Name = "生活費"
				j.UpdatedAt = j.UpdatedAt.Add(time.Second)

				return j
			},
		},
		{
			name: "duplicate name",
			setup: func(t *testing.T) *jar.Jar {
				first := newTestJar(
					"First",
					jar.AllocationTypePercentage,
					50,
				)
				second := newTestJar(
					"Second",
					jar.AllocationTypePercentage,
					50,
				)

				if err := jarRepo.Create(ctx, first); err != nil {
					t.Fatalf("first Create() error = %v", err)
				}

				if err := jarRepo.Create(ctx, second); err != nil {
					t.Fatalf("second Create() error = %v", err)
				}

				second.Name = first.Name

				return second
			},
			wantErr: jar.ErrJarNameExists,
		},
		{
			name: "not found",
			setup: func(t *testing.T) *jar.Jar {
				return newTestJar(
					"Missing",
					jar.AllocationTypePercentage,
					50,
				)
			},
			wantErr: jar.ErrJarNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := tt.setup(t)

			err := jarRepo.Save(ctx, j)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf(
						"Save() error = %v, want %v",
						err,
						tt.wantErr,
					)
				}
				return
			}

			if err != nil {
				t.Fatalf("Save() error = %v", err)
			}

			got := queryJar(t, j.ID)

			assertJar(t, got, j)
		})
	}
}

func TestJarUpdateAllocations(t *testing.T) {
	t.Cleanup(func() {
		truncateTables(t, ctx, db)
	})

	j1 := newTestJar(
		"Needs",
		jar.AllocationTypePercentage,
		50,
	)

	j2 := newTestJar(
		"Leisure",
		jar.AllocationTypeFixed,
		100000,
	)

	if err := jarRepo.Create(ctx, j1); err != nil {
		t.Fatalf("Create() j1 error = %v", err)
	}

	if err := jarRepo.Create(ctx, j2); err != nil {
		t.Fatalf("Create() j2 error = %v", err)
	}

	if err := j1.UpdateAllocation(
		jar.AllocationTypeFixed,
		500000,
	); err != nil {
		t.Fatalf("UpdateAllocation() j1 error = %v", err)
	}

	if err := j2.UpdateAllocation(
		jar.AllocationTypePercentage,
		25,
	); err != nil {
		t.Fatalf("UpdateAllocation() j2 error = %v", err)
	}

	updatedAt1 := j1.UpdatedAt
	updatedAt2 := j2.UpdatedAt

	if err := jarRepo.UpdateAllocations(
		ctx,
		[]*jar.Jar{j1, j2},
	); err != nil {
		t.Fatalf("UpdateAllocations() error = %v", err)
	}

	want1 := *j1
	want1.UpdatedAt = updatedAt1

	want2 := *j2
	want2.UpdatedAt = updatedAt2

	assertJar(t, queryJar(t, j1.ID), &want1)
	assertJar(t, queryJar(t, j2.ID), &want2)
}

func TestJarUpdateAllocations_NotFound(t *testing.T) {
	t.Cleanup(func() {
		truncateTables(t, ctx, db)
	})

	j := newTestJar(
		"Needs",
		jar.AllocationTypePercentage,
		50,
	)
	j.ID = uuid.New()

	err := jarRepo.UpdateAllocations(
		ctx,
		[]*jar.Jar{j},
	)

	if !errors.Is(err, jar.ErrJarNotFound) {
		t.Errorf(
			"UpdateAllocations() error = %v, want %v",
			err,
			jar.ErrJarNotFound,
		)
	}
}

func TestJarUpdateAllocations_Empty(t *testing.T) {
	t.Cleanup(func() {
		truncateTables(t, ctx, db)
	})

	if err := jarRepo.UpdateAllocations(
		ctx,
		[]*jar.Jar{},
	); err != nil {
		t.Fatalf(
			"UpdateAllocations() error = %v, want nil",
			err,
		)
	}
}

func TestTxManager_WithinTransaction_Commits(t *testing.T) {
	t.Cleanup(func() {
		truncateTables(t, ctx, db)
	})

	repo := postgres.NewJarRepository(db)
	txManager := postgres.NewTxManager(db)

	j := newTestJar(
		"Transaction Commit Test",
		jar.AllocationTypePercentage,
		100,
	)

	err := txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		return repo.Create(txCtx, j)
	})
	if err != nil {
		t.Fatalf("WithinTransaction() error = %v", err)
	}

	got := queryJar(t, j.ID)

	assertJar(t, got, j)
}

func TestTxManager_WithinTransaction_RollsBack(t *testing.T) {
	t.Cleanup(func() {
		truncateTables(t, ctx, db)
	})

	repo := postgres.NewJarRepository(db)
	txManager := postgres.NewTxManager(db)

	j := newTestJar(
		"Transaction Rollback Test",
		jar.AllocationTypePercentage,
		100,
	)

	expectedErr := errors.New("intentional failure")

	err := txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := repo.Create(txCtx, j); err != nil {
			return err
		}

		return expectedErr
	})

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"WithinTransaction() error = %v, want %v",
			err,
			expectedErr,
		)
	}

	_, err = jarRepo.FindByID(ctx, j.ID)
	if !errors.Is(err, jar.ErrJarNotFound) {
		t.Fatalf(
			"FindByID() error = %v, want %v",
			err,
			jar.ErrJarNotFound,
		)
	}
}

func TestTxManager_WithinTransaction_RollsBackMultipleRepositories(
	t *testing.T,
) {
	t.Cleanup(func() {
		truncateTables(t, ctx, db)
	})

	accountRepo := postgres.NewAccountRepository(db)
	jarRepo := postgres.NewJarRepository(db)
	txManager := postgres.NewTxManager(db)

	acc := newTestAccount("Transaction Test Account")
	j := newTestJar(
		"Transaction Test Jar",
		jar.AllocationTypePercentage,
		100,
	)

	expectedErr := errors.New("intentional failure")

	err := txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := accountRepo.Create(txCtx, acc); err != nil {
			return err
		}

		if err := jarRepo.Create(txCtx, j); err != nil {
			return err
		}

		return expectedErr
	})

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"WithinTransaction() error = %v, want %v",
			err,
			expectedErr,
		)
	}

	_, err = accountRepo.FindByID(ctx, acc.ID)
	if !errors.Is(err, account.ErrAccountNotFound) {
		t.Fatalf(
			"FindByID() account error = %v, want %v",
			err,
			account.ErrAccountNotFound,
		)
	}

	_, err = jarRepo.FindByID(ctx, j.ID)
	if !errors.Is(err, jar.ErrJarNotFound) {
		t.Fatalf(
			"FindByID() jar error = %v, want %v",
			err,
			jar.ErrJarNotFound,
		)
	}
}
