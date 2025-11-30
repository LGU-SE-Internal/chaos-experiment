package systemconfig

import (
	"testing"
)

func TestSetCurrentSystem(t *testing.T) {
	// Reset to default before tests
	_ = SetCurrentSystem(SystemTrainTicket)

	tests := []struct {
		name        string
		system      SystemType
		wantErr     bool
		expectedSys SystemType
	}{
		{
			name:        "set TrainTicket system",
			system:      SystemTrainTicket,
			wantErr:     false,
			expectedSys: SystemTrainTicket,
		},
		{
			name:        "set OtelDemo system",
			system:      SystemOtelDemo,
			wantErr:     false,
			expectedSys: SystemOtelDemo,
		},
		{
			name:        "set invalid system",
			system:      "invalid-system",
			wantErr:     true,
			expectedSys: SystemOtelDemo, // Should remain unchanged from previous test
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetCurrentSystem(tt.system)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetCurrentSystem() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && GetCurrentSystem() != tt.expectedSys {
				t.Errorf("GetCurrentSystem() = %v, want %v", GetCurrentSystem(), tt.expectedSys)
			}
		})
	}
}

func TestGetCurrentSystem(t *testing.T) {
	// Reset to default
	_ = SetCurrentSystem(SystemTrainTicket)

	if got := GetCurrentSystem(); got != SystemTrainTicket {
		t.Errorf("GetCurrentSystem() = %v, want %v", got, SystemTrainTicket)
	}

	_ = SetCurrentSystem(SystemOtelDemo)
	if got := GetCurrentSystem(); got != SystemOtelDemo {
		t.Errorf("GetCurrentSystem() = %v, want %v", got, SystemOtelDemo)
	}
}

func TestIsTrainTicket(t *testing.T) {
	// Reset to TrainTicket
	_ = SetCurrentSystem(SystemTrainTicket)

	if !IsTrainTicket() {
		t.Error("IsTrainTicket() should return true when system is TrainTicket")
	}

	_ = SetCurrentSystem(SystemOtelDemo)
	if IsTrainTicket() {
		t.Error("IsTrainTicket() should return false when system is OtelDemo")
	}
}

func TestIsOtelDemo(t *testing.T) {
	// Reset to TrainTicket
	_ = SetCurrentSystem(SystemTrainTicket)

	if IsOtelDemo() {
		t.Error("IsOtelDemo() should return false when system is TrainTicket")
	}

	_ = SetCurrentSystem(SystemOtelDemo)
	if !IsOtelDemo() {
		t.Error("IsOtelDemo() should return true when system is OtelDemo")
	}
}

func TestSystemTypeString(t *testing.T) {
	if SystemTrainTicket.String() != "ts" {
		t.Errorf("SystemTrainTicket.String() = %v, want %v", SystemTrainTicket.String(), "ts")
	}

	if SystemOtelDemo.String() != "otel-demo" {
		t.Errorf("SystemOtelDemo.String() = %v, want %v", SystemOtelDemo.String(), "otel-demo")
	}
}

func TestGetAllSystemTypes(t *testing.T) {
	types := GetAllSystemTypes()
	if len(types) != 2 {
		t.Errorf("GetAllSystemTypes() returned %d types, want 2", len(types))
	}

	found := make(map[SystemType]bool)
	for _, st := range types {
		found[st] = true
	}

	if !found[SystemTrainTicket] {
		t.Error("GetAllSystemTypes() should include SystemTrainTicket")
	}
	if !found[SystemOtelDemo] {
		t.Error("GetAllSystemTypes() should include SystemOtelDemo")
	}
}

func TestParseSystemType(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    SystemType
		wantErr bool
	}{
		{
			name:    "parse ts",
			input:   "ts",
			want:    SystemTrainTicket,
			wantErr: false,
		},
		{
			name:    "parse otel-demo",
			input:   "otel-demo",
			want:    SystemOtelDemo,
			wantErr: false,
		},
		{
			name:    "parse invalid",
			input:   "invalid",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSystemType(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSystemType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseSystemType() = %v, want %v", got, tt.want)
			}
		})
	}
}
