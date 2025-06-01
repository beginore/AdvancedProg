package validator

import (
	"testing"
)

func TestValidator(t *testing.T) {
	t.Run("NotBlank", func(t *testing.T) {
		tests := []struct {
			value string
			want  bool
		}{
			{"", false},
			{"   ", false},
			{"a", true},
		}

		for _, tt := range tests {
			got := NotBlank(tt.value)
			if got != tt.want {
				t.Errorf("NotBlank(%q) = %v, want %v", tt.value, got, tt.want)
			}
		}
	})

	t.Run("ValidatePassword", func(t *testing.T) {
		tests := []struct {
			password string
			want     bool
		}{
			{"Short1!", false},                    // слишком короткий
			{"longpasswordwithoutnumspec", false}, // нет цифр и спецсимволов
			{"ValidPass123!", true},
			{"invalidpass", false}, // нет заглавных и спецсимволов
		}

		for _, tt := range tests {
			got := ValidatePassword(tt.password)
			if got != tt.want {
				t.Errorf("ValidatePassword(%q) = %v, want %v", tt.password, got, tt.want)
			}
		}
	})
}
