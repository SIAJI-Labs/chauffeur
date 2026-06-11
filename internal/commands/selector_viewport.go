package commands

func computeSelectorViewport(totalRows, activeRow, terminalHeight, reservedLines int) (start, end int, showAbove, showBelow bool) {
	if totalRows <= 0 {
		return 0, 0, false, false
	}

	if activeRow < 0 {
		activeRow = 0
	} else if activeRow >= totalRows {
		activeRow = totalRows - 1
	}

	visibleRows := terminalHeight - reservedLines
	if visibleRows < 1 {
		visibleRows = 1
	}

	if totalRows > visibleRows {
		if activeRow >= visibleRows {
			start = activeRow - visibleRows + 1
		}

		maxStart := totalRows - visibleRows
		if start > maxStart {
			start = maxStart
		}
	}

	end = start + visibleRows
	if end > totalRows {
		end = totalRows
	}

	showAbove = start > 0
	showBelow = end < totalRows

	return start, end, showAbove, showBelow
}
