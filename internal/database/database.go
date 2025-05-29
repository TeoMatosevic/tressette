package database

import (
	"database/sql"
	"sync"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/mattn/go-sqlite3"
)

type Service struct {
	db         *sql.DB
	m          *sync.Mutex
	table_name string
}

var (
	tableName  = "tressette"
	dbInstance *Service
)

func New() Service {
	var err error
	db, err := sql.Open("sqlite3", "./tressette.db")
	if err != nil {
		panic(err)
	}

	sqlStmt := `
	create table if not exists tressette (
		id string not null primary key,
		created_at string,
		player1 string,
		player2 string,
		player3 string,
		player4 string,
		player1_team string,
		player2_team string,
		player3_team string,
		player4_team string,
		team1_score integer,
		team2_score integer
	);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		panic(err)
	}

	dbInstance = &Service{
		db:         db,
		table_name: tableName,
		m:          &sync.Mutex{},
	}

	return *dbInstance
}

func (s *Service) Close() error {
	return s.db.Close()
}

func (s *Service) TableName() string {
	return s.table_name
}

func (s *Service) GetAll() ([]GameResult, error) {
	s.m.Lock()
	defer s.m.Unlock()
	rows, err := s.db.Query("SELECT * FROM " + s.table_name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []GameResult
	for rows.Next() {
		var result GameResult
		if err := rows.Scan(
			&result.ID,
			&result.CreatedAt,
			&result.Player1,
			&result.Player2,
			&result.Player3,
			&result.Player4,
			&result.Player1Team,
			&result.Player2Team,
			&result.Player3Team,
			&result.Player4Team,
			&result.Team1Score,
			&result.Team2Score); err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

func (s *Service) GetByID(id string) (GameResult, error) {
	s.m.Lock()
	defer s.m.Unlock()
	var result GameResult
	err := s.db.QueryRow("SELECT * FROM "+s.table_name+" WHERE id = ?", id).Scan(
		&result.ID,
		&result.CreatedAt,
		&result.Player1,
		&result.Player2,
		&result.Player3,
		&result.Player4,
		&result.Player1Team,
		&result.Player2Team,
		&result.Player3Team,
		&result.Player4Team,
		&result.Team1Score,
		&result.Team2Score)
	if err != nil {
		return GameResult{}, err
	}
	return result, nil
}

func (s *Service) Insert(result GameResult) error {
	s.m.Lock()
	defer s.m.Unlock()
	_, err := s.db.Exec("INSERT INTO "+s.table_name+
		" (id, created_at, player1, player2, player3, player4, player1_team, player2_team, player3_team, player4_team, team1_score, team2_score) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		result.ID,
		result.CreatedAt,
		result.Player1,
		result.Player2,
		result.Player3,
		result.Player4,
		result.Player1Team,
		result.Player2Team,
		result.Player3Team,
		result.Player4Team,
		result.Team1Score,
		result.Team2Score)

	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetByPlayer(player_name string) ([]GameResult, error) {
	s.m.Lock()
	defer s.m.Unlock()
	rows, err := s.db.Query("SELECT * FROM "+s.table_name+
		" WHERE player1 = ? OR player2 = ? OR player3 = ? OR player4 = ?",
		player_name,
		player_name,
		player_name,
		player_name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []GameResult
	for rows.Next() {
		var result GameResult
		if err := rows.Scan(
			&result.ID,
			&result.CreatedAt,
			&result.Player1,
			&result.Player2,
			&result.Player3,
			&result.Player4,
			&result.Player1Team,
			&result.Player2Team,
			&result.Player3Team,
			&result.Player4Team,
			&result.Team1Score,
			&result.Team2Score); err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	if len(results) == 0 {
		return nil, sql.ErrNoRows // No results found
	}

	return results, nil
}
