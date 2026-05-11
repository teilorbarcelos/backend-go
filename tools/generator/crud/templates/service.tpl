package {{.LowerName}}

import (
	"context"
	"backend-go/pkg/database"
)

type {{.Name}}Service struct {
	Repo *{{.Name}}Repository
}

func New{{.Name}}Service(repo *{{.Name}}Repository) *{{.Name}}Service {
	return &{{.Name}}Service{Repo: repo}
}

func (s *{{.Name}}Service) Create(ctx context.Context, entity *{{.Name}}) error {
	return s.Repo.WithContext(ctx).Create(entity)
}

func (s *{{.Name}}Service) List(ctx context.Context, params database.FilterParams) ([]{{.Name}}, int64, error) {
	allowed := map[string]bool{
		"name": true,
	}
	return s.Repo.WithContext(ctx).SearchPaginated(params, allowed)
}

func (s *{{.Name}}Service) GetByID(ctx context.Context, id string) (*{{.Name}}, error) {
	return s.Repo.WithContext(ctx).FindByID(id)
}

func (s *{{.Name}}Service) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	return s.Repo.WithContext(ctx).Update(id, updates)
}

func (s *{{.Name}}Service) Delete(ctx context.Context, id string) error {
	return s.Repo.WithContext(ctx).Delete(id)
}
