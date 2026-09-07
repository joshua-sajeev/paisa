package jar_test

import (
	"errors"
	"testing"

	"github.com/joshu-sajeev/paisa/internal/domain/jar"
)

func TestNewJar(t *testing.T) {
	tests := []struct {
		name            string
		inputName       string
		allocationType  jar.AllocationType
		allocationValue int64
		wantName        string
		wantErr         error
	}{
		{
			name:            "valid percentage jar",
			inputName:       "Needs",
			allocationType:  jar.AllocationTypePercentage,
			allocationValue: 50,
			wantName:        "Needs",
		},
		{
			name:            "valid minimum percentage",
			inputName:       "Needs",
			allocationType:  jar.AllocationTypePercentage,
			allocationValue: 1,
			wantName:        "Needs",
		},
		{
			name:            "valid maximum percentage",
			inputName:       "Needs",
			allocationType:  jar.AllocationTypePercentage,
			allocationValue: 100,
			wantName:        "Needs",
		},
		{
			name:            "valid fixed jar",
			inputName:       "Insurance",
			allocationType:  jar.AllocationTypeFixed,
			allocationValue: 500000, // ₹5,000
			wantName:        "Insurance",
		},
		{
			name:            "valid remainder jar",
			inputName:       "Remainder",
			allocationType:  jar.AllocationTypeRemainder,
			allocationValue: 0,
			wantName:        "Remainder",
		},
		{
			name:            "trims whitespace",
			inputName:       "  Needs  ",
			allocationType:  jar.AllocationTypePercentage,
			allocationValue: 50,
			wantName:        "Needs",
		},
		{
			name:            "empty name",
			inputName:       "",
			allocationType:  jar.AllocationTypePercentage,
			allocationValue: 50,
			wantErr:         jar.ErrInvalidName,
		},
		{
			name:            "whitespace only name",
			inputName:       "   ",
			allocationType:  jar.AllocationTypePercentage,
			allocationValue: 50,
			wantErr:         jar.ErrInvalidName,
		},
		{
			name:            "invalid allocation type",
			inputName:       "Needs",
			allocationType:  "invalid",
			allocationValue: 50,
			wantErr:         jar.ErrInvalidAllocationType,
		},
		{
			name:            "percentage below minimum",
			inputName:       "Needs",
			allocationType:  jar.AllocationTypePercentage,
			allocationValue: 0,
			wantErr:         jar.ErrInvalidAllocationVal,
		},
		{
			name:            "percentage above maximum",
			inputName:       "Needs",
			allocationType:  jar.AllocationTypePercentage,
			allocationValue: 101,
			wantErr:         jar.ErrInvalidAllocationVal,
		},
		{
			name:            "negative percentage",
			inputName:       "Needs",
			allocationType:  jar.AllocationTypePercentage,
			allocationValue: -1,
			wantErr:         jar.ErrInvalidAllocationVal,
		},
		{
			name:            "remainder must be zero",
			inputName:       "Remainder",
			allocationType:  jar.AllocationTypeRemainder,
			allocationValue: 1,
			wantErr:         jar.ErrInvalidAllocationVal,
		},
		{
			name:            "remainder cannot be negative",
			inputName:       "Remainder",
			allocationType:  jar.AllocationTypeRemainder,
			allocationValue: -1,
			wantErr:         jar.ErrInvalidAllocationVal,
		},
		{
			name:            "fixed amount must be positive",
			inputName:       "Insurance",
			allocationType:  jar.AllocationTypeFixed,
			allocationValue: 0,
			wantErr:         jar.ErrInvalidAllocationVal,
		},
		{
			name:            "fixed amount cannot be negative",
			inputName:       "Insurance",
			allocationType:  jar.AllocationTypeFixed,
			allocationValue: -500000,
			wantErr:         jar.ErrInvalidAllocationVal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jar.NewJar(
				tt.inputName,
				tt.allocationType,
				tt.allocationValue,
			)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf(
					"NewJar() error = %v, wantErr = %v",
					err,
					tt.wantErr,
				)
			}

			if tt.wantErr != nil {
				if got != nil {
					t.Fatal("NewJar() returned jar, want nil")
				}
				return
			}

			if got == nil {
				t.Fatal("NewJar() returned nil jar")
			}

			if got.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", got.Name, tt.wantName)
			}

			if got.AllocationType != tt.allocationType {
				t.Errorf(
					"AllocationType = %q, want %q",
					got.AllocationType,
					tt.allocationType,
				)
			}

			if got.AllocationValue != tt.allocationValue {
				t.Errorf(
					"AllocationValue = %d, want %d",
					got.AllocationValue,
					tt.allocationValue,
				)
			}

			if got.ID.String() == "" {
				t.Error("ID should not be empty")
			}

			if got.IsArchived {
				t.Error("new jar should not be archived")
			}

			if got.CreatedAt.IsZero() {
				t.Error("CreatedAt should be set")
			}

			if got.UpdatedAt.IsZero() {
				t.Error("UpdatedAt should be set")
			}
		})
	}
}

