package commands

import "testing"

func TestComputeSelectorViewport(t *testing.T) {
	tests := []struct {
		name          string
		totalRows     int
		activeRow     int
		terminalHeight int
		reservedLines int
		wantStart     int
		wantEnd       int
		wantShowAbove bool
		wantShowBelow bool
	}{
		{
			name:           "zero rows returns empty viewport",
			totalRows:      0,
			activeRow:      0,
			terminalHeight: 10,
			reservedLines:  8,
			wantStart:      0,
			wantEnd:        0,
			wantShowAbove:  false,
			wantShowBelow:  false,
		},
		{
			name:           "negative total rows returns empty viewport",
			totalRows:      -5,
			activeRow:      3,
			terminalHeight: 10,
			reservedLines:  8,
			wantStart:      0,
			wantEnd:        0,
			wantShowAbove:  false,
			wantShowBelow:  false,
		},
		{
			name:           "no clipping when rows fit",
			totalRows:      3,
			activeRow:      1,
			terminalHeight: 12,
			reservedLines:  8,
			wantStart:      0,
			wantEnd:        3,
			wantShowAbove:  false,
			wantShowBelow:  false,
		},
		{
			name:           "short terminal clips to one visible row",
			totalRows:      5,
			activeRow:      3,
			terminalHeight: 6,
			reservedLines:  8,
			wantStart:      3,
			wantEnd:        4,
			wantShowAbove:  true,
			wantShowBelow:  true,
		},
		{
			name:           "does not scroll at last visible row",
			totalRows:      12,
			activeRow:      1,
			terminalHeight: 10,
			reservedLines:  8,
			wantStart:      0,
			wantEnd:        2,
			wantShowAbove:  false,
			wantShowBelow:  true,
		},
		{
			name:           "starts scrolling at first row beyond viewport",
			totalRows:      12,
			activeRow:      2,
			terminalHeight: 10,
			reservedLines:  8,
			wantStart:      1,
			wantEnd:        3,
			wantShowAbove:  true,
			wantShowBelow:  true,
		},
		{
			name:           "shows only bottom indicator near top",
			totalRows:      12,
			activeRow:      1,
			terminalHeight: 10,
			reservedLines:  8,
			wantStart:      0,
			wantEnd:        2,
			wantShowAbove:  false,
			wantShowBelow:  true,
		},
		{
			name:           "shows only top indicator at bottom",
			totalRows:      12,
			activeRow:      11,
			terminalHeight: 10,
			reservedLines:  8,
			wantStart:      10,
			wantEnd:        12,
			wantShowAbove:  true,
			wantShowBelow:  false,
		},
		{
			name:           "clamps negative active row to top",
			totalRows:      12,
			activeRow:      -3,
			terminalHeight: 10,
			reservedLines:  8,
			wantStart:      0,
			wantEnd:        2,
			wantShowAbove:  false,
			wantShowBelow:  true,
		},
		{
			name:           "clamps active row beyond total rows to bottom",
			totalRows:      12,
			activeRow:      99,
			terminalHeight: 10,
			reservedLines:  8,
			wantStart:      10,
			wantEnd:        12,
			wantShowAbove:  true,
			wantShowBelow:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, showAbove, showBelow := computeSelectorViewport(tt.totalRows, tt.activeRow, tt.terminalHeight, tt.reservedLines)

			if start != tt.wantStart || end != tt.wantEnd || showAbove != tt.wantShowAbove || showBelow != tt.wantShowBelow {
				t.Fatalf(
					"computeSelectorViewport(%d, %d, %d, %d) = (%d, %d, %t, %t), want (%d, %d, %t, %t)",
					tt.totalRows,
					tt.activeRow,
					tt.terminalHeight,
					tt.reservedLines,
					start,
					end,
					showAbove,
					showBelow,
					tt.wantStart,
					tt.wantEnd,
					tt.wantShowAbove,
					tt.wantShowBelow,
				)
			}
		})
	}
}

func TestSelectorViewportWithIndicators(t *testing.T) {
	tests := []struct {
		name           string
		totalRows      int
		activeRow      int
		terminalHeight int
		reservedLines  int
		wantStart      int
		wantEnd        int
		wantShowAbove  bool
		wantShowBelow  bool
	}{
		{
			name:           "bottom indicator only reduces visible rows",
			totalRows:      12,
			activeRow:      1,
			terminalHeight: 10,
			reservedLines:  4,
			wantStart:      0,
			wantEnd:        5,
			wantShowAbove:  false,
			wantShowBelow:  true,
		},
		{
			name:           "top indicator only reduces visible rows",
			totalRows:      12,
			activeRow:      11,
			terminalHeight: 10,
			reservedLines:  4,
			wantStart:      7,
			wantEnd:        12,
			wantShowAbove:  true,
			wantShowBelow:  false,
		},
		{
			name:           "both indicators reduce visible rows twice",
			totalRows:      12,
			activeRow:      8,
			terminalHeight: 10,
			reservedLines:  4,
			wantStart:      5,
			wantEnd:        9,
			wantShowAbove:  true,
			wantShowBelow:  true,
		},
		{
			name:           "tiny terminal still shows one active row with both indicators",
			totalRows:      12,
			activeRow:      8,
			terminalHeight: 5,
			reservedLines:  4,
			wantStart:      8,
			wantEnd:        9,
			wantShowAbove:  true,
			wantShowBelow:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, showAbove, showBelow := selectorViewportWithIndicators(tt.totalRows, tt.activeRow, tt.terminalHeight, tt.reservedLines)

			if start != tt.wantStart || end != tt.wantEnd || showAbove != tt.wantShowAbove || showBelow != tt.wantShowBelow {
				t.Fatalf(
					"selectorViewportWithIndicators(%d, %d, %d, %d) = (%d, %d, %t, %t), want (%d, %d, %t, %t)",
					tt.totalRows,
					tt.activeRow,
					tt.terminalHeight,
					tt.reservedLines,
					start,
					end,
					showAbove,
					showBelow,
					tt.wantStart,
					tt.wantEnd,
					tt.wantShowAbove,
					tt.wantShowBelow,
				)
			}
		})
	}
}
