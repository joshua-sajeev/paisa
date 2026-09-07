package application_test

import (
	"context"
	"errors"
	"log/slog"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/application"
	"github.com/joshu-sajeev/paisa/internal/domain/jar"
)

var errRepository = errors.New("database error")

// MockJarRepository implements ports.JarRepository for testing.
type MockJarRepository struct {
	jars []*jar.Jar

	ListCalls              int
	CreateCalls            int
	FindByIDCalls          int
	SaveCalls              int
	UpdateAllocationsCalls int

	ListError              error
	CreateError            error
	FindByIDError          error
	SaveError              error
	UpdateAllocationsError error
}

func (m *MockJarRepository) List(ctx context.Context) ([]*jar.Jar, error) {
	m.ListCalls++

	if m.ListError != nil {
		return nil, m.ListError
	}

	active := make([]*jar.Jar, 0, len(m.jars))

	for _, j := range m.jars {
		if !j.IsArchived {
			active = append(active, j)
		}
	}

	return active, nil
}

func (m *MockJarRepository) Create(ctx context.Context, j *jar.Jar) error {
	m.CreateCalls++

	if m.CreateError != nil {
		return m.CreateError
	}

	m.jars = append(m.jars, j)
	return nil
}

func (m *MockJarRepository) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*jar.Jar, error) {
	m.FindByIDCalls++

	if m.FindByIDError != nil {
		return nil, m.FindByIDError
	}

	for _, existing := range m.jars {
		if existing.ID == id {
			return existing, nil
		}
	}

	return nil, jar.ErrJarNotFound
}

func (m *MockJarRepository) Save(
	ctx context.Context,
	j *jar.Jar,
) error {
	m.SaveCalls++

	if m.SaveError != nil {
		return m.SaveError
	}

	for i, existing := range m.jars {
		if existing.ID == j.ID {
			m.jars[i] = j
			return nil
		}
	}

	return jar.ErrJarNotFound
}

func (m *MockJarRepository) UpdateAllocations(
	ctx context.Context,
	jars []*jar.Jar,
) error {
	m.UpdateAllocationsCalls++

	if m.UpdateAllocationsError != nil {
		return m.UpdateAllocationsError
	}

	for _, updated := range jars {
		for i, existing := range m.jars {
			if existing.ID != updated.ID {
				continue
			}

			m.jars[i] = updated
			break
		}
	}

	return nil
}

func mustNewJar(
	t *testing.T,
	name string,
	allocationType jar.AllocationType,
	value int64,
) *jar.Jar {
	t.Helper()

	j, err := jar.NewJar(name, allocationType, value)
	if err != nil {
		t.Fatalf("failed to create test jar: %v", err)
	}

	return j
}

func requireError(t *testing.T, err, want error) {
	t.Helper()

	if want == nil {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		return
	}

	if !errors.Is(err, want) {
		t.Fatalf("expected error %v, got %v", want, err)
	}
}