func TestJarRename(t *testing.T) {
	j, err := jar.NewJar("Needs", jar.AllocationTypePercentage, 50)
	if err != nil {
		t.Fatal(err)
	}

	oldUpdatedAt := j.UpdatedAt

	if err := j.Rename("  Essentials  "); err != nil {
		t.Fatalf("Rename() error = %v", err)
	}

	if j.Name != "Essentials" {
		t.Errorf("Name = %q, want %q", j.Name, "Essentials")
	}

	if !j.UpdatedAt.After(oldUpdatedAt) {
		t.Error("UpdatedAt should be updated")
	}
}

func TestJarRename_InvalidName(t *testing.T) {
	j, err := jar.NewJar("Needs", jar.AllocationTypePercentage, 50)
	if err != nil {
		t.Fatal(err)
	}

	oldName := j.Name
	oldUpdatedAt := j.UpdatedAt

	err = j.Rename("   ")
	if !errors.Is(err, jar.ErrInvalidName) {
		t.Fatalf(
			"Rename() error = %v, want %v",
			err,
			jar.ErrInvalidName,
		)
	}

	if j.Name != oldName {
		t.Errorf(
			"Name = %q, want unchanged %q",
			j.Name,
			oldName,
		)
	}

	if !j.UpdatedAt.Equal(oldUpdatedAt) {
		t.Error("UpdatedAt should remain unchanged")
	}
}

func TestJarArchive(t *testing.T) {
	j, err := jar.NewJar("Needs", jar.AllocationTypePercentage, 50)
	if err != nil {
		t.Fatal(err)
	}

	oldUpdatedAt := j.UpdatedAt

	if err := j.Archive(); err != nil {
		t.Fatalf("Archive() error = %v", err)
	}

	if !j.IsArchived {
		t.Error("IsArchived = false, want true")
	}

	if !j.UpdatedAt.After(oldUpdatedAt) {
		t.Error("UpdatedAt should be updated")
	}
}

func TestJarArchive_AlreadyArchived(t *testing.T) {
	j, err := jar.NewJar("Needs", jar.AllocationTypePercentage, 50)
	if err != nil {
		t.Fatal(err)
	}

	if err := j.Archive(); err != nil {
		t.Fatal(err)
	}

	oldUpdatedAt := j.UpdatedAt

	err = j.Archive()
	if !errors.Is(err, jar.ErrJarAlreadyArchived) {
		t.Fatalf(
			"Archive() error = %v, want %v",
			err,
			jar.ErrJarAlreadyArchived,
		)
	}

	if !j.IsArchived {
		t.Error("jar should remain archived")
	}

	if !j.UpdatedAt.Equal(oldUpdatedAt) {
		t.Error("UpdatedAt should remain unchanged")
	}
}

func TestJarUnarchive(t *testing.T) {
	j, err := jar.NewJar("Needs", jar.AllocationTypePercentage, 50)
	if err != nil {
		t.Fatal(err)
	}

	if err := j.Archive(); err != nil {
		t.Fatal(err)
	}

	oldUpdatedAt := j.UpdatedAt

	if err := j.Unarchive(); err != nil {
		t.Fatalf("Unarchive() error = %v", err)
	}

	if j.IsArchived {
		t.Error("IsArchived = true, want false")
	}

	if !j.UpdatedAt.After(oldUpdatedAt) {
		t.Error("UpdatedAt should be updated")
	}
}

func TestJarUnarchive_NotArchived(t *testing.T) {
	j, err := jar.NewJar("Needs", jar.AllocationTypePercentage, 50)
	if err != nil {
		t.Fatal(err)
	}

	oldUpdatedAt := j.UpdatedAt

	err = j.Unarchive()
	if !errors.Is(err, jar.ErrJarNotArchived) {
		t.Fatalf(
			"Unarchive() error = %v, want %v",
			err,
			jar.ErrJarNotArchived,
		)
	}

	if j.IsArchived {
		t.Error("jar should remain active")
	}

	if !j.UpdatedAt.Equal(oldUpdatedAt) {
		t.Error("UpdatedAt should remain unchanged")
	}
}

