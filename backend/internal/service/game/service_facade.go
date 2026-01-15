package game

// Service is the entry point for game logic (facade)
type Service struct {
	Repo GameRepository
}

func NewService(repo GameRepository) *Service {
	return &Service{
		Repo: repo,
	}
}