func TestJarServiceCreate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		existingJars    []*jar.Jar
		jarName         string
		allocationType  jar.AllocationType
		allocationValue int64
		listErr         error
		createErr       error
		wantErr         error
		wantCreateCalls int
		wantListCalls   int
		wantName        string
	}{
		{
			name:            "create remainder",
			existingJars:    []*jar.Jar{},
			jarName:         "Savings",
			allocationType:  jar.AllocationTypeRemainder,
			allocationValue: 0,
			wantCreateCalls: 1,
			wantListCalls:   1,
			wantName:        "Savings",
		},
		{
			name: "create percentage with remainder",
			existingJars: []*jar.Jar{
				mustNewJar(t, "Remainder", jar.AllocationTypeRemainder, 0),
			},
			jarName:         "Needs",
			allocationType:  jar.AllocationTypePercentage,
			allocationValue: 50,
			wantCreateCalls: 1,
			wantListCalls:   1,
			wantName:        "Needs",
		},
		{
			name: "create remainder with archived remainder",
			existingJars: []*jar.Jar{
				func() *jar.Jar {
					j := mustNewJar(
						t,
						"Archived Remainder",
						jar.AllocationTypeRemainder,
						0,
					)
					j.IsArchived = true
					return j
				}(),
			},
			jarName:         "Remainder",
			allocationType:  jar.AllocationTypeRemainder,
			allocationValue: 0,
			wantCreateCalls: 1,
			wantListCalls:   1,
			wantName:        "Remainder",
		},
		{
			name: "create fixed with remainder",
			existingJars: []*jar.Jar{
				mustNewJar(t, "Remainder", jar.AllocationTypeRemainder, 0),
			},
			jarName:         "Insurance",
			allocationType:  jar.AllocationTypeFixed,
			allocationValue: 500000,
			wantCreateCalls: 1,
			wantListCalls:   1,
			wantName:        "Insurance",
		},
		{
			name: "create percentage and fixed with remainder",
			existingJars: []*jar.Jar{
				mustNewJar(t, "Needs", jar.AllocationTypePercentage, 50),
				mustNewJar(t, "Remainder", jar.AllocationTypeRemainder, 0),
			},
			jarName:         "Insurance",
			allocationType:  jar.AllocationTypeFixed,
			allocationValue: 500000,
			wantCreateCalls: 1,
			wantListCalls:   1,
			wantName:        "Insurance",
		},
		{
			name:            "empty name",
			existingJars:    []*jar.Jar{},
			jarName:         "",
			allocationType:  jar.AllocationTypeRemainder,
			allocationValue: 0,
			wantErr:         jar.ErrInvalidName,
			wantListCalls:   0,
			wantCreateCalls: 0,
		},
		{
			name:            "invalid percentage",
			existingJars:    []*jar.Jar{},
			jarName:         "Needs",
			allocationType:  jar.AllocationTypePercentage,
			allocationValue: 101,
			wantErr:         jar.ErrInvalidAllocationVal,
			wantListCalls:   0,
			wantCreateCalls: 0,
		},
		{
			name:            "invalid fixed amount",
			existingJars:    []*jar.Jar{},
			jarName:         "Insurance",
			allocationType:  jar.AllocationTypeFixed,
			allocationValue: 0,
			wantErr:         jar.ErrInvalidAllocationVal,
			wantListCalls:   0,
			wantCreateCalls: 0,
		},
		{
			name: "second remainder",
			existingJars: []*jar.Jar{
				mustNewJar(t, "Remainder", jar.AllocationTypeRemainder, 0),
			},
			jarName:         "Another",
			allocationType:  jar.AllocationTypeRemainder,
			allocationValue: 0,
			wantErr:         jar.ErrMultipleRemainders,
			wantListCalls:   1,
			wantCreateCalls: 0,
		},
		{
			name:            "percentage without remainder",
			existingJars:    []*jar.Jar{},
			jarName:         "Needs",
			allocationType:  jar.AllocationTypePercentage,
			allocationValue: 50,
			wantErr:         jar.ErrNoRemainder,
			wantListCalls:   1,
			wantCreateCalls: 0,
		},
		{
			name:            "fixed without remainder",
			existingJars:    []*jar.Jar{},
			jarName:         "Insurance",
			allocationType:  jar.AllocationTypeFixed,
			allocationValue: 500000,
			wantErr:         jar.ErrNoRemainder,
			wantListCalls:   1,
			wantCreateCalls: 0,
		},
		{
			name: "percentage total reaches 100",
			existingJars: []*jar.Jar{
				mustNewJar(t, "Remainder", jar.AllocationTypeRemainder, 0),
				mustNewJar(t, "Needs", jar.AllocationTypePercentage, 60),
			},
			jarName:         "Leisure",
			allocationType:  jar.AllocationTypePercentage,
			allocationValue: 40,
			wantErr:         jar.ErrPercentageExceedsLimit,
			wantListCalls:   1,
			wantCreateCalls: 0,
		},
		{
			name: "repository list error",
			existingJars: []*jar.Jar{
				mustNewJar(t, "Remainder", jar.AllocationTypeRemainder, 0),
			},
			jarName:         "Needs",
			allocationType:  jar.AllocationTypePercentage,
			allocationValue: 50,
			listErr:         errRepository,
			wantErr:         errRepository,
			wantListCalls:   1,
			wantCreateCalls: 0,
		},
		{
			name: "repository create error",
			existingJars: []*jar.Jar{
				mustNewJar(t, "Remainder", jar.AllocationTypeRemainder, 0),
			},
			jarName:         "Needs",
			allocationType:  jar.AllocationTypePercentage,
			allocationValue: 50,
			createErr:       jar.ErrJarNameExists,
			wantErr:         jar.ErrJarNameExists,
			wantListCalls:   1,
			wantCreateCalls: 1,
		},
		{
			name:            "whitespace is trimmed",
			existingJars:    []*jar.Jar{},
			jarName:         "  Savings  ",
			allocationType:  jar.AllocationTypeRemainder,
			allocationValue: 0,
			wantCreateCalls: 1,
			wantListCalls:   1,
			wantName:        "Savings",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockJarRepository{
				jars:        tt.existingJars,
				ListError:   tt.listErr,
				CreateError: tt.createErr,
			}

			service := application.NewJarService(repo, slog.Default())

			got, err := service.Create(
				context.Background(),
				tt.jarName,
				tt.allocationType,
				tt.allocationValue,
			)

			requireError(t, err, tt.wantErr)

			if repo.ListCalls != tt.wantListCalls {
				t.Fatalf(
					"expected %d List calls, got %d",
					tt.wantListCalls,
					repo.ListCalls,
				)
			}

			if repo.CreateCalls != tt.wantCreateCalls {
				t.Fatalf(
					"expected %d Create calls, got %d",
					tt.wantCreateCalls,
					repo.CreateCalls,
				)
			}

			if tt.wantErr == nil {
				if got == nil {
					t.Fatal("expected created jar, got nil")
				}

				if got.Name != tt.wantName {
					t.Fatalf(
						"expected name %q, got %q",
						tt.wantName,
						got.Name,
					)
				}
			}
		})
	}
}