func TestValidateAllocationConfiguration(t *testing.T) {
	mustNewJar := func(
		name string,
		allocationType jar.AllocationType,
		allocationValue int64,
	) *jar.Jar {
		j, err := jar.NewJar(name, allocationType, allocationValue)
		if err != nil {
			t.Fatalf("failed to create jar: %v", err)
		}
		return j
	}

	tests := []struct {
		name    string
		jars    []*jar.Jar
		wantErr error
	}{
		{
			name: "single remainder jar",
			jars: []*jar.Jar{
				mustNewJar(
					"Investment",
					jar.AllocationTypeRemainder,
					0,
				),
			},
			wantErr: nil,
		},
		{
			name: "percentage + remainder",
			jars: []*jar.Jar{
				mustNewJar("Needs", jar.AllocationTypePercentage, 50),
				mustNewJar(
					"Investment",
					jar.AllocationTypeRemainder,
					0,
				),
			},
			wantErr: nil,
		},
		{
			name: "fixed + remainder",
			jars: []*jar.Jar{
				mustNewJar(
					"Insurance",
					jar.AllocationTypeFixed,
					500000,
				),
				mustNewJar(
					"Investment",
					jar.AllocationTypeRemainder,
					0,
				),
			},
			wantErr: nil,
		},
		{
			name: "multiple percentages + remainder",
			jars: []*jar.Jar{
				mustNewJar("Needs", jar.AllocationTypePercentage, 50),
				mustNewJar("Leisure", jar.AllocationTypePercentage, 20),
				mustNewJar(
					"Investment",
					jar.AllocationTypeRemainder,
					0,
				),
			},
			wantErr: nil,
		},
		{
			name: "multiple fixed + remainder",
			jars: []*jar.Jar{
				mustNewJar(
					"Insurance",
					jar.AllocationTypeFixed,
					500000,
				),
				mustNewJar(
					"Charity",
					jar.AllocationTypeFixed,
					100000,
				),
				mustNewJar(
					"Investment",
					jar.AllocationTypeRemainder,
					0,
				),
			},
			wantErr: nil,
		},
		{
			name: "mixed percentage and fixed + remainder",
			jars: []*jar.Jar{
				mustNewJar("Needs", jar.AllocationTypePercentage, 50),
				mustNewJar("Leisure", jar.AllocationTypePercentage, 20),
				mustNewJar(
					"Insurance",
					jar.AllocationTypeFixed,
					500000,
				),
				mustNewJar(
					"Investment",
					jar.AllocationTypeRemainder,
					0,
				),
			},
			wantErr: nil,
		},
		{
			name: "percentage total equals 100%",
			jars: []*jar.Jar{
				mustNewJar("Needs", jar.AllocationTypePercentage, 50),
				mustNewJar("Leisure", jar.AllocationTypePercentage, 50),
				mustNewJar(
					"Investment",
					jar.AllocationTypeRemainder,
					0,
				),
			},
			wantErr: jar.ErrPercentageExceedsLimit,
		},
		{
			name: "percentage total exceeds 100%",
			jars: []*jar.Jar{
				mustNewJar("Needs", jar.AllocationTypePercentage, 60),
				mustNewJar("Leisure", jar.AllocationTypePercentage, 50),
				mustNewJar(
					"Investment",
					jar.AllocationTypeRemainder,
					0,
				),
			},
			wantErr: jar.ErrPercentageExceedsLimit,
		},
		{
			name:    "empty allocation",
			jars:    []*jar.Jar{},
			wantErr: jar.ErrEmptyAllocationConfiguration,
		},
		{
			name: "percentage allocation without remainder",
			jars: []*jar.Jar{
				mustNewJar("Needs", jar.AllocationTypePercentage, 50),
			},
			wantErr: jar.ErrNoRemainder,
		},
		{
			name: "fixed allocation without remainder",
			jars: []*jar.Jar{
				mustNewJar(
					"Insurance",
					jar.AllocationTypeFixed,
					500000,
				),
			},
			wantErr: jar.ErrNoRemainder,
		},
		{
			name: "two remainder jars",
			jars: []*jar.Jar{
				mustNewJar(
					"Investment",
					jar.AllocationTypeRemainder,
					0,
				),
				mustNewJar(
					"Savings",
					jar.AllocationTypeRemainder,
					0,
				),
			},
			wantErr: jar.ErrMultipleRemainders,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := jar.ValidateAllocationConfiguration(tt.jars)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf(
					"ValidateAllocationConfiguration() error = %v, wantErr = %v",
					err,
					tt.wantErr,
				)
			}
		})
	}
}

