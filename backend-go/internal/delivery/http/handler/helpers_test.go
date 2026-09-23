package handler

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCleanUUID(t *testing.T) {
	rawUUID := uuid.New()
	rawStr := rawUUID.String()

	tests := []struct {
		name    string
		input   string
		want    uuid.UUID
		wantErr bool
	}{
		{
			name:  "plain uuid",
			input: rawStr,
			want:  rawUUID,
		},
		{
			name:  "fav_ prefix",
			input: "fav_" + rawStr,
			want:  rawUUID,
		},
		{
			name:  "usr_ prefix with trailing slash",
			input: "usr_" + rawStr + "/",
			want:  rawUUID,
		},
		{
			name:  "pro_ prefix with spaces",
			input: "  pro_" + rawStr + "  ",
			want:  rawUUID,
		},
		{
			name:  "apt_ prefix",
			input: "apt_" + rawStr,
			want:  rawUUID,
		},
		{
			name:    "invalid input",
			input:   "not-a-valid-uuid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cleanUUID(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
