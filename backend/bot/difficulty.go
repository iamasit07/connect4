package bot

type BotDifficulty string

const (
	DifficultyEasy   BotDifficulty = "easy"
	DifficultyMedium BotDifficulty = "medium"
	DifficultyHard   BotDifficulty = "hard"
)

// ParseDifficulty validates and returns the bot difficulty
// Defaults to Medium if invalid or empty
func ParseDifficulty(difficulty string) BotDifficulty {
	switch difficulty {
	case "easy":
		return DifficultyEasy
	case "medium":
		return DifficultyMedium
	case "hard":
		return DifficultyHard
	default:
		return DifficultyMedium // Default to medium
	}
}
