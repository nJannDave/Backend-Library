package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPersonalInfo_ValidatePN(t *testing.T) {
	tests := []struct {
		name     string
		phone    string
		expected string
		hasError bool
	}{
		{"Prefix_0", "08123456789", "+628123456789", false},
		{"Prefix_62", "628123456789", "+628123456789", false},
		{"Prefix_Plus62", "+628123456789", "+628123456789", false},
		{"With_Format_Chars", "(0812) 345.678-9", "+628123456789", false},
		{"Too_Short", "0812", "", true},
		{"Too_Long", "08123456789012345", "", true},
		{"Invalid_Prefix", "123456789", "", true},
		{"Non_Numeric", "0812abc567", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PersonalInfo{PhoneNumber: tt.phone}
			var errMsg []string
			p.ValidatePN(&errMsg)

			if tt.hasError {
				assert.NotEmpty(t, errMsg)
			} else {
				assert.Empty(t, errMsg)
				assert.Equal(t, tt.expected, p.PhoneNumber)
			}
		})
	}
}

func TestAcademicInfo_ValidateClass(t *testing.T) {
	tests := []struct {
		name     string
		major    string
		class    string
		hasError bool
	}{
		{"SIJA_Class_XIII_OK", "SIJA", "XIII", false},
		{"IOP_Class_XIII_OK", "IOP", "XIII", false},
		{"Other_Class_XIII_Fail", "TKJ", "XIII", true},
		{"Normal_Class_XII_OK", "TKJ", "XII", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AcademicInfo{Major: tt.major, Class: tt.class}
			var errMsg []string
			a.ValidateClass(&errMsg)

			if tt.hasError {
				assert.Contains(t, errMsg, "class not available")
			} else {
				assert.Empty(t, errMsg)
			}
		})
	}
}

func TestLoan_ValidateDateFormat(t *testing.T) {
	l := &Loan{}
	t.Run("Valid_Format", func(t *testing.T) {
		err := l.ValidateDateFormat("17-08-2025")
		assert.NoError(t, err)
		assert.Equal(t, 2025, l.MustReturnedAt.Year())
	})

	t.Run("Invalid_Format", func(t *testing.T) {
		err := l.ValidateDateFormat("2025/08/17")
		assert.Error(t, err)
	})
}

func TestLoan_ValidateDate(t *testing.T) {
	t.Run("Date_Within_Limit", func(t *testing.T) {
		l := &Loan{MustReturnedAt: time.Now().AddDate(0, 0, 3)}
		err := l.ValidateDate()
		assert.NoError(t, err)
	})

	t.Run("Date_Over_Limit", func(t *testing.T) {
		l := &Loan{MustReturnedAt: time.Now().AddDate(0, 0, 10)}
		err := l.ValidateDate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "maximum loan limit")
	})

	t.Run("Date_In_Past", func(t *testing.T) {
		l := &Loan{MustReturnedAt: time.Now().AddDate(0, 0, -1)}
		err := l.ValidateDate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "past")
	})
}

func TestLdUpdate_GiveSanctions(t *testing.T) {
	t.Run("Late_Return_Sanction", func(t *testing.T) {
		mustReturn := time.Now().AddDate(0, 0, -2)
		returned := time.Now()
		var sanction int64

		ldu := &LdUpdate{
			MustReturnedAt: mustReturn,
			ReturnedAt:     &returned,
			Sanctions:      &sanction,
		}
		ldu.GiveSanctions()
		assert.True(t, *ldu.Sanctions > 0)
	})

	t.Run("On_Time_No_Sanction", func(t *testing.T) {
		mustReturn := time.Now().AddDate(0, 0, 2)
		returned := time.Now()
		var sanction int64

		ldu := &LdUpdate{
			MustReturnedAt: mustReturn,
			ReturnedAt:     &returned,
			Sanctions:      &sanction,
		}
		ldu.GiveSanctions()
		assert.Equal(t, int64(0), *ldu.Sanctions)
	})
}

func TestTableNames(t *testing.T) {
	assert.Equal(t, "students", Students{}.TableName())
	assert.Equal(t, "categories", Categories{}.TableName())
	assert.Equal(t, "books", Book{}.TableName())
	assert.Equal(t, "loan", Loan{}.TableName())
	assert.Equal(t, "connections", Connections{}.TableName())
}