func TestAllocateIncome(t *testing.T) {
	mustNewJar := func(
		name string,
		allocationType jar.AllocationType,
		allocationValue int64,
	) *jar.Jar {
		j, err := jar.NewJar(name, allocationType, allocationValue)
		if err != nil {
			t.Fatalf("failed to create jar: %v", err)
		}
		return j
	}

	tests := []struct {
		name            string
		jars            []*jar.Jar
		totalIncome     int64
		wantAllocations map[string]int64
		wantErr         error
	}{
		{
			name: "single remainder jar",
			jars: []*jar.Jar{
				mustNewJar(
					"Investment",
					jar.AllocationTypeRemainder,
					0,
				),
			},
			totalIncome: 1000000,
			wantAllocations: map[string]int64{
				"Investment": 1000000,
			},
			wantErr: nil,
		},
		{
			name: "percentage + remainder",
			jars: []*jar.Jar{
				mustNewJar("Needs", jar.AllocationTypePercentage, 50),
				mustNewJar(
					"Investment",
					jar.AllocationTypeRemainder,
					0,
				),
			},
			totalIncome: 1000000,
			wantAllocations: map[string]int64{
				"Needs":      500000,
				"Investment": 500000,
			},
			wantErr: nil,
		},
		{
			name: "fixed + remainder",
			jars: []*jar.Jar{
				mustNewJar(
					"Insurance",
					jar.AllocationTypeFixed,
					500000,
				),
				mustNewJar(
					"Investment",
					jar.AllocationTypeRemainder,
					0,
				),
			},
			totalIncome: 1000000,
			wantAllocations: map[string]int64{
				"Insurance":  500000,
				"Investment": 500000,
			},
			wantErr: nil,
		},
		{
			name: "complex allocation",
			jars: []*jar.Jar{
				mustNewJar("Needs", jar.AllocationTypePercentage, 50),
				mustNewJar("Leisure", jar.AllocationTypePercentage, 20),
				mustNewJar(
					"Insurance",
					jar.AllocationTypeFixed,
					500000,
				),
				mustNewJar(
					"Investment",
					jar.AllocationTypeRemainder,
					0,
				),
			},
			totalIncome: 5000000,
			wantAllocations: map[string]int64{
				"Needs":      2500000,
				"Leisure":    1000000,
				"Insurance":  500000,
				"Investment": 1000000,
			},
			wantErr: nil,
		},
		{
			name: "allocation exceeds income",
			jars: []*jar.Jar{
				mustNewJar(
					"Insurance",
					jar.AllocationTypeFixed,
					2000000,
				),
				mustNewJar(
					"Investment",
					jar.AllocationTypeRemainder,
					0,
				),
			},
			totalIncome:     1000000,
			wantAllocations: nil,
			wantErr:         jar.ErrAllocationExceedsIncome,
		},
		{
			name: "invalid configuration",
			jars: []*jar.Jar{
				mustNewJar("Needs", jar.AllocationTypePercentage, 50),
			},
			totalIncome:     1000000,
			wantAllocations: nil,
			wantErr:         jar.ErrNoRemainder,
		},
		{
			name:            "empty jars",
			jars:            []*jar.Jar{},
			totalIncome:     1000000,
			wantAllocations: nil,
			wantErr:         jar.ErrEmptyAllocationConfiguration,
		},
		{
			name: "negative income",
			jars: []*jar.Jar{
				mustNewJar(
					"Investment",
					jar.AllocationTypeRemainder,
					0,
				),
			},
			totalIncome:     -1000000,
			wantAllocations: nil,
			wantErr:         jar.ErrAllocationExceedsIncome,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jar.AllocateIncome(
				tt.jars,
				tt.totalIncome,
			)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf(
					"AllocateIncome() error = %v, wantErr = %v",
					err,
					tt.wantErr,
				)
			}

			if tt.wantErr != nil {
				if got != nil {
					t.Error(
						"AllocateIncome() returned allocations, want nil",
					)
				}
				return
			}

			if got == nil {
				t.Fatal("AllocateIncome() returned nil allocations")
			}

			if len(got) != len(tt.wantAllocations) {
				t.Errorf(
					"AllocateIncome() got %d allocations, want %d",
					len(got),
					len(tt.wantAllocations),
				)
			}

			for _, j := range tt.jars {
				want, ok := tt.wantAllocations[j.Name]
				if !ok {
					continue
				}

				gotAmount, ok := got[j.ID]
				if !ok {
					t.Errorf(
						"AllocateIncome() missing allocation for %s",
						j.Name,
					)
					continue
				}

				if gotAmount != want {
					t.Errorf(
						"AllocateIncome() %s = %d, want %d",
						j.Name,
						gotAmount,
						want,
					)
				}
			}
		})
	}
}