func TestJarServiceList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		jars    []*jar.Jar
		listErr error
		wantLen int
		wantErr error
	}{
		{
			name: "empty",
			jars: []*jar.Jar{},
		},
		{
			name: "multiple jars",
			jars: []*jar.Jar{
				mustNewJar(t, "Needs", jar.AllocationTypePercentage, 50),
				mustNewJar(t, "Remainder", jar.AllocationTypeRemainder, 0),
			},
			wantLen: 2,
		},
		{
			name: "excludes archived jars",
			jars: []*jar.Jar{
				mustNewJar(t, "Needs", jar.AllocationTypePercentage, 50),
				func() *jar.Jar {
					j := mustNewJar(
						t,
						"Archived",
						jar.AllocationTypePercentage,
						20,
					)
					j.IsArchived = true
					return j
				}(),
			},
			wantLen: 1,
		},
		{
			name:    "repository error",
			listErr: errRepository,
			wantErr: errRepository,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockJarRepository{
				jars:      tt.jars,
				ListError: tt.listErr,
			}

			service := application.NewJarService(repo, slog.Default())

			got, err := service.List(context.Background())

			requireError(t, err, tt.wantErr)

			if tt.wantErr == nil && len(got) != tt.wantLen {
				t.Fatalf(
					"expected %d jars, got %d",
					tt.wantLen,
					len(got),
				)
			}

			if repo.ListCalls != 1 {
				t.Fatalf(
					"expected 1 List call, got %d",
					repo.ListCalls,
				)
			}
		})
	}
}

