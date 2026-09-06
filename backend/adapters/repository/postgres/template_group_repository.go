package postgres

import (
	"database/sql"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
)

type TemplateGroupRepository struct {
	db *sql.DB
}

func NewTemplateGroupRepository(db *sql.DB) *TemplateGroupRepository {
	return &TemplateGroupRepository{db: db}
}

func (r *TemplateGroupRepository) SaveTemplateGroup(g domain.TemplateGroup) (int, error) {
	query := `INSERT INTO template_groups (user_id, name) VALUES ($1, $2) RETURNING id`
	var id int
	err := r.db.QueryRow(query, g.UserID, g.Name).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *TemplateGroupRepository) FindTemplateGroupsByUser(userID int) ([]domain.TemplateGroup, error) {
	rows, err := r.db.Query("SELECT id, user_id, name FROM template_groups WHERE user_id = $1 ORDER BY id ASC", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []domain.TemplateGroup
	for rows.Next() {
		var g domain.TemplateGroup
		if err := rows.Scan(&g.ID, &g.UserID, &g.Name); err != nil {
			return nil, err
		}
		res = append(res, g)
	}
	return res, nil
}

func (r *TemplateGroupRepository) UpdateTemplateGroup(g domain.TemplateGroup) error {
	query := `UPDATE template_groups SET name = $2 WHERE id = $1`
	_, err := r.db.Exec(query, g.ID, g.Name)
	return err
}

func (r *TemplateGroupRepository) DeleteTemplateGroup(id int) error {
	_, err := r.db.Exec("DELETE FROM template_groups WHERE id = $1", id)
	return err
}

func (r *TemplateGroupRepository) DeleteAllByUser(userID int) error {
	_, err := r.db.Exec("DELETE FROM template_groups WHERE user_id = $1", userID)
	return err
}
