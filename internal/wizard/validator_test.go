package wizard

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateProjectName(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{"valid simple", "my-app", false},
		{"valid letters only", "backend", false},
		{"valid with numbers", "app2go", false},
		{"too short", "a", true},
		{"starts with number", "1app", true},
		{"contains spaces", "my app", true},
		{"contains underscore", "my_app", true},
		{"empty string", "", true},
		{"non-string type", 123, true},
		{"valid long name", "my-really-long-project-name", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProjectName(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateGroupID(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{"valid two parts", "com.example", false},
		{"valid three parts", "com.example.myapp", false},
		{"valid org", "org.springframework", false},
		{"single part", "com", true},
		{"starts with uppercase", "Com.example", true},
		{"contains hyphen", "com.my-app", true},
		{"empty", "", true},
		{"just dots", "...", true},
		{"non-string", 42, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGroupID(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateArtifactID(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{"valid simple", "my-app", false},
		{"valid no hyphens", "backend", false},
		{"valid with numbers", "app2", false},
		{"starts with number", "2app", true},
		{"contains uppercase", "MyApp", true},
		{"contains spaces", "my app", true},
		{"empty", "", true},
		{"non-string", 42, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateArtifactID(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePackageName(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{"valid", "com.example.myapp", false},
		{"valid two parts", "com.example", false},
		{"contains hyphen", "com.example.my-app", true},
		{"single part", "com", true},
		{"uppercase", "Com.Example", true},
		{"empty", "", true},
		{"non-string", 42, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePackageName(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateNotEmpty(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{"non-empty", "hello", false},
		{"empty string", "", true},
		{"whitespace only", "   ", true},
		{"non-string", 42, true},
		{"single char", "a", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNotEmpty(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDerivePackageName(t *testing.T) {
	assert.Equal(t, "com.example.myapp", DerivePackageName("com.example", "myapp"))
	assert.Equal(t, "com.example.myservice", DerivePackageName("com.example", "my-service"))
	assert.Equal(t, "org.acme.shopapi", DerivePackageName("org.acme", "shop-api"))
}