func TestJarServiceUpdate(t *testing.T) {
	t.Parallel()

	name := "Updated"

	tests := []struct {
		name         string
		setup        func(t *testing.T) ([]*jar.Jar, uuid.UUID)
		newName      *string
		isArchived   *bool
		findErr      error
		saveErr      error
		wantErr      error
		wantName     string
		wantArchived bool
		wantFindCall int
		wantListCall int
		wantSaveCall int
	}{
		{
			name: "update name",
			setup: func(t *testing.T) ([]*jar.Jar, uuid.UUID) {
				j := mustNewJar(
					t,
					"Original",
					jar.AllocationTypeRemainder,
					0,
				)

				return []*jar.Jar{j}, j.ID
			},
			newName:      &name,
			wantName:     "Updated",
			wantFindCall: 1,
			wantSaveCall: 1,
		},
		{
			name: "archive jar",
			setup: func(t *testing.T) ([]*jar.Jar, uuid.UUID) {
				j := mustNewJar(
					t,
					"Original",
					jar.AllocationTypePercentage,
					50,
				)
				remainder := mustNewJar(
					t,
					"Remainder",
					jar.AllocationTypeRemainder,
					0,
				)

				return []*jar.Jar{j, remainder}, j.ID
			},
			isArchived:   func() *bool { v := true; return &v }(),
			wantName:     "Original",
			wantArchived: true,
			wantFindCall: 1,
			wantListCall: 1,
			wantSaveCall: 1,
		},
		{
			name: "unarchive jar",
			setup: func(t *testing.T) ([]*jar.Jar, uuid.UUID) {
				needs := mustNewJar(
					t,
					"Needs",
					jar.AllocationTypePercentage,
					50,
				)

				remainder := mustNewJar(
					t,
					"Remainder",
					jar.AllocationTypeRemainder,
					0,
				)

				archived := mustNewJar(
					t,
					"Archived",
					jar.AllocationTypePercentage,
					20,
				)
				archived.IsArchived = true

				return []*jar.Jar{
					needs,
					remainder,
					archived,
				}, archived.ID
			},
			isArchived:   func() *bool { v := false; return &v }(),
			wantName:     "Archived",
			wantArchived: false,
			wantFindCall: 1,
			wantListCall: 1,
			wantSaveCall: 1,
		},
		{
			name: "archive state unchanged",
			setup: func(t *testing.T) ([]*jar.Jar, uuid.UUID) {
				j := mustNewJar(
					t,
					"Remainder",
					jar.AllocationTypeRemainder,
					0,
				)

				return []*jar.Jar{j}, j.ID
			},
			isArchived:   func() *bool { v := false; return &v }(),
			wantName:     "Remainder",
			wantArchived: false,
			wantFindCall: 1,
			wantListCall: 0,
			wantSaveCall: 0,
		},
		{
			name: "update name and archive",
			setup: func(t *testing.T) ([]*jar.Jar, uuid.UUID) {
				j := mustNewJar(
					t,
					"Original",
					jar.AllocationTypePercentage,
					50,
				)
				remainder := mustNewJar(
					t,
					"Remainder",
					jar.AllocationTypeRemainder,
					0,
				)

				return []*jar.Jar{j, remainder}, j.ID
			},
			newName:      &name,
			isArchived:   func() *bool { v := true; return &v }(),
			wantName:     "Updated",
			wantArchived: true,
			wantFindCall: 1,
			wantListCall: 1,
			wantSaveCall: 1,
		},
		{
			name: "jar not found",
			setup: func(t *testing.T) ([]*jar.Jar, uuid.UUID) {
				j := mustNewJar(
					t,
					"Original",
					jar.AllocationTypeRemainder,
					0,
				)

				return []*jar.Jar{j}, uuid.New()
			},
			wantErr:      jar.ErrJarNotFound,
			wantFindCall: 1,
		},
		{
			name: "find repository error",
			setup: func(t *testing.T) ([]*jar.Jar, uuid.UUID) {
				j := mustNewJar(
					t,
					"Original",
					jar.AllocationTypeRemainder,
					0,
				)

				return []*jar.Jar{j}, j.ID
			},
			findErr:      errRepository,
			wantErr:      errRepository,
			wantFindCall: 1,
		},
		{
			name: "save repository error",
			setup: func(t *testing.T) ([]*jar.Jar, uuid.UUID) {
				j := mustNewJar(
					t,
					"Original",
					jar.AllocationTypeRemainder,
					0,
				)

				return []*jar.Jar{j}, j.ID
			},
			newName:      &name,
			saveErr:      errRepository,
			wantErr:      errRepository,
			wantFindCall: 1,
			wantSaveCall: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jars, id := tt.setup(t)

			repo := &MockJarRepository{
				jars:          jars,
				FindByIDError: tt.findErr,
				SaveError:     tt.saveErr,
			}

			service := application.NewJarService(
				repo,
				slog.Default(),
			)

			err := service.Update(
				context.Background(),
				id,
				tt.newName,
				tt.isArchived,
			)

			requireError(t, err, tt.wantErr)

			if repo.FindByIDCalls != tt.wantFindCall {
				t.Fatalf(
					"expected %d FindByID calls, got %d",
					tt.wantFindCall,
					repo.FindByIDCalls,
				)
			}

			if repo.ListCalls != tt.wantListCall {
				t.Fatalf(
					"expected %d List calls, got %d",
					tt.wantListCall,
					repo.ListCalls,
				)
			}

			if repo.SaveCalls != tt.wantSaveCall {
				t.Fatalf(
					"expected %d Save calls, got %d",
					tt.wantSaveCall,
					repo.SaveCalls,
				)
			}

			if tt.wantErr != nil {
				return
			}

			var updated *jar.Jar

			for _, j := range repo.jars {
				if j.ID == id {
					updated = j
					break
				}
			}

			if updated == nil {
				t.Fatal("updated jar not found")
			}

			if updated.Name != tt.wantName {
				t.Fatalf(
					"expected name %q, got %q",
					tt.wantName,
					updated.Name,
				)
			}

			if updated.IsArchived != tt.wantArchived {
				t.Fatalf(
					"expected archived=%v, got %v",
					tt.wantArchived,
					updated.IsArchived,
				)
			}
		})
	}
}

