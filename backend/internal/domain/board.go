package domain

func NewBoard() [][]PlayerID {
	board := make([][]PlayerID, Rows)
	for i := range board {
		board[i] = make([]PlayerID, Columns)
	}
	return board
}

func IsValidMove(board [][]PlayerID, column int) bool {
	if column < 0 || column >= Columns {
		return false
	}

	// here board[0] represents the top row (0 -> top and 5 -> bottom)
	if board[0][column] != 0{
		return false
	}

	return true
}

func DropDisk(board [][]PlayerID, column int, player PlayerID) (int, error) {
	// shifting all the disk from top to bottom till it 
	// reaches the end or another disk
	for row := Rows - 1; row >= 0; row-- {
		if board[row][column] == Empty {
			board[row][column] = player
			return row, nil
		}
	}

	return -1, ErrColumnFull
}

func IsBoardFull(board [][]PlayerID) bool {
	for c := 0; c < Columns; c++ {
		if board[0][c] == Empty {
			return false
		}
	}

	return true
}

// this creates a deep copy of the board
func CopyBoard(board [][]PlayerID) [][]PlayerID {
	newBoard := make([][]PlayerID, len(board))
	for i := range board {
		newBoard[i] = make([]PlayerID, len(board[i]))
		copy(newBoard[i], board[i])
	}
	return newBoard
}

// this is a helper function that will later be used by the bot
func GetValidMoves(board [][]PlayerID) []int {
	validMoves := []int{}
	for col := 0; col < Columns; col++ {
		if board[0][col] == Empty && IsValidMove(board, col) {
			validMoves = append(validMoves, col)
		}
	}
	return validMoves
}

// this will simulate a move and give the result to the coller
func SimulateMove(board [][]PlayerID, column int, player PlayerID) ([][]PlayerID, int, error) {
	newBoard := CopyBoard(board)
	row, err := DropDisk(newBoard, column, player)
	if err != nil {
		return nil, -1, err
	}
	return newBoard, row, nil
}

// this counts the number of disks in a specific direction
func CountDiskInDirection(board [][]PlayerID, row, columns int, deltaRow, deltaCol int, player PlayerID) int {
	count := 0
	r, c := row+deltaRow, columns+deltaCol
	for r >= 0 && r < Rows && c >= 0 && c < Columns && board[r][c] == player {
		count++
		r += deltaRow
		c += deltaCol
	}
	return count
}
