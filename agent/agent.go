package agent

import (
	"context"

	"encore.dev/storage/sqldb"
)

// Type that represents a note.
type Agent struct {
	ID         string `json:"id"`
	Bio        string `json:"bio"`
	ProfileUrl string `json:"profile_url"`
}

//encore:api public method=POST path=/agent
func SaveAgent(ctx context.Context, agent *Agent) (*Agent, error) {
	// Save the agent to the database.
	// If the agent already exists (i.e. CONFLICT), we update the agents text and the cover URL.
	_, err := sqldb.Exec(ctx, `
		INSERT INTO agent (id, bio, profile_url) VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET bio=$2, profile_url=$3
	`, agent.ID, agent.Bio, agent.ProfileUrl)

	// If there was an error saving to the database, then we return that error.
	if err != nil {
		return nil, err
	}

	// Otherwise, we return the agent to indicate that the save was successful.
	return agent, nil
}

//encore:api public method=GET path=/agent/:id
func GetAgent(ctx context.Context, id string) (*Agent, error) {
	agent := &Agent{ID: id}

	// We use the agent ID to query the database for the agent's text and cover URL.
	err := sqldb.QueryRow(ctx, `
		SELECT bio, profile_url FROM agent
		WHERE id = $1
	`, id).Scan(&agent.Bio, &agent.ProfileUrl)

	// If the agent doesn't exist, we return an error.
	if err != nil {
		return nil, err
	}

	// Otherwise, we return the agent.
	return agent, nil
}