func TestJarServiceUpdateAllocations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func(t *testing.T) ([]*jar.Jar, []jar.JarAllocationUpdate)
		repoErr   error
		wantErr   error
		wantCalls int
		wantType  jar.AllocationType
		wantValue int64
	}{
		{
			name: "change percentage",
			setup: func(t *testing.T) ([]*jar.Jar, []jar.JarAllocationUpdate) {
				needs := mustNewJar(
					t,
					"Needs",
					jar.AllocationTypePercentage,
					50,
				)
				remainder := mustNewJar(
					t,
					"Remainder",
					jar.AllocationTypeRemainder,
					0,
				)

				return []*jar.Jar{needs, remainder}, []jar.JarAllocationUpdate{
					{
						ID:              needs.ID,
						AllocationType:  jar.AllocationTypePercentage,
						AllocationValue: 60,
					},
				}
			},
			wantCalls: 1,
			wantType:  jar.AllocationTypePercentage,
			wantValue: 60,
		},
		{
			name: "change percentage to fixed with remainder",
			setup: func(t *testing.T) ([]*jar.Jar, []jar.JarAllocationUpdate) {
				needs := mustNewJar(
					t,
					"Needs",
					jar.AllocationTypePercentage,
					50,
				)
				remainder := mustNewJar(
					t,
					"Remainder",
					jar.AllocationTypeRemainder,
					0,
				)

				return []*jar.Jar{needs, remainder}, []jar.JarAllocationUpdate{
					{
						ID:              needs.ID,
						AllocationType:  jar.AllocationTypeFixed,
						AllocationValue: 500000,
					},
				}
			},
			wantCalls: 1,
			wantType:  jar.AllocationTypeFixed,
			wantValue: 500000,
		},
		{
			name: "invalid remainder value",
			setup: func(t *testing.T) ([]*jar.Jar, []jar.JarAllocationUpdate) {
				remainder := mustNewJar(
					t,
					"Remainder",
					jar.AllocationTypeRemainder,
					0,
				)

				return []*jar.Jar{remainder}, []jar.JarAllocationUpdate{
					{
						ID:              remainder.ID,
						AllocationType:  jar.AllocationTypeRemainder,
						AllocationValue: 100,
					},
				}
			},
			wantErr:   jar.ErrInvalidAllocationVal,
			wantCalls: 0,
		},
		{
			name: "remove remainder",
			setup: func(t *testing.T) ([]*jar.Jar, []jar.JarAllocationUpdate) {
				needs := mustNewJar(
					t,
					"Needs",
					jar.AllocationTypePercentage,
					50,
				)
				remainder := mustNewJar(
					t,
					"Remainder",
					jar.AllocationTypeRemainder,
					0,
				)

				return []*jar.Jar{needs, remainder}, []jar.JarAllocationUpdate{
					{
						ID:              remainder.ID,
						AllocationType:  jar.AllocationTypePercentage,
						AllocationValue: 50,
					},
				}
			},
			wantErr:   jar.ErrNoRemainder,
			wantCalls: 0,
		},
		{
			name: "percentage reaches 100",
			setup: func(t *testing.T) ([]*jar.Jar, []jar.JarAllocationUpdate) {
				needs := mustNewJar(
					t,
					"Needs",
					jar.AllocationTypePercentage,
					50,
				)
				leisure := mustNewJar(
					t,
					"Leisure",
					jar.AllocationTypePercentage,
					30,
				)
				remainder := mustNewJar(
					t,
					"Remainder",
					jar.AllocationTypeRemainder,
					0,
				)

				return []*jar.Jar{needs, leisure, remainder}, []jar.JarAllocationUpdate{
					{
						ID:              leisure.ID,
						AllocationType:  jar.AllocationTypePercentage,
						AllocationValue: 50,
					},
				}
			},
			wantErr:   jar.ErrPercentageExceedsLimit,
			wantCalls: 0,
		},
		{
			name: "percentage fixed and remainder allowed",
			setup: func(t *testing.T) ([]*jar.Jar, []jar.JarAllocationUpdate) {
				needs := mustNewJar(
					t,
					"Needs",
					jar.AllocationTypePercentage,
					50,
				)
				insurance := mustNewJar(
					t,
					"Insurance",
					jar.AllocationTypeFixed,
					500000,
				)
				remainder := mustNewJar(
					t,
					"Remainder",
					jar.AllocationTypeRemainder,
					0,
				)

				return []*jar.Jar{
						needs,
						insurance,
						remainder,
					}, []jar.JarAllocationUpdate{
						{
							ID:              insurance.ID,
							AllocationType:  jar.AllocationTypeFixed,
							AllocationValue: 600000,
						},
					}
			},
			wantCalls: 1,
			wantType:  jar.AllocationTypeFixed,
			wantValue: 600000,
		},
		{
			name: "multiple updates",
			setup: func(t *testing.T) ([]*jar.Jar, []jar.JarAllocationUpdate) {
				needs := mustNewJar(
					t,
					"Needs",
					jar.AllocationTypePercentage,
					50,
				)
				leisure := mustNewJar(
					t,
					"Leisure",
					jar.AllocationTypePercentage,
					20,
				)
				insurance := mustNewJar(
					t,
					"Insurance",
					jar.AllocationTypeFixed,
					500000,
				)
				remainder := mustNewJar(
					t,
					"Remainder",
					jar.AllocationTypeRemainder,
					0,
				)

				return []*jar.Jar{
						needs,
						leisure,
						insurance,
						remainder,
					}, []jar.JarAllocationUpdate{
						{
							ID:              needs.ID,
							AllocationType:  jar.AllocationTypePercentage,
							AllocationValue: 40,
						},
						{
							ID:              insurance.ID,
							AllocationType:  jar.AllocationTypeFixed,
							AllocationValue: 700000,
						},
					}
			},
			wantCalls: 1,
		},
		{
			name: "repository error",
			setup: func(t *testing.T) ([]*jar.Jar, []jar.JarAllocationUpdate) {
				needs := mustNewJar(
					t,
					"Needs",
					jar.AllocationTypePercentage,
					50,
				)
				remainder := mustNewJar(
					t,
					"Remainder",
					jar.AllocationTypeRemainder,
					0,
				)

				return []*jar.Jar{needs, remainder}, []jar.JarAllocationUpdate{
					{
						ID:              needs.ID,
						AllocationType:  jar.AllocationTypePercentage,
						AllocationValue: 60,
					},
				}
			},
			repoErr:   errRepository,
			wantErr:   errRepository,
			wantCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jars, updates := tt.setup(t)

			repo := &MockJarRepository{
				jars:                   jars,
				UpdateAllocationsError: tt.repoErr,
			}

			service := application.NewJarService(repo, slog.Default())

			err := service.UpdateAllocations(
				context.Background(),
				updates,
			)

			requireError(t, err, tt.wantErr)

			if repo.ListCalls != 1 {
				t.Fatalf(
					"expected 1 List call, got %d",
					repo.ListCalls,
				)
			}

			if repo.UpdateAllocationsCalls != tt.wantCalls {
				t.Fatalf(
					"expected %d UpdateAllocations calls, got %d",
					tt.wantCalls,
					repo.UpdateAllocationsCalls,
				)
			}

			if tt.wantErr == nil && len(updates) == 1 {
				var updated *jar.Jar

				for _, j := range repo.jars {
					if j.ID == updates[0].ID {
						updated = j
						break
					}
				}

				if updated == nil {
					t.Fatal("updated jar not found")
				}

				if tt.wantType != "" &&
					updated.AllocationType != tt.wantType {
					t.Fatalf(
						"expected allocation type %q, got %q",
						tt.wantType,
						updated.AllocationType,
					)
				}

				if tt.wantValue != 0 &&
					updated.AllocationValue != tt.wantValue {
					t.Fatalf(
						"expected allocation value %d, got %d",
						tt.wantValue,
						updated.AllocationValue,
					)
				}
			}
		})
	}
}

