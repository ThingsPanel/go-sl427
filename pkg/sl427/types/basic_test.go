package types

import (
	"testing"
	"time"
)

func TestTimeStamp(t *testing.T) {
	// 测试时间戳编解码
	tests := []struct {
		name    string
		time    time.Time
		want    string
		wantErr bool
	}{
		{
			name:    "normal time",
			time:    time.Date(2024, 3, 15, 14, 30, 0, 0, time.Local),
			want:    "240315143000",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := NewTimeStamp(tt.time)
			bytes := ts.Bytes()
			if string(bytes) != tt.want {
				t.Errorf("TimeStamp.Bytes() = %v, want %v", string(bytes), tt.want)
			}

			// 测试解析
			parsed, err := ParseTimeStamp(bytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTimeStamp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !parsed.Equal(ts.Time) {
				t.Errorf("ParseTimeStamp() = %v, want %v", parsed, ts)
			}
		})
	}
}