func TestJarServiceAllocateIncome(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		income  int64
		setup   func(t *testing.T) ([]*jar.Jar, map[uuid.UUID]int64)
		repoErr error
		wantErr error
	}{
		{
			name:   "remainder only",
			income: 1000000,
			setup: func(t *testing.T) ([]*jar.Jar, map[uuid.UUID]int64) {
				remainder := mustNewJar(
					t,
					"Remainder",
					jar.AllocationTypeRemainder,
					0,
				)

				return []*jar.Jar{remainder}, map[uuid.UUID]int64{
					remainder.ID: 1000000,
				}
			},
		},
		{
			name:   "percentage and remainder",
			income: 1000000,
			setup: func(t *testing.T) ([]*jar.Jar, map[uuid.UUID]int64) {
				needs := mustNewJar(
					t,
					"Needs",
					jar.AllocationTypePercentage,
					50,
				)
				remainder := mustNewJar(
					t,
					"Remainder",
					jar.AllocationTypeRemainder,
					0,
				)

				return []*jar.Jar{needs, remainder}, map[uuid.UUID]int64{
					needs.ID:     500000,
					remainder.ID: 500000,
				}
			},
		},
		{
			name:   "fixed and remainder",
			income: 1000000,
			setup: func(t *testing.T) ([]*jar.Jar, map[uuid.UUID]int64) {
				insurance := mustNewJar(
					t,
					"Insurance",
					jar.AllocationTypeFixed,
					300000,
				)
				remainder := mustNewJar(
					t,
					"Remainder",
					jar.AllocationTypeRemainder,
					0,
				)

				return []*jar.Jar{insurance, remainder}, map[uuid.UUID]int64{
					insurance.ID: 300000,
					remainder.ID: 700000,
				}
			},
		},
		{
			name:   "percentage fixed and remainder",
			income: 1000000,
			setup: func(t *testing.T) ([]*jar.Jar, map[uuid.UUID]int64) {
				needs := mustNewJar(
					t,
					"Needs",
					jar.AllocationTypePercentage,
					50,
				)
				leisure := mustNewJar(
					t,
					"Leisure",
					jar.AllocationTypePercentage,
					20,
				)
				insurance := mustNewJar(
					t,
					"Insurance",
					jar.AllocationTypeFixed,
					100000,
				)
				remainder := mustNewJar(
					t,
					"Remainder",
					jar.AllocationTypeRemainder,
					0,
				)

				return []*jar.Jar{
						needs,
						leisure,
						insurance,
						remainder,
					}, map[uuid.UUID]int64{
						needs.ID:     500000,
						leisure.ID:   200000,
						insurance.ID: 100000,
						remainder.ID: 200000,
					}
			},
		},
		{
			name:   "fixed allocations exceed income",
			income: 100000,
			setup: func(t *testing.T) ([]*jar.Jar, map[uuid.UUID]int64) {
				insurance := mustNewJar(
					t,
					"Insurance",
					jar.AllocationTypeFixed,
					150000,
				)
				remainder := mustNewJar(
					t,
					"Remainder",
					jar.AllocationTypeRemainder,
					0,
				)

				return []*jar.Jar{insurance, remainder}, nil
			},
			wantErr: jar.ErrAllocationExceedsIncome,
		},
		{
			name:   "invalid configuration",
			income: 1000000,
			setup: func(t *testing.T) ([]*jar.Jar, map[uuid.UUID]int64) {
				needs := mustNewJar(
					t,
					"Needs",
					jar.AllocationTypePercentage,
					50,
				)

				return []*jar.Jar{needs}, nil
			},
			wantErr: jar.ErrNoRemainder,
		},
		{
			name:   "zero income",
			income: 0,
			setup: func(t *testing.T) ([]*jar.Jar, map[uuid.UUID]int64) {
				remainder := mustNewJar(
					t,
					"Remainder",
					jar.AllocationTypeRemainder,
					0,
				)

				return []*jar.Jar{remainder}, map[uuid.UUID]int64{
					remainder.ID: 0,
				}
			},
		},
		{
			name:   "repository error",
			income: 1000000,
			setup: func(t *testing.T) ([]*jar.Jar, map[uuid.UUID]int64) {
				return nil, nil
			},
			repoErr: errRepository,
			wantErr: errRepository,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jars, want := tt.setup(t)

			repo := &MockJarRepository{
				jars:      jars,
				ListError: tt.repoErr,
			}

			service := application.NewJarService(repo, slog.Default())

			got, err := service.AllocateIncome(
				context.Background(),
				tt.income,
			)

			requireError(t, err, tt.wantErr)

			if tt.wantErr != nil {
				return
			}

			if !reflect.DeepEqual(got, want) {
				t.Fatalf(
					"unexpected allocations:\nwant: %#v\ngot:  %#v",
					want,
					got,
				)
			}

			if repo.ListCalls != 1 {
				t.Fatalf(
					"expected 1 List call, got %d",
					repo.ListCalls,
				)
			}
		})
	}
}

func TestJarServiceIntegration_CreateAndAllocate(t *testing.T) {
	t.Parallel()

	repo := &MockJarRepository{
		jars: []*jar.Jar{},
	}

	service := application.NewJarService(repo, slog.Default())
	ctx := context.Background()

	remainder, err := service.Create(
		ctx,
		"Remainder",
		jar.AllocationTypeRemainder,
		0,
	)
	if err != nil {
		t.Fatalf("failed to create remainder jar: %v", err)
	}

	needs, err := service.Create(
		ctx,
		"Needs",
		jar.AllocationTypePercentage,
		50,
	)
	if err != nil {
		t.Fatalf("failed to create needs jar: %v", err)
	}

	allocations, err := service.AllocateIncome(ctx, 1000000)
	if err != nil {
		t.Fatalf("failed to allocate income: %v", err)
	}

	if allocations[needs.ID] != 500000 {
		t.Fatalf(
			"expected needs allocation 500000, got %d",
			allocations[needs.ID],
		)
	}

	if allocations[remainder.ID] != 500000 {
		t.Fatalf(
			"expected remainder allocation 500000, got %d",
			allocations[remainder.ID],
		)
	}
}

func TestJarServiceIntegration_CreateUpdateAllocate(t *testing.T) {
	t.Parallel()

	repo := &MockJarRepository{
		jars: []*jar.Jar{},
	}

	service := application.NewJarService(repo, slog.Default())
	ctx := context.Background()

	remainder, err := service.Create(
		ctx,
		"Remainder",
		jar.AllocationTypeRemainder,
		0,
	)
	if err != nil {
		t.Fatalf("failed to create remainder jar: %v", err)
	}

	needs, err := service.Create(
		ctx,
		"Needs",
		jar.AllocationTypePercentage,
		50,
	)
	if err != nil {
		t.Fatalf("failed to create needs jar: %v", err)
	}

	err = service.UpdateAllocations(ctx, []jar.JarAllocationUpdate{
		{
			ID:              needs.ID,
			AllocationType:  jar.AllocationTypePercentage,
			AllocationValue: 60,
		},
	})
	if err != nil {
		t.Fatalf("failed to update allocations: %v", err)
	}

	allocations, err := service.AllocateIncome(ctx, 1000000)
	if err != nil {
		t.Fatalf("failed to allocate income: %v", err)
	}

	if allocations[needs.ID] != 600000 {
		t.Fatalf(
			"expected needs allocation 600000, got %d",
			allocations[needs.ID],
		)
	}

	if allocations[remainder.ID] != 400000 {
		t.Fatalf(
			"expected remainder allocation 400000, got %d",
			allocations[remainder.ID],
		)
	}
}
